# Native iOS port audit

Reviewed: 2026-10-13
Scope: `/Users/dungxbuif/Downloads/stitch_my_pocket.zip`, the current web
routes, and the native iOS target in this worktree.

## Current-source verification — 2026-10-13

- Native unit/contract suite: 148/148 pass; Release warnings-as-errors,
  project/base-component guards and `git diff --check` pass.
- Overview week-report totals now share one financial rule across platforms:
  account-local Monday–Sunday bounds and only included income/expense rows;
  transfer, adjustment and report-excluded rows are omitted. Deterministic
  web boundary/filter tests and a native policy regression pass. Evidence is in
  the [validation matrix](VALIDATION_MATRIX.md).
- Reports parity is now verified against current web behavior: safe
  income/expense summary chart and three metrics, always-visible trend card,
  category rows, activity, and Money Insider comparison ring. Focused
  API-backed Reports UI tests pass 1/1 for hierarchy/donut reachability and
  1/1 for account-local persisted trend values/report-exclusion handling. The
  disposable DB and Redis test database were cleaned; screenshots and commands
  are recorded in the [validation matrix](VALIDATION_MATRIX.md).
- The approved feedback-dock implementation now partitions the shell into a
  clipped route viewport, fixed feedback lane and persistent tabs. Focused
  iPhone 16 collision, capture, category/notification viewport, More final-row
  reachability, compact iPhone SE collision/capture and Account hit-testing,
  and independent form-inset regressions pass; evidence and
  remaining full-gate status are in the
  [validation matrix](VALIDATION_MATRIX.md).
- Native unit tests pass 147/147; Release warnings-as-errors build, project
  verification, base-component guards and `git diff --check` pass.
- Fresh device check on 2026-10-13: the owner's iPhone 15 Plus is paired,
  connected, unlocked, and in Developer Mode. The installed Apple Development
  certificate for `bui.dung.2k@icloud.com` expired on 2025-01-12, and the local
  provisioning-profiles directory is empty. Therefore Ad Hoc signing/install
  is blocked until a current Apple Developer signing identity and a device-
  registered provisioning profile are available. No Developer Portal change,
  signing attempt, or device install was performed.
- Fresh verification on 2026-10-13: native unit/contract tests pass 148/148;
  Release builds pass for generic iOS Simulator and generic iOS device with
  Swift warnings-as-errors. Project/base-component guards, web design/auth/
  transaction tests, web typecheck/build and `git diff --check` pass. Backend
  `go test ./... -count=1` also passes with database URLs unset; PostgreSQL
  integration tests requiring `TEST_DATABASE_URL` are skipped in this safe
  no-database run. Web production build retains its existing >500 KB chunk
  advisory.
- The full port still requires authenticated owner visual/device acceptance in
  [the UAT checklist](NATIVE-PORT-100-OWNER-UAT.md). This UI-only fix does not
  close that broader parity/UAT goal.

## AI-entry ambiguous-request recovery — 2026-10-08

- Source comparison found that web retains a submitted `request_id` after a
  transport/5xx failure and permits only GET recovery, while iOS exposed “Soạn
  nội dung mới” and could submit a second request before the first outcome was
  known.
- iOS now retains the same ID and draft, locks input/file actions, and exposes
  only `GET /ai/entry/requests/{id}`. A found result is shown without reposting;
  only a GET 404 re-enables explicit submission. Removed the inaccurate claim
  that no transaction was saved on an ambiguous outcome.
- RED/GREEN: regression policy tests first failed to compile before the policy
  existed. Then native unit suite passes 146/146 and isolated iPhone 16 UI
  recovery passes 1/1. Evidence:
  `/tmp/mypocket-ai-entry-recovery-unit-full-20261008.xcresult` and
  `/tmp/mypocket-ai-entry-recovery-final-20261008.xcresult`.
- This closes only the AI-entry retry-parity defect. It does not close the
  complete native port, populated-state visual parity, or owner-device UAT.
- No API/backend/schema/migration/provider or production change; public API
  docs are unchanged because endpoint contract is unchanged.

Owner-authenticated production/device checks are tracked in the prioritized
[native port UAT checklist](NATIVE-PORT-100-OWNER-UAT.md); this simulator and
API-backed evidence does not substitute for that acceptance.

## Core financial business-flow API-backed sweep — 2026-10-08

- Ran four end-to-end native UI tests on iPhone 16 Simulator against the local
  API and a fresh migrated PostgreSQL database: wallet create/update; create
  and update a transaction from wallet detail; create/read/delete an internal
  transfer pair; and persist/reload report trend values while excluding a
  transaction from reporting.
- Result: 4/4 passed, zero failures, in
  `/tmp/mypocket-core-business-20261008.xcresult`.
- The selected tests use their own disposable Redis container and database;
  provider keys are disabled. Database, Redis container and DerivedData were
  removed after the run. Production and shared Redis DB 15 were not used.
- This is targeted business-flow coverage, not evidence that every feature is
  complete. Full matrix, the unresolved feedback dock overlap, populated-state
  visual review and physical-device UAT remain open.

## iOS app icon packaging — 2026-10-08

- **Observed gap:** the iOS target names `AppIcon` in build settings but has no
  asset catalog, so an installed native build has no product icon. The web
  PWA's `app/public/pwa-icon.svg` is the current canonical artwork.
- **Intended behavior:** package that same wallet artwork as a 1024×1024 iOS
  AppIcon, using a square background so iOS applies its standard icon mask;
  keep the asset catalog wired through XcodeGen and the generated project.
- **Verification:** fail the project guard when the icon asset/catalog or
  XcodeGen resource wiring is missing; build Release and inspect the compiled
  app asset catalog for `AppIcon`. No business behavior, API, schema or
  production configuration changes.
- **Result:** `scripts/generate-app-icon.sh` deterministically derives the
  square-background 1024×1024 PNG from the canonical PWA SVG, avoiding a second
  artwork source. The project guard regenerates and byte-compares the PNG, in
  addition to checking dimensions and resource wiring. Release build and
  compiled `Assets.car` inspection confirm the `AppIcon` rendition exists. No
  backend/API/schema or production changes.

## Secondary-route navigation testability — 2026-10-08

- The 15-destination More-menu XCTest did not finish within 328 seconds and
  was manually interrupted while still traversing routes. Its partial result
  is cancellation evidence, not a product failure or a pass.
- A focused iPhone 16 UI test opens AI Entry from More, verifies the shared
  `screen-back` identifier is hittable, taps it, and returns to More (1/1 in
  23 seconds). Inspection confirmed `MPIconButton` assigns this identifier for
  the “Quay lại” label; no product bug or runtime change was needed.
- The broad test was reordered to match the actual More sections and now uses
  the menu's preserved scroll position. It passes 1/1 in 240 seconds and
  exports 15 named route screenshots. Evidence:
  `/tmp/mypocket-port100-secondary-routes-optimized-20261008.xcresult` and
  `/tmp/mypocket-port100-secondary-routes-attachments-20261008/`.
- This establishes fixture route/title/back-action coverage only, not
  populated-state API behavior, pixel-level parity, or owner/device acceptance.

## Reports exclusion persistence and VoiceOver trend values — 2026-10-08

- API-backed trend verification exposed an excluded expense in the Reports
  summary. A PostgreSQL repository regression reproduced the cause: normal
  `TransactionPostgresRepository.Create` submitted `included_in_reports=false`
  but GORM's `default:true` handling persisted `true`.
- The repository now explicitly persists the caller's report-inclusion choice
  in the same transaction. PostgreSQL regression verifies both `false` and
  `true`; full backend suite passes against the disposable DB clone.
- Native `MPTrendChart` VoiceOver value now reports each existing daily point's
  income/expense totals chronologically, using shared amount masking. The
  API-backed iPhone 16 test persists income 123,456, expense 34,567 and excluded
  expense 888,888 VND, then verifies the persisted exclusion and exact chart
  value. Result: `/tmp/mypocket-report-trend-green-20261008.xcresult`.
- Cross-flow revalidation after the repository correction passes standard
  wallet-detail transaction create/update and recurring CRUD 2/2 at
  `/tmp/mypocket-create-regression-ui-20261008.xcresult`; generated recurring
  ledger creation remains compatible. Fixture SQL found zero wallets,
  transactions or schedules; the local API and exact disposable clone were
  stopped/removed.
- No route, schema, migration or public API contract changed. Production and
  port 8080 were not used. This closes the report-flag persistence and chart
  accessibility gap only; the full iOS port and owner/device acceptance remain
  open.

## Travel Mode transaction link/unlink state proof — 2026-10-08

- Closed a missing E2E proof: the native UI regression creates a Travel Mode
  and ordinary transaction, selects the event from transaction detail,
  verifies its persisted `travel_event_id`, clears the selection, and verifies
  the persisted null state.
- iPhone 16 Simulator passes 1/1 at
  `/tmp/mypocket-travel-link-final-20261008.xcresult`; backend auto-link and
  clear-contract tests pass 2/2. Native unit/contract suite passes 144/144 at
  `/tmp/mypocket-travel-link-unit-20261008.xcresult`.
- Isolation: cloned local disposable database
  `mypocket_travel_link_verify_20261008`, Redis DB 15 and loopback API 8081
  only. Teardown verified zero synthetic wallet, event and transaction rows;
  API was stopped and that exact database dropped. Production and the unrelated
  API 8080 were not touched.
- Contract/schema/backend implementation are unchanged; public API and
  AI-readable endpoint docs need no update. This closes only the native
  link/unlink state-proof gap, not the wider Travel Mode acceptance, full port,
  visual parity or physical-device gates.

## Category-picker copy and clear-action parity — 2026-10-08

- The prior iPhone capture showed every category row repeating `Chi` and the
  budget clear action as a filled/centered control. The native picker now uses
  the web copy (`Nhóm mặc định`, `Nhóm cá nhân`, or `Không áp dụng cho lựa
  chọn này`) and the shared transparent, action-colored, leading-aligned ghost
  button.
- Eligibility is distinct from ancestry/search context. An eligible root kept
  above a matching child remains selectable, as in web; wallet/type-ineligible
  parents remain visible but non-interactive.
- API-backed budget CRUD verifies the exact accessibility label and persisted
  category scope; recurring and wallet-detail transaction flows pass. The
  screenshot and full test evidence are recorded in `VALIDATION_MATRIX.md`.
  No backend/API/schema/migration change was needed. Overall category parity,
  authenticated device acceptance and the complete port remain in progress.

## Category-picker root-card parity — 2026-10-08

- Source comparison found web `CategoryTreeSelector` composes one
  `BaseCategoryTree` surface per root, while iOS had wrapped all root groups in
  a single card. iOS now partitions the existing depth-first selection result
  into root groups and renders each through the shared selection-tree base.
- Pure tests verify stable root order and parent-context preservation. The
  API-backed iPhone 16 budget create/update/delete flow passes 1/1 at
  `/tmp/mypocket-category-root-cards-ui-pass-20261008.xcresult`; it checks
  multiple root groups, searchable selection and persisted category ID.
- Shared-component regressions for recurring create/update/delete/run-due and
  wallet-detail transaction create/update pass 2/2 at
  `/tmp/mypocket-category-picker-crossflow-ui-final-20261008.xcresult` and
  verify the selected category survives the API write.
- API base URL for both app and fixture helper was explicitly overridden to
  local loopback port 8081. Only the disposable local database/Redis fixture
  was used; production was not accessed. No API, schema, migration or public
  endpoint behavior changed. Native unit/contract tests pass 143/143, Release
  warnings-as-errors passes, and project/base-component guards plus diff check
  pass. Physical-device and owner visual acceptance stay open.

## Populated goal-wallet ledger CTA — 2026-10-08

- A real persisted goal-wallet income row exposed the CTA overlap missed by
  the earlier empty-ledger screenshot: CTA frame y=780.67, shell navigation
  y=767, raised add affordance y=730. The 104pt → 160pt cushion-only
  experiment had no effect.
- Root cause: `TransactionsView` embedded `WalletDetailView`'s own `ScrollView`
  inside the Transactions outer `ScrollView`; the goal footer was anchored to
  content height. The goal route now renders as the Transactions root and owns
  its viewport scroll/safe-area CTA. Other wallet scopes are unchanged.
- API-backed iPhone 16 simulator: scoped populated-ledger and explicit detail
  tests pass 2/2; the scoped test scrolls until the complete transaction row is
  above the fixed CTA and confirms CTA clearance above the shell add button.
  See `/tmp/mypocket-goal-ledger-final-20261008.xcresult` and latest
  `/tmp/mypocket-goal-ledger-reachability-final-20261008.xcresult`.
- Screenshot reviewed at
  `/tmp/mypocket-goal-ledger-reachability-attachments/D2901CC9-C45F-4EEC-8CED-92C14702B4F5.png`.
  API fixture cleanup verified (0 matching wallets and transactions). Local
  stage API/database/Redis only; no backend, production or API contract change.
- The two normal transaction regressions (weekly ledger and wallet selection)
  also pass 2/2; native unit/contract tests pass 142/142. Release
  warnings-as-errors, project verification, both base-component checks and
  `git diff --check` pass. This does not close the full API-backed UI suite,
  feedback-bubble collision, or device/owner visual acceptance.
- This closes the populated scoped-ledger CTA overlap only; full port, physical
  device and visual sign-off remain open. At this 2026-10-08 checkpoint the
  independent floating feedback overlap was still unfixed; the 2026-10-13
  shell-dock implementation resolves it in automated simulator coverage.

## Feedback-bubble collision regression — 2026-10-08

- A focused default UI regression asserts no overlap on visible More menu rows
  and category-editor save/cancel actions. It ran 1 test and failed 2 overlap
  assertions: the More wallets row and category-editor cancel button intersect
  the bubble. Xcode then hung while finalizing its xcresult; the session log
  contains both failures. The ineffective shared-inset increase was removed.
- Evidence locates the bubble at `(325, 600, 44, 44)` and category cancel at
  `(201.7, 557, 175.3, 48)` on the iPhone 16 viewport. The shell-overlay and
  scroll-view boundary needs an architecture review before another layout
  attempt; do not repeat route-wrapper or inset-only changes.
- Backend `go test ./...` passes 231 tests across 16 packages; PostgreSQL is
  stopped locally, so database-backed tests were skipped. This visual issue
  does not require an API/backend change. After removing the attempt, native
  unit tests pass 138/138, Release warnings-as-errors passes, and project/base
  guards plus diff check pass.
- Compact-device, another long-list and a finalized screenshot remain open.
  Do not call the collision fixed or the overall port complete.
- Current-source revalidation after the goal-ledger viewport change still fails
  at `/tmp/mypocket-feedback-collision-revalidate-20261008.xcresult`: the More
  wallets row, the More Assistant row after scroll, and the category-editor
  cancel action overlap the feedback button. No dock architecture change has
  been made pending owner review of `NATIVE-FEEDBACK-BUBBLE-DOCK`.

## Current-source verification refresh — 2026-10-08

- Backend without database configuration: `rtk go test ./...` passes 240
  tests across 16 packages. The local stage database was initially at schema
  version 29, so DB-backed tests reported missing `categories.is_active` until
  the already-defined additive migration `000030` was applied locally. Stage
  now reports version 30, and `TEST_DATABASE_URL=<local-stage-DSN> rtk go test
  -race ./... -count=1` passes 265 tests across 16 packages. A separate fresh
  disposable database also passed the same suite and was dropped afterward.
  No production DB/API was accessed or changed.
- Native `MyPocketIOSTests`: 142/142, zero skips/failures on iPhone 16
  Simulator; Release warnings-as-errors build and base/project guards pass.
- The physical iPhone 15 Plus reports unavailable through `devicectl`; no
  physical-device acceptance is claimed. At this 2026-10-08 checkpoint the
  feedback-collision UI regression remained red; the 2026-10-13 resolution is
  recorded in the current-source section above.
- Evidence: `/tmp/mypocket-port100-unit-20261008.xcresult` and
  `/tmp/mypocket-port100-release-20261008`.

## More destination navigation and isolated API render — 2026-10-08

- The ordered iPhone 16 XCTest traversed all 15 More destinations and checked
  screen titles and back actions (1/1 test, 240 seconds). This is route
  navigation evidence only. The test used `-UITestAuthenticated`, which gives
  the app an in-memory synthetic session, not an API session.
- **Evidence correction:** the exported screenshots at
  `/tmp/mypocket-port100-secondary-routes-attachments-20261008/` show API-load
  error states on data-backed routes. At capture time neither local API port
  8080 nor 8081 had a listener. These images are not evidence of live API
  rendering, valid empty states, populated state, or visual parity; the earlier
  description of them as API-backed was incorrect.
- The selected test's route/title/back assertions remain valid for navigation,
  but the API-backed claim for that first screenshot bundle is withdrawn. A
  separately built API-backed route run is recorded below.
- Repeated with a fresh `DEBUG LOCAL_API_UI` build and real local OAuth fixture
  exchange against an empty database migrated from source to version 30. All
  15 More destinations passed navigation/title/back assertions again (1/1,
  240.705 seconds); 15 screenshots were exported to
  `/tmp/mypocket-port-routes-api-attachments-20261008/`. API logs show the
  fixture exchange and owner-scoped data reads; visual samples of wallet,
  categories, budgets, reports, jars, credit, months, Travel, recurring,
  portfolio, notifications, account and feedback show valid empty states with
  no network-error placeholder.
- Test isolation: database `mypocket_port_routes_20261008` was created from
  `template0` and migrated to version 30; Redis used a dedicated loopback-only
  container on 16385. AI/OCR/storage credentials were overridden blank, so
  Finance Assistant shows its expected disabled state and AI Entry reports OCR
  unconfigured; this is not OCR/provider verification. After the run, the API
  process, exact test database and dedicated Redis container were stopped or
  removed. Existing stage DB, Redis DB 15, port 8080 and production were not
  touched.
- Screenshot inspection confirms the feedback bubble still overlaps category
  and assistant content; the docked-shell fix remains gated on owner review of
  the detail contract. This run proves route reachability and valid empty-state
  data only, not populated-data behavior, pixel parity or physical-device
  acceptance.
- **Still open:** populated per-domain scenarios, systematic Stitch/web
  comparison, feedback-dock collision, and
  owner acceptance on device.

## Finance Assistant structured-card density — 2026-10-08

- **Observed mismatch:** web caps transaction-list cards at five returned
  rows, and wallet, goal, budget, and jar cards at eight. Native renders every
  returned item. This changes card density and allows long responses to
  dominate the persistent chat.
- **Intended behavior:** use the same visible prefixes on iOS while preserving
  `total_count` and summary/footer values from the complete API response.
  Unknown card types and the read-only/write-confirmation policy are unchanged.
- **Scope:** native presentation only; no API, schema, provider or public
  endpoint changes. Verify a list larger than each cap and normalized legacy
  aliases before closing the gap.
- **Result:** native now limits transaction cards to five and wallet, goal,
  budget and jar cards to eight, with alias normalization applied before the
  cap. Missing `total_count` falls back to the visible five valid transaction
  rows, matching web; server totals remain unchanged. Focused simulator tests,
  Release warnings-as-errors, project/base guards and diff check pass.

## Wallet management parity follow-up — 2026-10-08

- **Observed mismatch:** native lists each wallet as an independent card with
  always-visible Chi tiết/Sửa/Xóa controls and a header add button. The web
  `WalletSelectionList` groups wallets by `is_in_total`, shows the included
  balance as “Tổng cộng”, and reveals wallet editing/deletion through a Sửa / Xong
  mode plus footer actions.
- **Intended behavior:** bring the native wallet-management list into parity
  with those grouped/aggregate/edit-mode behaviors while preserving the native
  Chi tiết extension and selecting a wallet (or aggregate) to open its scoped
  ledger. “Liên kết dịch vụ” remains present but disabled, matching web.
- **Technical decision:** add one reusable DesignSystem composition for the
  grouped wallet selection list and compose only existing MP base rows,
  surfaces, text, buttons, and headers. Aggregate selection maps to the existing
  unscoped ledger callback (`nil`); per-wallet selection remains owner-scoped.
  No API, schema, authorization, or financial calculations change.
- **Verification plan:** first add a failing API-backed simulator test with
  included and excluded wallet fixtures. Verify groups, aggregate row, edit-mode
  toggle, add/service actions, and that Chi tiết remains reachable. Then make it
  pass on isolated `mypocket_stage_v1` + Redis DB 15 / API 8081, run native unit
  tests, Release warnings-as-errors and project/base/diff guards, and reconcile
  wallet design, base-component, validation, backlog and changelog docs.
- **Data-integrity finding from the failing UI proof:** the fixture explicitly
  sent `is_in_total=false`, but the API persisted and returned `true`. GORM's
  `default:true` zero-value behavior in repository `Create` erased an explicit
  false. The included/excluded groups and total therefore could not be correct
  for users opting a wallet out of the total.
- **Decision:** preserve the existing request/schema/default contract and make
  repository creation explicitly persist all supplied wallet fields, including
  false. Add a backend regression and rerun the same API-backed UI test. No
  migration or public API change is intended.
- **Result:** the isolated PostgreSQL regression reproduced the bug before the
  fix. `WalletPostgresRepository.Create` now persists `is_in_total` explicitly
  within the same transaction, preserving the controller's default-true
  behavior when the request omits the field. The API-backed iPhone 16 test now
  verifies the included/excluded groups, exact included-only total, aggregate
  route to the unscoped ledger, edit mode, confirmed API deletion, add and
  disabled-link rows, and retained Chi tiết route. Screenshot:
  `/tmp/mypocket-wallet-management-delete-attachments/46DF2A08-454E-45E5-BE0C-8167142E8351.png`.
- **Verification:** backend `go test ./...` passes with
  `TEST_DATABASE_URL=mypocket_stage_v1`; native unit/contract tests pass
  130/130; API-backed wallet-management UI passes 1/1; Release
  warnings-as-errors and project/base-component/diff guards pass. The first
  rendered list screenshot caught narrow-column wrapping; moving Chi tiết to
  its own row fixed it and the inspected rerun is legible. Stage fixtures are
  zero after teardown, temporary API 8081 is stopped, and API 8080 is untouched.
- **Documentation scope:** update internal wallet behavior/base/backlog,
  validation, port audit, context, ticket and changelog docs. No OpenAPI,
  public endpoint, schema, migration, ERD or configuration surface changed, so
  public API and AI-readable endpoint references do not need edits.

## Wallet selection-row Stitch parity — 2026-10-08

- **Observed mismatch:** the Stitch/web wallet rows stack the title above the
  current balance and expose a check for the active scope. Native had a
  horizontal management row, generic wallet-type subtitle and no selected
  check, including on “Tổng cộng”.
- **Change:** `MPWalletSelectionList` now composes a shared native selection row
  with title/balance stacking, aligned leading edges and semantic checkmarks.
  Aggregate remains selected on entry; selecting a wallet updates the local
  selection before invoking the same existing scoped-ledger callback. Edit,
  goal-progress, credit-availability and Chi tiết paths are retained.
- **RED/GREEN:** API-backed iPhone 16 UI regression failed on missing layout
  IDs/checkmark, then passed 1/1 after the base change. Screenshot:
  `/tmp/mypocket-wallet-rows-final-attachments/4F92374E-3F0E-49FF-A097-675913C3EB34.png`.
  The fixture test asserts aggregate check, vertical title/balance order and
  aligned edges for aggregate and wallet rows, as well as prior CRUD/route
  assertions.
- **Final verification:** native unit/contract tests pass 130/130 at
  `/tmp/mypocket-wallet-rows-unit.xcresult`; Release warnings-as-errors build
  passes at `/tmp/mypocket-wallet-rows-release-final`; project and both
  base-component guards pass. Stage fixture wallet count is 0 after teardown;
  temporary API 8081 is stopped, and API 8080 remains untouched.
- **Density follow-up:** the first post-fix screenshot showed the preserved
  native `Chi tiết` action taking a separate row and increasing card height.
  Its action is now a compact trailing `MPTextLinkButton` aligned with the
  selection row. API-backed UI asserts the two controls share a centerline;
  RED/GREEN results are `/tmp/mypocket-wallet-detail-row-red.xcresult` and
  `/tmp/mypocket-wallet-detail-row-green.xcresult`. Final captured list:
  `/tmp/mypocket-wallet-detail-row-attachments/1DD156EC-692D-4DC9-8CCE-B541B2E74EEF.png`.
- **Post-density gates:** native unit/contract 130/130 at
  `/tmp/mypocket-wallet-density-unit.xcresult`; Release warnings-as-errors at
  `/tmp/mypocket-wallet-density-release`; project/base/diff guards pass.
- **Scope:** presentation and local selection state only. No API, schema,
  migration, backend, owner-isolation or balance-calculation change. Public
  endpoint docs and AI-agent API catalog are unchanged by design.
- **Still open:** the screen adds MyPocket's persistent shell and Chi tiết
  extension beyond the Stitch capture; full pixel approval, wallet glyph
  artwork parity, populated goal/credit visual review and device UAT remain.

## Wallet aggregate glyph parity — 2026-10-08

- **Observed mismatch:** “Tổng cộng” reused the orange basic-wallet glyph,
  while the approved Stitch/web contract uses a globe IconBadge in the
  categoryTeal tone.
- **Change:** added the shared `MPWalletAggregateGlyph` base and replaced only
  the aggregate icon. It uses the web teal foreground, soft surface and border;
  individual wallet glyphs and aggregate calculations are unchanged.
- **RED/GREEN:** API-backed iPhone 16 UI regression failed on the absent
  aggregate-glyph semantic identifier and passed 1/1 after the fix. Captured
  screenshot: `/tmp/mypocket-total-glyph-attachments/3D11766A-DDD7-459D-BC11-C16FA65867C9.png`.
- **Scope:** native visual/accessibility contract only; public API and
  AI-readable endpoint references do not change. No financial data behavior
  was changed.

## Category editor route/layout parity — 2026-10-08

- **Historical observed gap, now resolved:** web exposes `/account/groups/new` and
  `/account/groups/:categoryId/edit` as dedicated editor routes. The Stitch
  edit-group screen also has its own title/header, metadata card, system-group
  explanation and wallet applicability card. At the time of this audit, native
  embedded its create/edit form above the category tree and used a menu for
  wallet applicability. The dedicated editor and visible wallet rows have
  since been implemented and verified; do not treat this historical finding
  as an open source-code gap.
- **Constraints:** preserve category CRUD, system-metadata read-only policy,
  owner-scoped wallet applicability, parent-cycle validation, shell navigation
  and the existing public API payload contract. No endpoint/schema/migration
  change is indicated.
- **Implemented direction:** make create/edit a dedicated native destination and
  render applicable wallets as visible shared-base rows with selection
  controls, matching the web composition and Stitch hierarchy. This replaces
  the native-inline divergence documented in the account-groups screen spec.
- **Approval:** the owner's standing instruction to complete the 100% native
  port, now reaffirmed by the request to continue the goal and update backend
  contracts if needed, authorizes matching this established web/Stitch route
  and form interaction. Implement the bounded native navigation/presentation
  change under
  [NATIVE-CATEGORY-EDITOR-ROUTE](tickets/NATIVE-CATEGORY-EDITOR-ROUTE-DETAIL_DESIGN.md);
  preserve current API and finance rules.
- **Current evidence:** web routes in `app/src/router.tsx`; web composition in
  `app/src/atomic/organisms/GroupEditorPage.tsx` and
  `app/src/atomic/molecules/CategoryEditForm.tsx`; native dedicated editor in
  `CategoriesView`. `LaunchTests.testAPIBackedCategoryCreateUpdateDeleteFromNativeUI`
  asserts the editor route, visible wallet choices, scope readback, filters,
  rename, archive and delete. Focused proof is recorded in
  `docs/work/VALIDATION_MATRIX.md` under “Native category editor route, wallet
  scope and wire contract”. Reference capture:
  `/tmp/stitch_my_pocket/stitch_my_pocket/s_a_nh_m_edit_category/screen.png`.
- **Result:** native create/edit now uses the dedicated editor route and shared
  visible wallet checkbox card. API-backed iPhone 16 CRUD passes 1/1 at
  `/tmp/mypocket-category-editor-final/Logs/Test/Test-MyPocketIOS-2026.10.08_10-45-32-+0700.xcresult`.
  The test exposed `walletIDs` serializing to `wallet_i_ds`; the Go handler
  expected `wallet_ids`, so create silently lost its wallet scope. Native now
  pins the exact wire key and XCTest asserts it. The backend stores the category
  and validated wallet links atomically. No public contract/schema/migration
  changed. Focused proof and cleanup are in `VALIDATION_MATRIX.md`.

### Additional category-list and picker parity findings — 2026-10-08

- **Category list — implemented in this slice:** native `CategoriesView` now
  exposes Khoản chi / Khoản thu / Vay-Nợ, wallet scope, search and the shared
  root/child tree using `MPCategoryTree`; local filtering preserves applicable
  child ancestry. The category API already returns `wallet_ids`, so no backend
  field or migration was necessary. The web page still lacks list wallet scope
  and search. Stitch `img_7793` also shows an inactive-group toggle, but neither
  client has a reliable active-state contract.
- **Transaction and budget category pickers:** web `CategoryTreeSelector`
  provides localized search and grouped root/child selection, and is used by
  `TransactionFields` and `BudgetEditorForm`. Native forms now match the
  searchable hierarchy, per-root cards, selectable system/personal copy and
  clear action through shared `MPCategoryPicker`,
  `MPCategorySelectionTree` and `MPButton.ghost` bases. Latest behavior and
  accessibility proof is in the parity follow-up above and
  `NATIVE-CATEGORY-PICKER-PARITY-DETAIL_DESIGN.md`.
- **Remaining parity:** native transaction and budget category selectors now
  compose `MPCategoryPicker`/`MPCategorySelectionTree`: transaction type and
  wallet eligibility determine selectable IDs, while budget clear preserves
  the existing nil `category_id` contract. API-backed iPhone UI verifies each
  selected ID in a fresh API read. Keep current CRUD and owner-scoped API
  behavior; do not imply unsupported debt-transaction writes. The inactive
  toggle needs a separate product/API decision because categories expose no
  active/usage state. The Stitch list's wallet/search controls are present on
  native; web does not currently render those extra filters, so they are a
  native reference-driven extension rather than a missing web API capability.
- **Evidence:** `app/src/atomic/organisms/GroupManagementPanel.tsx`,
  `app/src/atomic/molecules/BaseCategoryTree.tsx`,
  `app/src/atomic/molecules/CategoryTreeSelector.tsx`,
  `app/src/atomic/molecules/TransactionFields.tsx`,
  `app/src/atomic/molecules/BudgetEditorForm.tsx`, and native
  `CategoriesView`, transaction editor and `BudgetsView` in
  `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift` and
  `Features/Transactions/TransactionsView.swift`. The category-list capture is
  `/tmp/stitch_my_pocket/stitch_my_pocket/img_7793.png/screen.png`.

- **Verification:** local API-backed create/update/delete plus kind, search and
  wallet-filter UI proof passes on iPhone 16 Simulator. The first rerun failed
  only because the test cleared search and then expected a newly created row in
  the accessibility tree while `MPCategoryTree`'s `LazyVStack` had not
  materialized the offscreen row. Keeping the row filtered through edit/delete
  makes the test deterministic; no product/backend fix was indicated.
- **Picker verification:** transaction category and budget category-scope
  selection/persistence pass 2/2; recurring selection and generated ledger
  category pass 1/1 on the same simulator against the isolated stage API. No
  backend change is required; current category and mutation contracts already
  carry the parent and selected category IDs.
- **Visual/action parity RED:** the authenticated category capture showed the
  kind selector below the add button, trailing pencil/delete actions on list
  rows, and transaction-kind subtitles rather than wallet-scope text. The new
  API-backed UI regression failed on those expectations before the native fix.
- **Result:** native now orders kind → add → wallet/search → tree, makes each
  eligible row navigate to the dedicated editor, leaves delete in the editor,
  and derives the scope subtitle from existing `wallet_ids`. The shared
  `MPButton` now accepts an optional underlying-button accessibility ID so
  row navigation can be tested stably. No backend/API/schema/migration change
  was needed; current API scope data is sufficient.
- **Evidence:** focused iPhone 16 API UI test passes 1/1 at
  `/tmp/mypocket-category-list-parity-green2.xcresult`; the 138/138 native
  unit/contract suite passes at
  `/tmp/mypocket-category-list-parity-unit.xcresult`; Release
  warnings-as-errors, project/base-component guards and `git diff --check`
  pass. Captures are in
  `/tmp/mypocket-category-list-parity-green2-attachments/`.
- **Superseded 2026-10-08:** category archive semantics and the API contract are
  now defined in `docs/superpowers/specs/2026-10-08-category-archive-design.md`
  and implemented across backend, web and native. The unrelated global
  feedback-bubble overlap in the long-list screenshot remains tracked in
  [NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION](bugs/NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md).

### Native/backend contract inventory — 2026-10-08

- The owner authorized backend/API changes when required to complete the
  native port. This is implementation scope authorization, not a decision to
  change financial-data semantics.
- Compared the native `FinanceAPI` calls with the registered Go routes for
  auth/profile, wallets, categories, transactions/transfers, budgets, months,
  reports, jars, recurring, travel, portfolio, credit, AI, feedback and
  notifications. No missing route or HTTP-method mismatch was found in this
  pass. Native transfer create/update/delete supplies an idempotency key, and
  the backend implements the corresponding replay-safe mutation contract.
- **Superseded 2026-10-08:** category active/archive state is now defined and
  implemented locally across backend, web and native, including the reversible
  owner-scoped status API, history-preserving behavior and active-only new
  assignments. See the linked category archive design, plan and validation
  evidence; production migration/deployment remains intentionally unperformed.
- Impact audit: category assignment is validated in ordinary transaction,
  credit-entry, budget, recurring-schedule and AI-proposal flows; reports,
  month summaries and transaction details also resolve historical category
  data. A correct archive contract must filter creation choices without
  erasing history, and define whether already-configured recurring schedules
  and budgets remain valid after a category is archived. This pass made no
  runtime change.

## Full UI suite wallet-editor fixture — 2026-10-08

- **Observed failure:** the default non-`LOCAL_API_UI` UI suite attempted to
  open Wallet Management's add editor even though the authenticated shell had
  no API-backed wallet list; the screen correctly showed its retryable network
  error instead. This made `testWalletEditorReservesShellInset` fail before it
  could inspect the editor.
- **Intended behavior:** classify this test as an API-fixture test, and skip it
  in the default UI configuration with the same explicit fixture requirement
  used by the other API-backed editor checks. Run it in the isolated
  `LOCAL_API_UI` suite where Wallet Management can load and the editor can
  actually be reached.
- **Verification plan:** add the compile-time skip, run the focused default UI
  test to verify it reports skipped (not failed), then retain API-backed wallet
  editor and grouped-wallet proofs as the actual interaction coverage. Rerun
  both native unit/contract and build/base guards. No production behavior or
  public API contract changes.
- **Result:** the focused test reports `Skipped` 1/1 without compile warnings
  at `/tmp/mypocket-wallet-editor-skip-final.xcresult`. The complete default
  UI suite passes at `/tmp/mypocket-wallet-glyph-default-ui-fixed.xcresult`:
  20 passed, 31 fixture-dependent skipped, 0 failed. API-backed editor and
  grouped-wallet proofs remain the actual wallet interaction evidence. No
  production code, API, or data was changed.
- **Final gates:** native unit/contract tests pass 130/130 at
  `/tmp/mypocket-wallet-glyph-final-unit2.xcresult`; Release warnings-as-errors
  build passes at `/tmp/mypocket-wallet-glyph-final-release`; project
  verification, both base-component guards/tests, and `git diff --check` pass.

## Credit wallet action presentation follow-up — 2026-10-08

- **Observed mismatch:** web `CreditWalletPanel` opens add-entry and payment
  forms in `BaseBottomSheet`; native `CreditView` inserts the form inline after
  the statement, making it scroll-dependent and visually unlike the web flow.
- **Intended behavior:** tapping either credit action opens a dismissible,
  shared-base native form sheet with the same fields, validation and primary
  save action. Cancel dismisses without writing. Successful writes continue to
  use the existing owner-scoped credit APIs and refresh the statement; failure
  retains the open form and its draft.
- **Technical decision:** add/reuse one DesignSystem form-sheet presentation
  base and compose it with the existing reference header, amount field,
  selectors, status and button bases. No API, persistence, permission, or
  finance-rule changes.
- **Verification plan:** add a failing API-backed UI assertion that the form is
  presented as a sheet and can be canceled, run it red before implementation,
  then green against the disposable local API; run focused native tests,
  Release warnings-as-errors and project/base/diff guards.
- **Result:** the red UI test reproduced the inline-form mismatch. `MPFormBottomSheet`
  now owns the shared reference header, scrollable field area, pinned footer,
  large native sheet, canvas and dismissal lock. Both entry and payment actions
  use it; success still closes and refreshes, while validation/provider errors
  leave the form available. The screenshot is
  `/tmp/mypocket-credit-sheet-attachments-green/01BF787C-CB37-4C43-BDB1-9C9357671A7B.png`.
- **Verification:** focused sheet route 1/1 and actual entry+payment statement
  mutation 1/1 pass on the iPhone 16 simulator. Native unit/contract tests pass
  128/128, Release warnings-as-errors and project/base/diff guards pass. The
  isolated stage fixture wallets and their transaction rows are both 0 after
  cleanup; temporary API 8081 is stopped, and port 8080 was not used.
- **Documentation scope:** this audit and the wallet design contract are
  internal implementation docs; public API docs and machine-readable API
  surfaces do not change because the request/payload contract is unchanged.

## Native credit-entry category tree — 2026-10-08

- **Gap:** the dedicated credit entry sheet used a flat `Menu` for category
  selection, unlike the shared transaction/budget/recurring hierarchy. It also
  provided no search or parent/child context.
- **Change:** the “Nhóm” row now opens the shared `MPCategoryPicker` inside the
  same sheet (no nested presentation); selecting a category returns to the
  preserved entry kind, amount and note. Purchase/fee/interest use applicable
  expense groups; refunds use applicable income groups. Payment is unchanged.
- **Backend decision:** verified the existing `/categories` response and
  credit-entry validation already supply and enforce owner-scoped category
  hierarchy, kind and wallet scope. No backend, API, migration or permissions
  change was needed.
- **RED/GREEN:** the focused API-backed iPhone test failed before the picker
  existed, then passed after implementation. It selects an eligible category,
  creates the credit entry, reads the statement through the fixture owner API
  and compares the persisted category ID; the same test also creates a payment.
- **Verification:** native unit/contract tests pass 137/137, backend `go test
  ./...` passes against isolated local stage, Release warnings-as-errors passes,
  and project/base/diff guards pass. Detailed result bundles and fixture
  cleanup are recorded in the validation matrix. No production database was
  used.

## Native transaction amount follow-up — 2026-10-08

- Comparison against Stitch `image.png_2` found a real gap: the iOS editor
  previously placed VND after a bordered 48pt input and rendered raw digits.
- Added a transaction-only `MPAmountField` presentation with a leading VND
  chip, large unboxed input and `vi-VN` grouping in the display layer. The
  stored binding and API payload remain integer digits. Other amount consumers
  keep their standard control style.
- RED/GREEN UI proof: the new VND-position assertion failed before the change
  (`VND.midX=292.17`, input minX=36.00) and passes after. Display assertion
  confirms `1.500.000`; the focused iPhone 16 editor flow passes 1/1 and the
  screenshot is `/tmp/mypocket-amount-layout-final-attachments/`.
- The global segmented-control base was also verified to use its documented
  white selected pill; an older screenshot showing the green selected state
  was stale and not a current defect.
- Native tests/build/guards are recorded in the validation matrix. Full
  cross-screen visual/UAT sign-off remains open; this one-screen result does
  not close the 100% port audit.

## Native ordinary-ledger wallet selector follow-up — 2026-10-08

- **Observed gap:** the web ordinary ledger places a compact, centered wallet
  capsule directly below the screen header. Native ordinary ledger currently
  renders the generic full-width `MPOptionSelector` form row (`Ví: …`) instead.
- **Intended behavior:** aggregate and basic-wallet scopes place the same
  centered capsule anatomy as the web selector immediately below the screen
  header, before period controls. Wallet identity glyph/name and aggregate
  label remain visible, and all existing scope choices continue to select the
  same owner-scoped ledger. Goal and credit retain their specialized bodies.
- **Technical decision:** add a DesignSystem wallet-scope selector composed
  from the existing `MPMenuSelector`, selector-specific `MPWalletScopeGlyph`,
  `MPText`, and shared tokens. Keep selection and ledger filtering unchanged;
  no API, schema, persistence, permission, or finance-rule changes.
- **Verification plan:** add an API-backed UI assertion for the compact
  selector accessibility contract and relative order, rerun the wallet-scope
  route test, native unit/contract tests, Release warnings-as-errors,
  project/base guards, and `git diff --check`. Capture the selected basic-wallet
  state for visual inspection.
- **Result:** aggregate/basic ledger now uses the centered capsule with
  selector-specific neutral/goal/credit glyph tones. The existing wallet
  selection and period/filter state are untouched. The selector now renders
  before the period controls, matching the web hierarchy. The API-backed route
  assertion targets the stable selector ID, exact selected-wallet label, and
  relative position before the week control. The API-backed iPhone 16 test
  passes 1/1 at
  `/tmp/mypocket-wallet-scope-api-ui/Logs/Test/Test-MyPocketIOS-2026.10.08_03-41-37-+0700.xcresult`;
  its screenshot is
  `/tmp/mypocket-wallet-scope-attachments/B8303648-1B89-4E52-A027-F4269B21D189.png`.
  Native unit/contract passes 129/129, Release warnings-as-errors and
  project/base/diff guards pass. The test used only `mypocket_stage_v1`, Redis
  DB 15, and loopback API 8081; cleanup confirmed zero fixture wallets and
  transactions. API 8081 was stopped afterward and the unrelated API on 8080
  remains running and untouched.
- **Documentation scope:** transaction screen design and this port audit are
  updated. Public API docs and machine-readable API surfaces remain unchanged
  because no endpoint or request/response contract changed.

Latest automated checkpoint (auth parity rerun 2026-10-07): the native
unit/contract target passes 127/127; the corresponding Release
warnings-as-errors build also
passes. The transaction day-group, category glyph, ended-budget preview,
wallet-card, segmented-control, transaction-editor icon, category-icon picker,
integer-safe amount-keypad, shared calendar-picker, budget-date parity and
account-timezone trend aggregation tests are included in this count; the Money
Insider comparison-ring contract is also included. Result evidence is
`/tmp/mypocket-auth-unit-rerun.xcresult` and the Release build derived data is
`/tmp/mypocket-auth-release`. The prior 129-test checkpoint preceded
removal of two obsolete Google-auth tests.

The bounded disposable/API UI bundles and the release warnings-as-errors build
also pass. The Xcode 27 runner can hang when all stateful UI cases share one
invocation, so stateful cases remain intentionally rerun in bounded groups.
Project/base-component/diff guards, backend race tests with the isolated local
stage database, web typecheck/build/design checks all pass. The matrix below
deliberately keeps visual UAT and live API/provider evidence open; passing a
route smoke test is not treated as pixel or production proof.

The current live provider checkpoint closes the text/tool/schema portion of the
AI gate against the supplied FPT-compatible endpoint and `DeepSeek-V4-Flash`:
text extraction, one strict tool-call probe and five strict JSON Schema cases
all pass without writes. OCR/S3 image UAT remains open because this worktree
does not contain an OCR credential; the earlier synthetic OCR pipeline proof is
retained in the validation matrix.

The transaction-ledger follow-up fixed bulk-delete request ownership: native
now sends one sorted, de-duplicated `/transactions/bulk-delete` request and
updates local rows only after success. The focused regression passed 1/1 at
`/tmp/mypocket-bulk-delete-green/Logs/Test/`; all project/base-component/diff
guards passed. The all-target Xcode invocation was stopped after the known
UI-runner hang, so it is not used as evidence.

The latest disposable local API UI sweep also passed independently for wallet
scope routing, credit wallet actions, portfolio asset/price refresh, atomic
transfer pair mutation, credit entry/payment statement refresh, and wallet
detail transaction create/update. The AI entry capability/attachment states
also pass in isolated local-fixture runs. Exact result bundles are recorded in
`docs/work/VALIDATION_MATRIX.md`; these are
behavior evidence only and do not close authenticated device/production visual
UAT.

The latest authenticated visual capture gate now includes More/account and the
transaction editor. The editor mismatch found by inspecting the fixture-backed
screenshot was actionable: it rendered under the shell bottom navigation. Both
ledger and wallet-detail editor entry points now use a full-screen presentation;
the shared reference header owns the Stitch close capsule and the shared footer
image action routes to the existing AI-entry screen. Focused editor UI and local
API fixture UI checks pass 1/1 each. Evidence and exact paths are recorded in
`docs/work/VALIDATION_MATRIX.md` and the visual-capture plan. Physical-device
pixel approval and live provider/OCR/production feedback remain open.

The latest visual sweep also fixed a shared base-input defect: the UIKit-backed
email control could expand vertically because its representable had no bounded
height. `MPTextField` now constrains that bridge to the shared 48px control
height, restoring the Stitch/web form rhythm for authentication and every
other screen that uses the base input.

### Authenticated secondary-route visual evidence — 2026-10-08

- **Gap:** `testAuthenticatedMoreMenuOpensEveryDestination` asserted that all
  15 secondary routes opened and returned, but captured no screenshot for most
  destinations. This proved reachability, not inspectable screen layout.
- **Plan:** attach a named screenshot after each route's title and shell
  clearance assertions, before navigating back. In `LOCAL_API_UI` builds use
  the isolated stage API on port 8081, never the unrelated port 8080 or
  production. No ledger mutations are part of this route sweep.
- **Evidence boundary:** these API-backed screenshots can prove rendered
  owner-scoped empty/populated screen state. They do not prove pixel equivalence
  without a matched Stitch/web comparison or owner/device acceptance.
- **Verification:** run the focused iPhone 16 simulator route sweep, export and
  inspect all 15 attachments, then reconcile this audit, validation matrix and
  visual-capture plan. Preserve the existing broad test/build gates.

The latest Auth visual comparison now matches the web/Stitch anatomy: a shared
64pt circular `MPBrandMark`, direct canvas title/subtitle, flat shared fields
and full-width email/password sign-in. Native auth no longer exposes Google
login or carries an OAuth callback configuration. The launch contract checks
that the Google action and divider are absent; authenticated device sign-in and
pixel-level cross-screen UAT remain open.

An additional authenticated Overview launch using the local disposable API
fixture produced `/tmp/mypocket-overview-api.png`; the owner-scoped balance,
wallet, Money Insider, reports and persistent shell all render in one native
frame. After wallet-glyph parity, the same fixture was captured at
`/tmp/mypocket-overview-wallet-glyph.png`. This is anatomy/API evidence only:
populated multi-wallet icon artwork, real provider/OCR and full pixel sign-off
remain open.

The Overview visual pass also aligned wallet identity badges with the web
contract: `MPWalletGlyph` now owns 48pt geometry and basic/goal/credit semantic
tones (warning/success/teal). The wallet API, labels, balances and row scope
are unchanged; focused/full native tests, Release and web guards pass.

The More/profile parity pass also exposes the reference/web `Quản lý tài khoản`
action directly in the native profile card. It reuses `MPMenuRow` and routes to
the existing account settings screen; no new backend contract is required.
The adjacent `Trợ giúp` action now routes to the existing feedback/changelog
surface instead of being a no-op, also without a backend contract change.
The overview header bell is wired through the shell to the existing
owner-scoped notifications screen; the search icon remains intentionally
presentation-only because no web/API search contract exists.

The earlier post-input-fix unit result is
`/tmp/mypocket-auth-input-tests-final/Logs/Test/Test-MyPocketIOS-2026.10.07_09-52-13-+0700.xcresult`
(82/82, no failures or skips); the unauthenticated visual launch proof is
`/tmp/mypocket-auth-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_09-51-00-+0700.xcresult`
(1/1).

The profile-card parity proof is
`/tmp/mypocket-profile-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-14-12-+0700.xcresult`:
the native More profile exposes `more-profile-account` and opens the existing
account settings route (1/1 UI test).

The profile help action now has its own base-button route proof; it opens the
existing feedback/changelog screen instead of being a no-op.

The transaction-editor parity slice now composes shared `MPFormRow` and
`MPDateStepper` bases for wallet/category/note/date controls inside the Stitch
reference surface. The editor keeps the existing owner-scoped API payload,
category/jar validation, account timezone serialization and safe-area save
footer. Unit coverage is 84/84 and the focused editor UI contract is 1/1;
pixel-level authenticated device sign-off remains open. The subsequent
Finance Assistant structured-card parity slice raises the current unit gate to
86/86 and adds a focused Assistant UI smoke (1/1). Its capability-policy
follow-up raises the gate to 87/87 and keeps composer/send disabled until the
provider is enabled; provider-backed send/history and authenticated visual
sign-off remain open. The disposable local API Assistant capability/overview
smoke is 1/1 at `/tmp/mypocket-assistant-policy-api-ui`. The native goal
progress renderer now also matches the web card's clamped percentage contract;
its post-change full unit gate is 88/88. The follow-up jar-progress resolver now
reads the backend `allocation_amount` contract and omits a misleading bar when
no valid target exists; the latest full unit gate is 89/89. The metadata
follow-up adds web-compatible source/count/period context; the latest full
unit gate is 90/90. The privacy-mask follow-up now makes the shell-owned
persisted mask state flow into Overview and Finance Assistant; card money,
summary count metrics and free-form assistant text are redacted consistently.
The latest full native unit gate is 91/91, the focused and API-backed Assistant
UI smokes are 1/1 each, and the Release warnings-as-errors, project,
base-component and diff guards pass. Authenticated visual comparison and
provider-backed Assistant send/history remain external gates. The subsequent
secondary-screen mask routing raises the native unit gate to 92/92 and covers
the web-masked budgets, reports, jars and month-detail surfaces. The latest
Reports follow-up adds `MPReportsInsiderCard`, so the native Reports route now
renders the full web Money Insider result (period delta, daily average,
spending/income ratio, top categories and top expenses) with the same mask
boundary. The latest native unit gate is 93/93 and the API-backed Reports UI
smoke is 1/1.

The Budgets parity follow-up now matches the web summary and status behavior:
native renders overall spent/limit/remaining/progress, active daily-spend
guidance, ended/future timing and 75%/over-budget warnings through shared base
components. The focused contract is 23/23, the API-backed CRUD/status UI smoke
is 1/1, and the latest full native unit gate is 94/94. Release
warnings-as-errors, project/base-component and diff guards pass. No API or
database migration changed; authenticated visual UAT remains an external gate.

The AI Entry parity follow-up adds the missing native header-level `Lưu tất cả`
action. It selects only pending proposals and approves them sequentially via
the existing versioned owner-scoped endpoint, retaining partial statuses and
surfacing the first error. The focused unit contract is 24/24, the pending
review UI smoke is 1/1, and the latest full native unit gate is 95/95. Release
warnings-as-errors, project/base-component and diff guards pass; no backend or
schema change was made.

The Overview recent-title parity follow-up now loads categories as optional
metadata and follows the web category-first rule (`category.name`, then
`Khoản thu`/`Khoản chi`). Notes and wallet metadata remain unchanged. The
focused regression is 1/1 and the latest full native unit gate is 96/96;
Release warnings-as-errors, project/base-component and diff guards pass.

The Overview wallet subtitle follow-up extends shared `MPWalletRow` with an
optional subtitle and feeds the web-equivalent `description ?? type label`
rule from Overview only. Other consumers keep their existing defaults. The
focused regression is 1/1 and the latest full native unit gate is 97/97;
Release warnings-as-errors and all guards pass.

The Overview wallet-action follow-up now uses the web-equivalent `Quản lý`
label on the shared `MPSectionHeader` while preserving the existing wallet
management callback. The focused regression is 1/1 and the latest full native
unit gate is 99/99; Release warnings-as-errors and all guards pass. No API,
schema, route or owner-scope behavior changed.

The Reports month-action follow-up now uses the selected month as the trailing
label on the `Tổng quan báo cáo` summary card, matching the web report panel;
the explicit `Xem báo cáo` control remains responsible for reload. The focused
copy contract is 1/1 and the latest full native unit gate is 100/100;
Release warnings-as-errors and all guards pass. No API, schema, route or
owner-scope behavior changed.

The Budgets empty-structure follow-up now withholds the overall summary card
when the backend returns no budget items and shows only the shared web-equivalent
`Chưa có ngân sách.` empty state. The pure visibility contract is 1/1 and the
latest full native unit gate is 101/101; Release warnings-as-errors and all
guards pass. No API, schema, route or owner-scope behavior changed.

### Global add long-press regression investigation — 2026-10-08

- Reproduced the isolated UI failure in
  `/tmp/mypocket-longpress-investigation.xcresult`: long press opens AI Entry
  and Back returns to the Account (`Mở rộng`) tab; the subsequent short-tap
  assertion then fails.
- The first failure attachment showed the add action is intentionally disabled
  on Account. After correcting the test to select Transactions before the
  short-tap assertion, the test still failed: the post-tap hierarchy continued
  to report Account selected. This exposed a product navigation defect, not
  merely a stale test sequence.
- Root cause: `MainTabView` applies `.allowsHitTesting(false)` to the entire
  bottom navigation while Account is selected. `MoreView` already reserves a
  shared `MPBottomNavigationInset`; the additional shell-wide hit-test block
  therefore prevents users from leaving Account via Overview/Transactions/
  Assistant tabs after opening a More destination.
- Intended behavior: keep the shared tab bar interactive on Account, retain
  the disabled center-add affordance there, and retain the existing narrow
  blocker over only the Account tab to prevent taps falling through to the
  final More row. The regression test must prove long-press → AI Entry → Back →
  Transactions tab → short-tap → transaction editor.
- PASS: `MainTabView` keeps shell tab hit-testing enabled on Account while
  preserving the disabled center-add affordance and narrow blocker over the
  selected Account tab. `MoreView` retains the shared bottom-navigation inset.
- PASS: the corrected long-press/return/tab/short-tap flow, the normal
  transaction add form and all 15 More destinations pass together on the
  iPhone 16 Simulator, 3/3, in
  `/tmp/mypocket-longpress-fixed.xcresult`.
- SCOPE: no API, schema, permission or finance behavior changed. Release,
  project/base guards, full native unit suite and diff verification all pass:
  127/127 in `/tmp/mypocket-shell-navigation-unit.xcresult`, Release
  warnings-as-errors at `/tmp/mypocket-shell-navigation-release`, and
  `verify-project.sh`, `check-base-components.test.sh`, and `git diff --check`.

The Categories action follow-up now matches the web Groups route: the native
header is navigation-only and the create action is a full-width outline
`MPButton` labeled `Nhóm mới` below it. The outline geometry is centralized in
the shared button base and exposed as `category-add-group`; category CRUD/tree,
owner scope and API payloads are unchanged. The latest native unit checkpoint
is 102/102, the Release warnings-as-errors build and guards pass. Authenticated
visual comparison of the full category tree remains an external gate.

The Budgets action follow-up now matches the web `BudgetsPanel` header: native
uses the shared compact `MPButton` with copy `Thêm` and stable ID `budget-add`
instead of a screen-local icon-only action. The focused contract, 103/103 full
native unit suite, Release build and base/project/diff guards pass; budget CRUD,
validation, owner scope and API behavior remain unchanged. Authenticated visual
comparison remains open.

The Jars action follow-up now matches the web monthly section: the native page
header is navigation-only and the shared compact `MPButton` `Thêm hũ` sits beside
`Hũ tháng` with stable ID `jar-add-month`. The focused contract, 104/104 full
native unit suite, Release build and base/project/diff guards pass; jar month,
cumulative, CRUD, validation and owner-scope behavior remain unchanged.
Authenticated visual comparison remains open.

The Recurring action follow-up now matches the web summary card: the native page
header is navigation-only and the shared primary `MPButton` `Thêm` sits beside
`Giao dịch định kỳ` with stable ID `recurring-add`. The API-backed create,
update, delete and run-due UI flow now passes 1/1 at
`/tmp/mypocket-recurring-ui-final2.xcresult`; the stable ID is asserted with
the current `Thêm` label instead of the stale `Thêm lịch` lookup. The 127/127
native unit/contract suite, Release build and base/project/diff guards pass;
run-due, schedule CRUD, validation, timezone and owner-scope behavior remain
unchanged. Authenticated visual comparison remains open.

The Travel Mode action follow-up now matches the web summary card: the native
page header is navigation-only and the shared primary `MPButton` `Thêm` sits
beside `Travel Mode` with stable ID `travel-add`. The focused contract, 106/106
full native unit suite, Release build and base/project/diff guards pass; event
CRUD, active state, validation and owner-scope behavior remain unchanged.
Authenticated visual comparison remains open.

The Portfolio action-anatomy follow-up now matches the web panel: the header
keeps only the shared `Làm mới` action, and loaded summaries expose shared-base
`Thêm tài sản` and `Ghi giao dịch` cards. Per-asset `Giao dịch` selects the
shared trade card; no portfolio API, validation, timezone or owner-scope
contract changed. The focused contract and full native unit target pass 1/1 and
107/107 respectively, Release/guards/web checks pass, and the bounded launch
smoke is green. Authenticated disposable API mutation and pixel-level visual
comparison remain open.

The Credit ledger summary follow-up now matches the web panel's anatomy: one
shared reference card presents current debt and available credit as the primary
pair, followed by limit and payment status/due details. `Thêm giao dịch` and
`Thanh toán thẻ` use the web-equivalent shared actions and stable IDs; statement,
entry, payment, empty-state, timezone and owner-scope contracts are unchanged.
The focused contract is 1/1, the full native unit target is 108/108, and
Release/guard/web checks pass. The bounded authenticated credit API UI smoke
passes 1/1. A follow-up section-order contract also passes 1/1: actions now sit
before the statement and the empty message is plain text outside the list card;
authenticated visual UAT remains open.

The Portfolio summary follow-up now adds the web subtitle and wraps the
headline metrics in a shared `Tổng danh mục` reference card. Portfolio API,
trade/price validation, timezone and owner scope are unchanged. The focused
copy contract is 1/1, the full native unit target is 108/108 and Release/guard/
web checks pass; prior disposable portfolio mutation proof remains recorded,
while the latest rerun hit the known Xcode runner hang and visual UAT remains
open.

The Credit ledger empty-state follow-up keeps the shared `Giao dịch tín dụng`
card visible and renders the web-equivalent `Chưa có giao dịch tín dụng.`
message when the selected credit wallet has no statement rows. The pure copy
contract is 1/1 and the API-backed credit route asserts the empty state before
entry/payment actions. The latest native unit gate before this copy-only
follow-up was 98/98; no backend/API or
migration changed.

The earlier native Google sign-in implementation was superseded by the
password-only migration decision. The backend Google routes still serve the
web app; native Release contains no Google PKCE client or callback scheme.

## Privacy-mask parity boundary

The native shell persists `mypocket.privacy.masked` and passes the value to
Overview, Finance Assistant and the web-masked secondary surfaces: budgets,
reports, jars and month detail. Assistant money, summary count metrics, wallet
opening balances and secondary monetary labels render as `••••••`; free-form
assistant prose is replaced by the safe privacy message used by the web. The
transaction, recurring, travel, portfolio and account routes retain their
screen-specific contracts because the web shell does not pass this prop to
those panels. RED evidence is
`/tmp/mypocket-assistant-privacy-red`; full native unit
`/tmp/mypocket-assistant-privacy-unit` is 91/91, both Assistant UI smokes are
1/1, and Release/project/base/diff gates pass.

The latest native unit/contract result is
`/tmp/mypocket-final-unit.xcresult` (83/83, no failures or skips) on
`MyPocket-iPhone-16`.

The earlier Google-auth wiring smoke (`/tmp/mypocket-auth-final3.xcresult`) is
historical and its assertion that the Google action is present is superseded by
the password-only migration decision and current launch contract.

The base-component audit also found one remaining screen-level typography
override in the Overview metric and moved it to the centralized `MPFont` role,
so the native screen cannot drift from the Stitch/web type scale.

The page-margin parity pass also moved the shared native inset from 20pt to the
Stitch/web 16px contract. Because every affected screen already composes
`MPMetrics.pagePadding`, the correction stays centralized and does not add
screen-local spacing. The post-change unit gate is
`/tmp/mypocket-margin-unit.xcresult` (83/83), the unauthenticated launch smoke
is `/tmp/mypocket-margin-ui.xcresult` (1/1), and the Release warnings-as-errors
build is `/tmp/mypocket-margin-release`.

The action-shape parity pass now gives filled `MPButton` styles the shared
pill radius from the Stitch/web contract while leaving row and icon controls
unchanged. This remains a DesignSystem-only change. The follow-up unit gate is
`/tmp/mypocket-typography-unit.xcresult` (83/83), add/auth UI smoke is
`/tmp/mypocket-typography-ui.xcresult` (2/2), and the Release build is
`/tmp/mypocket-typography-release`. The convenience label initializer also now
inherits the base `MPFont.label` instead of forcing a system body font.

The base-component guard now rejects system-font modifiers in both feature
screens and DesignSystem bases; a negative fixture test proves the guard catches
`.font(.body)` before it can reintroduce typography drift.
The focused DesignSystem test bundle after the UIFont assertion is
`/tmp/mypocket-typography-final.xcresult` (12/12).

The final selector/date-field typography sweep also applies `MPFont.body` to
wallet, option, multi-option and compact date controls (including menu rows),
so no shared form control inherits a platform font by accident. The complete
post-sweep unit gate is `/tmp/mypocket-typography-complete.xcresult` (83/83),
the focused auth/add UI smoke is `/tmp/mypocket-typography-complete-ui.xcresult`
(2/2), and the Release warnings-as-errors build is
`/tmp/mypocket-typography-complete-release` (the final incremental rebuild is
`/tmp/mypocket-typography-final-release`).

The follow-up overview header parity pass keeps the amount and visibility
control together in the leading row, while search and notifications remain in
the trailing cluster as in the Stitch capture. Existing accessibility IDs and
the notification route are unchanged. The post-change unit gate is
`/tmp/mypocket-header-unit.xcresult` (83/83), header/notification UI smoke is
`/tmp/mypocket-header-ui.xcresult` (2/2), and the Release warnings-as-errors
build is `/tmp/mypocket-header-release`.

The More/profile parity pass now uses shared `MPProfileCard` and
`MPHeaderSupportButton` bases: centered identity/avatar, account subtitle and
the Stitch-style support pill. The existing account/help route checks remain
green: `/tmp/mypocket-more-profile-ui.xcresult` (2/2), with full native units
at `/tmp/mypocket-more-profile-unit.xcresult` (83/83) and Release build at
`/tmp/mypocket-more-profile-release`.

The full More destination sweep also remains green after the profile-card
recomposition: `/tmp/mypocket-more-profile-sweep.xcresult` reports 1/1, with
all 15 account-menu destinations opened and returned successfully.

The profile-card UI assertions now also verify the centered fixture identity,
account subtitle and shared support-pill identifier; the focused rerun is
`/tmp/mypocket-more-profile-ui2.xcresult` (2/2).

After that cleanup, the warnings-as-errors Debug build is
`/tmp/mypocket-base-audit-build` and the full native unit/contract result is
`/tmp/mypocket-base-audit-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_12-34-43-+0700.xcresult`
(83/83).

The post-routing native unit/contract result is
`/tmp/mypocket-profile-notification-unit/Logs/Test/Test-MyPocketIOS-2026.10.07_12-24-15-+0700.xcresult`
(82/82, no failures).

The fresh More-menu route sweep after that profile-card change passed all 15
secondary destinations and returned to the More screen after each route. The
proof bundle is
`/tmp/mypocket-more-sweep-profile/Logs/Test/Test-MyPocketIOS-2026.10.07_12-15-17-+0700.xcresult`
(1/1 UI test, 344.968 seconds).

The native shell visual sweep also changed the shared bottom navigation from a
full-width strip to the Stitch inset rounded surface with a centered elevated
add action. Primary-route and long-press add UI tests passed 2/2 after this
base-component change.

The scroll-content safe area is now also owned by the shared
`MPBottomNavigationInset` base component. Overview, More and secondary feature
screens use the same `safeAreaInset` spacer, so long content cannot finish
under the persistent shell navigation. A Debug simulator build and the
notification-route UI regression passed after this change.

The final hit-test pass fixed two compact-phone collisions: the wallet-detail
“Thêm giao dịch” CTA is a nested safe-area action with a 64px cushion above
the shell, and account-menu destination rows own hit-testing while the shell
navigation is disabled there. A narrow transparent guard covers only the
selected account-tab footprint, so a repeat tap cannot fall through to a
nearby More row while centered rows such as Hũ chi tiêu remain tappable. The
wallet detail editor and mutation UI cases, jar mutation UI case and the full
15-destination More route sweep pass after this change.

The local API health endpoint is reachable. Its normal configuration keeps the
Google fixture exchange disabled (`501 GOOGLE_MOBILE_LOGIN_UNAVAILABLE`), while
the explicitly enabled disposable `OAUTH_FIXTURE_MODE=true` run returns a
fixture session for authenticated UI tests. No migration force, database reset
or production change was used.

The earlier safe-area build is `/tmp/mypocket-safearea-build` and the targeted
route proof is
`/tmp/mypocket-safearea-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_12-27-39-+0700.xcresult`
(1/1). The warnings-as-errors Release build also passes at
`/tmp/mypocket-safearea-release`; the current Release build is
`/tmp/mypocket-final-release`. Project verification, base-component checks,
web design/type/build checks and `git diff --check` pass.

The fixture-backed reachability proof now scrolls the Overview to
`overview-section-recent-transactions` and asserts its frame stays above the
persistent shell navigation. Result:
`/tmp/mypocket-safearea-reachability/Logs/Test/Test-MyPocketIOS-2026.10.07_12-30-45-+0700.xcresult`
(1/1).

## Decision

The ZIP is a visual reference; the web/API contracts are the behavioral source
of truth. “Port 100%” means that every user-visible web capability has a native
entry point, uses the shared native base components, calls the owner-scoped
API, and has at least one automated interaction proof. A route that only opens
is not sufficient.

## Current parity matrix

| Capability | Native entry point | API surface | Interaction proof | Status |
| --- | --- | --- | --- | --- |
| Auth/password, refresh, Face ID | `AuthView`, `BiometricUnlockView` | login/session | email/password launch + refresh + biometric contracts | password-only native login; device sign-in UAT remains open |
| Overview, month summary, insider, notification entry | `OverviewView` | overview/month/insider/notifications | unit + shell UI + notification route + Stitch section-order contract | automated behavior/order proof; visual UAT open |
| Transaction ledger and editor | `TransactionsView`, `TransactionDetailView` | transactions/categories/jars | ledger, editor, transfer, adjustment, detail-overlay shell and exactly-once bulk-delete tests | automated mutation/detail-overlay/bulk-delete proof; visual UAT open |
| Wallet CRUD/detail | `WalletsView`, `WalletDetailView`, `WalletEditorView` | wallets/transactions | CRUD, detail editor, scoped ledger, editor section and shell-clearance contracts | automated proof including wallet-editor inset; visual UAT open |
| Goal wallet | `WalletDetailView` | wallet/ledger | explicit detail plus Stitch-style embedded savings ledger, selector/search/overflow, progress/advice/month grouping/add action | API-backed simulator visual capture PASS; authenticated device/browser UAT remains open |
| Credit wallet | `CreditView` | credit statement/entry/payment/categories | summary/card, route, statement, entry/payment and shared category-tree selection/persistence tests | automated behavior proof including persisted selected category; authenticated visual UAT open |
| Budgets | `BudgetsView` | budgets/wallets/categories | API-backed create/update/delete UI test plus account-timezone date-field contract | automated proof at `/tmp/mypocket-budget-date-ui2.xcresult`; run-time UAT open |
| Categories | `CategoriesView` | categories/wallets/status | API-backed create/update/archive/restore/delete UI test; active-state API/model contracts | archive/restore API/UI state assertions pass; the older selected run was also red on the now-resolved feedback-dock overlap, with the independent dock geometry regression green; populated-data/device visual UAT remains open |
| Jars | `JarsView` | jars/month report | create/update/remove UI test | automated proof; cumulative-report UAT open |
| Month detail/notes | `MonthsView` | month summary/note | save/delete UI test | automated proof; visual UAT open |
| Recurring | `RecurringView` | recurring/run-due | create/update/delete plus run-due API/UI state proof | automated API and ledger-state proof passes; visual UAT remains open |
| Travel Mode | `TravelView`, `TransactionDetailView` | travel events/link | create/toggle/update/delete; manual transaction link/unlink persisted-state test | event CRUD and transaction link/unlink are API-backed; active auto-link has backend contract proof; report narrative and authenticated visual/device UAT remain open |
| Portfolio | `PortfolioView` | portfolio | action/card contract + asset create/price + trade create/history UI tests; trade/history model + validator tests; owner-path/snake_case/idempotency API contract | native buy UI flow passes 1/1 at `/tmp/mypocket-portfolio-trade-ui4.xcresult`, including refreshed position/title and a fresh owner-scoped API assertion for side/quantity/price/fee; unit 127/127; production/device visual UAT remains open |
| Reports | `ReportsView` | insider/month | route + month filter/reload UI test + Stitch section-order contract | automated reference-card/filter/reload proof; API-backed Money Insider UI 1/1 at `/tmp/mypocket-reports-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-41-41-+0700.xcresult`; authenticated visual UAT open |
| Finance Assistant | `AssistantView`, `AssistantStructuredCard` | advisor conversation/tools | route + overview API test + empty-history suggestion UI test + structured-card vocabulary tests | overview API UI 1/1; DeepSeek provider-backed stage send/history E2E completed with synthetic prompt; persisted model usage keys verified and all fixture history cleaned; visual and production UAT remain open |
| AI transaction entry/OCR | `AIEntryView`, `EntryProposalEditor` | entry capabilities/process/proposals | fixture, attachment cap, proposal review, proposal-editor shell clearance and 500ms add long-press tests | deterministic attachment UI 1/1, explicit-type review 1/1 and proposal-editor shell proof 1/1; synthetic live OCR+private S3+DeepSeek evaluation passes 24/24 fields; approval/UI attachment download remains separate owner UAT |
| Feedback/changelog | `FeedbackView`, shell bubble | feedback/changelog | capture/submit and route tests | local API submit UI 1/1 at `/tmp/mypocket-feedback-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-43-30-+0700.xcresult`; production credential/storage UAT open |
| Account/API keys/timezone | `AccountView` | profile/timezone/api keys | route + profile-card management action + timezone contract + stable editor UI test + local API create/revoke UI + request contracts | API-key behavior verified; local create/revoke UI 1/1 at `/tmp/mypocket-account-key-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_15-40-22-+0700.xcresult`; production credential UAT remains open |
| Notifications | `NotificationsView` | notifications/read | route + model contract + owner-session mark-read API contract + stable identifier | owner-scope/idempotent-read integration and seeded authenticated inbox mark-read UI test pass 1/1 against isolated local stage API; visual owner UAT remains open; no public notification-create route is intended |

### Assistant wallet confirmation duplicate-submit parity — 2026-10-08

The web confirmation card disables confirmation while `createWallet` is in
flight and after it succeeds; native previously fired a new task on every tap.
Parity requires allowing confirmation when idle, locking while saving, staying
locked after success, and allowing retry after an error. Keep this confined to
the existing shared `MPButton` card, owner session API call and presentation
state; no backend idempotency or API/schema change is part of this UI parity
fix. A focused state-transition contract was added before changing the card.

Verified: `AssistantWalletConfirmationState` now gates the async action and the
card renders the success state or a retryable error through shared status
components. The regression passes in the 128/128 native unit suite; Release
warnings-as-errors and all project/base-component/diff guards pass. No API,
schema or migration changed.

### Month-summary ledger-to-screen proof — 2026-10-08

The month route's design contract requires live report totals to reflect
owner-scoped ledger changes. The new API-backed simulator flow adds one
disposable-wallet expense, verifies the month API's expense/count delta, and
checks that native renders the updated expense total. It passed 1/1 at
`/tmp/mypocket-month-summary-green/Logs/Test/Test-MyPocketIOS-2026.10.08_03-02-16-+0700.xcresult`.
The inspected screenshot is
`/tmp/mypocket-month-summary-attachments/6C29E050-644F-4899-A4EF-69B3E1D7B3BE.png`.
The isolated stage database contains zero remaining `Fixture Month Wallet`
wallets and `Fixture month summary` transactions; the temporary 8081 API is
stopped. No API, schema or business-rule change was needed. Browser and owner
visual UAT remain separate.

### Recurring run-due ledger-state proof — 2026-10-08

The native integration test now exercises `Chạy các giao dịch đến hạn`, checks
the returned created-count message, fetches the generated transaction through
the same owner session, and asserts its amount, default recurring note and
fixture wallet. It then opens the native transaction tab and asserts that row
is visible. Teardown removes the fixture wallet's transactions and schedule.
This is a test-only proof improvement: no runtime/API/schema behavior changes
and no production data is touched.

The updated `LaunchTests.swift` compiles with
`SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG LOCAL_API_UI'` and the focused test
passes 1/1 on iPhone 16 Simulator at
`/tmp/mypocket-recurring-state-ui/Logs/Test/Test-MyPocketIOS-2026.10.08_02-43-21-+0700.xcresult`.
It ran only against the disposable `mypocket_stage_v1` database and Redis DB
15 on a temporary loopback API at `127.0.0.1:8081`; the fixture wallet,
schedule and generated ledger row were removed by teardown, and that API was
stopped. The existing API at `8080` was not used or changed. Pixel-level and
physical-device visual UAT remain open.

### Notification inbox UI proof plan

Keep the production contract unchanged: there is intentionally no public
notification-create endpoint. Close the UI proof gap by inserting one clearly
synthetic row directly into the disposable `mypocket_stage_v1` database for the
existing `fixture@example.com` local OAuth owner, then exercise the real native
list and mark-read action against a dedicated local API process pointed at the
stage database. Do not reuse the already-running port 8080 API: inspection
showed it uses the default `postgres` database, not `mypocket_stage_v1`. The
test API runs on loopback port 8081 with a separate Redis database index. The
app build setting and every `LocalAPIFixture` request must target that isolated
API; the API UI harness must never default to the already-running port 8080,
which uses the `postgres` database. The UI
test must verify the row appears, tap its stable
`notification-mark-read-<id>` action, and confirm through a fresh API list that
the same row has `read_at` set. Delete only the synthetic row afterward.
Refuse to seed unless the explicitly supplied test DSN targets the exact local
stage database; never use production credentials or data. No production/API/
auth/schema change is in scope. The affected internal docs are this audit and
the validation matrix; public API docs need no change because the contract is
unchanged. Changelog is not applicable to a test-only proof improvement.

## Work order

1. DONE: native portfolio buy execution and history refresh are covered on the
   simulator; the exact synthetic asset/trade was removed from the disposable
   stage DB after proof. Continue monitoring this contract in bounded tests.
2. DONE for provider recognition: the opt-in synthetic evaluation successfully
   uploads/reads/deletes one private object, extracts a 45,000 VND receipt with
   OCR, and gets 24/24 expected fields across five DeepSeek cases. It never
   approves proposals or writes ledger data. Assistant provider send/history
   are separately covered against the isolated stage API. Browser attachment
   download and owner review/approval remain separate acceptance gates.
3. Keep the bounded disposable API UI runner documented and repeat it after
   behavior changes; then run native verification, web typecheck, design
   checks and production build.
4. Only then perform authenticated visual UAT against the Stitch screenshots
   and mark this audit complete.

## Proof boundary

Tests use disposable owner-scoped fixtures or UI fixtures. They must not mutate
production data. Live provider/OCR, production feedback credentials and
physical-device signing remain manual release gates. The Stitch/web reference
specifies Manrope; native now bundles the OFL-licensed Manrope variable font
(`Manrope-Variable.ttf`) and routes all shared typography through `MPFont`.
The font registration and resource guard are verified at build time.
Pixel-level visual sign-off still requires an authenticated device comparison
against the supplied Stitch screenshots.
### Goal-wallet visual follow-up — 2026-10-07

The first scoped-goal capture showed the generic transaction shell around the
goal card. The native route now follows the supplied Stitch savings reference
when a goal is selected: ordinary period controls/title are removed, the
wallet/search/overflow header is compact and interactive, savings balance and
remaining target are centered, and the progress/advice plus month ledger are
rendered from shared bases. Focused API-backed evidence is recorded in
`/tmp/mypocket-goal-focused-final-4/Logs/Test/Test-MyPocketIOS-2026.10.07_17-15-08-+0700.xcresult`
and the inspected PNGs under `/tmp/mypocket-goal-focused-final-4-attachments/`.
This is still simulator and
local-fixture evidence; authenticated device/provider visual gates remain open.
The follow-up ordering regression is recorded at
`/tmp/mypocket-goal-order-green4/Logs/Test/` with inspected capture under
`/tmp/mypocket-goal-order-green4-attachments/`; the transaction scope now sits
between the centered balance and the savings progress block as in the Stitch
reference. A subsequent red/green hit-test regression found the goal CTA was
still covered by the shell's raised add button; the CTA now uses a nested
safe-area region with a 104px cushion and 56px visual lift. Green evidence is
`/tmp/mypocket-goal-cta-green3/Logs/Test/` with inspected capture under
`/tmp/mypocket-goal-cta-green3-attachments/`. Full unit/Release verification
after this last correction is green: 110/110 at
`/tmp/mypocket-goal-cta-full-unit/Logs/Test/`, Release at
`/tmp/mypocket-goal-cta-release`, with project/base/diff guards passing.

### Ledger and secondary shell visual follow-up — 2026-10-07

The ordinary transaction ledger now has a stable row reachability assertion
and an inspected capture proving its final rows remain above the persistent
shell after scrolling. Secondary routes use the shared
`MPBottomNavigationInset`; Categories, Portfolio, Credit and Notifications no
longer finish underneath the shell. The `SecondaryScreen` host hides
SwiftUI's default navigation back control so the custom Stitch header is the
only back affordance. Focused red/green evidence and screenshot paths are
recorded in `docs/work/VALIDATION_MATRIX.md` and the two visual-capture plans.
These are simulator/local-fixture checks, not production or physical-device
sign-off.

The same shell audit subsequently found four scrollable secondary screens that
were not reserving the shared inset: Finance Assistant, AI Entry, Feedback and
Account. Each now composes `MPBottomNavigationInset`, with a red/green focused
route regression proving the custom header and bottom shell remain usable.
This is a layout contract only; provider-backed content, seeded notifications,
physical-device pixel approval and production UAT remain open.

Wallet management had the same issue in its list/editor branch: the form save
footer could finish below the shell. The list and `WalletEditorView` now use
the shared inset, with a focused geometry assertion proving the save action is
above the shell. Wallet CRUD behavior and owner scope are unchanged.

The transaction overflow forms had the same scroll-clearance risk. Transfer
and balance-adjustment overlays now reserve the shared inset, and the focused
route test opens both actions to prove the form surfaces remain above the
persistent shell.

### Release configuration wiring audit — 2026-10-07

- The API process reads provider settings from process environment
  (`AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`, `OCR_API_URL`, `OCR_API_KEY`) and
  only loads the ignored `backend/.env.local` convenience file for local
  development. The iOS bundle receives none of these secrets.
- `GET /api/v1/ai/entry/capabilities` reports `files_configured` only when
  both OCR credentials and private attachment storage are available. This
  prevents the native/web pickers from claiming that an upload is ready when
  the backend cannot persist or OCR the file.
- The native AI Entry flow consumes that capability field before opening the
  picker and before submitting a request; it does not manufacture a successful
  capability state in Release builds. The `true` values in the DEBUG UI-test
  fixtures are isolated to deterministic UI fixtures.
- Production must inject OCR and S3/storage secrets into the API process and
  keep them out of Git, web bundles, and the mobile app. No production
  mutation was made during this audit.

### Feedback/changelog parity follow-up — 2026-10-07

The web feedback surface renders each user's description and the changelog
entry description, while the native list previously rendered only title,
status/date and changelog version/title. The native surface now uses the shared
`MPFeedbackStatus` presentation base and renders the same descriptive fields,
including the linked changelog version for fixed feedback. This is a client-
only parity change: the existing owner-scoped feedback/changelog APIs and
screenshot fallback remain unchanged. Focused RED→GREEN coverage plus the
full native unit target (113/113) and Release warnings-as-errors build pass.

The month detail parity review also found that the web shows the account
timezone and full report-boundary timestamps, while native showed only the
calendar dates. The next client-only correction will reuse the shared
timezone/date formatter and keep the existing month API contract unchanged.

### Transaction day-group header parity follow-up — 2026-10-07

- Web behavior: the ordinary transaction ledger renders each date group with a large day number plus localized weekday and month/year metadata. The group key is interpreted in the account timezone.
- Native gap: the iOS ledger previously rendered only one relative/short-date label, so the same ledger lost the visual hierarchy and month context present on web/Stitch.
- Intended behavior: add a pure, timezone-aware `TransactionsPresentation.dayGroup` formatter and render the same three-part header while retaining category grouping unchanged.
- Verification: add formatter contract tests first, then run the focused native test and project/base-component/design guards.

### Transaction category glyph parity follow-up — 2026-10-07

- Web behavior: transaction rows use the category presentation registry for icon/badge tone and reserve green for income, danger for expenses, and neutral text for adjustments/transfers.
- Native gap: rows used generic income/expense arrows and did not surface the category's semantic icon/tone, despite category metadata already being loaded for titles/grouping.
- Intended behavior: move the mapping into a shared base/presentation helper, map the canonical `system_key` families used by web, and keep the row's owner/API behavior unchanged.
- Verification: add pure mapping tests first, then run the focused native test and the existing project/base-component/design guards.

### Ended-budget preview parity follow-up — 2026-10-07

- Web behavior: ended budgets expose `Xem / Xóa`; the sheet is read-only and still permits deletion after confirmation.
- Native gap: ended budgets rendered a terminal label and a delete action but could not open the budget's details, so the native route lost the web inspection flow.
- Intended behavior: open a read-only preview for ended budgets, preserve the existing delete confirmation/API call, and keep save/edit controls unavailable.
- Verification: add a pure policy contract first, run the focused native test, then rerun native guards and the relevant build/test gates.

### Wallet card metadata/tone parity follow-up — 2026-10-07

- Web behavior: wallet cards show the wallet description or type label and render negative current balances in the danger tone.
- Native gap: the wallet-management list called `MPWalletRow` without a subtitle and the shared row always rendered the balance as ink.
- Intended behavior: keep this presentation in the shared wallet row base, pass the real description/type fallback and explicit negative-balance state from API-backed callers, and preserve masking.
- Verification: add a pure display-policy test first, then rerun focused/full native tests and design/base guards.
- PASS: `MPWalletRowPolicy` now owns description trimming/type fallback and
  negative-balance tone selection. Overview and wallet-management callers pass
  the API-backed description and balance sign while preserving the existing
  masked-value behavior and shared wallet row base.
- PASS: focused wallet policy test is 1/1 at
  `/tmp/mypocket-wallet-row-final/Logs/Test/`; the unit-only target is 117/117
  at `/tmp/mypocket-port-wallet-final-unit-only/Logs/Test/`, and the Release
  warnings-as-errors build is `/tmp/mypocket-port-wallet-final-release`.
- OPEN: authenticated device visual approval and production data UAT remain
  external gates.

### Reference segmented-control parity follow-up — 2026-10-07

- Web/Stitch behavior: segmented controls use a muted track with a white
  active pill and action-colored label; the active option is not a solid green
  primary CTA.
- Native gap: `MPSegmentedControl` reused `MPButtonStyle.primary`, producing a
  green active segment in the transaction editor and report/detail switches.
- PASS: the shared `MPButtonStyle.segmentedSelected` variant now owns the
  reference active surface and `MPSegmentedControl` selects it through a pure
  base policy. No feature screen adds local styling.
- PASS: focused base contract is 1/1 at
  `/tmp/mypocket-segmented-reference/Logs/Test/`; the unit-only target is
  118/118 at `/tmp/mypocket-port-segmented-final-unit/Logs/Test/`, Release
  warnings-as-errors is `/tmp/mypocket-port-segmented-final-release`, and the
  refreshed transaction-editor API fixture screenshot is recorded at
  `/tmp/mypocket-visual-wallet-editor-final-attachments/2DA367F3-AFFC-4C51-9BE5-5FE060546DA2.png`.
- OPEN: authenticated physical-device pixel comparison remains external.

### Transaction-editor category icon parity follow-up — 2026-10-07

- RED: the native editor always showed a generic tag for a selected category,
  while the Stitch/web form uses the category presentation registry.
- PASS: `TransactionsPresentation.glyph` now accepts canonical `iconKey`
  values as a fallback when a user-created category has no `systemKey`; the
  editor and ledger therefore resolve the same SF Symbol/tone through shared
  policy. No provider-specific bank logo is fabricated when the API has no
  logo asset contract.
- PASS: glyph fallback regression is included in the 118/118 unit target; the
  API-backed editor UI test passes and its refreshed screenshot is
  `/tmp/mypocket-visual-wallet-editor-icon2-attachments/0BFB1ECA-9040-426B-ABE8-1F95A365B1DC.png`.
- OPEN: provider-specific wallet artwork and physical-device pixel approval
  remain external gates.

### Category icon-picker parity follow-up — 2026-10-07

- RED: native category create/edit preserved `iconKey` but exposed no picker,
  so a user could not reproduce the web category icon workflow.
- PASS: `MPCategoryIconSelector` is a shared DesignSystem base with the same
  ten canonical keys/labels as the web picker. Editable category forms now
  bind the selection and send it through the existing owner-scoped
  `CategoryModelInput.iconKey`; category rows reuse the selected glyph and
  system categories remain metadata read-only.
- PASS: key/fallback contract is covered in the full 119/119 unit target at
  `/tmp/mypocket-native-final-unit3.xcresult`; Release warnings-as-errors
  build is `/tmp/mypocket-native-final-release2`.
- PASS: authenticated local API UI create/update/delete flow now selects
  `category-icon-expense_food`, persists `icon_key`, and completes 1/1 at
  `/tmp/mypocket-category-picker-ui4.xcresult`.
- OPEN: authenticated category visual UAT remains external.

### Amount keypad parity follow-up — 2026-10-07

- RED: the Stitch/web amount field exposed a calculator/keypad flow while
  native only opened the platform number keyboard.
- PASS: `MPAmountCalculator` now evaluates a bounded, non-`eval` grammar with
  checked integer arithmetic, precedence, divide-by-zero and fractional-result
  rejection. `MPAmountKeypad` is a shared base composed from `MPButton`,
  `MPSurfaceCard` and `MPReferenceScreenHeader`; every native amount field can
  open it without feature-local styling.
- PASS: the full native unit/contract target is 120/120 at
  `/tmp/mypocket-native-keypad-final-unit.xcresult`; Release
  warnings-as-errors passes at `/tmp/mypocket-native-keypad-final-release`;
  the transaction editor keypad UI flow passes 1/1 at
  `/tmp/mypocket-keypad-ui6.xcresult` (enter `1,000,000 + 500,000`, evaluate,
  commit and return to the form). Project/base-component/diff guards pass.
- SCOPE: decimal comma remains visible for reference parity but cannot be
  committed because transaction amounts are integer VND; no API/schema change.
- OPEN: authenticated physical-device pixel approval remains external.

### Calendar picker parity follow-up — 2026-10-07

- RED: the Stitch transaction editor exposed a calendar modal from the date
  row while native only exposed previous/next date actions.
- PASS: `MPDateCalendarSheet` is now a shared base composed from
  `MPSurfaceCard`, tokenized buttons and a Monday-first Vietnamese calendar
  grid. It includes adjacent-month cells, month navigation, selected/today
  states and stable day identifiers. `MPDateStepper` opens it from a stable
  date-button identifier; confirming updates the editor date and cancelling
  preserves the current value. Standalone `MPDateField` consumers in wallet,
  transfer/adjustment, portfolio, budget and transaction custom-range forms
  reuse the same sheet with the account timezone.
- PASS: the focused transaction-editor calendar UI flow is 1/1 at
  `/tmp/mypocket-calendar-custom-ui2.xcresult` (open, cancel, reopen and
  confirm; screenshot attachment exported under
  `/tmp/mypocket-calendar-custom-ui2-attachments/`); the full native
  unit/contract target is 121/121 at
  `/tmp/mypocket-native-calendar-field-final-unit.xcresult`; Release
  warnings-as-errors passes at `/tmp/mypocket-native-calendar-field-final-release`;
  project/base-component/diff guards pass.
- PASS: the API-backed ledger UI flow opens the custom range selector and the
  shared start-date calendar, then cancels back to the range form in 1/1 at
  `/tmp/mypocket-calendar-range-ui.xcresult`.
- SCOPE: the platform calendar supplies selection UI; account-local date
  serialization and validation remain owned by the existing feature forms.
- OPEN: authenticated physical-device pixel approval remains external.

### Budget date picker parity follow-up — 2026-10-07

- RED: the native Budgets form used raw `YYYY-MM-DD` text inputs while the web
  period selector and the other native date forms use a date-only picker.
- PASS: `BudgetsView` now uses the shared account-timezone `MPDateField` for
  start/end dates. Loaded and edited API strings are parsed into date-only
  state; save/update serializes the selected local calendar day back to
  `yyyy-MM-dd`, preserving existing validation, owner scope and API payloads.
- PASS: the API-backed create/update/delete flow is 1/1 at
  `/tmp/mypocket-budget-date-ui2.xcresult`; the prior harness failure was fixed
  by using the stable `budget-add` identifier (the UI label is localized as
  `Thêm`). Native unit/Release evidence is recorded in
  `/tmp/mypocket-budget-date-unit-final.xcresult` and
  `/tmp/mypocket-budget-date-release-final`.
- OPEN: authenticated physical-device pixel approval and live production API
  UAT remain external gates.

### Date-time picker parity follow-up — 2026-10-07

- RED: recurring schedules and AI proposal editing still exposed SwiftUI's
  compact `DatePicker`, bypassing the shared Stitch-style date base.
- PASS: `MPDateTimeField` now opens the shared `MPDateCalendarSheet` with an
  opt-in account-timezone hour/minute section. Selecting a new day preserves
  the existing clock, while cancel leaves the binding unchanged. Recurring
  schedule and AI proposal consumers pass their account/session timezone.
- PASS: focused design-system coverage is 43/43 at
  `/tmp/mypocket-datetime-unit-redgreen.xcresult`; the full 123-test and
  Release evidence is recorded in the current checkpoint paths above.
- PASS: the fixture-backed AI proposal editor opens the shared date+time sheet,
  exposes its hour/minute control and cancels cleanly in 1/1 at
  `/tmp/mypocket-datetime-ui2.xcresult` (screenshot attachment exported under
  `/tmp/mypocket-datetime-ui2-attachments/`).
- OPEN: authenticated physical-device pixel approval and live provider UAT
  remain external gates.

### Money Insider comparison-ring parity follow-up — 2026-10-07

- RED: the Stitch Money Insider reference exposes a compact period-comparison
  ring beside the current-period average. The web had a reusable donut base but
  the mounted card did not render it; native had no equivalent base.
- PASS (implementation): web `InsiderCard` now mounts `BaseDonutChart` from the
  real `expense_change_bps` response. Native `MPMoneyInsiderCard` mounts the
  shared `MPDonutChart`; both preserve the actual percentage label, clamp only
  the visual stroke, and use danger/action/neutral semantic tones.
- SCOPE: missing `expense_change_bps` renders the existing fallback copy and no
  ring. No API, schema, owner-scope, masking or provider behavior changed.
- Verification: focused native ring contracts, full native unit/Release,
  project/base/diff guards, web design checks and production build are recorded
  in the validation matrix below.

### Native Reports trend chart parity follow-up — 2026-10-07

- RED: the Stitch Reports reference exposes a real “Xu hướng chi tiêu” view,
  while the native segmented control previously repeated top expenses as a
  placeholder bar chart.
- PASS (implementation): `MPTrendData` aggregates eligible income/expense rows
  by account-local day and `MPTrendChart` renders separate semantic income and
  expense lines with a safe empty state. `ReportsView` loads the existing
  owner-scoped transaction endpoint in parallel with report summaries; no API or
  schema change was needed.
- SCOPE: excluded rows and transfer/adjustment types are ignored; masking stays
  in the shared screen contract. The web now mounts the matching
  `BaseTrendChart`/aggregation contract, so both clients consume the same
  owner-scoped source.
- Verification: `/tmp/mypocket-trend-full-unit.xcresult` was the preceding
  Reports-trend checkpoint at 126/126; the latest base-component checkpoint is
  `/tmp/mypocket-final-full.xcresult` at 129/129. The trend UI smoke passes 1/1,
  the latest Release warnings-as-errors build passes, native base/project/diff
  guards pass, and web trend/base SSR/design tests plus production build pass.

### Shared base-component completion — 2026-10-07

- `MPBudgetGauge` now owns the native 180-degree budget summary arc, clamps
  only visual progress, preserves real spent/limit labels and exposes a stable
  VoiceOver value. Web `BudgetGauge` remains the matching SVG base.
- `StatusBadge`/`MPStatusBadge` now own compact success/muted/danger status
  presentation and are mounted in recurring and budget status surfaces.
- Selector rows already expose controlled callbacks/identity IDs, and report
  bar charts are mounted with owner-scoped summary data; the inventory was
  reconciled instead of keeping those rows falsely marked partial.
- Verification: native DesignSystem tests are 129/129 in
  `/tmp/mypocket-final-full.xcresult`; web base SSR contract, typecheck, build,
  design links and both native base guards pass. No API/schema/migration changed.
- Post-harness verification: the same native target remains 129/129 in
  `/tmp/mypocket-final-full-rerun.xcresult`; Release warnings-as-errors,
  project/base-component guards and `git diff --check` pass.
