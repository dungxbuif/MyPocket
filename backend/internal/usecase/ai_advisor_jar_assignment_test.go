package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type jarAssignSourceStub struct {
	rows map[string]*JarAssignmentTransaction
	jars map[string]string // jar|month -> name
}

func (s *jarAssignSourceStub) Transaction(_ context.Context, owner, id string) (*JarAssignmentTransaction, error) {
	if row := s.rows[owner+"|"+id]; row != nil {
		return row, nil
	}
	return nil, repository.ErrNotFound
}
func (s *jarAssignSourceStub) ActiveJarName(_ context.Context, _, jarID, month string) (string, error) {
	if name, ok := s.jars[jarID+"|"+month]; ok {
		return name, nil
	}
	return "", repository.ErrNotFound
}

func runAssignTool(t *testing.T, source JarAssignmentSource, args map[string]any) (map[string]any, error) {
	t.Helper()
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
	registry.SetJarAssignmentSource(source)
	tool, ok := registry.Lookup("propose_assign_jar")
	if !ok {
		t.Fatal("propose_assign_jar not registered")
	}
	raw, _ := json.Marshal(args)
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "o1"}, raw)
	if err != nil {
		return nil, err
	}
	if result.ViewKind != "jar_assignment_proposal" || result.Status != "proposed" {
		t.Fatalf("result=%+v", result)
	}
	var view map[string]any
	_ = json.Unmarshal(result.View, &view)
	return view, nil
}

func TestProposeAssignJarValidatesEachRowAndKeepsValidOnes(t *testing.T) {
	jar, other := uuid.NewString(), uuid.NewString()
	ids := map[string]string{}
	source := &jarAssignSourceStub{rows: map[string]*JarAssignmentTransaction{}, jars: map[string]string{jar + "|2026-09": "Ăn uống"}}
	add := func(name, reject string, jarID *string) {
		ids[name] = uuid.NewString()
		source.rows["o1|"+ids[name]] = &JarAssignmentTransaction{ID: ids[name], Amount: 50000, Note: "Phở " + name, CategoryName: "Ăn", OccurredAt: time.Date(2026, 9, 10, 3, 0, 0, 0, time.UTC), LocalMonth: "2026-09", Version: 3, RejectReason: reject, CurrentJarID: jarID}
	}
	add("ok", "", nil)
	add("income", "Khoản thu không được gắn hũ.", nil)
	add("transfer", "Giao dịch chuyển ví không được gắn hũ.", nil)
	add("credit", "Giao dịch tín dụng không được gắn hũ.", nil)
	add("unassign", "", &jar)
	source.rows["o2|"+uuid.NewString()] = &JarAssignmentTransaction{}
	row := func(name, jarID string) map[string]any {
		return map[string]any{"transaction_id": ids[name], "jar_id": jarID, "reason": "Ghi chú 'Phở'"}
	}
	view, err := runAssignTool(t, source, map[string]any{"month": "2026-09", "assignments": []any{
		row("ok", jar), row("income", jar), row("transfer", jar), row("credit", jar), row("unassign", ""), row("ok", other),
		map[string]any{"transaction_id": uuid.NewString(), "jar_id": jar, "reason": "x"}, // not owned / unknown
	}})
	if err != nil {
		t.Fatal(err)
	}
	rows := view["assignments"].([]any)
	status := func(i int) string { return rows[i].(map[string]any)["status"].(string) }
	want := []string{"proposed", "rejected", "rejected", "rejected", "proposed", "rejected", "rejected"}
	for i, expected := range want {
		if status(i) != expected {
			t.Fatalf("row %d status=%s want %s: %v", i, status(i), expected, rows[i])
		}
	}
	first := rows[0].(map[string]any)
	if first["jar_name"] != "Ăn uống" || first["expected_version"].(float64) != 3 || first["category_name"] != "Ăn" || first["note"] == "" {
		t.Fatalf("row facts missing: %v", first)
	}
	if !strings.Contains(rows[1].(map[string]any)["reject_reason"].(string), "thu") {
		t.Fatalf("income reason: %v", rows[1])
	}
	if !strings.Contains(rows[5].(map[string]any)["reject_reason"].(string), "nhiều lần") {
		t.Fatalf("duplicate/inactive reason: %v", rows[5])
	}
	if !strings.Contains(rows[6].(map[string]any)["reject_reason"].(string), "Không tìm thấy") {
		t.Fatalf("foreign transaction must look missing: %v", rows[6])
	}
	// Inactive jar in the transaction's month.
	inactive, _ := runAssignTool(t, source, map[string]any{"month": "2026-09", "assignments": []any{row("unassign", other)}})
	if inactive["assignments"].([]any)[0].(map[string]any)["status"] != "rejected" {
		t.Fatalf("inactive jar accepted: %v", inactive)
	}
}

func TestProposeAssignJarEnforcesMaxRowsAndRequiresSource(t *testing.T) {
	source := &jarAssignSourceStub{rows: map[string]*JarAssignmentTransaction{}}
	rows := make([]any, 0, 51)
	for i := 0; i < 51; i++ {
		rows = append(rows, map[string]any{"transaction_id": uuid.NewString(), "jar_id": "", "reason": fmt.Sprint(i)})
	}
	if _, err := runAssignTool(t, source, map[string]any{"month": "2026-09", "assignments": rows}); err == nil {
		t.Fatal("51 rows must be rejected")
	}
	if _, err := runAssignTool(t, source, map[string]any{"month": "2026-09", "assignments": rows[:50]}); err != nil {
		t.Fatalf("50 rows must be allowed: %v", err)
	}
	if _, err := runAssignTool(t, source, map[string]any{"month": "bad", "assignments": rows[:1]}); err == nil {
		t.Fatal("bad month must be rejected")
	}
	if _, err := runAssignTool(t, nil, map[string]any{"month": "2026-09", "assignments": rows[:1]}); err == nil {
		t.Fatal("tool without a validation source must refuse")
	}
}

func TestJarAssignmentPartRegistrationAndEditFields(t *testing.T) {
	if part, ok := advisorPartType("jar_assignment_proposal"); !ok || part != "jar_assignment_confirmation" || !isAdvisorProposalView("jar_assignment_proposal") {
		t.Fatalf("part=%s ok=%v", part, ok)
	}
	if fields := advisorSupportedEditFields("assign_jar"); len(fields) != 1 || fields[0] != "assignments" {
		t.Fatalf("fields=%v", fields)
	}
	if advisorProposalViewKind("assign_jar") != "jar_assignment_proposal" {
		t.Fatal("revision view kind missing")
	}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
	tool, _ := registry.Lookup("propose_assign_jar")
	if tool.Definition().Parameters["required"] == nil || !strings.Contains(BuildAdvisorSystemMessage(Principal{}).Content, "propose_assign_jar") {
		t.Fatal("definition/handbook missing")
	}
	_ = entity.AdvisorPart{}
}
