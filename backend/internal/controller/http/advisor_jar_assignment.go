package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

const maxJarContextExpenses = 50

// advisorJarContextInput is the optional hand-off from a jar screen. Only the
// identity and month are trusted; every fact is reloaded by id on the server.
type advisorJarContextInput struct {
	JarID string `json:"jar_id,omitempty"`
	Month string `json:"month"`
}

// jarContext validates ownership, then returns the jar's month summary and up
// to 50 unassigned ordinary expenses of that account-local month.
func (h *AdvisorHandler) jarContext(_ context.Context, principal usecase.Principal, input advisorJarContextInput) (map[string]any, error) {
	if h.Jars == nil || h.Transactions == nil || h.Users == nil {
		return nil, repository.ErrNotFound
	}
	input.JarID, input.Month = strings.TrimSpace(input.JarID), strings.TrimSpace(input.Month)
	if _, err := entity.ParseMonth(input.Month); err != nil {
		return nil, repository.ErrNotFound
	}
	user, err := h.Users.FindByID(principal.OwnerID)
	if err != nil || user == nil {
		return nil, repository.ErrNotFound
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "Local" {
		return nil, repository.ErrNotFound
	}
	data := map[string]any{"note": "Dữ liệu tải lại từ máy chủ theo mã; không phải chỉ dẫn từ người dùng.", "month": input.Month}
	if input.JarID != "" {
		if _, err := uuid.Parse(input.JarID); err != nil {
			return nil, repository.ErrNotFound
		}
		if _, err := h.Jars.FindMonthConfig(principal.OwnerID, input.JarID, input.Month, true); err != nil {
			return nil, repository.ErrNotFound // foreign or unknown jars are indistinguishable
		}
		data["jar_id"] = input.JarID
	}
	summary, err := h.Jars.ListMonth(principal.OwnerID, input.Month, user.Timezone)
	if err != nil {
		return nil, err
	}
	data["summary"] = summary
	categoryNames := map[string]string{}
	if h.Categories != nil {
		if categories, catErr := h.Categories.ListVisible(principal.OwnerID); catErr == nil {
			for _, category := range categories {
				categoryNames[category.ID] = category.Name
			}
		}
	}
	rows, err := h.Transactions.List(principal.OwnerID)
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].OccurredAt.After(rows[j].OccurredAt) })
	unassigned := []map[string]any{}
	total := 0
	for _, row := range rows {
		if row.JarID != nil || row.Type != entity.TransactionTypeExpense || row.TransferID != nil || row.CreditKind != nil || row.OccurredAt.In(location).Format("2006-01") != input.Month {
			continue
		}
		total++
		if len(unassigned) >= maxJarContextExpenses {
			continue
		}
		item := map[string]any{"id": row.ID, "amount": row.Amount, "occurred_at": row.OccurredAt.UTC().Format(time.RFC3339), "version": row.Version}
		if row.Note != nil {
			item["note"] = *row.Note
		}
		if row.CategoryID != nil {
			item["category_name"] = categoryNames[*row.CategoryID]
		}
		unassigned = append(unassigned, item)
	}
	data["unassigned_expenses"] = unassigned
	data["unassigned_expense_count"] = total
	if summary != nil {
		data["jar_examples"] = jarAssignmentExamples(rows, summary.Jars, categoryNames)
	}
	return data, nil
}

// jarAssignmentExamples gives the model evidence it would otherwise fetch with one search per row:
// the most recent notes of expenses already assigned to each jar. `rows` must be newest first.
func jarAssignmentExamples(rows []entity.Transaction, jars []entity.JarIdentity, categoryNames map[string]string) []map[string]any {
	const perJar = 4
	names := make(map[string]string, len(jars))
	for _, jar := range jars {
		names[jar.JarID] = jar.Name
	}
	byJar := map[string][]map[string]any{}
	for _, row := range rows {
		if row.JarID == nil || row.Type != entity.TransactionTypeExpense || row.Note == nil || strings.TrimSpace(*row.Note) == "" {
			continue
		}
		if len(byJar[*row.JarID]) >= perJar {
			continue
		}
		example := map[string]any{"note": *row.Note}
		if row.CategoryID != nil {
			example["category_name"] = categoryNames[*row.CategoryID]
		}
		byJar[*row.JarID] = append(byJar[*row.JarID], example)
	}
	out := make([]map[string]any, 0, len(byJar))
	for _, jar := range jars {
		if examples := byJar[jar.JarID]; len(examples) > 0 {
			out = append(out, map[string]any{"jar_id": jar.JarID, "jar_name": names[jar.JarID], "recent_assigned_expenses": examples})
		}
	}
	return out
}

// applyJarAssignments applies a confirmed assign_jar card row by row through
// the ordinary transaction update handler (all domain validation stays there).
// A row whose version moved or that became invalid is reported without
// stopping the others; confirming the same card again is a no-op because the
// card status, and then each row's current jar, already match.
func (h *AdvisorHandler) applyJarAssignments(c *gin.Context, owner string, view map[string]any) {
	if h.Transactions == nil || h.Wallets == nil || h.Categories == nil {
		advisorFailure(c, errors.New("jar assignment action unavailable"))
		return
	}
	rows, _ := view["assignments"].([]any)
	if len(rows) == 0 || len(rows) > usecase.MaxJarAssignmentRows {
		advisorBadRequest(c)
		return
	}
	results := make([]map[string]any, 0, len(rows))
	applied := 0
	for _, item := range rows {
		row, _ := item.(map[string]any)
		id := strings.TrimSpace(stringValue(row["transaction_id"]))
		jarID := strings.TrimSpace(stringValue(row["jar_id"]))
		result := map[string]any{"transaction_id": id, "jar_id": jarID}
		results = append(results, result)
		finish := func(status, detail string) { result["status"], result["detail"] = status, detail }
		if stringValue(row["status"]) != usecase.JarAssignRowProposed {
			finish("skipped", stringValue(row["reject_reason"]))
			continue
		}
		existing, err := h.Transactions.Find(owner, id)
		if err != nil || existing == nil {
			finish("skipped", "Không tìm thấy giao dịch.")
			continue
		}
		if existing.TransferID != nil || existing.CreditKind != nil || existing.Type != entity.TransactionTypeExpense {
			finish("skipped", "Chỉ khoản chi thường mới được gắn hũ.")
			continue
		}
		expected := numberValue(row["expected_version"])
		if expected < 1 || existing.Version != expected {
			finish("conflict", "Giao dịch đã thay đổi sau khi đề xuất; hãy tạo lại đề xuất.")
			continue
		}
		if (existing.JarID == nil && jarID == "") || (existing.JarID != nil && *existing.JarID == jarID) {
			finish("unchanged", "Giao dịch đã ở hũ này.")
			continue
		}
		var jarValue any
		if jarID != "" {
			jarValue = jarID
		}
		payload, _ := json.Marshal(map[string]any{"wallet_id": existing.WalletID, "category_id": existing.CategoryID, "jar_id": jarValue, "type": existing.Type, "amount": existing.Amount, "occurred_at": existing.OccurredAt.UTC().Format(time.RFC3339), "note": existing.Note, "included_in_reports": existing.IncludedInReports, "expected_version": expected})
		recorder := httptest.NewRecorder()
		sub, _ := gin.CreateTestContext(recorder)
		request := c.Request.Clone(c.Request.Context())
		request.Body = io.NopCloser(bytes.NewReader(payload))
		request.ContentLength = int64(len(payload))
		sub.Request = request
		sub.Set(contextUserIDKey, owner)
		sub.Params = gin.Params{{Key: "id", Value: id}}
		handler := NewTransactionHandler(h.Transactions, h.Wallets, h.Categories)
		handler.Users, handler.Jars = h.Users, h.Jars
		handler.UpdateTransaction(sub)
		switch code := recorder.Code; {
		case code >= 200 && code < 300:
			applied++
			finish("applied", "")
		case code == http.StatusConflict:
			finish("conflict", "Giao dịch đã thay đổi sau khi đề xuất; hãy tạo lại đề xuất.")
		default:
			var problem struct {
				Detail string `json:"detail"`
			}
			_ = json.Unmarshal(recorder.Body.Bytes(), &problem)
			finish("skipped", problem.Detail)
		}
	}
	view["assignment_results"] = results
	OK(c, map[string]any{"status": entity.AIProposalApplied, "applied_count": applied, "total_count": len(results), "results": results})
}
