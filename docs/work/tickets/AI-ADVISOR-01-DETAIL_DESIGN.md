---
artifact_type: detail_design
id: AI-ADVISOR-01
status: in_review
owner: shared
updated: 2026-09-22
approval: pending_owner_review
implementation: local_slice_in_progress
source: FB-004
---

# Finance Assistant — thiết kế tab AI riêng

## 1. Quyết định đề xuất

Giữ Go modular monolith và React/base components. Xây **trợ lý đọc dữ liệu thật, sau đó mới thêm hành động có xác nhận**, không nhúng một chatbot toàn quyền vào app.

Hai luồng độc lập:

- **Nhập nhanh:** giữ hold Add → text/ảnh/PDF → OCR bên thứ ba → proposal → user sửa/duyệt. Không thêm history/chat tài chính vào flow này.
- **Trợ lý tài chính `/assistant`:** một hội thoại bền vững/account, dashboard thu gọn, hỏi đáp nhiều lượt và card tương tác. V1 chỉ đọc; V2 mới đề xuất thay đổi. Không tích hợp ngân hàng/thanh toán.

Thiết kế này cập nhật phần advisor của [DESIGN-09-AI](TICKET-09-DETAIL_DESIGN.md); không ghi đè contract nhập nhanh đã triển khai. Request đã chuyển sang **Implement**; local slice đã có query layer, migrations, tool loop, JWT/API-key JSON routes và tab UI.

Trace: [intake FB-004](../FEEDBACK_LOG.md), [ticket hỏi đáp](TICKET-09-02-hoi-dap-tai-chinh.md), [backlog](../BACKLOG.md), [implementation plan V1](../../superpowers/plans/2026-09-22-finance-assistant-v1.md), [validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md). Không có active phase. ADR sẽ được tạo khi chốt kiến trúc triển khai, không coi đề xuất này là ADR đã duyệt.

Owner follow-up “Writing plan chi tiết và chỉ dẫn kỹ thuật” được đáp ứng bằng [Technical Implementation Guide](../../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md). Các contract SSE/recovery/Redis-audit vẫn là release gate; runtime hiện tại deliberately dùng synchronous JSON để khóa vertical slice trước. API-key contract đã được implement locally, nhưng chưa có public deployment claim.

### Implementation checkpoint — 2026-09-22

- **Đã chạy local:** semantic finance query layer (owner scope, report semantics, keyset search, category subtree, zero baseline, wallet/goal/budget/jar read views), migrations `000018_advisor_conversations` and `000019_user_api_keys`, durable conversation/run repository, tám read-only tools, OpenAI-compatible provider adapter, bounded orchestration loop, JWT/API-key JSON routes và `/assistant` tab.
- **Đã kiểm chứng:** `TEST_DATABASE_URL=... go test -race ./... -count=1`, `go generate ./cmd/api`, `npm run typecheck`, `npm run check:design`, `npm run build`; PostgreSQL fixture cũng kiểm tra `get_transaction` owner isolation.
- **Chưa release-ready:** SSE/reconnect, clear/purge/lease recovery đầy đủ, server-built typed card parts, overview/drilldown endpoints, durable Redis audit delivery/alerting, configured-provider browser E2E và deploy. Không gọi đây là public third-party API.

## 2. Hiện trạng đã đối chiếu code

| Năng lực | Hiện tại | Việc cần làm |
| --- | --- | --- |
| Frontend | React 19, Vite 7, Tailwind 4, TanStack Router; đã có thin Finance Assistant shell dùng bases | Typed parts/cards, SSE state machine và browser state proof còn lại |
| AI entry | `infrastructure/ai/client.go` dùng JSON schema; **từ chối tool_calls** | Thêm provider interface advisor, không sửa extractor thành agent |
| History/context/tool loop | Durable history/run + bounded tool loop đã có local | Summary anchors, lease recovery và final fact refresh đầy đủ còn lại |
| Báo cáo tháng | `controller/http/month_handler.go` tính từ `Transactions.List(owner)` | Tách query/aggregate có filter, phân trang và scope dùng chung |
| Ví/budget/hũ/goal | Có API; goal là ví loại `goal` | Tool gọi use case dùng chung, không tạo Goals domain trùng |
| Đọc hũ | `JarPostgresRepository.ListMonth` tự ensure/copy tháng trong DB transaction | Tool advisor phải có SELECT-only path; inherited config chỉ là preview, không materialize |
| Chuyển ví | Có paired transfer trong working tree; edit/delete/idempotency còn open | Không thêm tool sửa từng transfer leg |
| Merchant | `entity.Transaction` chưa có merchant ID/name chuẩn | V1 chỉ tìm text trong note; không gọi đó là merchant breakdown |
| API key của user | Local middleware now accepts owner-scoped `mpk_...` keys for advisor read/chat scopes; best-effort redacted Redis audit is wired | Local contract is implemented and tested; public third-party release still requires durable audit/provider/browser/deploy evidence |
| Recurring/rules/forecast/notification | Chưa phải domain hoàn chỉnh trong scope khảo sát | Các phase sau có prerequisite riêng |
| Feedback → Fix → Changelog | Đã wire persistence/auth/services/HTTP/UI local | Browser round-trip/deploy vẫn độc lập; không cho advisor dùng developer service token |

Khảo sát giới hạn ở các file trên, repository transaction, account timezone, report requirements, AI/assistant specs, middleware/router, dependency manifest và base inventory. Không chứng nhận toàn bộ app đã test. Không gọi provider thật trong lượt lập plan.

## 3. Kiến trúc và quyền sở hữu

```text
React /assistant + shared Finance cards
                 | authenticated HTTP (local JSON; SSE release gate)
                 v
Go HTTP handlers -> Advisor use case -> LLM adapter
                       |                   |
                       +<-- tool requests--+
                       |
                 Allowlisted registry
                       |
              Finance query/use case layer
                       |
            Owner-scoped repositories -> PostgreSQL

PostgreSQL: ledger, conversation, runs, facts, action state
Redis: session integration, audit delivery; never ledger authority
```

Tool **không phải gọi vòng lại HTTP API bằng token của user**. Tool và HTTP handler cùng gọi use case; auth principal do server truyền vào. LLM chỉ chọn tên tool/tham số schema; không biết SQL, DB credentials, internal token hoặc URL tùy ý.

Giữ cấu trúc repo: entities tại `backend/internal/entity/`, orchestration tại `usecase/`, ports tại `repository/`, adapters tại `infrastructure/`, routes tại `controller/http/`. Không tạo một `/internal/ai` song song chồng trách nhiệm. Python, microservice, vector DB, multi-agent chưa cần cho V1.

## 4. Finance Intelligence: đúng nghĩa trước khi đẹp

`FinanceQueryService` là nguồn số cho advisor và monthly summary; các màn report/dashboard mới phải tái dùng cùng contract. Không tính tổng từ trang search 50 dòng.

- Money: VND nguyên đồng, kiểm tra overflow và giới hạn safe integer phía UI; tổng vượt giới hạn trả lỗi rõ, không mất độ chính xác.
- Khoảng ngày user chọn là inclusive calendar dates; normalize về UTC `[start_at, end_exclusive)` bằng IANA timezone account. Không lấy timezone server. Response luôn có timezone/range.
- Thu/chi báo cáo: chỉ dòng `included_in_reports=true`; loại transfer ID và system transfer categories, kể cả dữ liệu cũ. Phí thực sự có giao dịch chi riêng mới tính; không tự suy ra phí.
- Số dư ví tính từ ledger đầy đủ. `is_in_total=false` chỉ điều khiển tổng ví mặc định, không tự xóa chi tiêu khỏi report. Số dư ví không được gọi là net worth đầy đủ khi credit/debt/portfolio chưa được tính.
- Nhóm cha bao gồm chính nó và con, khử trùng ID. Tìm tên trùng phải hỏi hoặc trả lựa chọn ID có nhãn cha/ví. Uncategorized là bucket thật.
- Budget giữ đúng kỳ ngày, ví/nhóm và quy tắc hiện hành. Hũ là theo dõi chi, không là ví có số dư hay carryover. Goal dùng ví goal, không trừ thêm khoản tiền đã dành trong ví khi tính affordability.
- Read-only bao gồm không tạo/sửa cấu hình hũ. Tháng chưa khởi tạo chỉ render inherited preview có nhãn nguồn; các bảng hũ/cấu hình/note cũng phải giữ nguyên sau chat, không chỉ kiểm tra ledger.
- So sánh mặc định current-month-to-date với cùng tiến độ tháng trước, cắt nếu tháng trước ngắn hơn; explicit full-period được hỗ trợ với nhãn riêng. Baseline 0 → `change_percent=null`, có absolute delta; không Infinity/100% tùy tiện.
- `income - expense` gọi **thu trừ chi**, không mặc định là “đã tiết kiệm”. Transaction note không chứng minh chi tiết món hàng: “Shopee mua gì?” chỉ trả nội dung thật/thiếu chi tiết.
- Distinguish `ok`, `empty`, `unsupported`, `unavailable`, `partial`. Mảng rỗng là `[]`, không null. Không convert unavailable thành số 0.

### Snapshot nhất quán

Không giữ DB transaction khi đợi LLM. Trong tool loop, lưu normalized read requests và kết quả tạm. Trước câu trả lời cuối, chạy lại toàn bộ tối đa 4 requests trong **một read-only REPEATABLE READ transaction ngắn**, cấp `fact_bundle_id/as_of` mới rồi đóng transaction. Answer/card chỉ dùng bundle cuối. Nếu model yêu cầu tool mới sau bước này, kết thúc lượt bằng câu hỏi thu hẹp, không trộn bundle. `as_of` là thời điểm snapshot, không phải cam kết ledger không đổi sau đó.

## 5. Tool catalog theo năng lực

V1 registry chỉ đăng ký những tool có service + contract test. Không có SQL/query_finance generic.

| Tool V1 | Input chính | Kết quả |
| --- | --- | --- |
| `search_transactions` | ngày, wallet/category IDs, include_children, loại, note_contains, min/max, cursor | trang ≤50, count/total riêng toàn scope, report_scope rõ |
| `get_transaction` | transaction_id | chi tiết owner-scoped, transfer label và link nguồn hợp lệ |
| `get_finance_summary` | ngày, wallet/category IDs | income, expense, net, count, category breakdown |
| `compare_spending_periods` | hai normalized ranges, cùng filters | tổng/delta, category contributors, baseline-zero state |
| `get_wallet_balances` | wallet IDs hoặc included-in-total | số dư theo ví + coverage, không giả tài sản ròng |
| `get_budget_progress` | budget IDs hoặc active-on date | limit/spent/remaining + scope |
| `get_goal_progress` | goal wallet IDs | target/current/remaining/target_date |
| `get_jar_progress` | tháng, jar IDs | allocation nullable, actual, coverage |

Search mặc định `report_scope=all`, các aggregate báo cáo mặc định `included`; drilldown từ card kế thừa chính xác filters của aggregate. Period max 12 calendar months/lượt; date range lớn hơn đề nghị chia kỳ. Đây là bounded query, không phải quota usage user.

V1 chưa có `get_spending_by_merchant`: trước hết thêm merchant entity/normalization, unknown bucket và manual correction. Không dùng fuzzy note để khẳng định merchant. `get_spending_by_category`/income/cashflow là typed views của cùng summary service, không cần ba cách tính riêng.

V2 tools chỉ `propose_*` cho nghiệp vụ đã hoàn thiện: transaction create/edit/category/note, budget create/update, goal wallet create. Rules, transfer sửa/xóa, credit và các đối tượng thiếu service giữ `unsupported`. Xóa luôn là proposal riêng, không auto-execute. V3 mới thêm detection/forecast khi nguồn đủ; xem roadmap mục 12.

## 6. Conversation, context và memory

Một active conversation/account, không thread picker/branching/regenerate V1. PostgreSQL lưu toàn bộ lịch sử đến khi user xóa; tải cursor 50 messages/trang. Thêm “Xóa lịch sử” có xác nhận; xóa chat không xóa ledger/AI-entry evidence. Không âm thầm đặt TTL lịch sử. Chính sách backup/provider retention phải ghi riêng, không hứa xóa tức thì ở hệ thống bên ngoài.

Mô hình đề xuất (migration number chọn lúc thực thi, không đụng 16/17 đang dùng):

| Record | Trường/ràng buộc chính |
| --- | --- |
| `advisor_conversations` | id, owner_id unique, generation, summary, summary_through_seq, summary_version |
| `advisor_messages` | id, conversation_id, seq unique trong conversation, role, parts_version, parts JSONB, run_id, created_at |
| `advisor_runs` | id, owner, conversation generation, client_request_id + payload_hash unique owner, status, lease_token, lease_until, last_event_seq |
| `advisor_events` | run_id + seq unique, type, payload, created_at; replay active run, compact sau 24h |
| `advisor_fact_bundles` | id, run_id, scope, as_of, facts/provenance JSONB; cùng vòng đời history |
| `advisor_actions` (V2) | owner, run, kind, canonical payload/hash, version, expiry, expected target versions, status, result_refs |
| `advisor_preferences` (V2) | key/value, source_message_id, user_confirmed_at, version; user xem/sửa/xóa |

Context mỗi lượt: policy → account timezone/currency/current date → summary có watermark → tối đa 12 recent messages → current question. Không load toàn bộ giao dịch, attachments hay toàn bộ tool payload cũ. Tổng input giới hạn 16k tokens hoặc context window provider thấp hơn, dành chỗ 4k output; adapter phải có estimator bảo thủ và kiểm tra provider capability. Tool payload tối đa 64 KiB, user message 8 KiB, summary 2k tokens. Khi vượt budget, giảm history/result verbosity, không âm thầm cắt rules/current question.

Summary chỉ giữ intent, entity references và unresolved question; ví dụ lần hỏi trước là Food/tháng 9. V1 dùng structured context anchors dựng deterministic từ tool scopes/selection đã lưu, không phát sinh một LLM-summary call ngoài budget; không hứa tóm tắt đầy đủ mọi lời nói cũ. Số dư/số chi trong summary **không authoritative**. Câu “so tháng trước?” phải resolve khoảng tương đối từ kỳ đang nói đến, không luôn từ hôm nay. Summary cập nhật compare-and-swap watermark/generation; clear history hoặc lượt mới ngăn late result ghi lại context cũ.

Memory dài hạn chỉ lưu preference user duyệt, không tự chép “mua trà sữa 45k” từ ledger. V1 chỉ dùng profile hiện có và conversation summary; chưa thêm tự động ghi nhớ. Summary, notes, OCR và user preference đều là dữ liệu không tin cậy, không nâng quyền.

## 7. API và streaming contract

Prefix giữ `/api/v1`, không thêm `/ai/chat` không version song song.

| Endpoint | Contract |
| --- | --- |
| `GET /ai/advisor/capabilities` | enabled, read tools, supported parts, write=false V1; không secret |
| `GET /ai/advisor/overview?month=YYYY-MM` | implemented local read-only summary; không gọi LLM |
| `GET /ai/advisor/conversation/messages?before_seq=...` | history owner-scoped, 50/trang |
| `POST /ai/advisor/messages` | local: `{client_request_id, text}` → synchronous JSON `{run_id, conversation_id, text, status}`; planned release: 202 + events URL |
| `GET /ai/advisor/runs/{id}` | authoritative status + final message refs |
| `GET /ai/advisor/runs/by-request/{client_request_id}` | planned recovery endpoint; chưa mount |
| `GET /ai/advisor/runs/{id}/events?after_seq=...` | planned authenticated SSE; chưa mount |
| `GET /ai/advisor/fact-bundles/{id}/transactions?scope_ref=...&cursor=...` | planned owner-scoped drilldown; chưa mount |
| `POST /ai/advisor/runs/{id}/cancel` | cancellation persisted; terminal replay no-op |
| `DELETE /ai/advisor/conversation` | planned JWT-only clear/purge; chưa mount |
| `POST /ai/advisor/actions/{id}/confirm` (V2) | version + idempotency key; server stored payload only |
| `POST /ai/advisor/actions/{id}/reject` (V2) | expected version; terminal state |

Ordinary responses use existing `{data,meta}` and problem details. SSE envelope `{v:1,run_id,seq,type,payload}`; types `run.started`, `tool.started`, `tool.finished`, `part.upsert`, `run.completed`, `run.failed`, `run.cancelled`. Event `id` is seq; persist before sending. Cards only sent as fully schema-validated `part.upsert` (not half JSON). V1 streams progress and complete answer parts; token-by-token unvalidated financial prose is deliberately not shipped.

Reconnect fetches missing events; old compacted event cursor gets 410 + run status URL. FE uses authenticated fetch streaming (Bearer header), not token in URL/native EventSource. Proxy buffering off, cache off, heartbeat 15s; no sensitive content in access-log query. GET replay never calls model again.

One active run/conversation via PostgreSQL claim/unique constraint, concurrent different requests → 409 busy. Same request ID/same payload returns same run; changed payload → 409. Lease heartbeat 10s, expiry 45s, run deadline 90s. In-process Go runner V1 claims before execution; on restart, expired executing runs become interrupted, no blind provider retry. Claim pending work on startup; no cron service required. Cancellation/lease token/generation checked before every tool and final commit.

Loop limits: at most 4 tool executions, at most 6 provider requests including final response/one repair, 30s per provider HTTP call within 90s run deadline. Return grounded partial result with warning when capped; no infinite retries. No per-user daily/monthly usage quota. Client can explicitly send a new request after an interrupted run; warn provider work may have been billed.

## 8. Authorization, API keys, audit

V1 browser pilot can run with JWT, and the local integration surface accepts server-hashed high-entropy user keys with one-time secret display and owner/revocation/scopes checked at the advisor boundary. Supported advisor scopes are `finance:read`, `advisor:read`, `advisor:chat`; V2 action confirmation additionally needs `advisor:confirm` plus the relevant domain write scope. Key creation cannot grant scopes user lacks; keys cannot create/manage other keys. User-session-only destructive history clear. Do not silently switch a key failure to JWT acceptance. Redis audit and public deployment evidence are still required before calling this public-ready.

Tools receive authenticated Principal, never `user_id` supplied by model. Validate every wallet/category/jar/goal ID and all derived result IDs. Cross-owner resource is 404. Revocation checked before tool execution and before sending final financial result, not only initial POST. Capability listing filters tools by credential scopes. Service key for Feedback has no finance/advisor access.

Redis audit carries IDs, actor credential ID/type, operation, allowed/denied, status/reason, latency, model/prompt version and usage counts. No keys, prompts, notes, OCR text, monetary rows or signed URLs. Add bounded audit delivery with durable PostgreSQL outbox for V2 writes; enqueue audit and ledger change in same transaction, relay to Redis. Redis outage must not cause ambiguous duplicate writes; retain backlog/alert. Read audit delivery failure is explicit operational warning, not data leakage or silent success in telemetry. Audit never substitutes PostgreSQL authorization/idempotency.

No tools for network fetch, shell, SQL, arbitrary UI, API-key management, permission changes or confirmation. Prompt injection through notes must fail even if model attempts a disallowed tool. Redact provider error bodies. Raw model reasoning is not persisted/rendered; UI shows tool status and short evidence-backed explanation only. Explain what data is sent to model before first chat; keys remain backend-only.

## 9. Structured cards và UX contract

LLM chọn **reference tới facts**, không tự cấp số tiền/chart payload. Backend kiểm tra references rồi build immutable message parts từ fact bundle. Quantitative sentences dùng allowlisted sentence templates + fact references được server format (chi tiết trong technical guide), không chỉ regex tìm số trong prose. Số/percent/comparison chưa có fact không được render như kết luận. Free text cho qualitative explanation không được coi là nguồn số, phải phân biệt nhận định AI; unknown fact → fail closed + card dữ liệu thật/thông báo không đủ dữ liệu. Không HTML/JSX/SQL từ model.

```json
{
  "id": "message-server-id",
  "parts_version": 1,
  "parts": [
    {"id":"p1","type":"text","text":"Chi tiêu trong kỳ đã chọn:"},
    {"id":"p2","type":"metric_group","fact_bundle_id":"bundle-server-id",
     "data":{"currency":"VND","items":[{"key":"expense","value":3250000}]},
     "source":{"timezone":"Asia/Ho_Chi_Minh","from":"2026-09-01","to":"2026-09-22","as_of":"2026-09-22T08:00:00Z"}}
  ]
}
```

Registry whitelist V1: `text`, `metric_group`, `transaction_list`, `transaction_detail`, `category_breakdown`, `period_comparison`, `budget_status`, `goal_progress`, `jar_progress`, `warning`. V2 adds `confirmation`, `action_result`; V3 adds `time_series`, `insight`, `forecast`. Unknown type/version shows accessible “Nội dung chưa được hỗ trợ” and keeps message/history usable; null/bad money rejected before renderer, never blank page.

| Region / action | Base-first composition |
| --- | --- |
| Tab, title, monthly scope | BaseNavigationItem, Heading/Text, DateField / existing month selector |
| Monthly metrics | SurfaceCard + numeric Text; new MetricGroup molecule |
| Suggested prompts | BaseButton chip; click sends visible user message |
| History and tool states | new AssistantMessageList/ToolStatus molecules using Text, StatusMessage; no custom screen controls |
| Composer | reuse AssistantComposer with attachment disabled for advisor V1; enter sends, shift-enter newline, IME composition safe |
| Transactions / budgets / goals | TransactionItem, BudgetProgressItem, GoalCard; validated adapter data |
| Categories / comparisons | base progress/bar primitives + accessible table/list of same numbers; no color-only meaning |
| Confirm actions | BaseBottomSheet + BaseButton, shared field selectors; diff, scope/count, cancel/confirm |

Initial screen shows deterministic current-month summary and suggestions **only for enabled capabilities**. No fake insight with no history. Existing chat collapses dashboard to compact header, history scroll with load older, composer sticky above keyboard/safe area. Preserve scroll anchor on pagination; new message does not force scroll if user reads old history. Visible jump-to-latest button uses base. 390px mobile and desktop keyboard/screen-reader proof required.

Navigation refinement: current base shell has four tabs plus centered Add. Proposed six-slot layout keeps overview/transactions/Add/budgets/assistant/account reachable without overlapping FAB; verify 360/390/430px and ≥44px hit targets. Extend AssistantComposer with advisor mode so labels are financial Q&A and upload is hidden, while entry defaults remain unchanged. Không nhét tab mới vào grid cũ mà không test layout.

Card “Xem giao dịch” passes typed `{kind:"show_transactions", fact_bundle_id, scope_ref}`. Backend rechecks owner, resolves normalized filters; FE opens list/detail or sends an explicit chat drilldown intent. It never follows model-provided arbitrary URL/method. Old card shows snapshot time; a fresh drilldown may differ and says so. Selection for future bulk category update must show exact selected IDs/count and preview, not silently select every query match.

Mask balance mode also masks amounts in assistant text/cards and accessible labels; free-form user/AI text is hidden in this mode because it may contain sensitive amounts. Do not hide only headline while exposing the same number in prose. History is private, not public cached/service-worker content. Streaming errors preserve sent message and show retry/recovery without re-submitting automatically.

### Library choice, verified against official docs 2026-09-22

- Prefer a **small MyPocket chat shell using existing bases**, plus a bounded `assistant-ui` ExternalStoreRuntime spike for lifecycle/accessibility reuse. Its adapter accepts externally owned state/custom message conversion, so domain parts stay ours. Adopt only if no second visual system, no breaking native-control guards, no forced thread model; otherwise retain the thin shell. [Official ExternalStoreRuntime](https://www.assistant-ui.com/docs/runtimes/custom/external-store).
- AI SDK supports a custom-backend SSE protocol; Go is possible, Node backend is not mandatory. It is an alternative transport adapter, not the domain message schema. Do not install both chat runtimes. [Official stream protocol](https://ai-sdk.dev/docs/ai-sdk-ui/stream-protocol).
- AI Elements targets React 19/Tailwind 4, but documented setup assumes Next.js/AI SDK/shadcn. Our Vite app already owns equivalents: no wholesale import or Next.js migration. [Official prerequisites](https://elements.ai-sdk.dev/docs).
- CopilotKit supports interactive UI/approvals with agent backends, but the broader integration does not remove our ownership, ledger and confirm requirements. Not selected for initial bounded Go orchestration. [Official overview](https://docs.copilotkit.ai/).
- Existing BaseBarChart/BaseDonutChart are preview-level. V1 simple category/compare cards use accessible shared primitives; evaluate Recharts behind a base chart in V3 when actual time-series/interactions justify it. Do not add a chart package just for static bars.

## 10. V2 action safety contract

All writes, including notes/category, require explicit confirmation initially. No unreviewed SAFE WRITE exception. LLM creates a proposed intent; backend normalizes/validates and persists `advisor_actions`, server ID/hash/version and 30-minute expiry. Cards show affected objects, before/after, financial impact and warnings. Editing payload creates a new version and invalidates previous confirmation.

Confirm takes only action ID/version and idempotency key, never fresh arbitrary payload. In one DB transaction lock action + sorted target IDs, revalidate owner/scopes/current entity versions/business rules, execute shared domain service, store result + committed audit outbox. Unique action execution prevents duplicate ledger effects even with new request keys. Response loss is recovered by GET run/message state. No provider call inside DB transaction.

States: proposed → applied/rejected/expired/stale/failed; applying is a transaction-internal lock, not an ambiguous terminal result. Unknown commit outcome must be read back, not labeled failed and re-executed. Different confirmed payload/version returns 409. Batch stores independent actions and displays each outcome; within one transfer/bulk operation the promised atomicity is enforced by its domain service. Old messages rehydrate action state before allowing confirm.

## 11. Verification / release gates

Run on an isolated migrated PostgreSQL database with two disposable owners; never mutate the real seeded owner's month note or timezone. Fake model supports deterministic tool-call/error/injection sequences. Real provider evaluation is a separate, recorded compatibility/quality gate, not replaced by mock UI screenshots.

- Golden accounting: income 1,000; Food parent expense 100, Food child 50, excluded expense 20, transfer 200 between two wallets. Expected report income 1,000/expense 150/net 850; Food subtree 150; wallets conserve transfer. Add >50 rows to prove totals cover all pages.
- Time: Asia/Ho_Chi_Minh UTC boundary, DST IANA account, leap day, 31-day→short-month comparison, account timezone change, baseline zero, uncategorized, overflow, duplicate category names and unsupported credit.
- Consistency: transaction edit during tools; final bundle internally matches all cards/drilldowns; old history remains timestamped; no mixed snapshots as current totals.
- Security: JWT and keys; owner mismatch for every ID; revoked keys mid-run; unknown tools/SQL requests/injected notes; Feedback token denied; no prompts/financial rows in audit.
- Reliability: duplicate POST, changed-payload replay, multi-tab busy, refresh/reconnect ordered events, disconnect after completion, cancellation, restart/expired lease, clear-history during generation, Redis outage, provider malformed JSON/timeout.
- UI: actual browser + API + DB; read-only chat causes zero ledger deltas; data shown equals SQL golden fixture; reload restores history; card drilldown filter matches; empty/error/null/unknown part never crashes; keyboard/mobile/masking covered; manual/AI-entry regressions pass.
- Provider evaluation: versioned ≥30 synthetic advisor cases including 10 multi-turn, 10 accounting/scope, 10 refusal/ambiguity/injection/error cases, each run 3 times. All displayed figures must equal facts, zero cross-owner/write executions; ≥90% of answerable cases complete required read tasks, refusal is not counted as success. Record failures, latency and provider tool support; no bank documents/secrets committed.

Release requires backend race/DB tests, frontend design/type/build, integrated/E2E state proof, public docs and actual configured-provider tool-calling evaluation. Public API release additionally requires user-key prerequisite. No deployment in this planning task.

## 12. Delivery roadmap / non-goals

| Stage | Deliverable / exit condition | Prerequisite |
| --- | --- | --- |
| V1 | Read-only persistent advisor, 8 tools, facts/cards/drilldown; golden and real-browser proof | Query layer, conversation, provider tool-call support; keys before third-party launch |
| V2 | Multi-action drafts, deterministic confirmation, user-controlled preferences | Shared write use cases + concurrency/versioning + durable audit outbox |
| V3 | Merchant/rules groundwork, recurring/subscription candidates, anomalies, scenarios | Merchant normalization, explicit coverage, repeatable detection tests |
| V4 | Opt-in budget/price-change insights and notifications | WORKER-01, notification channel, dedupe/cooldown and user controls |
| V5 | Ongoing user-defined financial goals with monitoring | V2–V4 proof, explicit user policy; never bank/payments by implication |

V3 affordability is a scenario, not a promise: user supplies purchase date/price, reserve definition and known obligations; service computes assumptions/ranges. Separate recorded recurring schedules from inferred subscriptions. Insufficient history/unknown debts must produce incomplete assessment, not “you can afford it”. Do not double-count goal allocations or use portfolio/credit shells as complete assets/liabilities. No investment/tax advice scope is implied.

Detailed executable tasks now cover **V1 only**. Later stages have boundaries and gates above; each needs its own implementation plan before execution. Feedback loop continues independently and does not wait for an autonomous financial agent.

## 13. Docs reconciliation and review

- This task changes planning docs only; API/ERD/architecture master docs and Swagger are **not** updated to claim proposed endpoints exist.
- Human-facing guide and machine-readable tool/parts JSON schemas, endpoint catalog and plain Markdown are required with implementation. Add entry links from `docs/README.md`; integrate actual publication path only after confirming current hosting configuration. The old Docusaurus under `refereces/disappointed_app` is not assumed to publish the current app.
- Internal docs at implementation: API, ERD, ARCHITECTURE, report rules, assistant screen/base contracts, runbook (SSE proxy/lease/cancel/audit), ADR, validation/changelog/context. Preserve quick-add screen spec independently.
- Source capability observations are dated; provider tool-calling support and assistant-ui compatibility remain measured gates, not assertions of readiness.
- Documentation verification: check local Markdown links, whitespace and roadmap/tool/card coverage. Runtime tests are not required for this docs-only change; no runtime success claimed.

### Implementation/documentation verification — 2026-09-22

- Follow-up technical plan: 8 tasks / 42 ordered subtask gates plus technical guide. Checked 12 local Markdown links, 4 JSON examples and 27 fenced blocks across the three docs with a Node filesystem/JSON checker; no failures. Technical examples are instructional and not compiled/applied. The jar read/write trap, recovery routes, message templates, deterministic context summary and six-slot navigation proposal are reconciled above.
- PASS: `rtk git diff --check` (no whitespace errors).
- PASS: `rtk proxy node --input-type=module -e ...` local-link check over this design and the V1 plan: 8 relative Markdown targets, 0 missing files. External references were inspected through official documentation separately.
- PASS: `rtk proxy sh -c 'rg -n "TBD|TODO|implement later|fill in details" docs/work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md docs/superpowers/plans/2026-09-22-finance-assistant-v1.md'` found no placeholder in the authored plan/design before this command was recorded.
- Reviewed scope coverage: V1 has query/history/provider/orchestration/API/UI/eval/docs tasks; V2–V5 are explicitly roadmap, not included implementation. API-key prerequisite is not hidden by browser JWT support.
- PASS: local backend race/DB suite and frontend type/design/build listed in the implementation checkpoint above. `get_transaction` is owner-scoped and reads report-excluded detail without changing ledger state.
- NOT RUN: configured-provider browser chat, SSE/reconnect, durable audit delivery/alerting, public deployment and owner visual UAT. The synchronous JSON endpoint is a local pilot surface, not a claim of release readiness.

### Bug-fix checkpoint — 2026-09-22

- Reproduced and fixed three provider-contract defects: generic read-tool schemas now expose typed properties/required fields; assistant tool calls serialize as OpenAI-compatible `type=function` envelopes; the orchestration cap now allows a final model answer after four tool calls while still rejecting additional calls.
- Reproduced and fixed UI contract defects: `finance_summary`, `spending_comparison`, `transaction_search` and `budget_progress` are normalized to rendered cards; wallet, budget, goal and jar views now use shared `SurfaceCard`, `Progress`, `Text` and `BaseButton` primitives; transaction counts are no longer formatted as money; the history control reloads instead of clearing the conversation.
- Added credential revalidation at provider/tool boundaries and before assistant persistence. API-key principals are re-read from storage, so expiry/revocation stops an in-flight advisor run. Session principals are rechecked through the existing session verifier. Cancelled runs cannot accept a late assistant message because repository persistence requires queued/running status.
- Cancellation now propagates to the in-process provider context through a service-owned run registry; the database status remains authoritative and the repository lease guard prevents late commits. Cross-process immediate interruption remains a release follow-up requiring durable pub/sub or worker ownership.
- Conversation history now pages backward through the existing `before_seq` route, merges by message ID and preserves chronological order in the UI; the newest page remains bounded at 50 messages and reload does not clear visible state.
- Regression evidence: the new schema, wire, four-call, principal-revalidation and cancelled-run tests were red before the fixes and green after them. Remaining release gates are configured-provider browser proof, SSE/reconnect, durable audit delivery and owner UAT.
