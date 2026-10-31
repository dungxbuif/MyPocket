# Accumulated Assets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Nâng portfolio thành Tài sản tích lũy, nhập số lượng × đơn giá rõ ràng, dễ sửa và tự cập nhật giá vàng từ catalog hệ thống.

**Architecture:** Giữ holdings/trades/correction engine và `/portfolio` API. Bổ sung unit/price mode/catalog binding, bảng quote dùng chung và worker lấy giá mỗi 10 phút. Summary resolve quote hiện tại; custom và tài sản cũ tiếp tục manual.

> **Reconciliation 2026-11-01:** Tasks 1–3 and the native integration are
> implemented and covered by the fresh backend/native matrix. Task 4 is
> deployed: BTMC/BTMH adapters use the production n8n source contract,
> migration `000049` is included in the live schema (now at `000052`), and the
> worker reported a successful two-provider pass after rollout. Remaining
> acceptance is provider freshness/owner review, not an unimplemented local
> code path.
>
> **Unit catalog decision 2026-11-01:** The native create/edit flow now loads
> a backend-owned catalog with only `Vàng` and selectable units `chỉ`, `lượng`,
> `phân`. Automatic pricing is offered only when a configured market product
> matches the selected unit; the current seeded gold products are `chỉ`, so
> `lượng` and `phân` remain valid manual-price choices. Future asset types and
> units can be added by extending the backend catalog without accepting free
> text from the client.

**Tech Stack:** Go/Gin/GORM/PostgreSQL, worker hiện có, SwiftUI/XCTest.

**Spec:** [Accumulated-assets design](../specs/2026-10-31-accumulated-assets-design.md). Companion tổng hợp các quyết định trong hội thoại; các default kỹ thuật chưa có trong API nguồn là lựa chọn của người thiết kế.

## Global Constraints

- Tên chung “Tài sản tích lũy”; vàng tự cập nhật + custom nhập tay. Không tạo holding trùng vì mục đích đầu tư/tích lũy.
- Tài sản cũ mặc định manual, unit chưa xác nhận; không đoán đơn vị, đổi quantity/price hoặc tự ghép nguồn.
- Giá thị trường không sửa lịch sử mua, giá vốn hoặc giao dịch tiền mặt.
- Hai nguồn giữ riêng theo product mapping; valuation dùng giá cửa hàng mua vào. Poll10phút; đề xuất TTL60phút, skew5phút, timeout10giây, max body1MiB.
- Product system unit cố định; unit mismatch chặn binding; conversion holdings hiện có ngoài scope.
- API thật chờ owner cung cấp; không gọi fixtures là live integration. Không cron AI, automatic stocks, price charts/alerts hoặc giao dịch tự động.
- Native light mode, màu/base hiện tại; density, fields, status và card qua DesignSystem; mask cả VoiceOver.

## Review Focus

- Hai nguồn cùng tên sản phẩm nhưng khác đơn vị không được ghép bằng tên — Task2/3.
- Tài sản cũ chưa biết đơn vị không được tự định giá bằng quote hệ thống — Task1/2.
- Quote đến trễ hoặc provider trả timestamp tương lai không được thay giá mới hơn — Task3.
- Quote refresh không làm correction đang mở conflict chỉ vì asset version bị tăng — Task3/5.
- Tổng portfolio có asset unknown/stale phải nói rõ độ đầy đủ, không coi unknown bằng0 — Task2/5.

## File ownership và verification

Go commands tại `backend/`; xcodebuild tại repo root. Dùng simulator UDID `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`, register Swift sources/test vào project và kiểm tra số test >0. Preserve dirty user changes; project file chỉ stage registration hunks thuộc task. Không sửa migration đã áp dụng; lấy số forward migration tiếp theo lúc execution, không giữ chỗ số trùng report/chat plans.

Mỗi task RED→GREEN rồi commit các file riêng task, reviewer kiểm spec+quality. Trước task mới đọc ledger; không chạy implementation song song trên portfolio handler/repository/SecondaryFeatureViews/project file. Không merge/push/deploy trong plan này.

### Task 1: Asset unit và migration bảo toàn dữ liệu

**Files:** Modify `backend/internal/entity/portfolio.go`, `backend/internal/repository/portfolio.go`, `backend/internal/infrastructure/repository/portfolio_postgres.go`, `backend/internal/controller/http/portfolio_handler.go`, their existing `_test.go` files, `backend/migrations/migrations_test.go`, `backend/migrations/README.md`. Create next forward pair `backend/migrations/<next>_asset_units.up.sql`/`.down.sql` and `docs/decisions/ADR-028-accumulated-assets-pricing.md` (check ADR ID free).

**Interfaces:** additive `PortfolioAsset.UnitLabel *string`, `UnitConfirmed bool`, `PriceMode string` default manual. Add owner/version-scoped `ConfirmAssetUnit(ownerID, assetID string, expectedVersion int64, unitLabel string)`. PATCH `/portfolio/assets/:id/unit` accepts only expected_version/unit_label; confirms legacy interpretation, never converts numbers. Custom create accepts unit_label; old create body remains valid as unconfirmed manual.

- [ ] Write `TestAssetUnitMigrationPreservesTradesAndCorrections`, `TestConfirmAssetUnitRejectsForeignOwnerAndStaleVersion`, `TestCannotChangeConfirmedUnitWithTrades`. Golden migration fixture includes fractional quantity, buy/sell, price and correction receipts; assert values/receipts identical after migration.
- [ ] Run `rtk proxy go test ./internal/entity ./internal/controller/http ./internal/infrastructure/repository ./migrations -run 'AssetUnit|Migration' -count=1`; expect RED for missing behavior/schema.
- [ ] Implement fields/migration/version guard and confirmation endpoint. Existing rows manual/unconfirmed with unit null; don't backfill gold guesses. Define DTO compatibility for omitted new fields and legacy routes. Use canonical repository transactions.
- [ ] Run command GREEN with disposable migrated PostgreSQL; record DB tests actually ran, plus legacy create/read/correction tests. Commit scoped changes `feat: preserve holdings while adding explicit asset units`.

### Task 2: System catalog, binding and valuation provenance

**Files:** Create `backend/internal/entity/asset_market.go`, `asset_market_test.go`; `backend/internal/repository/asset_market.go`; `backend/internal/infrastructure/repository/asset_market_postgres.go`, `asset_market_postgres_test.go`; `backend/internal/controller/http/asset_market_handler.go`, `asset_market_handler_test.go`; next forward migration pair `<next>_asset_market.up.sql`/`.down.sql`. Modify portfolio entity/repo/handler/summary and `backend/cmd/api/main.go`.

**Interfaces:** `MarketProduct{ID,ProviderKey,ExternalProductID,Name,Kind,Unit,Currency,ValuationSide,Enabled}`; `MarketQuote{ProductID,Price,ObservedAt,FetchedAt}`. GET `/portfolio/catalog` lists enabled/configured products. PATCH `/portfolio/assets/:id/price-source` accepts expected_version, price_mode, product_id; validates owner/confirmed-unit/exact-unit/currency. Creating new system holding references catalog product; no name-based match. Summary adds price source/as_of/status and unknown/stale counts.

- [ ] Write `TestMarketBindingRequiresConfirmedMatchingUnit`, `TestMarketCatalogSeparatesSameNameDifferentSources`, `TestMarketSummaryUnknownIsNotZero`, `TestLegacyPriceEndpointCannotOverrideAutomaticAsset`. Run `rtk proxy go test ./internal/entity ./internal/controller/http ./internal/infrastructure/repository -run 'Market|PriceSource' -count=1` RED.
- [ ] Implement catalog/quote storage, explicit binding and manual-unlink flow with expected_version. Quote lookup owner-authorized through holding binding; product catalog public-to-authenticated scope contains no owner data. Unlink keeps displayed last quote as manual price with provenance; existing trades/unit unchanged. Product change with existing trades blocked in first version.
- [ ] Extend summary valuation through existing fixed-point math; unknown full total remains unavailable, known subtotal explicitly labeled. Stale values retained and flagged; price timestamp not replaced by fetch timestamp. Test same quote used by2 owners with isolated holdings.
- [ ] Run command GREEN plus existing portfolio suite on disposable DB; commit `feat: add catalog-bound asset valuation`.

### Task 3: Shared scheduled quote polling with provider adapter

**Files:** Create `backend/internal/usecase/market_prices.go`, `market_prices_test.go`; `backend/internal/infrastructure/marketdata/provider.go`, `provider_test.go`; `backend/cmd/worker/market_prices.go`, `market_prices_test.go`; next migration pair `<next>_market_poll_claims.up.sql`/`.down.sql`. Modify market repository/adapter, `backend/cmd/worker/main.go`, runtime config docs.

**Interfaces:** `MarketPriceProvider.Fetch(ctx context.Context) ([]ProviderQuote,error)` returns external product ID, unit/currency, buyback price and source observed_at. `MarketPriceService.PollDue(ctx, now time.Time) (PollResult,error)` claims provider+30minute slot; DB claim lease token fences completion. Provider registry backend-configured; no user URL.

- [ ] Write `TestMarketPollSharedAcrossOwners`, `TestMarketPollConcurrentWorkers`, `TestMarketPollRejectsOutOfOrderAndFutureQuote`, `TestMarketPollFailureKeepsLastPrice`, `TestMarketPollDoesNotBumpAssetVersion`. Run `rtk proxy go test ./internal/usecase ./internal/infrastructure/marketdata ./internal/infrastructure/repository ./cmd/worker -run Market -count=1` RED.
- [ ] Implement HTTP transport timeout10s/body1MiB with bounded errors; parser rejects wrong unit/currency/nonpositive/unknown product/time>now+5min. Keep observed_at and fetched_at separate. Fetch once/provider/slot, atomically update accepted quotes, no owner asset rewrites; lease expiry/restart can repeat GET but no price rollback. Failed slot retries bounded to next scheduled slot; expired in-flight claim reclaimable.
- [ ] Add deterministic adapter fixtures, injection-free static product mapping and disabled/unconfigured-source capability. Run preceding command GREEN and no-provider worker regression. Record metrics source/latency/status/count without secrets. Commit `feat: poll shared asset prices on schedule`.

### Task 4: Two real gold-source adapters — implemented locally, rollout pending

**Files:** Create `backend/internal/infrastructure/marketdata/gold_source_a.go`, `gold_source_a_test.go`, `gold_source_b.go`, `gold_source_b_test.go`; `docs/work/tickets/GOLD-SOURCE-CONTRACT.md`. Modify registry/config and catalog seed forward migration only after actual product list known.

**Consumes:** the production n8n deployment record and its sanitized parser contract for BTMC and BTMH. Tasks1–3/5 can use fixtures; production automatic pricing still awaits the database/worker rollout gate.

- [x] Record each real source's mapping in contract ticket; never commit credentials/raw private headers. No invented API URL/schema or fake live status.
- [x] Write sanitized fixture tests covering product identity, buyback unit mapping, observed_at parsing and malformed/partial responses.
- [x] Implement parsers against the shared provider contract. No unit conversion beyond the source's declared BTMC `1 = 1.000 VNĐ` scale, and no cross-source averaging/fallback.
- [x] Run fixture tests and a bounded read-only live GET through both adapters; the check returned two validated quotes per source on 2026-11-01. Migration rollout and worker runtime smoke remain pending.

### Task 5: Native accumulation UI, trade totals and discoverable editing

**Files:** Create `mobile-ios/MyPocketIOS/Features/Portfolio/AccumulatedAssetsView.swift`, `AssetTradeEditor.swift`, `AssetPriceSourceView.swift`; `mobile-ios/MyPocketIOSTests/AccumulatedAssetsTests.swift`. Modify `Core/Models/FinanceModels.swift`, `Core/API/FinanceAPI.swift`, `Features/Secondary/SecondaryFeatureViews.swift`, `Features/More/MoreView.swift`, project file, existing portfolio tests and `MyPocketIOSUITests/LaunchTests.swift`. Shared variants only in `DesignSystem/MPFields.swift`, `MPFinanceCards.swift` if required.

**Interfaces:** additive optional native unit/mode/source/quote status DTO fields; `FinanceAPI.marketCatalog()`, `confirmAssetUnit(...)`, `updateAssetPriceSource(...)`; keep existing create trade/correction endpoints. Extract portfolio composition from SecondaryFeatureViews only within portfolio range; existing shell route stays stable, visible label becomes “Tài sản tích lũy”.

- [ ] Write `testTradeTotalUsesQuantityTimesUnitPrice`, `testLegacyHoldingPromptsForUnitBeforeAutomaticPrice`, `testQuoteRefreshDoesNotResetCorrectionDraft`, `testUnknownValuationMasking`, `testEditTradeUsesExistingCorrectionAPI`. Example2units×8.000.000=16.000.000 before fee; test0.5units, buy+fee/sell−fee and overflow. Register sources then run `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -only-testing:MyPocketIOSTests/AccumulatedAssetsTests test` RED.
- [ ] Implement “Chọn từ hệ thống”/“Tạo tài sản riêng”; empty/unconfigured catalog explains manual option. Existing holdings show manual/unconfirmed badge and explicit unit/source setup with preview. Auto assets show buyback source/as_of/stale; manual assets expose update-price action.
- [ ] Implement shared form anatomy: số lượng with unit, đơn giá đ/unit, phí, date, computed gross/net. History main amount is total, subtitle per-unit price; unconfirmed legacy unit gets neutral label, not guessed gold unit. Match server rounding; server totals authoritative. Keep purchase price independent of latest quote.
- [ ] Open trade detail on row tap with visible “Sửa giao dịch”; reuse correction preview+confirm, source trade ID/version/idempotency and before/after position. Keep rename asset/update-market-price separate. Preserve conflict messages/draft and complete masking/VoiceOver.
- [ ] Run command GREEN and API-backed UI fixtures custom create/buy/sell/edit, legacy unit confirmation, source link/mismatch, stale/failure/unknown pricing. Capture small/regular iPhone and large type light; register new tests and prove count>0. Commit scoped changes `feat: simplify accumulated asset entry and editing`.

### Task 6: End-to-end migration/quote proof and docs reconciliation

**Files:** Create `docs/work/audits/ACCUMULATED-ASSETS-UAT-20261031.md`; update portfolio screen/base specs, API/ERD, ADR, migration runbook, `docs/work/VALIDATION_MATRIX.md`, `docs/work/BACKLOG.md`, `docs/CONTEXT.md`, `docs/releases/CHANGELOG.md`.

- [ ] Run portfolio entity/handler/repository and market worker/provider tests with disposable migrated DB. Verify before/after legacy rows unchanged, quote updates valuation only, source failures retain price and correction replay remains exactly once.
- [ ] Run native focused tests, `mobile-ios/scripts/verify-project.sh`, `mobile-ios/scripts/check-base-components.sh` and `rtk proxy git diff --check`. Inspect xcresult/screenshot and source proof; fixture success is separate from real API source success.
- [ ] Record owner UAT for “2chỉ×đơn giá”, visible trade edit, old-unit confirmation, source freshness, custom manual price and partial portfolio total. Keep Task4 blocked on missing contract if necessary; don't claim automatic pricing live or whole feature done until actual integration evidence exists.
- [ ] Reconcile API/ERD/screen/base/ADR/context/backlog/changelog; link report plan for later net-worth integration without adding holdings to cash totals. Commit `docs: record accumulated asset validation` with only reconciled task docs. Deployment approval separate.

## Self-review

Migration/units Task1; catalog/binding/provenance/summary Task2; worker/claims/provider transport Task3; actual sources Task4; density/input/edit/price UI Task5; proof/docs Task6. All5 Review Focus risks have named tests. Types/calls use one market provider service and canonical portfolio correction math; no second ledger. Dependencies1→2→3→4, Task5 after1/2 can run with fixtures alongside3; Task6 integrates. Task4's external dependency is explicit and doesn't prevent manual UX work. No runtime code, migrations or credentials changed during planning.

Execution preference retained: agent implementation + separate review per task, one owner per shared file/migration. Review this plan and companion design together before implementation.
