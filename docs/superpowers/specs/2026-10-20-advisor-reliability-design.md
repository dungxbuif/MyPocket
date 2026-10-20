---
artifact_type: detail_design
id: AI-ADVISOR-RELIABILITY-20261020
status: implemented_with_feature_flags
owner: shared
implementation_approval: owner_approved
created: 2026-10-20
trace:
  backlog: ../../work/BACKLOG.md
  phase: ../../work/phases/PHASE-002-native-ios-swiftui.md
  plan: ../plans/2026-10-20-advisor-reliability.md
  validation: ../../work/VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
  catalog: ../../architecture/AI_ADVISOR_TOOLS.md
---

# Finance Assistant: reliability, tools và context

## 1. Mục tiêu và trạng thái

Tài liệu này là contract triển khai sau review code ngày 2026-10-20. Local
candidate đã triển khai các slice lõi nhưng chưa rollout production. Mục tiêu:
trợ lý hiểu đúng tài nguyên của
user, đưa ra proposal có thể xác nhận, không kẹt sau restart, trả lời có nguồn
số liệu và tiếp tục hội thoại ổn định trên iOS.

Giữ các quyết định của chủ app: một hội thoại trợ lý/user; xóa nội dung sau
30 phút không có tin nhắn mới; AI entry/OCR là luồng riêng; user xác nhận mọi
proposal trong chat; JWT và API key có cùng quyền trên cùng owner. Không thêm
scope, quota user, ngân hàng hoặc tự động thực hiện giao dịch.

Runtime leases/atomic completion, bounded context, native run observation và
typed evidence đã được triển khai trong local candidate. Hai flag
`AI_ADVISOR_RUN_LEASES_ENABLED` và `AI_ADVISOR_GROUNDING_ENABLED` mặc định
`false`; bật theo thứ tự sau khi apply migration 000041 và xác nhận provider hỗ
trợ `finish_answer`. Không deploy hoặc thay đổi production trong tài liệu này.

## 2. Hiện trạng được đối chiếu với code

| Thành phần | Candidate hiện tại | Cổng còn mở |
| --- | --- | --- |
| Model adapter | OpenAI-compatible HTTP, provider call timeout 30s | Live provider latency/error-class evidence |
| Orchestrator | Go, bounded 6 provider calls/8 tools | Disposable kill-executor acceptance |
| Tools | Owner-scoped catalog/version/snapshot validation | Full public route/catalog parity check |
| Context | 12-message/24 KiB bounded request context on both service paths + validated entity refs | Repository summary inspection |
| Run | PostgreSQL lease/heartbeat/reaper behind flag | Migration and multi-process race evidence |
| Completion | Atomic message/facts/event/usage/context | Production rollout remains off |
| iOS | Dedicated observer, metadata, reconnect/backoff | Physical device UAT |
| Số liệu | Typed facts, `finish_answer`, one repair, safe card fallback | Live repair/fallback metrics |

Nguồn: `backend/internal/usecase/ai_advisor*.go`,
`backend/internal/infrastructure/ai/advisor_client.go`,
`backend/internal/infrastructure/repository/{ai_advisor,finance_query}_postgres.go`,
`mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`.
Deterministic candidate evidence: full backend suite, vet, selected race suite,
native observer/API tests, base-component/project guards, and the 31-case
reliability dataset schema check pass locally (31 prompts have not been
executed as a live quality matrix). Migration 000041 and the
repository lease/completion race suite and expired-lease reaper simulation also
pass against the disposable local PostgreSQL instance. Đây chưa phải live
quality acceptance; provider matrix và physical-device UAT vẫn là release gates
riêng.

Review correction: the earlier implementation snapshot overstated integration
coverage. The [review-fix ticket](../../work/tickets/AI-ADVISOR-REVIEW-FIX-20261020.md)
tracks production-loop snapshots, current-generation context loading and
reference revalidation, strict repair transcripts, terminal-tool budgeting,
server-owned final labels/help and locked lease rechecks. Dataset validity,
scripted behavioral regression, real PostgreSQL integration and live-model
intent accuracy are separate evidence classes.

## 3. Quyết định kiến trúc

Giữ stack Go 1.26 + Gin + GORM/PostgreSQL + Redis auth/audit + SwiftUI.
Không thay orchestrator bằng agent framework; không thêm embeddings/vector DB.
MCP official Go SDK là adapter bên ngoài; assistant nội bộ tiếp tục gọi semantic
tools trực tiếp. Không đưa API key/provider secrets vào iOS.

```text
SwiftUI chat / typed cards / run observer
             |
             v
Auth API -> AdvisorService -> PostgreSQL run/conversation/evidence
                 |                   ^
                 v                   | heartbeat / atomic completion
        Context + Tool Orchestrator --+
           |                 |
           v                 v
   OpenAI-compatible      FinanceQueryService -> owner-scoped repositories
           |
           v
   validated answer -> persisted parts -> iOS

Independent reaper -> expired run lease -> interrupted, no provider replay
```

Lựa chọn khác: queue chạy model trong worker riêng hỗ trợ tiếp quản nhưng tăng
chi phí vận hành và rủi ro gọi provider lại. Đợt này giữ API executor, dùng
lease để kết thúc run mất executor an toàn. Tiếp quản/replay provider là future
scope. Tăng model hoặc prompt đơn thuần không sửa được thiếu catalog/version.

## 4. Contract đọc tài nguyên và proposal

### 4.1 Tools và payload

- Bổ sung `version:int64 >= 1` vào transaction detail/search. Không thay
  `expected_version` của confirmation thành optional; không mặc định version=1.
- Thêm `list_categories` qua FinanceQueryService, arguments:
  `{kind?:income|expense|debt, wallet_ids?:uuid[], include_archived?:bool,
  limit?:int, cursor?:string}`. Default limit=50, max=100, max 20 wallet IDs.
  Cursor phải khóa vào owner và canonical filter; sort ổn định `id ASC`.
- Thêm `get_category` arguments `{category_id:uuid}`; result gồm `id`, `name`,
  `kind`, `parent_id`, `wallet_ids`, `icon_key`, `is_active`, `is_system`,
  `editable`, `version`. System category có thể đọc, `editable=false`.
- List trả cùng trường trên và `next_cursor`; tên trùng trả nhiều lựa chọn,
  không tự chọn resource khi không đủ ngữ cảnh. Archive không bị chọn cho
  giao dịch mới. Search catalog tối đa 2 trang/lượt, sau đó hỏi user thu hẹp.
- Owner là Principal server-side; tools không nhận owner/user ID. Danh mục
  foreign và not-found không làm lộ tên/ID. Wallet scope được xác minh owner.

### 4.2 Snapshot trước proposal

Mọi proposal update/delete/recategorize phải đối chiếu ID và version với tool
result đã đọc trong chính run hiện tại. Model nhập version khác snapshot,
ID chưa đọc, hoặc system category không editable: trả tool error có cấu trúc
`snapshot_required`/`snapshot_mismatch`, cho model đọc lại; không tạo card lỗi.
Proposal create có wallet/category reference cũng phải đối chiếu catalog đã
đọc. Không yêu cầu version cho create nếu domain hiện tại không có contract đó.

Confirmation vẫn tra card đã lưu theo owner/conversation/generation/action,
chạy canonical handler với idempotency key=action ID. Khi dữ liệu đổi sau
proposal: trả CAS 409, không tự thay version hoặc retry write. User có thể tạo
proposal mới. API key/MCP dùng cùng domain contract; không thêm quyền phụ.

## 5. Lifecycle và phục hồi run

### 5.1 Bounds đề xuất

| Giá trị | Contract |
| --- | --- |
| Total run deadline | 240s từ `created_at`, tính bằng server/DB clock |
| Provider call timeout | 30s; bị giới hạn thêm bởi deadline còn lại |
| Provider/tool calls | Giữ 6/8; repair calls cũng tính vào 6 |
| Lease | 45s; heartbeat mỗi 10s, DB operation timeout 2s |
| Reaper | mỗi 15s, batch tối đa 100, operation timeout 5s |
| Terminal DB operation | timeout 5s; không dùng unbounded WithoutCancel |
| Idle reset | Giữ 30 phút, chỉ accepted user message gia hạn |

Đây là giới hạn kỹ thuật mỗi request, không phải quota sử dụng của user.

### 5.2 Lease semantics

`StartRun` giữ dedupe owner/request ID như hiện tại và ghi `deadline_at`.
`ClaimRun(owner, runID, token)` đổi queued -> running, chỉ một executor claim;
lease until = min(DB now+45s, deadline). Claim và heartbeat phải match
owner/generation/token/status. Token không trả client, không ghi log.

Mỗi executor hủy provider/tool context ngay khi heartbeat thất bại hoặc mất
lease; fail closed khi DB không xác nhận còn quyền ghi. Watch cancel hiện tại
giữ làm fast path và polling liên process, thêm `failed`/`completed` terminal.

Reaper xử lý queued chưa claim quá 45s hoặc running hết lease/deadline:
`interrupted` với `executor_lost`, hoặc `failed` với `deadline_exceeded`.
Queued -> interrupted/failed phải được thêm vào state transition contract.
Run cũ được giữ làm receipt; gửi lại cùng request ID trả trạng thái cũ, không
gọi model lại. User chủ động gửi lại bằng request ID mới sau khi terminal.

Reaper chạy bằng timer riêng trong worker, không bị các recurring/budget pass
chặn; StartRun cũng thực hiện owner-local stale-run check trước busy check.
Read run chỉ phản ánh trạng thái; không kích hoạt provider. Không có replay tự
động sau restart, vì một provider request đã nhận có thể tính phí.

### 5.3 Lock order và completion

Mọi thao tác nhiều row theo `conversation -> run`; heartbeat chỉ lock run và
không lock conversation sau đó. Reaper chọn candidate không giữ run lock qua
bước lock conversation, rồi recheck điều kiện dưới lock đúng thứ tự.

`CompleteRun` trong một DB transaction: xác minh owner/generation/token/status/
lease/deadline; insert assistant message, fact bundle và run.completed event;
lưu sanitized usage; tăng sequence; chuyển run completed; cập nhật compact
context. Một lỗi -> rollback toàn bộ. Retry cùng token/run trả message đã có
nếu completed cùng generation; không insert duplicate. Receipt lookup sau
commit-ack timeout xác định đã hoàn thành hay chưa, không gọi provider lại.
Cancel/reset/reaper thắng lock trước -> completion bị từ chối, không late card.

## 6. iOS quan sát run

Tách `AdvisorRunObserver` khỏi screen để kiểm thử fake clock/transport.
Foreground poll: sau 1s, 2s, rồi 3s; reset backoff khi status đổi. Network/5xx
dùng backoff 3/5/10s và giữ run ID; user có nút Tải lại/Hủy. Không tự gửi POST.
Background ngừng polling, foreground fetch run ngay; owner đổi/logout hủy task
và xóa handle owner cũ. Sau app restart đọc history metadata có `active_run`.

Run response thêm `server_time`, `deadline_at`, `retry_after_ms`; history meta
thêm `active_run:{id,status,created_at,deadline_at}` hoặc null. Client cũ bỏ qua
field mới. Không trả credential/lease token. Capabilities thêm
`run_timeout_seconds:240`, `grounding_version:1` khi feature tương ứng bật.

Không cắt quan sát theo 30s cố định. Khi vượt deadline server, tiếp tục GET đến
terminal, không tự unlock gửi lượt mới khi server còn running. Hiện trạng thái
“Đang đối chiếu dữ liệu”, “Kết nối gián đoạn. Yêu cầu vẫn được theo dõi”, hoặc
“Phiên xử lý bị gián đoạn. Bạn có thể gửi lại”. Dùng base MPStatusMessage,
MPButton, MPComposer và card renderer hiện có; giữ palette và typography.
SSE event replay đang có được giữ tương thích; token streaming không thuộc scope.

## 7. Context hữu ích nhưng nhỏ

Giữ tối đa 12 message text, tổng text UTF-8 24 KiB (giới hạn kỹ thuật context).
Luôn giữ trọn current user message; cắt oldest messages trước, không cắt giữa
UTF-8 rune. Run có tool transcript tối đa 96 KiB serialized; quá budget kết thúc
bằng kết quả có nguồn đã lấy, không âm thầm cắt JSON tool response.

Dùng `advisor_conversations.summary` có sẵn cho `AdvisorContextState v1`:
generation, through_seq, last_query_kind, explicit date range, timezone,
wallet/category/transaction references (mỗi loại tối đa 8), selected topic.
Tổng state <=4 KiB. Chỉ server rút references từ validated completed results;
không lưu số dư, số tiền, version dùng lại, note nguyên văn hoặc model summary.
Mỗi message mới revalidate quyền/tình trạng entity trước khi đưa vào context.
Version và số tiền luôn đọc mới bằng tool. Chỉ completion cùng generation mới
cập nhật state; cancel/failed/reset/idle expiry không để state cũ ghi đè.

“So với tháng trước?” kế thừa filter rõ ràng gần nhất, rồi query lại số liệu.
Nếu nhiều ví/danh mục cùng hợp lý, hỏi một câu để chọn. Đổi tháng/năm phải tính
calendar theo timezone; không suy diễn ngày bằng cộng trừ 30 ngày. Không thêm
keyword router cưỡng bức chọn tool.

## 8. Grounding và lỗi provider

### 8.1 Evidence và câu trả lời

Mỗi tool call có evidence key duy nhất trong run (`call_id/query_key`); không
dùng tất cả `q1` khiến fact IDs đụng nhau. Chỉ whitelisted typed financial facts
từ query result được đưa vào EvidenceIndex. Budget/goal/jar cần export typed
facts tương ứng; không đọc số tùy ý từ JSON path do model chọn.

Thêm provider-facing terminal tool `finish_answer` (không tính finance tool
execution; request chứa nó vẫn tính provider call). Input schema:
`{mode:finance|product_help|clarification, template_id?:string,
claims?:[{slot:string,fact_ref:string}], explanation?:string}`.
Mode finance dùng template allowlist của server: summary, period comparison,
transaction results, wallet/budget/goal/jar status và proposal ready. Claims
trỏ đúng evidence trong run hiện tại; currency/scope/period phải tương thích.
Server tự render mọi giá trị định lượng trong câu trả lời và các cards.
Tỷ lệ/delta/rounding do helper Go tính, model không tự cấp arithmetic expression.

Trong mode finance, `explanation` từ model không được publish như financial
narrative: chỉ dùng để debug redacted shape, không lưu nội dung raw. Server
templates cung cấp giải thích có thể chứng minh; phần suy luận tài chính tự do
là scope tương lai. Như vậy không giả vờ regex/LLM judge có thể chứng minh mọi
số tiền viết bằng chữ là đúng. Tradeoff: câu trả lời định lượng ít tự do hơn.
Mode product_help dùng source IDs từ catalog có sẵn; clarification không đưa
ra claim về tài chính user. Mode labels từ model không được coi là bằng chứng;
các câu trả lời ngoài template finance phải hiện là hướng dẫn/câu hỏi, không
được dùng để cập nhật metric, suggestion amount hoặc action payload.

Nếu provider trả final text tài chính legacy, missing/foreign fact hoặc sai
template: cho một repair request nếu còn budget; sau đó trả text template an
toàn + validated cards hiện có. Không publish raw text đó. Nếu chưa có evidence:
hỏi clarification, không invent zero hoặc “không có dữ liệu” chỉ vì chưa query.
Guarantee 100% áp dụng số liệu server render/claim references; chất lượng ngữ
nghĩa và nhận diện intent phải đo live eval, không tuyên bố chứng minh tuyệt đối.

### 8.2 Lỗi và quan sát

Tool input/snapshot/not-found lỗi có thể sửa trả tool result typed error và
cho model tiếp tục trong budget. Auth revoked/foreign access/lease loss/cancel
là fatal; không gửi tên entity hoặc chi tiết DB lỗi sang model.
Provider 401/403/429/5xx, timeout, malformed envelope, oversized response được
phân loại nội bộ và map stable public error. Không tự retry inference transport
ambiguous; repair không phải retry cùng inference đã mất kết quả.

Log có run ID, correlation ID, tool name, status/error class, latency, byte
count, token usage khi có. Không raw prompt, argument, financial payload,
credential, attachment. Usage đã biết của failed run vẫn lưu sanitized trong
terminal transaction; thiếu usage là unknown, không ghi zero như đã đo.
Usage/log exporter lỗi không làm câu trả lời đã hoàn tất thành failed.

## 9. Schema, API và tương thích

- Migration forward additive: thêm `advisor_runs.deadline_at timestamptz`, index
  active status/lease/deadline và partial unique assistant message per run
  (`WHERE role='assistant'`). Dùng số migration kế tiếp sau max hiện tại khi
  implementation bắt đầu; không sửa migrations đã chạy.
- Preflight kiểm tra duplicate assistant messages; nếu có, dừng migration và
  lập remediation riêng, không tự xóa. Backfill deadline=created_at+240s; run
  đang treo cũ chỉ được reaper kết thúc, không replay provider.
- Reuse lease fields, summary JSON và fact bundle table hiện có; không thêm
  session/token/key tables. Prompt protocol mới nằm server-side; iOS giữ parts
  version 1 và các card types đang hỗ trợ.
- API additions: metadata và category semantic tools; update OpenAPI/MCP
  inventory nếu thêm canonical routes. Đợt này dùng routes hiện có cho categories;
  không cần thêm REST route chỉ để phục vụ provider-facing tools.
- Feature flags server-only: `AI_ADVISOR_RUN_LEASES_ENABLED`,
  `AI_ADVISOR_GROUNDING_ENABLED`, default false cho mixed rollout. Backend phải
  báo đúng capability. Context state có version; old reader bỏ qua version lạ.

## 10. Acceptance và release gates

| ID | Điều kiện pass |
| --- | --- |
| AC1 | Đọc transaction/category rồi đề xuất update với đúng persisted version; stale 409, không write |
| AC2 | Tên danh mục trùng/archived/system/foreign đều xử lý đúng; model không đoán ID/version |
| AC3 | Kill executor: run terminal sau lease expiry + một reaper tick; không gọi provider lại |
| AC4 | Cancel/reset/reaper vs complete chỉ một terminal outcome, không late message/event |
| AC5 | Inject lỗi từng bước completion: all-or-nothing, replay không duplicate |
| AC6 | Run 90s: iOS vẫn quan sát; offline/foreground/relaunch hồi phục bằng GET |
| AC7 | Multi-turn filter/date refs đúng timezone; 30m reset xóa summary/facts/messages |
| AC8 | Mọi rendered financial claim match evidence; fabricated fact/template bị chặn |
| AC9 | JWT/API key owner-isolation/revocation, confirmation và idempotency không suy giảm |
| AC10 | Không raw financial/secret logs; failed usage không báo số 0 giả |

Deterministic gates: security/schema/CAS/grounding rendered values đạt 100% trên
case pack. Live model: tối thiểu 30 synthetic scenarios x 3 lần, đúng tool/intent
>=95%, không cross-owner hoặc write ngoài confirmation. Provider-unavailable
hoặc skipped không tính pass. So với baseline cùng model/data trước khi rollout.
Latency report p50/p95 theo nhóm, ghi token/repair/fallback rate và không tự hạ
assertion để qua release. Chạy staging/disposable DB; real iPhone dùng UAT owner
riêng đã có trên prod sau deploy, không ghi vào account chính.

## 11. Rollout, alternatives và docs

Schema trước, API/worker cùng build sau. Flags tắt khi deploy code; bật lease
chỉ sau khi mọi executor có fencing (không mixed old writer). Bật grounding sau
deterministic/live gates; test capability + old/new iOS. Rollback code dùng bản
hiểu fencing khi leases đã bật; không rollback về old writer còn active run.
Không migration down production, không reset tài chính hoặc feedback statuses.
Backup theo chính sách/ủy quyền hiện hành tại thời điểm rollout; exception cũ
không phải mặc định vĩnh viễn. Không deploy trong task viết spec này.

Docs cần reconcile lúc implementation: AI_ADVISOR_TOOLS, API, ARCHITECTURE,
ERD, INTEGRATIONS, guide finance-assistant-ios, screen finance-assistant,
PROVIDER_UAT, DEPLOYMENT/ROLLOUT_READINESS, OpenAPI, MCP contract, eval README,
CONTEXT/BACKLOG/VALIDATION_MATRIX/CHANGELOG. Markdown tool catalog và guide là
surface đọc được cho AI agent; cập nhật cùng human docs. ADR mới cho lifecycle
lease/atomic completion và grounding protocol khi triển khai, không gán số ADR
trùng với công việc khác.

Known unknowns phải kiểm chứng bằng implementation tests: provider target có
tuân thủ finish tool, duplicate data trước index, long-run iOS foreground trên
máy thật. Không coi các điều này là lỗi production đã tái hiện.
