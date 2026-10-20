# Portfolio, AI actions, push và physical UAT — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Hoàn thiện năm hạng mục user yêu cầu: correction portfolio, AI actions mở rộng, push/signing, offline/device regression và conversation UAT.

**Architecture:** Go giữ toàn bộ tính toán và quyền owner; iOS dùng base components để preview/confirm. Portfolio lưu revisions bất biến, AI gọi cùng domain service sau confirmation; UAT kiểm cả UI và state backend.

**Tech Stack:** Go, PostgreSQL, Redis, SwiftUI/XCTest, OpenAPI, APNs HTTP/2.

**Spec:** Kế hoạch này đề xuất cụ thể phần còn thiếu trong [portfolio contract](../../work/tickets/FEEDBACK-20261019-PORTFOLIO-CORRECTION-CONTRACT.md), [AI contract](../../work/tickets/FEEDBACK-20261019-AI-ACTIONS-CONTRACT.md), [remaining work](../../work/tickets/FEEDBACK-20261019-REMAINING.md). Chưa triển khai các contract mới trong lượt lập plan.

## Global constraints

- Giữ dữ liệu production; migration additive, backup và kiểm restore trước rollout.
- JWT và API key dùng cùng owner scope; owner lấy từ principal, không từ model/request.
- Không thêm ngân hàng, thanh toán thật hoặc chuyển tiền ra ngoài. Transfer ở đây là bút toán giữa ví trong app.
- Money/quantity giữ quy tắc integer/scaled integer hiện có; không tính tiền bằng floating point.
- Tất cả AI mutation phải persisted proposal → preview → explicit confirmation → domain service → receipt.
- Giữ bảng màu và dùng base components hiện có. Một đợt test chỉ có một agent điều khiển iPhone.
- Kết quả test tách local/mock, provider thật, simulator, physical và production; HTTP 200 không đủ chứng minh nghiệp vụ.
- Không tự mua Apple membership. Phần code/mock APNs có thể làm trước external credentials.

## Review focus

1. Sửa giao dịch cũ làm một lần bán về sau âm số lượng: reject toàn bộ, không ghi một phần (P1).
2. Correction cùng timestamp đổi thứ tự tính giá vốn: giữ thứ tự logic của trade gốc (P1).
3. Confirm sau khi dữ liệu đổi hoặc sau response loss: conflict hoặc replay receipt, không double write (P2).
4. Đổi tài khoản trong lúc offline/sending: cache/outbox/token không chạy dưới owner mới (P3/P4).
5. Chat hết hạn trong UAT: ghi nhận generation/410, không tự gửi lại yêu cầu tốn phí (P5).

## P1 — Portfolio sửa/hủy bằng revision và replay

**Quyết định đề xuất:** sửa = thêm revision thay thế; xóa = thêm revision void. Giữ trade gốc và toàn bộ lịch sử. Không tạo lệnh mua/bán ngược chiều để giả lập undo: với weighted-average cost, một lệnh ngược chiều có thể tự sinh realized P&L và không khôi phục đúng giá vốn.

**Files:** modify `backend/internal/entity/portfolio.go`, `backend/internal/repository/portfolio.go`, `backend/internal/infrastructure/repository/portfolio_postgres.go`, `backend/internal/controller/http/portfolio_handler.go`; create `backend/internal/entity/portfolio_correction.go`, `backend/internal/usecase/portfolio_correction.go`, corresponding tests and next free additive migration. Native: `Features/Secondary/SecondaryFeatureViews.swift`, `Core/API/FinanceAPI.swift`, `Core/Models/FinanceModels.swift`; extract a focused portfolio editor if needed.

**Contract:**

- Revision stores owner, source trade, revision number, operation `replace|void`, full replacement fields, reason, actor, creation time. Existing trades represent revision 0.
- Replace accepts side, quantity, unit_price, fee, occurred_at, note. Reason required. Cannot change asset/owner. Void has no replacement. Second void with another key conflicts.
- Replay selects latest effective revision for each logical trade. Void excludes it from financial calculation but leaves it in audit history. Preserve original CreatedAt/ID as tie-breakers; changed occurred_at intentionally changes effective chronology.
- Use existing `ComputePortfolioPosition` rounding/fee semantics. Recompute complete ordered history, including later trades; reject any negative inventory or overflow. No automatic modification of later sells.
- New `POST /api/v1/portfolio/assets/:id/trade-corrections/preview`: source_trade_id, expected_asset_version, operation, reason, replacement. Return before/after position, realized P&L/cost/quantity delta and affected later trade IDs.
- New `POST .../trade-corrections`: same payload plus `Idempotency-Key`. Lock owned asset; check version; validate/replay; append revision, increment asset version, persist receipt and audit atomically. Return current position and revision ID.
- All trade creation and valuation/metadata mutations must follow the same asset lock/version rule so preview cannot silently commit against changed inputs.
- Same owner/key/body returns original receipt; same key/different body = 409. Wrong owner = 404. Stale version = 409; invalid resulting inventory = 422.
- Portfolio correction affects portfolio history/reports only; no wallet cash entry is created implicitly. Latest manual quote remains unchanged.

- [ ] Write entity tests: buy 10@100, buy 10@200, sell 5@300 gives quantity 15, cost 2250, realized 750. Replace first price with 120 → quantity 15, cost 2400, realized 700. Void first buy → quantity 5, cost 1000, realized 500. Void the only buy before a sell → reject.
- [ ] Add same-timestamp ordering, fractional quantity, fee, zero-price, overflow, repeated correction and backdated tests. Run `cd backend && go test ./internal/entity -run Portfolio -count=1`; record failures before implementation.
- [ ] Implement revision projection and additive schema, lock/version/receipt transaction. Test concurrent correction versus sell, rollback after failed replay, owner isolation and lost-response retry on disposable PostgreSQL.
- [ ] Add HTTP schema/status tests and native preview/history badges (`Đã sửa`, `Đã hủy`). Confirm pins reviewed asset version; conflict reloads preview. Verify exact backend totals after native save/relaunch.
- [ ] Update `docs/guides/portfolio-ios.md`, `docs/architecture/API.md`, database docs and generated OpenAPI; run relevant Go race and XCTest suites; commit P1 independently.

## P2 — AI actions dùng chung domain contracts

**Files:** `backend/internal/usecase/ai_advisor_tools.go`, `backend/internal/controller/http/ai_advisor_handler.go`, focused new proposal/execution tests; native `Features/Assistant/AssistantViews.swift` and existing confirmation bases. Add a focused action executor instead of growing handler switch indefinitely.

**Scope đề xuất:** add proposals for transaction create/update/delete; internal wallet transfer; category create/update; transaction recategorization. Preserve five current wallet/jar/budget actions. AI portfolio correction is a subsequent adapter dependent on P1; category rules and bank/payment integrations remain outside this plan because no current domain contract was established for them.

**Interfaces:** each action definition declares input JSON schema, domain adapter, preview fields, expected resource version and result schema. Confirmation still accepts only action/conversation/generation binding; financial fields come from persisted proposal. Model cannot supply owner or confirm itself.

- [ ] First inventory existing transaction/category/transfer domain methods and map each proposal to them in AI contract doc. Reuse AI-entry validation/defaults where applicable, without merging the two chat experiences.
- [ ] Write tests for create/edit/delete/transfer/category proposals: no write before confirm, bad/foreign IDs, missing category, transfer exclusion, stale snapshots, forged confirmation amounts, reset/cancel generation and API-key parity.
- [ ] Add exact expected_version support where target domain lacks it. Legacy transfer halves, credit payments and linked special transactions must use their established domain operations; never edit one leg as an ordinary transaction.
- [ ] Persist execution receipt atomically with domain writes. Response loss retry returns the same resource IDs. Delete receipt remains replayable after resource deletion. Prevent executing the same persisted action twice across API processes.
- [ ] Support multiple action cards with independent state. Execute dependent actions sequentially after prior result is known; no unpromised batch atomicity. Bulk proposals record fixed resource IDs/versions, not a query reevaluated later.
- [ ] Add native before/after preview, confirm/cancel/result/error states using common bases; include deletion impact. Run tool/schema/owner tests and synthetic Vietnamese evaluation prompts. Amount correctness is checked against tools, not an LLM judge.
- [ ] Reconcile AI capability response, public assistant guide and generated OpenAPI. Commit per adapter group, then run combined backend/native gate.

## P3 — APNs implementation and signing profiles

**Files:** create `backend/internal/infrastructure/push/apns.go` and tests; extend `backend/cmd/worker/notifications.go`, config, push-token repository/usecase; add durable push-delivery schema separate from audit stream. Native token registration wiring in `App/MyPocketIOSApp.swift`; signing under `mobile-ios/MyPocketIOS/Signing/`.

**Contract:** enqueue delivery from an existing inbox event with dedupe `(notification_id, token_id, environment)`. Lease jobs; recheck owner, token active state and opt-out before sending. Generic push payload contains notification ID and app-owned route; no amount, account name or transaction note on lock screen by default.

- [ ] Implement mock HTTP/2 APNs transport tests: 200 sent; 410 invalidates matching token registration only; 429/5xx bounded retry honoring Retry-After; invalid credentials/topic produces operator-visible failure. Token refresh races must not revoke a newer registration.
- [ ] Configure APNs team/key ID, topic, secret key and environment server-side. Redact key/token. At-least-once transport can duplicate after ambiguous response; use stable collapse ID and inbox event ID, never promise exactly-once banners.
- [ ] Separate Personal Team entitlement override from paid profile. Keep production entitlement source intact; select override at build time. Release API endpoint is independent of signing tier. A paid development profile can test APNs sandbox; TestFlight/distribution uses production APNs.
- [ ] Keep Personal Team install and inbox functional without APNs. Local notification is an optional distinct feature, not proof of server push delivery.
- [ ] Once membership/profile/key are available, test permission grant/deny, token upload/rotation, logout/revoke, real foreground/background delivery and authenticated deep-link opening. Record delivery/device evidence.
- [ ] Update push/API docs and signing runbook. Gate only real APNs/distribution acceptance on external setup; do not block P1/P2/P4/P5.

Apple references: [capabilities](https://developer.apple.com/help/account/reference/supported-capabilities-ios), [signing workflow](https://help.apple.com/xcode/mac/current/en.lproj/dev60b6fbbc7.html).

## P4 — Offline and physical regression

**Files:** `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`, focused offline XCTest files, `scripts/mypocket-production-uat-regression.sh`, `docs/operations/IOS-PHYSICAL-UAT-20261019.md`. Keep fault server/proxy in local test tooling, never a production fault endpoint.

- [ ] Prepare synthetic UAT owner fixtures: two ordinary wallets, savings wallet, categories, budgets/jars, portfolio buys/sells; record IDs and baseline balances. Do not use personal account financial data for writes.
- [ ] Automate simulator plus disposable backend faults: request dropped before acceptance; commit followed by lost response; 401/refresh; 422; 503; kill/relaunch pending queue. Assert queued row state AND exact backend count/balance/report totals.
- [ ] Acceptance arithmetic: initial A=1,000,000 and B=0; expense 100,000 and internal transfer 200,000 → A=700,000, B=200,000, total=900,000, reported expense=100,000. Retrying after lost response must preserve these totals.
- [ ] Check account switch prevents previous owner's queued writes/cache display; same-owner relogin resumes safely. Pending edits/retry/cancel use shared queue menu. Offline unsupported edit/delete must explain unavailability rather than silently queue an unsupported mutation.
- [ ] Inspect availability of current physical automation helper/MCP; use it for app navigation if connected. A screenshot-only tool is insufficient for full touch coverage. Run only one device controller.
- [ ] On physical Release build, exercise airplane-mode/reconnect, force-quit/relaunch, pending draft recovery, cached navigation without empty-page flashes, auth expiry and Face ID. Biometric match/passcode and trust dialogs remain human-only if tools cannot perform them.
- [ ] Maintain screen/function matrix: auth/account, overview, wallets, categories, transactions/transfers, savings, budgets/recurrence, jars, reports/month/travel, portfolio, AI entry/advisor, feedback resolution, notifications, API-key copy. For each record UI action, persisted state readback and pass/fail/blocked evidence. An opened screen alone is not a functional pass.

## P5 — Obtain real conversation ID and complete advisor UAT

**Files:** extend `scripts/mypocket-production-uat-regression.sh` with conversation read verification; create `scripts/mypocket-advisor-uat.sh` for explicit live-provider runs; update provider/UAT runbooks. Never silently change the GET-only harness into a paid POST runner.

- [ ] With UAT credentials held outside repo, submit `POST /api/v1/ai/advisor/messages/async` using a new client_request_id and a synthetic question about that owner's fixtures.
- [ ] Capture conversation_id/run_id from response into permission-restricted temporary evidence; poll existing run endpoint until terminal with bounded timeout. No automatic resubmission on timeout.
- [ ] Pass returned ID into authenticated history GET tests; verify persisted user/assistant turns, order, pagination, cards and numeric grounding against fixture totals.
- [ ] Verify run/event/history inaccessible to a second UAT owner. Test cancel, reset and 30-minute idle expiry in deterministic local tests; on prod inspect actual contract without modifying clocks or deleting other users' histories.
- [ ] Run native chat reopen/relaunch, follow-up context, Stop, action draft/confirm/replay and duplicate-send cases with matching IDs. Label any provider-dependent failure explicitly; never fabricate an answer to make UAT pass.
- [ ] Record live model/configuration identifier, latency, usage and redacted error stages. Do not log credentials or personal finance contents. This conversation ID is obtained by the harness; user need not locate it manually.

## Execution order and closure

1. Run P5 and prepare P4 fixtures first to replace assumed blockers with real evidence.
2. Parallel independent lanes: P1 portfolio, P2 domain-adapter inventory/tests, P3 push/mock/signing. Coordinator serializes migrations, shared router/docs changes and device control. P2 portfolio adapter waits for P1.
3. Integrate changes; execute relevant tests once per meaningful change, then combined Go race/native regression. Backup/migrate/rollout matched API-worker only after gates pass; install matched Release iPhone artifact.
4. Complete P4 physical sweep and P5 live flows; attach evidence per feedback. User's resolved/unresolved choice stays user-controlled. Changelog states shipped behavior and remaining limitations.

Internal docs: affected tickets, architecture/database, deployment/provider/device runbooks, backlog and validation matrix. Public docs: portfolio, assistant actions, notifications/offline guides plus OpenAPI as the AI-readable contract. This planning turn changes only this plan; no new endpoint, migration, production write or release claim.

External release-aware RAG is not among the five requested items; it remains explicitly open in the broader backlog.
