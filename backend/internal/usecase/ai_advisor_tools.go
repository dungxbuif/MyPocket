package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAdvisorToolInvalid     = errors.New("advisor tool arguments are invalid")
	ErrAdvisorToolUnsupported = errors.New("advisor tool is unsupported")
)

type Principal struct {
	OwnerID        string
	CredentialID   string
	CredentialKind string
	ExpiresAt      time.Time
	Scopes         []string
	Timezone       string
}

func (p Principal) HasScope(scope string) bool {
	if p.CredentialKind == "session" {
		return true
	}
	for _, candidate := range p.Scopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type AdvisorTool interface {
	Name() string
	Definition() ToolDefinition
	Execute(context.Context, Principal, json.RawMessage) (entity.FinanceResult, error)
}

type AdvisorToolRegistry struct {
	tools map[string]AdvisorTool
}

func NewAdvisorToolRegistry(finance *FinanceQueryService) *AdvisorToolRegistry {
	registry := &AdvisorToolRegistry{tools: make(map[string]AdvisorTool)}
	for _, tool := range []AdvisorTool{
		&summaryTool{finance: finance}, &searchTool{finance: finance}, &compareTool{finance: finance},
		&walletTool{finance: finance}, &budgetTool{finance: finance}, &goalTool{finance: finance}, &jarTool{finance: finance},
		&transactionTool{finance: finance},
	} {
		registry.tools[tool.Name()] = tool
	}
	return registry
}

func (r *AdvisorToolRegistry) Lookup(name string) (AdvisorTool, bool) {
	if r == nil {
		return nil, false
	}
	tool, ok := r.tools[name]
	return tool, ok
}

func DecodeAdvisorToolArgs(raw json.RawMessage, destination any) error {
	if len(raw) == 0 || len(raw) > 16*1024 {
		return ErrAdvisorToolInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return ErrAdvisorToolInvalid
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrAdvisorToolInvalid
	}
	return nil
}

type summaryArgs struct {
	Range           entity.DateRange `json:"range"`
	WalletIDs       []string         `json:"wallet_ids"`
	CategoryIDs     []string         `json:"category_ids"`
	IncludeChildren bool             `json:"include_children"`
}

type searchArgs struct {
	Range           entity.DateRange `json:"range"`
	WalletIDs       []string         `json:"wallet_ids"`
	CategoryIDs     []string         `json:"category_ids"`
	IncludeChildren bool             `json:"include_children"`
	Type            string           `json:"type"`
	ReportScope     string           `json:"report_scope"`
	NoteContains    string           `json:"note_contains"`
	MinAmount       *int64           `json:"min_amount"`
	MaxAmount       *int64           `json:"max_amount"`
	Limit           int              `json:"limit"`
	Cursor          string           `json:"cursor"`
}

type compareArgs struct {
	Current         entity.DateRange `json:"current"`
	Previous        entity.DateRange `json:"previous"`
	WalletIDs       []string         `json:"wallet_ids"`
	CategoryIDs     []string         `json:"category_ids"`
	IncludeChildren bool             `json:"include_children"`
}

type objectArgs struct {
	IDs []string `json:"wallet_ids"`
}

type budgetArgs struct {
	BudgetIDs []string `json:"budget_ids"`
	ActiveOn  string   `json:"active_on"`
}

type jarArgs struct {
	Month  string   `json:"month"`
	JarIDs []string `json:"jar_ids"`
}

type transactionArgs struct {
	TransactionID string `json:"transaction_id"`
}

type summaryTool struct{ finance *FinanceQueryService }

func (t *summaryTool) Name() string { return "get_finance_summary" }
func (t *summaryTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Read report-included income, expense and net totals for a calendar range", map[string]any{
		"range": rangeSchema(), "wallet_ids": stringArraySchema(), "category_ids": stringArraySchema(), "include_children": boolSchema(),
	}, []string{"range"})
}
func (t *summaryTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args summaryArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	filter, err := normalizeToolFilter(args.Range, principal.Timezone, entity.FinanceFilter{WalletIDs: args.WalletIDs, CategoryIDs: args.CategoryIDs, IncludeChildren: args.IncludeChildren})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), Filter: filter})
}

type searchTool struct{ finance *FinanceQueryService }

func (t *searchTool) Name() string { return "search_transactions" }
func (t *searchTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Search owner-scoped transactions with bounded keyset pagination", map[string]any{
		"range": rangeSchema(), "wallet_ids": stringArraySchema(), "category_ids": stringArraySchema(), "include_children": boolSchema(),
		"type": stringSchema(32), "report_scope": stringSchema(32), "note_contains": stringSchema(256),
		"min_amount": integerSchema(), "max_amount": integerSchema(), "limit": integerSchema(), "cursor": stringSchema(256),
	}, []string{"range"})
}
func (t *searchTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args searchArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	filter, err := normalizeToolFilter(args.Range, principal.Timezone, entity.FinanceFilter{WalletIDs: args.WalletIDs, CategoryIDs: args.CategoryIDs, IncludeChildren: args.IncludeChildren, Type: args.Type, ReportScope: args.ReportScope, NoteContains: args.NoteContains, MinAmount: args.MinAmount, MaxAmount: args.MaxAmount, Limit: args.Limit})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), Filter: filter, Cursor: strings.TrimSpace(args.Cursor)})
}

type compareTool struct{ finance *FinanceQueryService }

func (t *compareTool) Name() string { return "compare_spending_periods" }
func (t *compareTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Compare report-included spending between two calendar ranges", map[string]any{
		"current": rangeSchema(), "previous": rangeSchema(), "wallet_ids": stringArraySchema(), "category_ids": stringArraySchema(), "include_children": boolSchema(),
	}, []string{"current", "previous"})
}
func (t *compareTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args compareArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	filter, err := normalizeToolFilter(args.Current, principal.Timezone, entity.FinanceFilter{WalletIDs: args.WalletIDs, CategoryIDs: args.CategoryIDs, IncludeChildren: args.IncludeChildren})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), Filter: filter, CompareRange: &args.Previous})
}

type walletTool struct{ finance *FinanceQueryService }

func (t *walletTool) Name() string { return "get_wallet_balances" }
func (t *walletTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Read current balances for owner wallets", map[string]any{"wallet_ids": stringArraySchema()}, nil)
}
func (t *walletTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args objectArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), ObjectIDs: args.IDs})
}

type budgetTool struct{ finance *FinanceQueryService }

func (t *budgetTool) Name() string { return "get_budget_progress" }
func (t *budgetTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Read budget limits and report-included spending", map[string]any{"budget_ids": stringArraySchema(), "active_on": dateSchema()}, nil)
}
func (t *budgetTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args budgetArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), ObjectIDs: args.BudgetIDs, ActiveOn: args.ActiveOn})
}

type goalTool struct{ finance *FinanceQueryService }

func (t *goalTool) Name() string { return "get_goal_progress" }
func (t *goalTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Read goal-wallet balances and target dates", map[string]any{"wallet_ids": stringArraySchema()}, nil)
}
func (t *goalTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args objectArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), ObjectIDs: args.IDs})
}

type jarTool struct{ finance *FinanceQueryService }

func (t *jarTool) Name() string { return "get_jar_progress" }
func (t *jarTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Read jar allocations and spending without creating month rows", map[string]any{"month": monthSchema(), "jar_ids": stringArraySchema()}, []string{"month"})
}
func (t *jarTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args jarArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	if strings.TrimSpace(args.Month) == "" {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), ObjectIDs: args.JarIDs, Month: args.Month})
}

type transactionTool struct{ finance *FinanceQueryService }

func (t *transactionTool) Name() string { return "get_transaction" }
func (t *transactionTool) Definition() ToolDefinition {
	return ToolDefinition{
		Name:        t.Name(),
		Description: "Read one owner-scoped transaction, including transfer/report metadata",
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"transaction_id"},
			"properties": map[string]any{
				"transaction_id": map[string]any{"type": "string", "minLength": 1, "maxLength": 128},
			},
		},
	}
}
func (t *transactionTool) Execute(ctx context.Context, principal Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args transactionArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil || strings.TrimSpace(args.TransactionID) == "" || len(args.TransactionID) > 128 {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	return t.finance.Execute(ctx, principal.OwnerID, entity.NormalizedQuery{Key: "q1", Kind: t.Name(), ObjectIDs: []string{strings.TrimSpace(args.TransactionID)}})
}

func readDefinition(name, description string, properties map[string]any, required []string) ToolDefinition {
	parameters := map[string]any{"type": "object", "additionalProperties": false, "properties": properties}
	if len(required) > 0 {
		parameters["required"] = required
	}
	return ToolDefinition{Name: name, Description: description, Parameters: parameters}
}

func rangeSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"from", "to"}, "properties": map[string]any{
		"from": dateSchema(), "to": dateSchema(),
	}}
}

func dateSchema() map[string]any {
	return map[string]any{"type": "string", "pattern": "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"}
}
func monthSchema() map[string]any {
	return map[string]any{"type": "string", "pattern": "^[0-9]{4}-[0-9]{2}$"}
}
func stringSchema(maxLength int) map[string]any {
	return map[string]any{"type": "string", "maxLength": maxLength}
}
func stringArraySchema() map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string", "maxLength": 128}, "maxItems": 100}
}
func boolSchema() map[string]any    { return map[string]any{"type": "boolean"} }
func integerSchema() map[string]any { return map[string]any{"type": "integer", "minimum": 0} }

func normalizeToolFilter(dateRange entity.DateRange, timezone string, input entity.FinanceFilter) (entity.NormalizedFinanceFilter, error) {
	if strings.TrimSpace(timezone) == "" || timezone == "Local" {
		return entity.NormalizedFinanceFilter{}, ErrAdvisorToolInvalid
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return entity.NormalizedFinanceFilter{}, ErrAdvisorToolInvalid
	}
	return entity.NormalizeFinanceFilter(entity.FinanceFilter{Range: dateRange, WalletIDs: input.WalletIDs, CategoryIDs: input.CategoryIDs, IncludeChildren: input.IncludeChildren, Type: input.Type, ReportScope: input.ReportScope, NoteContains: input.NoteContains, MinAmount: input.MinAmount, MaxAmount: input.MaxAmount, Limit: input.Limit}, location)
}
