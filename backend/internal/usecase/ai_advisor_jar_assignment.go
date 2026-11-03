package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

// MaxJarAssignmentRows bounds one confirmation card (AI-JAR-01).
const MaxJarAssignmentRows = 50

const (
	JarAssignRowProposed = "proposed"
	JarAssignRowRejected = "rejected"
)

// JarAssignmentTransaction is the server-read fact set for one proposed row.
// RejectReason is non-empty when the row can never carry a jar.
type JarAssignmentTransaction struct {
	ID           string
	Type         string
	Amount       int64
	Note         string
	CategoryName string
	OccurredAt   time.Time
	LocalMonth   string
	Version      int64
	CurrentJarID *string
	RejectReason string
}

// JarAssignmentSource reads owner-scoped facts for validation. Every lookup is
// scoped by owner; a foreign or missing id is reported as not found.
type JarAssignmentSource interface {
	Transaction(ctx context.Context, ownerID, transactionID string) (*JarAssignmentTransaction, error)
	// ActiveJarName returns the jar's name only when it has an ACTIVE config in month.
	ActiveJarName(ctx context.Context, ownerID, jarID, month string) (string, error)
}

type repositoryJarAssignmentSource struct {
	transactions repository.TransactionRepository
	categories   repository.CategoryRepository
	jars         repository.JarRepository
	users        repository.UserRepository
}

func NewJarAssignmentSource(transactions repository.TransactionRepository, categories repository.CategoryRepository, jars repository.JarRepository, users repository.UserRepository) JarAssignmentSource {
	return &repositoryJarAssignmentSource{transactions: transactions, categories: categories, jars: jars, users: users}
}

func (s *repositoryJarAssignmentSource) Transaction(_ context.Context, ownerID, id string) (*JarAssignmentTransaction, error) {
	row, err := s.transactions.Find(ownerID, id)
	if err != nil || row == nil {
		return nil, repository.ErrNotFound
	}
	user, err := s.users.FindByID(ownerID)
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "Local" {
		return nil, errors.New("account timezone unavailable")
	}
	out := &JarAssignmentTransaction{ID: row.ID, Type: row.Type, Amount: row.Amount, OccurredAt: row.OccurredAt, LocalMonth: row.OccurredAt.In(location).Format("2006-01"), Version: row.Version, CurrentJarID: row.JarID}
	if row.Note != nil {
		out.Note = *row.Note
	}
	switch {
	case row.Type == entity.TransactionTypeIncome:
		out.RejectReason = "Khoản thu không được gắn hũ."
	case row.TransferID != nil:
		out.RejectReason = "Giao dịch chuyển ví không được gắn hũ."
	case row.CreditKind != nil:
		out.RejectReason = "Giao dịch tín dụng không được gắn hũ."
	case row.Type != entity.TransactionTypeExpense:
		out.RejectReason = "Chỉ khoản chi thường mới được gắn hũ."
	}
	if row.CategoryID != nil {
		if category, catErr := s.categories.FindVisible(ownerID, *row.CategoryID); catErr == nil && category != nil {
			out.CategoryName = category.Name
			if category.SystemKey != nil && *category.SystemKey == "expense_transfer_out" && out.RejectReason == "" {
				out.RejectReason = "Giao dịch chuyển nội bộ không được gắn hũ."
			}
		}
	}
	return out, nil
}

func (s *repositoryJarAssignmentSource) ActiveJarName(_ context.Context, ownerID, jarID, month string) (string, error) {
	config, err := s.jars.FindMonthConfig(ownerID, jarID, month, false)
	if err != nil || config == nil {
		return "", repository.ErrNotFound
	}
	return config.Name, nil
}

// SetJarAssignmentSource enables propose_assign_jar's server-side validation.
func (r *AdvisorToolRegistry) SetJarAssignmentSource(source JarAssignmentSource) {
	if tool, ok := r.tools["propose_assign_jar"].(*jarAssignProposalTool); ok {
		tool.source = source
	}
}

type jarAssignProposalArgs struct {
	Month       string `json:"month"`
	Assignments []struct {
		TransactionID string `json:"transaction_id"`
		JarID         string `json:"jar_id"`
		Reason        string `json:"reason"`
	} `json:"assignments"`
}

// jarAssignProposalTool reads (never writes) and returns a confirmation card.
// Invalid rows are kept in the card as rejected with a Vietnamese reason so one
// bad id cannot discard the other proposals.
type jarAssignProposalTool struct{ source JarAssignmentSource }

func (t *jarAssignProposalTool) Name() string { return "propose_assign_jar" }
func (t *jarAssignProposalTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Prepare assigning jars to ordinary expenses (up to 50 rows) for review and confirmation; never writes. The reason must cite the transaction note, category or similar already-assigned rows. Empty jar_id means unassign. Income, wallet transfers and credit rows are rejected by the server.", map[string]any{
		"month": monthSchema(),
		"assignments": map[string]any{"type": "array", "minItems": 1, "maxItems": MaxJarAssignmentRows, "items": map[string]any{
			"type": "object", "additionalProperties": false, "required": []string{"transaction_id", "jar_id", "reason"},
			"properties": map[string]any{
				"transaction_id": map[string]any{"type": "string", "format": "uuid"},
				"jar_id":         map[string]any{"type": "string", "maxLength": 128},
				"reason":         map[string]any{"type": "string", "maxLength": 300},
			},
		}},
	}, []string{"month", "assignments"})
}

func (t *jarAssignProposalTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args jarAssignProposalArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	args.Month = strings.TrimSpace(args.Month)
	if _, err := entity.ParseMonth(args.Month); err != nil || len(args.Assignments) == 0 || len(args.Assignments) > MaxJarAssignmentRows || t.source == nil {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	rows := make([]map[string]any, 0, len(args.Assignments))
	seen := map[string]bool{}
	for _, item := range args.Assignments {
		id, jarID, reason := strings.TrimSpace(item.TransactionID), strings.TrimSpace(item.JarID), strings.TrimSpace(item.Reason)
		if len(reason) > 300 {
			reason = reason[:300]
		}
		row := map[string]any{"transaction_id": id, "jar_id": jarID, "reason": reason, "status": JarAssignRowProposed}
		reject := func(message string) { row["status"], row["reject_reason"] = JarAssignRowRejected, message }
		rows = append(rows, row)
		if _, err := uuid.Parse(id); err != nil {
			reject("Mã giao dịch không hợp lệ.")
			continue
		}
		if seen[id] {
			reject("Giao dịch xuất hiện nhiều lần trong đề xuất.")
			continue
		}
		seen[id] = true
		tx, err := t.source.Transaction(ctx, principal.OwnerID, id)
		if err != nil || tx == nil {
			reject("Không tìm thấy giao dịch.")
			continue
		}
		row["note"], row["amount"], row["occurred_at"], row["category_name"] = tx.Note, tx.Amount, tx.OccurredAt.UTC().Format(time.RFC3339), tx.CategoryName
		row["expected_version"], row["current_jar_id"] = tx.Version, tx.CurrentJarID
		if tx.RejectReason != "" {
			reject(tx.RejectReason)
			continue
		}
		if jarID == "" {
			row["jar_name"] = ""
			continue
		}
		if _, err := uuid.Parse(jarID); err != nil {
			reject("Mã hũ không hợp lệ.")
			continue
		}
		name, err := t.source.ActiveJarName(ctx, principal.OwnerID, jarID, tx.LocalMonth)
		if err != nil {
			reject("Hũ không có cấu hình đang dùng trong tháng " + tx.LocalMonth + " của giao dịch.")
			continue
		}
		row["jar_name"] = name
	}
	return proposalResult("assign_jar", "jar_assignment_proposal", map[string]any{"action_id": uuid.NewString(), "action": "assign_jar", "month": args.Month, "assignments": rows})
}
