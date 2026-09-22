# Finance Assistant — Technical Implementation Guide

Status: proposed / in review, 2026-09-22. **Documentation, not implemented code.**

Read with the [V1 task plan](2026-09-22-finance-assistant-v1.md) and [product/detail design](../../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md). This guide defines the proposed interfaces and algorithms for those tasks; it is not a second runtime architecture. All backend-relative paths below start at `backend/internal/`. V2 appendices describe boundaries only; V1 must not register write tools.

## 1. Integration map and verified traps

| Current file / behavior | Implementation instruction |
| --- | --- |
| `controller/http/middleware.go` only puts `userID` into Gin context | Add typed principal context with session ID and expiry for advisor; retain old `userID` compatibility. Never persist bearer tokens for background work. |
| `controller/http/response.go` provides OK/Created, no Accepted | Return `c.JSON(202, Response{Data: accepted, Meta: nil})`; continue shared Problem format. |
| `infrastructure/ai/client.go` rejects tool calls | New advisor adapter; do not weaken entry extraction's contract. |
| `repository.TransactionRepository.List` loads all owner rows | New bounded read repository; no list-all → prompt or list-all → in-memory pagination. |
| `infrastructure/repository/jar_postgres.go` `ListMonth` calls `lockOwner` and `ensureMonth` | **Not a read-only tool.** Implement `ReadJarProgress` on the snapshot handle. If month config is absent, compute effective inherited config in memory with `config_origin=inherited_preview`, do not INSERT. |
| `entity.CalculateBudgetsInLocation` is existing budget arithmetic | Characterize with tests; reuse its semantics, moving owner/range filtering to SQL. Prevent transfer IDs/system transfer categories entering the shared report set. Update existing budget consumer if parity exposes this gap. |
| `app/src/services/api.ts` calls `response.text()` | Use it for JSON endpoints only. New streaming reader uses fetch/readable body, not apiRequest. |
| `AssistantComposer` hardcodes transaction copy and file input | Extend via optional `mode="entry"|"advisor"` with default entry; advisor hides upload, changes labels and adds IME-safe keyboard behavior. Test entry defaults unchanged. |
| BottomNavigation has four destinations + centered FAB | Proposed layout: six equal slots: overview, transactions, Add, budgets, assistant, account. BaseFab stays ≥44px, no absolute overlap. Test at 360/390/430px; this is a planned navigation change, not merely adding an item to the old five-column grid. |
| `cmd/migrate` supports up/version/force, not down | Test down/up through golang-migrate `Steps(-1)` in an isolated integration test; do not document a nonexistent `go run ./cmd/migrate down`. |
| Existing DB tests call `t.Skip` without TEST_DATABASE_URL | New release runner must fail if required DB suite skips. A green unit run with skipped integration is not release evidence. |

Jar inheritance: select the latest earlier initialized jar_month, use only its active configs, retain stable jar IDs, mark allocation as preview; an explicitly initialized empty month stays empty. If current month has historical linked spend without config, show that spend and allocation unknown. Existing actual month initialization stays in existing user flows; advisor never materializes it. Whole tool execution uses DB `READ ONLY`, so a hidden write is an error caught by tests.

## 2. File ownership / build order

| Task | New files | Existing integration points |
| --- | --- | --- |
| 1 | `entity/finance_query.go`, `repository/finance_query.go`, `usecase/finance_query.go`, `infrastructure/repository/finance_query_postgres.go`, `_test.go` siblings | month/budget handlers; read-only jar helpers; `cmd/api/main.go` |
| 2 | `entity/ai_advisor.go`, `repository/ai_advisor.go`, `infrastructure/repository/ai_advisor_postgres.go`, `usecase/ai_advisor_context.go`, sibling tests; migration pair | account timezone/profile and existing database wiring |
| 3 | `usecase/ai_advisor_tools.go`, `usecase/ai_advisor_provider.go`, `infrastructure/ai/advisor_client.go`, `advisor_plan.v1.txt`, `advisor_answer.v1.txt`, sibling tests | existing HTTP config, no extractor behavioral change |
| 4 | `usecase/ai_advisor.go`, `usecase/ai_advisor_parts.go`, `usecase/ai_advisor_runner.go`, tests | task 1–3 ports only |
| 5 | `controller/http/ai_advisor_handler.go`, `routes_ai_advisor.go`, tests | middleware, config, main, Redis audit adapter, generated Swagger |
| 6 | `app/src/services/aiAdvisor{,Parts,Stream}.ts`; molecules from task plan; `AssistantPart.tsx` | AssistantComposer, bases, design inventory |
| 7 | `FinanceAssistantPanel.tsx`, fixture/integrated/browser tests | router, FinancePrototypePage, BottomNavigation, package scripts |
| 8 | live eval fixtures/test; docs guides/tool schemas/runbook/verification | docs publishing entry points and release evidence |

No import cycle: entity ← repository ports ← usecase; infrastructure implements ports; HTTP/main wires use cases. Go `httpapi` package naming remains as existing files. Dependency changes occur only in the task that owns them. Avoid moving unrelated modules.

## 3. Go contract: query boundary

These definitions go in `entity/finance_query.go`. JSON-visible types require explicit snake_case tags when implemented; internal normalized query types are not decoded directly from the model.

```go
type Principal struct {
    OwnerID string
    CredentialID string // session ID or user API-key record ID; never the secret
    CredentialKind string // session | user_api_key
    ExpiresAt time.Time
    Scopes []string
}

type DateRange struct {
    From string `json:"from"` // YYYY-MM-DD, inclusive
    To string `json:"to"`     // YYYY-MM-DD, inclusive
}

type FinanceFilter struct {
    Range DateRange `json:"range"`
    WalletIDs []string `json:"wallet_ids"`
    CategoryIDs []string `json:"category_ids"`
    IncludeChildren bool `json:"include_children"`
    Type string `json:"type"` // all | income | expense
    ReportScope string `json:"report_scope"` // all | included | excluded
    NoteContains string `json:"note_contains"`
    MinAmount *int64 `json:"min_amount"`
    MaxAmount *int64 `json:"max_amount"`
}

type NormalizedQuery struct {
    Key string // server assigned q1..q4, not caller-chosen SQL alias
    Kind string // one of the eight registered read tool names
    Filter FinanceFilter
    CompareRange *DateRange
    ObjectIDs []string // transaction/budget/jar/goal IDs, according to Kind
    Month string // YYYY-MM for jar only
    ActiveOn string // YYYY-MM-DD for budget listing
    Limit int
    Cursor string
}

type SourceScope struct {
    Ref string `json:"ref"`
    Timezone string `json:"timezone"`
    StartAt *time.Time `json:"start_at"`
    EndExclusive *time.Time `json:"end_exclusive"`
    Filter FinanceFilter `json:"filter"`
    AsOf time.Time `json:"as_of"`
}

type Fact struct {
    ID string `json:"id"` // bundle-relative q1.expense, q2.delta
    Kind string `json:"kind"` // money | count | percent_bps | text | date
    Integer *int64 `json:"integer"`
    Text *string `json:"text"`
    Currency *string `json:"currency"`
    SourceRef string `json:"source_ref"`
}

type FinanceResult struct {
    QueryKey string `json:"query_key"`
    Status string `json:"status"` // ok | empty | unsupported | unavailable | partial
    Facts []Fact `json:"facts"`
    ViewKind string `json:"view_kind"`
    View json.RawMessage `json:"view"` // server-built, validated by per-kind codec
    Source SourceScope `json:"source"`
    WarningCodes []string `json:"warning_codes"`
}

type FactBundle struct {
    ID string `json:"id"`
    AsOf time.Time `json:"as_of"`
    Results []FinanceResult `json:"results"`
}
```

Ports in `repository/finance_query.go` (import `context` and entity):

```go
type FinanceReader interface {
    ReadBundle(context.Context, string, []entity.NormalizedQuery) (entity.FactBundle, error)
}
```

Public use-case methods:

```go
func (s *FinanceQueryService) Execute(ctx context.Context, ownerID string,
    query entity.NormalizedQuery) (entity.FinanceResult, error)
func (s *FinanceQueryService) RefreshBundle(ctx context.Context, ownerID string,
    queries []entity.NormalizedQuery) (entity.FactBundle, error)
```

Execute delegates a one-query bundle. ReadBundle owns transaction; every query—including profile timezone, ID visibility, budgets, goals, jars and category joins—uses its `tx`, never outer `db`. Refresh resolves the original calendar labels again using current account timezone; if timezone changed, all queries in final bundle use the same new timezone and emit `timezone_changed`.

Constructor: `NewFinanceQueryPostgresRepository(db *gorm.DB) repository.FinanceReader`. Define advisor errors in `repository/ai_advisor.go` without repurposing existing domain errors: `ErrAdvisorInvalidInput`, `ErrAdvisorRequestConflict`, `ErrAdvisorBusy`, `ErrAdvisorPurged`, `ErrAdvisorLeaseLost`, `ErrAdvisorEventsExpired`. Each is an `errors.New` sentinel mapped by handler to the exact code in §7. Reuse existing `repository.ErrNotFound` for absence. Business/provider errors are not raw PostgreSQL/HTTP error strings returned to users.

### Tool arguments, exact defaults and bounds

Input IDs are nonempty strings ≤128 bytes; use actual catalog IDs, do not impose UUID-only on old system categories. Arrays max 100 IDs, deduped. Dates must round-trip parse, range ≤12 calendar months; note max 200 runes, treated as literal case-insensitive substring, not regex or wildcard. Min/max nullable positive safe integers, min≤max. No owner/timezone/SQL in tool input.

| Tool | Accepted fields | Defaults / validation |
| --- | --- | --- |
| search_transactions | `range`, `wallet_ids`, `category_ids`, `include_children`, `type`, `report_scope`, `note_contains`, `min_amount`, `max_amount`, `limit`, `cursor` | range required; arrays []; children=true; type=all; report_scope=all; note empty; min/max null; limit=20, max50; cursor empty |
| get_transaction | `transaction_id` | required, exact one; report exclusion does not hide detail |
| get_finance_summary | `range`, `wallet_ids`, `category_ids`, `include_children` | range required; report_scope forced included; type all |
| compare_spending_periods | `current`, `previous`, `wallet_ids`, `category_ids`, `include_children` | both ranges required; normalized same filters; expense only; previous shorter range explicitly labeled |
| get_wallet_balances | `wallet_ids`, `scope` | scope=included or selected; selected requires IDs, included disallows IDs; credit shown unsupported, not ordinary assets |
| get_budget_progress | `budget_ids`, `active_on` | exactly one selector, active_on is account calendar date; max50 budgets returned, truncated state otherwise |
| get_goal_progress | `wallet_ids` | explicit goal wallet IDs or [] means list goal wallets, max50 with truncated state |
| get_jar_progress | `month`, `jar_ids` | month required, IDs [] all for month, max50 with truncated state; read-only inherited preview |

Catalog metadata (wallet/category IDs/names and supported domains) is loaded by context builder via bounded scoped reads, not a ninth arbitrary tool. Caps: 100 wallets, 200 categories ordered deterministically; if truncated, require user pick via UI/search catalog before asking model to resolve an absent ID. Do not tell model to invent missing IDs. Account preference timezone missing/unconfirmed produces a configuration question rather than falling back to server timezone.

Strict Go decode helper for each typed tool input, then semantic validation:

```go
func DecodeToolArgs(raw json.RawMessage, dst any) error {
    if len(raw) == 0 || len(raw) > 16*1024 { return errors.New("invalid tool arguments") }
    raw = bytes.TrimSpace(raw)
    if len(raw) == 0 || raw[0] != '{' { return errors.New("object required") }
    decoder := json.NewDecoder(bytes.NewReader(raw))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(dst); err != nil { return errors.New("invalid tool arguments") }
    var extra any
    if err := decoder.Decode(&extra); err != io.EOF { return errors.New("one JSON object required") }
    return nil
}
```

This helper alone is **not** full JSON Schema validation: add required-fields checks, enum/range checks and reject duplicate keys at the schema boundary. Use one schema source for provider definitions/docs/tests; disallow additionalProperties on every object. Null is only legal where specified. Missing optional fields get the defaults above; required fields never silently zero-fill.

No new schema library is required for V1: typed decoders plus explicit validators implement these eight bounded schemas. Write a recursive `json.Decoder.Token` object-key duplicate check before typed decode, including nested objects; tests validate schema examples against the same defaults/bounds. Do not mistake DisallowUnknownFields for duplicate-key rejection.

## 4. Query implementation details

### Normalization and overflow

For calendar labels parse with `time.Parse("2006-01-02", label)`; recreate midnight via `time.Date(y,m,d,0,0,0,0,location)`. End is **next local date** via `AddDate(0,0,1)`, not `Add(24*time.Hour)` (DST). `to` earlier than `from` is 400. Comparisons use normalized date ranges; a current month includes today and comparison uses same count of local dates, clamped to previous-month end.

PostgreSQL `SUM(bigint)` is numeric: scan the aggregate as decimal string or pg numeric, parse with `math/big.Int`, compare to ±9007199254740991 before converting to int64/JSON number. Do not scan to float64. Check subtraction/percentage intermediate arithmetic with big.Int; null percent for denominator≤0. Percent stored as signed basis points rounded half away from zero; FE divides by100 for display. This avoids per-language rounding drift.

### SQL shape

Use bound parameters. Base scope for reporting:

```sql
SELECT t.id, t.type, t.amount, t.category_id, t.occurred_at
FROM transactions AS t
LEFT JOIN categories AS c ON c.id = t.category_id
  AND (c.owner_id = $1 OR c.owner_id IS NULL)
WHERE t.owner_id = $1
  AND t.occurred_at >= $2 AND t.occurred_at < $3
  AND t.included_in_reports = true
  AND t.transfer_id IS NULL
  AND COALESCE(c.system_key, '') NOT IN ('income_transfer_in','expense_transfer_out');
```

Append wallet/category/type/amount/note predicates to the same base scope used by **both** totals and page rows. Escape `\`, `%`, `_` when building ILIKE pattern, or use `strpos(lower(coalesce(t.note,'')), lower($N)) > 0` for literal contains. Validate selected IDs in the snapshot, not merely ignore foreign IDs to produce empty results. Category joins never surface another owner's name even if legacy bad data exists; return a data-integrity warning, not foreign label.

Pagination: `ORDER BY occurred_at DESC,id DESC`, next page predicate `(occurred_at,id) < ($cursor_time,$cursor_id)`; fetch limit+1. Cursor is base64url JSON `{v:1,occurred_at,id,scope_hash}` size≤2KiB; validate all fields and canonical scope hash. It is not auth—owner filter is always applied. Malformed/wrong-scope cursor →400. It is live keyset pagination, not a historical snapshot across multiple HTTP requests; show `as_of` and refresh notice if data changes.

Category breakdown returns leaf/direct category rows, totals cannot double count parent+child. Parent subtree total is separate requested scope; do not add parent rollup back into leaf total. Top-N category presentation includes an Other amount so sum of visible buckets remains equal to total.

Final bundle transaction idiom in `finance_query_postgres.go`:

```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Exec("SET LOCAL statement_timeout = '2000ms'").Error; err != nil { return err }
    // Read account timezone first, then all query kinds on tx.
    // Build FactBundle entirely in memory; persist it only after this transaction closes.
    return r.readBundleOnTx(ctx, tx, ownerID, queries, &bundle)
}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
```

`readBundleOnTx(ctx context.Context, tx *gorm.DB, ownerID string, queries []entity.NormalizedQuery, out *entity.FactBundle) error` is private and must never access `r.db`. It sets `out.AsOf` from the first snapshot read. Query validation error aborts whole bundle; source unavailable yields no pretend zero. Can use `SET LOCAL` because it changes transaction settings, not financial tables. Budget and jar readers are SELECT-only on tx. Do not call existing ListMonth.

PostgreSQL Repeatable Read provides a stable transaction snapshot, but not a promise against future commits after response. Use transaction-local handle throughout. [PostgreSQL isolation](https://www.postgresql.org/docs/current/transaction-iso.html), [GORM transaction guidance](https://gorm.io/docs/transactions.html).

## 5. Proposed persistence DDL

Use the next available migration number at implementation. Below is the core proposed schema, **not an instruction to run SQL now**. IDs remain text to match current user schema; create UUIDs in Go. No `AutoMigrate`.

```sql
CREATE TABLE advisor_conversations (
  id text PRIMARY KEY,
  owner_id text NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,
  generation bigint NOT NULL DEFAULT 1 CHECK (generation > 0),
  next_message_seq bigint NOT NULL DEFAULT 1,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  summary_through_seq bigint NOT NULL DEFAULT 0,
  summary_version bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (id, owner_id)
);
CREATE TABLE advisor_runs (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  conversation_id text NOT NULL,
  generation bigint NOT NULL CHECK (generation > 0),
  client_request_id text NOT NULL,
  payload_hash text,
  credential_kind text NOT NULL CHECK (credential_kind IN ('session','user_api_key')),
  credential_id text NOT NULL,
  credential_expires_at timestamptz,
  status text NOT NULL CHECK (status IN ('queued','running','completed','failed','cancelled','interrupted','purged')),
  lease_token text,
  lease_until timestamptz,
  last_event_seq bigint NOT NULL DEFAULT 0,
  error_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  FOREIGN KEY (conversation_id, owner_id) REFERENCES advisor_conversations(id, owner_id) ON DELETE CASCADE,
  UNIQUE (owner_id, client_request_id)
);
CREATE UNIQUE INDEX advisor_one_active_run ON advisor_runs(conversation_id)
  WHERE status IN ('queued','running');
CREATE INDEX advisor_run_lease ON advisor_runs(lease_until) WHERE status = 'running';
CREATE TABLE advisor_messages (
  id text PRIMARY KEY,
  conversation_id text NOT NULL REFERENCES advisor_conversations(id) ON DELETE CASCADE,
  generation bigint NOT NULL,
  seq bigint NOT NULL,
  run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('user','assistant')),
  parts_version integer NOT NULL DEFAULT 1,
  parts jsonb NOT NULL CHECK (jsonb_typeof(parts) = 'array'),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (conversation_id, seq),
  UNIQUE (run_id, role)
);
CREATE TABLE advisor_events (
  run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
  seq bigint NOT NULL,
  type text NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (run_id, seq)
);
CREATE TABLE advisor_fact_bundles (
  id text PRIMARY KEY,
  run_id text NOT NULL UNIQUE REFERENCES advisor_runs(id) ON DELETE CASCADE,
  as_of timestamptz NOT NULL,
  queries jsonb NOT NULL CHECK (jsonb_typeof(queries) = 'array'),
  results jsonb NOT NULL CHECK (jsonb_typeof(results) = 'array')
);
```

Enforce conversation/run/generation agreement for messages and owner on every repository operation inside the transaction; add a compound FK if later access bypasses the repository. Session credentials require expiry; a future non-expiring user key uses NULL/zero ExpiresAt but still requires live revocation check. JSON per message/result max256KiB by application validation; each complete part must fit64KiB (truncate rows with explicit cursor first), then emit one part per event. Validate sequence values remain JS safe integers. No raw model transcript table V1: persisted history includes only user input, user-visible answer parts and safe context anchors, not chain-of-thought.

Down migration drops fact_bundles → events → messages → runs → conversations, only on disposable migration test DB. New advisor indexes for transactions are evaluated separately using EXPLAIN with fake large datasets; avoid huge blanket indexes/migration rewrite on actual data.

### Repository command types and signatures

Define in `entity/ai_advisor.go`: `Conversation{ID,OwnerID string; Generation,NextMessageSeq,SummaryThroughSeq,SummaryVersion int64; Summary json.RawMessage}`, `Message{ID,RunID,Role string; Generation,Seq int64; PartsVersion int; Parts json.RawMessage; CreatedAt time.Time}`, `Event{RunID string; Seq int64; Type string; Payload json.RawMessage}`, `Run{ID,ConversationID,Status string; Generation int64; Principal Principal; LastEventSeq int64}`, `RunLease{Run Run; Token string; Until time.Time}`. Serialization structs for wire responses have explicit tags and omit Principal/lease.

Port in `repository/ai_advisor.go`:

```go
type AdvisorRepository interface {
    GetOrCreateConversation(context.Context, string) (entity.Conversation, error)
    ListMessages(context.Context, string, int64, int) ([]entity.Message, error)
    StartRun(context.Context, entity.Principal, string, string) (entity.Run, bool, error)
    ClaimRun(context.Context, string, time.Duration) (entity.RunLease, error)
    Heartbeat(context.Context, entity.RunLease, time.Duration) error
    AppendEvent(context.Context, entity.RunLease, string, json.RawMessage) (entity.Event, error)
    CompleteRun(context.Context, entity.RunLease, entity.FactBundle, []entity.MessagePart) error
    FailRun(context.Context, entity.RunLease, string) error
    CancelRun(context.Context, string, string) error
    ClearConversation(context.Context, string) error
    UpdateSummary(context.Context, string, int64, int64, int64, json.RawMessage) error
    ReadRun(context.Context, string, string) (entity.Run, error)
    ReadEvents(context.Context, string, string, int64, int) ([]entity.Event, error)
    InterruptExpired(context.Context) (int64, error)
}
```

StartRun parameters after principal are client_request_id and text; bool reports replay. ListMessages parameters are owner,before_seq,limit (0 cursor means latest). UpdateSummary parameters after owner: expected generation, expected summary_version, through_seq, summary JSON. Cancel/ReadRun use owner,runID. CompleteRun atomically persists final bundle, assistant message, parts events and completion status. UI sees nothing financial that cannot be reloaded from committed history.

### Transaction algorithms

**Accept:** authenticate → max-body/decode → BEGIN → get/create conversation → lock conversation row → check `(owner,client_request_id)` before active-run check → same canonical payload hash returns existing run, changed hash409, purged410 → check active run → insert run+user message+started event using sequence counters → COMMIT → signal runner. Sending an unchanged text with a new request ID is a new user action, not silently deduped by text.

**Claim:** short SELECT queued run `FOR UPDATE SKIP LOCKED`, set running/random lease token/until, COMMIT. Claim only queued, never expired running. Every write afterward checks token/status/generation and lease_until>now. Heartbeat every10s extends45s. Global bounded worker pool avoids unbounded goroutines; a capacity rejection503 occurs before inserting new run. This operational concurrency bound is not daily user usage quota.

**Finish/cancel/clear lock order:** conversation first, then run. Cancellation acquires these locks and makes run terminal; complete after cancel fails fenced update. Clear increments generation, interrupts active run, deletes history/facts/events, nulls hash, sets prior runs purged; retain only request-ID tombstones and safe status so a delayed resend cannot recreate cleared content. Explicit full account deletion cascades everything. Purged run status access returns410. No user financial/config rows deleted or created.

**Restart:** expire old running lease to interrupted, do not automatically re-execute its provider call. Queued runs may start after fresh auth check. Worker crashes after DB completion but before response are recovered by GET. No “exactly once model call” claim; only accepted-run dedupe and V2 ledger idempotency are local guarantees.

**History:** GET never creates a conversation; no history returns `{messages:[],generation:0,next_before_seq:null}`. StartRun creates it. Messages fetched DESC limit+1, delivered ascending within page; older cursor uses first sequence. Clearing increments generation and invalidates all old responses client-side.

## 6. Provider loop, context and facts

Types in `usecase/ai_advisor_provider.go`:

```go
type ProviderMessage struct {
    Role string `json:"role"`
    Content string `json:"content"`
    ToolCallID string `json:"tool_call_id,omitempty"`
    ToolCalls []ProviderToolCall `json:"tool_calls,omitempty"`
}
type ProviderToolCall struct {
    ID string `json:"id"`
    Type string `json:"type"`
    Function struct {
        Name string `json:"name"`
        Arguments string `json:"arguments"`
    } `json:"function"`
}
type ToolDefinition struct {
    Name string
    Description string
    Parameters json.RawMessage
}
type AdvisorRequest struct {
    Messages []ProviderMessage
    Tools []ToolDefinition
    Phase string // plan | answer
    MaxOutputTokens int
}
type AdvisorResponse struct {
    Content string
    ToolCalls []ProviderToolCall
    InputTokens int
    OutputTokens int
}
type AdvisorProvider interface {
    Chat(context.Context, AdvisorRequest) (AdvisorResponse, error)
}
```

Adapter builds OpenAI-compatible `tools:[{type:"function",function:{name,description,parameters}}]`. Planning tool_choice=auto; final answer omits tools and uses approved structured-answer schema. Do not assume provider supports tools and strict response_format simultaneously; separate calls and test each. Read tool-call arguments only after complete response. Unknown tool name is denied before any lookup to SQL/network.

Algorithm, max4 calls to domain tools/max6 model calls **per user run**, including invalid attempts/repair:

1. Fresh credential validation; load profile, bounded visible catalog, context summary and recent turns.
2. LLM plan call; if tools requested append the assistant tool_calls message and a matching tool response for **each** tool_call_id. If batch exceeds remaining budget return a budget-exhausted result for surplus IDs; never execute them.
3. Decode/validate/auth each allowed tool and execute provisional read; record normalized query, safe status and compact facts. Tool return size cap applied with `truncated=true`, not broken JSON. No SQL text/provider credentials returned.
4. Repeat until answer intent or budget boundary. Reserve one model call for final answer; avoid consuming all six in planning. If no financial tools were used, return clarification/general limitations without fabricated totals.
5. Refresh all collected queries in one new read-only snapshot. If any required query fails, return source-unavailable warning; do not blend old successful provisional facts with new data. Recheck auth before refresh and before final publication.
6. Discard provisional values from answer context. Final prompt gets the original question, resolved intent and final bundle only; `fact_id`s now resolve solely against that bundle. Build validated AnswerPlan; one repair only if budget/time remains.
7. Build parts → fenced CompleteRun transaction → event readers see committed events. Invalid final answer still allows deterministic data cards + warning using final facts; not a fabricated text fallback.

`tool.started/finished` labels are operational status, not private reasoning. Do not persist/log raw provider response or hidden reasoning. Prompt source files versioned with schemas; prompts describe untrusted data boundaries and unsupported capabilities, never hold secrets.

### Context budget / summary

V1 summary can be deterministic structured context, not another unbudgeted LLM call:

```json
{"v":1,"through_seq":24,"last_intent":"spending_compare","selected_category_ids":["food-id"],"last_range":{"from":"2026-09-01","to":"2026-09-22"},"unresolved":"choose_wallet"}
```

Append no balances/transaction descriptions. Rebuild from committed tool scopes and explicit user selections; do not pretend this is a complete semantic summary of old free-form messages. Keep most recent12 messages separately; when context is insufficient, ask, rather than reread unbounded history. Future LLM summaries need their own budget/eval task. Summary CAS version/generation/watermark prevents an older task overwriting later intent.

Configured provider context capacity must be known before enablement. Until tokenizer verified, use conservative `UTF-8 byte length + fixed message/schema overhead` as token estimate, including tool schemas/results; choose a provider-specific counter only after testing. Input budget ≤16k tokens and total input+reserved4k output≤provider context. Count serialized request, not just user text. If catalog too large shrink metadata first; never drop system rules or current question. Byte estimate deliberately overcounts Vietnamese and may ask to narrow early.

### Typed parts and numeric text

Store `MessagePart{ID string, Type string, FactBundleID string, Data json.RawMessage}` with explicit JSON tags. Per-kind payload codecs define required fields; envelope is not permission to store arbitrary LLM JSON. Final answer shape:

```json
{"v":1,"sentences":[{"template":"expense_total","fact_ids":["q1.expense"]}],"cards":[{"type":"category_breakdown","query_key":"q1"}],"followups":["compare_previous_period","show_transactions"]}
```

Quantitative sentences use an allowlisted template, **not** arbitrary text with money regex filtering. Templates initially: expense_total(money), income_total(money), net_flow(money), period_delta(money,percent nullable), matched_count(count), no_data(), source_unavailable(), choose_wallet(), choose_category(), scope_too_large(), unsupported_domain(). Labels/entities are validated display data and escaped. Renderer includes proper scope/range from bundle. Model may suggest qualitative commentary only if clearly AI interpretation; V1's guaranteed numeric answer path uses templates/cards, not unverifiable prose. This is a deliberate flexibility tradeoff; expand sentence vocabulary via schema/eval, not silent free-text bypass.

Money templates store references rather than a flattened raw monetary string, so masking is deterministic. When masked, suppress full user message/free-form AI commentary as potentially sensitive; card amount DOM text, title, aria-label, tooltip and copy buttons all use masked value. Masking is a shoulder-surfing feature, not authorization or encryption.

## 7. HTTP errors and public integration contract

All paths below prefix `/api/v1/ai/advisor`. Advisor feature flag false: capabilities200 with `enabled:false`, history still accessible to owner, new messages503 `ADVISOR_DISABLED` and no run insertion. Disabling feature cancels/drains active runs; manual entry/ledger unaffected.

```json
{"data":{"run_id":"run-1","message_id":"message-1","generation":1,"status":"queued","events_url":"/api/v1/ai/advisor/runs/run-1/events"},"meta":null}
```

Submit body `{client_request_id,text}`: request ID UUID, text≤8KiB nonblank; HTTP body≤12KiB, Content-Type application/json. Reject unknown fields and client-supplied messages/owner/tools. Same payload/id returns202 if active,200 if terminal; FE accepts both. Payload hash computed on decoded text bytes + API contract version (not arbitrary key order); do not normalize away intentional whitespace beyond rejecting whitespace-only input.

| Case | Status / code | Client action |
| --- | --- | --- |
| Missing/expired/revoked credential | 401 existing auth code | stop stream, clear local private state, sign in; no POST retry |
| Authenticated key lacks scope | 403 ADVISOR_SCOPE_DENIED | explain permission; no fallback credential |
| Wrong owner/missing run/bundle/object | 404 ADVISOR_NOT_FOUND | generic not found |
| Invalid input/range/cursor | 400 ADVISOR_INVALID_INPUT | correct input |
| Oversize body | 413 ADVISOR_PAYLOAD_TOO_LARGE | reduce input |
| Another run active | 409 ADVISOR_BUSY | recover active run, no second request ID loop |
| Same request ID different body | 409 ADVISOR_REQUEST_CONFLICT | do not overwrite first message |
| Purged request/history, expired replay cursor | 410 ADVISOR_HISTORY_CLEARED / ADVISOR_EVENTS_EXPIRED | purge local cache or GET run final result |
| Provider unavailable/global worker capacity | 503 ADVISOR_UNAVAILABLE | explicit retry, no auto POST |
| Active run deadline/provider failure after202 | GET200 with failed state + safe error code | show partial/warning and retry affordance |

Add missing direct drilldown contract: `GET /fact-bundles/{id}/transactions?scope_ref=q1&cursor=...` returns fresh result page + `source_bundle_as_of`, `as_of`, normalized scope and warning when snapshot may differ. It calls no model and performs no ledger writes. Owner validated via bundle→run. For summary q1, inherits report_scope=included; category drilldown must be a server-issued child scope_ref, not frontend-authored arbitrary category injection. Bundle stores these child scopes. Scope refs may not widen owner or credential scope.

### JWT and API-key principal

Session principal contains claims.SessionID, owner from verified Redis session and JWT expiry. The local implementation also accepts `mpk_...` user keys through the advisor middleware, with authoritative lookup, constant-time hash comparison, expiry/revocation checks and scope checks before tool reads. JWT principal gets only the user's permitted operations. Redis audit, long-lived revalidation and public deployment are still release gates; do not trust a cached active=true value for a future public rollout.

Local key API: JWT-only `POST /api/v1/api-keys`, `GET /api/v1/api-keys`, `DELETE /api/v1/api-keys/{id}`. Key-only `self` endpoints and run cancellation on revocation remain proposed follow-up work. Secret format `mpk_<lookup-id>.<32-byte-random-base64url>`; hash full secret, compare constant-time, never log/return it after creation. Record owner/name/scopes/created/revoked/last_used/optional expiry. Prefix selects verifier; invalid key must not fall back to JWT verifier. Management key access is 403. Advisor requires `finance:read` plus `advisor:chat` for POST; `finance:read` plus `advisor:read` for history/overview/run reads. Confirm scope remains disabled V1. Service Feedback token never matches this principal type.

Key creation enforces `advisor:chat` implies `advisor:read`; POST additionally verifies both scopes at runtime so externally provisioned malformed keys cannot start an unreadable run. All API-key/session auth failures occur before reading response payloads; a model-selected operation can only narrow effective scope, never expand it.

API-key storage is isolated in migration `000019_user_api_keys`, not advisor tables. Contract tests prove digest-only persistence, owner isolation, expiry/revocation, scope checks and the no-JWT-fallback claim; public release still needs Redis audit and deployment evidence.

## 8. SSE wire and frontend state machine

Persist event id, not HTTP chunk number. Example:

```text
id: 3
event: part.upsert
data: {"v":1,"run_id":"run-1","generation":1,"seq":3,"type":"part.upsert","payload":{"id":"p1","type":"metric_group","fact_bundle_id":"bundle-1","data":{"items":[]}}}

```

Headers `Content-Type:text/event-stream`, `Cache-Control:no-store`, `X-Accel-Buffering:no`; flush after each full event; heartbeat comment `: keepalive\n\n` every15s. SSE has blank-line-delimited records, UTF-8 and event IDs; do not assume a reader chunk is a record. [MDN SSE format](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events).

Protocol choice is MyPocket domain SSE v1, **not** AI SDK wire format. If a library adapter is adopted it converts at UI boundary; backend/persisted parts remain vendor-independent. No `[DONE]` marker: typed terminal event closes stream. Auth failure after headers: close without any more financial payload, FE GET revalidates; do not put JSON Problem into an already-open SSE stream.

Frontend types in `aiAdvisorStream.ts`:

```ts
export type AdvisorEvent = {
  v: 1; run_id: string; generation: number; seq: number;
  type: 'run.started' | 'tool.started' | 'tool.finished' |
    'part.upsert' | 'run.completed' | 'run.failed' | 'run.cancelled';
  payload: unknown;
};
export type AdvisorState = {
  generation: number; runID: string | null; lastSeq: number;
  phase: 'idle' | 'submitting' | 'running' | 'reconnecting' | 'failed';
  partOrder: string[]; parts: Record<string, unknown>;
};
```

Reducer: different generation/run →ignore; seq≤lastSeq →ignore; seq>lastSeq+1 →pause/apply none and reconnect from lastSeq; expected seq →validate payload/apply, then advance lastSeq. A valid unknown future part advances sequence while rendering fallback. A malformed event must not advance. `part.upsert` replaces payload by stable part ID, appends order only first time. Run status does not remove sent user text.

Stream adapter: create fetch with Bearer from getStoredToken, `redirect:"error"`, `cache:"no-store"`, AbortSignal. Only same-origin/API configured base URLs, never model URLs. Validate content type/status/body. Use TextDecoder `{stream:true}`, buffer split on blank lines including CRLF, join multiple data lines with newline, ignore comment/retry fields; parse max64KiB JSON event and bound carry buffer128KiB. Cancel reader on unmount/logout, not server run; explicit Stop invokes cancel endpoint.

Reconnection: GET only with jittered1/2/4/8/15s backoff, max5 automatic attempts then “Kết nối lại”; poll run status before retry if terminal may have been reached. Resume at lastSeq; expired events fetch final persisted message. In-flight POST recovery uses same request ID/body, never new UUID on automatic recovery. SessionStorage may hold only request/run IDs, no conversation text/token copy; unknown POST outcome after reload uses `GET /runs/by-request/{client_request_id}` (register this static route before `/:id`). Add this lookup endpoint to handler/catalog; owner-scoped404 does not mean provider definitely did nothing while a previous POST is in flight. Explicit same-key resubmit is safe when user opts to recover.

## 9. UI data contracts and navigation

Card data are domain DTOs, not component props supplied by LLM:

| Part | Required server fields | Click behavior |
| --- | --- | --- |
| metric_group | items `{key,label,value,currency}`, source range/as_of | show matching scope only if scope_ref exists |
| transaction_list | items transaction summaries, total_count, next_cursor, scope_ref | detail by ID; load more via scoped drilldown |
| transaction_detail | ID, wallet/category labels, amount/type/date/report flag/transfer marker | navigate existing detail; no automatic edit |
| category_breakdown | items `{id nullable,label,amount,count,scope_ref}`, total | select server scope_ref, not label |
| period_comparison | current/previous range and totals, delta, percent_bps nullable, category deltas | matching current/previous child scopes |
| budget_status | budget ID/limit/spent/remaining/start/end/scope | open budget; V1 no create button |
| goal_progress | goal-wallet ID/current/target nullable/remaining/target_date | existing goal details |
| jar_progress | jar ID/month/spent/allocation nullable/config_origin | jar/month page; no carryover label |
| warning | allowlisted code + localized text | retry only for recoverable source errors |
| text | template + fact refs resolved at server, or clearly qualified general text | no arbitrary links/HTML |

Base work first: specify AssistantComposer mode, tool status live-region, message list scroll anchoring, metric/comparison/category molecules; then screen. No `dangerouslySetInnerHTML`; plain text V1, no Markdown dependency needed. Unknown payload/type turns into warning component rather than cast `as` and crash.

Extend PrototypeTab with assistant, tabFromPath and onTabChange mapping, TanStack route. Header and global Add remain reachable; six-slot navigation composition uses BaseNavigationItem/BaseFab, while Composer respects bottom-nav height/safe area. Set a shared CSS layout variable for nav height rather than hardcode unrelated offsets. At 360px every touch target remains≥44px, long Vietnamese labels wrap/shorten accessibly. If six slots cannot meet measured layout gate, stop UI subtask and revise base layout—not overlap controls or drop a destination silently.

Opening Assistant does not call model: overview endpoint pulls real figures; suggestions are capability-derived. Empty history and empty financial data are separate. Provider disabled leaves history/overview available; no fake greetings with amounts. Existing history collapses dashboard header. On logout abort readers and clear private in-memory parts; no service-worker/HTTP cache for messages/facts.

## 10. Deterministic test fixtures and commands

Use disposable owners with fixed clock `2026-09-22T08:00:00Z`, timezone Asia/Ho_Chi_Minh, zero-opening basic wallets A/B. Amounts intentionally small integer đồng to detect accidental thousands conversion. Fixture G1:

| Row | Wallet | Type | Amount | Report | Category / link |
| --- | --- | --- | --- | --- | --- |
| salary | A | income | 1000 | true | income category |
| breakfast | A | expense | 100 | true | Food parent |
| coffee | A | expense | 50 | true | Food child |
| private | A | expense | 20 | false | Food parent |
| transfer-out | A | expense | 200 | false | transfer_id=X/system category |
| transfer-in | B | income | 200 | false | transfer_id=X/system category |

Expected report income1000/expense150/net850; wallet A630/B200/sum830; Food subtree150; all-search count6, report-search count3; month note untouched. Transfer alone changes total by0; private expense explains net-flow vs balance difference20. Previous aligned period Food100 → current150, delta50, 5000bps. Separate fixture G2 has51 ordinary expenses×1 so total51/page50; never mix G2 into G1 expectations.

Fixture insertion must persist false flags explicitly via parameterized SQL or a map insert: GORM entity fields tagged `default:true` may turn a zero-value false into the default during Create. Read back `included_in_reports`/`is_in_total` before assertions; fail fixture setup if values differ. This prevents a misleading red test caused by fixture defaults rather than finance logic.

Critical named tests (must be real test functions, no print-only evidence):

| Test | Fixture / expected |
| --- | --- |
| TestFinanceQueryGolden | G1 exact totals, all/report counts, category subtree, balances |
| TestFinanceQueryPaginationTotal | G2; count51, amount51, pages50+1, no duplicate ID |
| TestFinanceQueryJarDoesNotInitializeMonth | no current-month rows; tool returns inherited_preview, DB tables unchanged |
| TestFinanceQuerySnapshotIsolation | barrier after first query, writer commits; all final queries match one snapshot |
| TestFinanceQueryBoundaryAndZeroBaseline | local midnight, DST, leap/short month, zero previous→null percent |
| TestAdvisorStartRunReplayConcurrent | 20 concurrent submits same key→1 user message,1 run,1 provider claim |
| TestAdvisorClearFencesLateCompletion | clear during fake provider block→no history restoration/no private event |
| TestAdvisorUnknownAndWriteToolDenied | fake model requests SQL/create_transaction; executions0 |
| TestAdvisorRevokeDuringRun | revoke session/key before final reply→no financial event, terminal denied |
| TestAdvisorReadOnlyFinancialState | all financial/config tables unchanged before/after entire run; only advisor/audit records allowed |
| TestAdvisorEventsReconnect | drop after seq3, reconnect→seq4+, no POST/model call |
| TestAdvisorExpiredLeaseNoProviderReplay | kill claimed run, expire lease→interrupted; restart never calls provider for it |
| TestAdvisorMigrationRoundTrip | disposable DB up, step down one, up; unrelated schema/data unchanged |

Repository fixture helper must fail closed for release tests:

```go
func requireAdvisorDB(t *testing.T) *gorm.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" { t.Fatal("TEST_DATABASE_URL must point at an isolated migrated database") }
    parsed, err := url.Parse(dsn)
    if err != nil || !strings.HasPrefix(strings.TrimPrefix(parsed.Path, "/"), "mypocket_advisor_test_") {
        t.Fatal("refusing non-advisor-test database")
    }
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil { t.Fatal("cannot connect to isolated test database") }
    return db
}
```

Credential provisioning is local environment setup, not literals in committed commands. Set TEST_DATABASE_URL securely to an explicitly named disposable DB; migrate with DATABASE_URL pointing to **that same DB**. Do not source arbitrary env scripts, print DSNs, use default `postgres` database or run cleanup against a broad schema. Add cleanup immediately after first fixture insert via t.Cleanup; use exact fixture IDs, never global TRUNCATE. DDL rollback test owns its isolated DB and requires separate teardown from data tests.

Commands from backend after implementation, with test env already set:

```sh
rtk go test ./internal/entity -run 'TestFinance|TestAdvisor' -count=1
rtk go test ./internal/infrastructure/repository -run 'TestFinanceQuery|TestAdvisor' -count=1 -v
rtk go test ./internal/usecase ./internal/controller/http ./internal/infrastructure/ai -run TestAdvisor -count=1
rtk go test -race ./... -count=1
rtk proxy go generate ./cmd/api
```

Frontend after scripts exist:

```sh
rtk npm run test:advisor
rtk npm run test:advisor:integrated
rtk npm run test:advisor:e2e
rtk npm run test:ai
rtk npm run test:transfer:integrated
rtk npm run test:transfer:e2e
rtk npm run check:design
rtk npm run test:design
rtk npm run build
```

Browser test must use real Chrome interaction + real Go API + isolated DB; provider only is fake. Persist sample message, compare card's structured value150 against DB/API, click Food scope shows exactly breakfast/coffee, reload restores same message, Stop becomes cancelled. Add unknown card/null data, keyboard/IME, masked DOM+ARIA/copy and 360/390px screenshots. Static fixture screenshot tests alone do not prove backend permissions or read-only behavior.

Live provider eval is separately opt-in, approved config only, synthetic data, no financial table writes. Store pass counts/task-completion/cost usage/latency, not provider secrets; 30 cases×3 from spec. Technical test gate passing without this eval means fixture-ready, not configured-provider ready.

## 11. Operations and rollout

Proposed config: `AI_ADVISOR_ENABLED=false`, `AI_ADVISOR_MAX_CONCURRENT_RUNS=4`, `AI_ADVISOR_CONTEXT_WINDOW_TOKENS` required when enabled. Reuse AI_BASE_URL/API_KEY/MODEL; do not create frontend secrets. Limits90s/30s/4tools/6calls and event retention24h are versioned constants initially; no huge matrix of env knobs.

Event cleanup runs as bounded opportunistic maintenance/startup pass, deletes **only terminal** runs' events older24h in batches; history/facts stay until explicit clear. Active run events never expire out from under a stream. Dashboard/facts/history are no-store. In-process runner shutdown stops claims, requests cancellation, waits bounded time; remaining leases expire to interrupted. Do not add a second worker/Redis job queue for V1.

Audit read attempts: event IDs, owner/credential ID, allow/deny, tool name, run ID, reason/latency only. Redis XADD context timeout200ms, bounded queue, dropped/delivery-failure counter + redacted log. Do not silently claim durable audit on Redis outage; V1 is best-effort read audit with explicit observability. V2 write events require a PostgreSQL outbox committed with the mutation before rollout. Health distinguishes Redis session auth unavailable (fail closed) from audit delivery degraded.

Release checklist: migration clean; all DB tests actually ran; no ledger/config writes from read tools; fake-model E2E; live-provider eval; proxy SSE reconnect; 401/403/404/key revocation; mobile/masking; current public docs. User-key milestone mandatory for “third-party ready”. Turn flag off to roll back feature behavior; keep schema/data, never destructive down migration on real history.

## 12. V2 extension seams (not V1 work)

Add `PreparedAction` distinct from `NormalizedQuery`, so an LLM tool cannot smuggle a write into read registry. Interface `Prepare(ctx,principal,intent)→proposal`, `Confirm(ctx,principal,id,version,key)→result`. Proposal canonical payload lives server-side; confirm accepts no replacement payload. Kind-specific domain services own transaction/category/budget/goal validation; no handler-to-handler HTTP calls.

Before V2, add monotonic version to every mutable target used for stale checking (updated_at alone is not guaranteed sufficient), action unique execution constraint, sorted lock order, expiry, complete before/after diff, actor-scoped audit outbox. Confirm transaction = auth recheck + action lock + target version validation + domain mutation + action/result + outbox. Provider/network/audit Redis delivery happens outside. For batch, define per-action partial success visibly; never imply whole-batch atomicity when each action commits separately.

Rules/merchant/recurring/forecasts/proactive notifications still require their own domain slices. No bank/payment tool may be introduced as a “small” extension of this registry. Do not combine pending Feedback service-token endpoints with finance actions.

## 13. Required documentation per task

- Internal: proposed→implemented distinctions, exact routes/errors, table ownership, snapshot semantics, auth scope matrix, lifecycle/runbook and test evidence.
- Public humans: two AI flows, unsupported domains, data-sharing, history/clear, transient vs failed run, drilldown freshness and no usage quota vs request bounds.
- Public agents: versioned schemas/tool catalog/parts/events examples, API-key scopes, cursor/idempotency/error handling; all available as Markdown/JSON, linked from docs entry. No proposed endpoint in generated live Swagger before handler exists.
- Plan changes need context/backlog/validation note; runtime tests are not claimed for documentation. New ADRs only when reviewed choices are adopted. Commit explicit task files; never `git add .` over existing Feedback/transfer work.

Review checklist: method names/types match task plan; read-only jars separated; deterministic summary avoids uncounted LLM call; no undefined UI recovery endpoint; no SSE token in URL; no raw financial logs; no invented migration down CLI; no per-user quota; no claims that a mocked run proves actual provider integration.
