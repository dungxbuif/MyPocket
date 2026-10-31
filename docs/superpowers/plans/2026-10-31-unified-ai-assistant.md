# Unified AI Assistant Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Đưa hỏi đáp, đính kèm chứng từ, preview/sửa/duyệt giao dịch và theo dõi trạng thái vào một hội thoại AI native có thể khôi phục an toàn.

**Architecture:** Giữ nguyên các domain handler và pipeline `AIEntryService`/`AdvisorService`; thêm một lớp binding bền giữa conversation, generation và entry task. API mới chỉ nhận request, lưu task/attachment private rồi worker xử lý; lịch sử chat tham chiếu batch typed và hydrate proposal từ trạng thái server. iOS dùng một `MPChatSurface` làm chủ transcript + composer, với renderer typed và shared transaction editor.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, object storage private, SwiftUI, `FinanceAPI`, XCTest và XCUITest.

**Spec:** `docs/superpowers/specs/2026-10-31-unified-ai-assistant-design.md`

> **Reconciliation 2026-11-01:** Durable AI entry lifecycle, typed history,
> authoritative action decisions, attachment diagnostics, full-screen native
> Assistant, adaptive composer and shared preview cards are implemented and
> deployed with migration `000052`. Backend and native suites pass; production
> API/worker are on the same immutable digest and route parity is `96/96`.
> Provider formatting, attachment-provider behavior and owner visual review
> remain acceptance gates. The optional advisor lease/grounding flags stay off
> by design because the release contract documents them as mixed-rollout flags.

## Global Constraints

- Giữ light mode và bảng màu xanh hiện tại; không mở rộng dark mode trong lát này.
- Giữ feedback bubble trên mọi màn khác; ẩn riêng trên Assistant để chat dùng toàn bộ bề ngang.
- Mọi spacing, typography, surface, button, icon và row phải đi qua DesignSystem/base; không thêm override theo từng màn.
- Giữ API text/entry cũ hoạt động; endpoint mới là additive và không hồi sinh web archived.
- Không ghi sổ trước khi người dùng xác nhận proposal; confirm chỉ nhận identity/version, không nhận tùy ý các trường tài chính.
- Dùng giới hạn file hiện có: JPEG/PNG/PDF/XLSX, tối đa 20 file và 5 MiB/file; attachment private, không log raw OCR/prompt.
- Không làm voice, liên kết ngân hàng, cashback, health score, dự báo chắc chắn, đổi provider/model hoặc multi-conversation sidebar.
- UI copy tiếng Việt; tên trạng thái phải phân biệt đã giữ cục bộ, đang xử lý, cần xem lại, đã lưu và lỗi.
- Phản hồi local của thao tác gửi là mục tiêu kiểm thử 100 ms trên thiết bị kiểm thử; không hứa SLA cho provider/model khi chưa có số đo.

## Review Focus

- Cùng `client_request_id` nhưng payload khác phải trả conflict và không tạo task/proposal thứ hai — test ở Task 2.
- Completion đến muộn sau reset hoặc generation cũ không được chèn part mới vào hội thoại — test ở Task 3 và Task 4.
- Batch có dòng hợp lệ/lỗi/đã lưu phải giữ kết quả từng dòng, không tuyên bố atomic và không thay đổi ledger trước confirm — test ở Task 3 và Task 5.
- Part/action/version mà client chưa hiểu phải hiển thị read-only, không lộ nút ghi — test ở Task 1 và Task 4.
- Draft, attachment, keyboard, scroll anchor, masking và VoiceOver phải tồn tại qua lỗi mạng/relaunch — test ở Task 4 và Task 5.

---

### Task 1: Khóa contract, capability matrix và lifecycle schema

**Files:**
- Create: `docs/decisions/ADR-026-ai-conversation-entry-lifecycle.md`
- Create: `backend/internal/entity/advisor_entry.go`
- Create: `backend/internal/entity/advisor_entry_test.go`
- Create: `backend/internal/repository/advisor_entry.go`
- Create: `backend/migrations/000042_advisor_entry_tasks.up.sql`
- Create: `backend/migrations/000042_advisor_entry_tasks.down.sql`
- Modify: `backend/migrations/README.md`
- Modify: `backend/migrations/migrations_test.go`
- Modify: `mobile-ios/MyPocketIOS/Core/Models/AIModels.swift`
- Create: `mobile-ios/MyPocketIOSTests/AIModelsTests.swift`

**Interfaces:**
- Produces `entity.AdvisorEntryTask` với `ID`, `OwnerID`, `ConversationID`, `Generation`, `ClientRequestID`, `PayloadHash`, `SessionID`, `Status`, `Stage`, `BatchID`, `BatchVersion`, `ErrorCode`, `CreatedAt`, `UpdatedAt`, `CompletedAt`.
- Produces status constants `accepted`, `processing`, `review_ready`, `failed`, `cancelled` và proposal constants `pending`, `applying`, `applied`, `rejected`, `expired`, `conflict`, `superseded`.
- Produces `repository.AdvisorEntryTaskRepository` với các method owner-scoped: `CreateTask`, `GetTask`, `GetTaskByRequest`, `ClaimTask`, `HeartbeatTask`, `CompleteTask`, `FailTask`, `CancelTask`, `AppendEvent`, `ListEvents`, `ReapExpiredTasks`.
- Produces JSON contract cho `entry_batch`: chỉ chứa `task_id`, `batch_id`, `version`, item IDs/count và summary; draft authoritative vẫn đọc từ entry service.
- Produces capability fields `entry_integration`, `supported_part_kinds`, `action_kinds`, `attachment_formats`, `attachment_max_count`, `attachment_max_bytes` và `parts_version`.
- Produces Swift models `AdvisorEntryTaskModel`, `AdvisorEntryBatchModel` và các capability fields optional để client cũ vẫn decode được.

- [ ] **Step 1: Viết test contract thất bại** cho status transitions, owner/generation identity, version dương, hash bắt buộc và decode fixture có `entry_batch` cùng unknown part.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy go test ./internal/entity ./migrations`

  Expected: FAIL vì entity/schema/fixtures mới chưa tồn tại.

- [ ] **Step 3: Implement entity, repository interface và migration additive**

  Tạo bảng task, event và attachment binding với foreign key owner/conversation/session, unique key `(owner_id, conversation_id, generation, client_request_id)`, payload hash, lease/deadline, event sequence và index cho worker. Không sửa migration cũ; down migration chỉ phục vụ disposable fixture. Ghi rõ retention, private object ownership, reset fencing và worker lease trong ADR-026.

- [ ] **Step 4: Implement capability/fixture models**

  Mở rộng model Swift theo kiểu additive, giữ unknown part decode được; không cho capability mặc định tự bật action/format nếu server chưa công bố.

- [ ] **Step 5: Chạy test để xác nhận pass**

  Run: `rtk proxy go test ./internal/entity ./migrations && rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/AIModelsTests test`

  Expected: PASS; migration test nhận version mới và fixture round-trip không làm mất unknown part.

- [ ] **Step 6: Commit riêng contract/schema**

  Chỉ stage các file của Task 1; không stage các thay đổi styling đang có trong worktree.

### Task 2: Durable accept, private attachment storage và worker recovery

**Files:**
- Create: `backend/internal/usecase/ai_advisor_entry.go`
- Create: `backend/internal/usecase/ai_advisor_entry_test.go`
- Create: `backend/internal/infrastructure/repository/ai_advisor_entry_postgres.go`
- Create: `backend/internal/infrastructure/repository/ai_advisor_entry_postgres_test.go`
- Create: `backend/cmd/worker/advisor_entry.go`
- Create: `backend/cmd/worker/advisor_entry_test.go`
- Modify: `backend/internal/controller/http/ai_advisor_handler.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/internal/repository/ai_entry.go`
- Modify: `backend/cmd/worker/main.go`

**Interfaces:**
- Produces `usecase.AdvisorEntryInput{ClientRequestID, ConversationID, Generation, Text, Timezone, Files}`.
- Produces `AdvisorEntryService.Accept(ctx, principal, input) (*entity.AdvisorEntryTask, error)`, `Get(ctx, ownerID, taskID) (*entity.AdvisorEntryTaskView, error)`, `Cancel(ctx, ownerID, taskID) error` và `ProcessClaimed(ctx, lease) error`.
- Adds `POST /api/v1/ai/advisor/entries` returning `202` and `GET /api/v1/ai/advisor/entries/:id` as the authoritative recovery read; adds `POST /api/v1/ai/advisor/entries/:id/cancel` for unprocessed work.
- `POST` accepts multipart fields `client_request_id`, `conversation_id`, `generation`, `text`, `timezone`, `files`; same request/hash replays the existing task, different hash returns `409`.
- Worker claims one task with an expiring lease, calls the existing `AIEntryService` extraction/normalization/validation path, persists the existing entry session/proposals, and then emits a typed batch binding. It must never create a second OCR/model pipeline.

- [ ] **Step 1: Viết failing unit tests cho acceptance**

  Cover `202` accepted-before-provider, same request replay, payload mismatch conflict, owner isolation, file count/size/type validation, and cancel before processing. Assert provider/extractor is not called twice for a replay.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy go test ./internal/usecase ./internal/controller/http -run 'AdvisorEntry|AIEntry' -count=1`

  Expected: FAIL vì service/route/worker chưa có.

- [ ] **Step 3: Implement `AdvisorEntryService.Accept` và PostgreSQL repository**

  Validate conversation owner and generation, hash canonical fields plus file metadata, store uploads under an owner/task-scoped private key, create the existing AI entry session, and commit task + attachment metadata before returning. Never store raw prompt/OCR in logs or public task JSON.

- [ ] **Step 4: Refactor `AIEntryService` to expose a stored-input path**

  Keep `/ai/entry/process` behavior unchanged, but share its validation, wallet/category loading, OCR/provider call, normalization and proposal persistence with the worker path. The worker must hydrate private objects into `AIImage` only while processing and must not upload the same attachment again.

- [ ] **Step 5: Wire routes and worker pass**

  Add owner-scoped handlers and a bounded `RunAdvisorEntryPass(ctx, limit)` invoked by the existing worker loop. Claim/heartbeat/complete/fail/reap must be lease-safe across worker restart; a cancelled task cannot be completed by a stale lease.

- [ ] **Step 6: Run backend tests and disposable PostgreSQL tests**

  Run: `rtk proxy go test ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository ./cmd/worker -count=1`

  Expected: PASS, including response-loss recovery through GET and restart/lease recovery without duplicate model calls or attachment objects.

- [ ] **Step 7: Commit backend acceptance/recovery**

  Stage only Task 2 files and record the exact test output in the implementation log.

### Task 3: Hydrate typed batches and make every action server-authoritative

**Files:**
- Modify: `backend/internal/entity/ai_advisor.go`
- Modify: `backend/internal/entity/ai_entry.go`
- Modify: `backend/internal/usecase/ai_advisor.go`
- Modify: `backend/internal/usecase/ai_advisor_parts.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/internal/controller/http/ai_advisor_handler.go`
- Modify: `backend/internal/controller/http/ai_entry_handler.go`
- Modify: `backend/internal/infrastructure/repository/ai_advisor_postgres.go`
- Modify: `backend/internal/infrastructure/repository/ai_entry_postgres.go`
- Test: `backend/internal/controller/http/ai_advisor_handler_test.go`
- Test: `backend/internal/controller/http/ai_entry_handler_test.go`
- Test: `backend/internal/usecase/ai_advisor_parts_test.go`
- Test: `backend/internal/usecase/ai_entry_test.go`
- Test: `backend/internal/infrastructure/repository/ai_entry_recovery_test.go`

**Interfaces:**
- Produces `AdvisorEntryBatchView` and `AIEntryService.Batch(ctx, ownerID, batchID, version)` that return authoritative proposal state, per-row errors, resource IDs/receipts and supported edit fields.
- History loading resolves `entry_batch` references through the batch service; it never copies stale drafts into an independent advisor message.
- Adds server-validated `PATCH /api/v1/ai/advisor/actions/:action_id/replace` with conversation/generation/revision and an action-kind-specific allowed patch; old action becomes `superseded`, replacement gets a new action ID.
- Capabilities and typed data cover all currently advertised actions: `create_wallet`, `create_jar`, `update_jar`, `create_budget`, `update_budget`, `create_transaction`, `update_transaction`, `delete_transaction`, `transfer_wallet`, `create_category`, `update_category`, `recategorize_transaction`.
- Confirm response contains applied/rejected/conflict state and resource receipt; repeated confirm is idempotent and never renders a fresh Save button for an already applied action.

- [ ] **Step 1: Viết failing table-driven tests**

  Test every action kind for owner/conversation/generation binding, expected-version enforcement, replacement validation and applied receipt. Test entry batch partial success, duplicate approve, stale edit, unknown batch version, reset race and no ledger mutation before confirm.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy go test ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository -run 'Advisor|AIEntry|Batch|Proposal' -count=1`

  Expected: FAIL for missing batch hydration/state transitions and missing action replacement contract.

- [ ] **Step 3: Implement authoritative batch hydration**

  Store only the batch reference in advisor history, derive rows from the entry session/proposal repository, and expose each row's state/error/resource ID. Keep `approved` compatibility reads if old rows exist, but write the new applied/rejected vocabulary for new tasks.

- [ ] **Step 4: Implement generation fencing and action replacement**

  Reject expired/foreign generation, validate allowed fields by action kind, create a replacement proposal through domain validation, and make late completion unable to append to a reset generation.

- [ ] **Step 5: Run focused and full backend AI tests**

  Run: `rtk proxy go test ./internal/entity ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository -count=1`

  Expected: PASS with no regressions in legacy text advisor or standalone `/ai/entry/process` flows.

- [ ] **Step 6: Commit lifecycle/action parity**

  Stage only Task 3 files and retain the old API compatibility tests in the commit.

### Task 4: Native transport, recoverable chat state and base-first full-screen surface

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantChatState.swift`
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/AdvisorEntryObserver.swift`
- Create: `mobile-ios/MyPocketIOSTests/AssistantChatStateTests.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/Models/AIModels.swift`
- Modify: `mobile-ios/MyPocketIOS/Core/API/FinanceAPI.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AdvisorRunObserver.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPAttachmentViews.swift`
- Modify: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

**Interfaces:**
- Produces `FinanceAPI.beginAdvisorEntry(_:files:)`, `advisorEntry(id:)` and `cancelAdvisorEntry(id:)`, all preserving the stable request/task identity.
- Produces `AssistantChatState` transitions for `idle`, `sending`, `accepted`, `processing`, `reviewReady`, `failed`, `cancelled`, `uncertain` and `resetting`; a transport error after acceptance must recover with GET, not POST.
- Produces `AdvisorEntryObserver.observe(ownerID:taskID:)` with polling/reconnect and cancellation fencing analogous to `AdvisorRunObserver`.
- `MPChatSurface` owns `VStack(spacing: 0)` containing transcript and composer, including keyboard-safe layout; `AssistantView` only composes base components and does not add a second bubble/card wrapper.
- Unknown parts render as readable read-only status; capability-gated buttons do not appear for unsupported action/format.

- [ ] **Step 1: Viết failing state/policy tests**

  Cover immediate local message insertion, stable request ID after transport failure, no duplicate submit, task recovery after relaunch, reset invalidating old task updates, attachment append/remove, unknown-part read-only behavior, scroll anchor and masking copy.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/AssistantChatStateTests test`

  Expected: FAIL because the new state/transport contracts are absent.

- [ ] **Step 3: Add API models and methods**

  Encode multipart fields once, reuse current `MultipartFile`, decode both task recovery and typed batch references, and preserve legacy `AdvisorMessageModel`/`EntryProcessModel` decoding.

- [ ] **Step 4: Implement state and observers**

  Persist only owner-scoped conversation/task/request pointers needed for recovery; keep draft text and pending attachments until the server result is known. Use GET after uncertain response, server task status after relaunch, and generation token checks before mutating `messages`.

- [ ] **Step 5: Fix the base surface and compose the full-screen Assistant**

  Make `MPChatSurface` the single vertical layout owner, keep composer anchored above the keyboard, remove the Assistant feedback dock, move overview into a sheet or chat part rather than a separate dashboard strip, and keep `Tin mới`/older-message anchor behavior.

- [ ] **Step 6: Run focused native tests and compile**

  Run: `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/AssistantChatStateTests -only-testing:MyPocketIOSTests/DesignSystemTests test`

  Expected: PASS; the Assistant builds with one transcript/composer surface and no feedback bubble overlay.

- [ ] **Step 7: Commit native transport/base state**

  Stage only Task 4 files; do not include unrelated Overview/Account density edits.

### Task 5: Unified cards, attachment batch review and shared transaction editor

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantEntryBatchView.swift`
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantActionCard.swift`
- Create: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionEditorFields.swift`
- Create: `mobile-ios/MyPocketIOSTests/AssistantEntryBatchTests.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFields.swift`
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- Produces `AssistantEntryBatchView` with per-file progress/error and per-row `pending`, `invalid`, `saving`, `applied`, `rejected`, `conflict` rendering.
- Produces `AssistantActionCard` that selects a renderer by capability/action kind and routes confirm/edit/reject through server IDs only.
- Produces shared `TransactionEditorFields` used by both `TransactionEditorView` and assistant proposal editing; amount is primary, wallet/category/jar/date/note use the same validation and compact row bases as “Thêm giao dịch”.
- Produces `saveOne`, `saveValid`, `skip` behavior with stable proposal identity; applied rows show the real transaction/resource receipt and no save action.

- [ ] **Step 1: Viết failing tests cho row state và action mapping**

  Assert all 12 actions map to a safe card, unsupported kinds map to read-only, a batch with 3 rows can save one and leave two pending, and `saveValid` skips invalid rows without hiding their reasons.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/AssistantEntryBatchTests test`

  Expected: FAIL until the renderer/state helpers exist.

- [ ] **Step 3: Extract shared transaction fields**

  Move field composition and binding/validation out of the current large `TransactionEditorView` without changing its public save behavior. `EntryProposalEditor` and the new assistant batch editor must call the same `TransactionEditorPolicy`, category/jar eligibility and timezone normalization.

- [ ] **Step 4: Implement batch/attachment cards using existing bases**

  Use `MPSurfaceCard`, `MPTransactionRow`, `MPAttachmentThumb`, `MPStatusMessage`, `MPButton` and Money Lover-like regular/duotone `MPIcon` catalog entries. Do not nest a second card around a card that already owns a surface. Keep padding/radius/typography token-driven and compact.

- [ ] **Step 5: Integrate chat actions and entry route**

  Attach photo/PDF/XLSX before sending, show progress and file-specific errors in the same transcript, open the shared editor as an anchored sheet/inline review without losing draft or scroll position, and keep standalone `/ai/entry` using the same backend engine/base components.

- [ ] **Step 6: Add deterministic UI fixtures and run XCUITest**

  Add launch fixtures for empty chat, text action, attachment batch, partial result, applied receipt, error/retry and keyboard-visible states. Run: `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSUITests/LaunchTests test`

  Expected: PASS for the new fixture cases; inspect each xcresult screenshot and failure log before interpreting any runner issue.

- [ ] **Step 7: Commit unified cards/editor**

  Stage only Task 5 files and include fixture names in the commit message.

### Task 6: Latency evidence, whole-app visual review and documentation reconciliation

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Core/Offline/FinancialMutationQueue.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Modify: `mobile-ios/MyPocketIOSTests/FinancialMutationQueueTests.swift`
- Modify: `docs/design/pages/assistant/README.md`
- Modify: `docs/design/pages/finance-assistant/README.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`
- Create: `docs/work/audits/AI-ASSISTANT-UNIFIED-UAT-20261031.md`

**Interfaces:**
- Produces queue timing evidence for tap→local feedback, persist, refresh, drain, server acknowledgement and ledger receipt; this task does not alter ledger semantics or introduce optimistic balance.
- Produces visual/UAT evidence for empty/data-rich chat, long transcript, batch, keyboard, error, applied state, small/regular iPhone and large Dynamic Type in light mode.
- Reconciles API, base component, screen specs, validation matrix, backlog/context and changelog with actual implementation; unresolved items remain explicit rather than marked complete.

- [ ] **Step 1: Viết failing timing tests/diagnostic hooks**

  Assert `FinancialMutationQueue` records stage durations without changing ordering, and that a slow refresh is distinguishable from a slow server ledger write.

- [ ] **Step 2: Chạy test để xác nhận đang fail**

  Run: `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=iPhone 16' -only-testing:MyPocketIOSTests/FinancialMutationQueueTests test`

  Expected: FAIL until instrumentation/fixture clock is added.

- [ ] **Step 3: Add bounded instrumentation only**

  Record p50/p95-ready event fields with privacy-safe IDs, keep existing durable mutation ordering, and report whether the delay is local persistence, refresh, queue drain or server acknowledgement. Do not paper over latency with a fake applied state.

- [ ] **Step 4: Run backend/native verification matrix**

  Run backend focused/full AI tests, native unit tests, base/project guards, `rtk proxy git diff --check`, then install the debug build on the owner simulator with disposable seeded data. Capture screenshots and inspect xcresult; do not call a failed UI test “runner-only” without root-cause evidence.

- [ ] **Step 5: Perform whole-app visual audit**

  Review Overview, Wallets, Transactions, Add Transaction, Account, Jars/Budgets/Reports and Assistant using the same base tokens. Verify feedback bubble remains on non-Assistant screens, is absent on Assistant, rows/cards are denser, headings smaller, icons regular/duotone, and no floating control hides content.

- [ ] **Step 6: Reconcile docs and write owner UAT checklist**

  Record commands, simulator/device, fixture data, screenshots, known limitations and any failed evidence in `AI-ASSISTANT-UNIFIED-UAT-20261031.md`; update only claims proven by tests. Release/deployment remains a separate decision.

- [ ] **Step 7: Commit verification/docs**

  Stage only Task 6 files after evidence is attached; leave unrelated pre-existing worktree changes untouched.

## Self-review

- **Spec coverage:** Contract/capability is Task 1; durable accept/storage/recovery is Task 2; typed history, action parity, proposal states and generation fencing are Task 3; full-screen chat, draft/keyboard/privacy/unknown parts are Task 4; batch attachments, jars/budgets, shared editor and save-one/save-valid are Task 5; queue latency and whole-app visual/docs reconciliation are Task 6.
- **Step scan:** Every task has a failing test, a failure command, an implementation boundary, a passing command and a commit boundary. No step authorizes a new provider, direct ledger write or screen-specific styling.
- **Type consistency:** `AdvisorEntryTaskRepository` is the source of task leases; `AdvisorEntryService` owns accept/get/cancel/process; `AdvisorEntryTaskModel` is the native recovery model; `AdvisorEntryBatchView` is the authoritative proposal hydration contract; `AssistantChatState` consumes it without duplicating server state.
- **Review focus coverage:** All five failure modes are tied to tests in Tasks 1–5; Task 6 adds evidence review so passing unit tests cannot be mistaken for visual/UAT completion.
- **Proportion:** The plan names interfaces, files, state transitions and verification commands, but leaves implementation bodies to the executor and reuses existing domain pipelines/base components.
