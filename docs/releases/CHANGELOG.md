---
artifact_type: changelog
id: CHANGELOG
status: active
owner: shared
human_fields: [release_approval]
ai_fields: [change_entries, linked_releases]
shared_fields: [status]
---

# Changelog

## 1.1.0 patch — 2026-10-16

- Deploy API and worker image `release-20261016-7846945` (digest
  `sha256:71fdd4b4155972f3e6380ae02de846264e503347ab63b9693f65d2f0c713db43`).
- Expose the async Finance Assistant submit route so the first provider run
  can be cancelled immediately, while preserving the synchronous contract.
- Preserve multi-image upload order and delimit each OCR document before model
  extraction; no database migration or financial-data mutation was performed.
- Production smoke: health `200`, API/worker `1/1`, Swagger `78` paths.
- Known gates: authenticated provider/OCR UAT, native offline sender, and
  physical-device acceptance remain open.

## 1.1.0 — 2026-10-15

- Freeze the current native iOS and finance-platform snapshot as the first
  semantic minor release after the `1.0.0` baseline.
- Align iOS bundle metadata (`1.1.0`, build `2`), web package metadata, and
  the public OpenAPI document on the same release version.
- Include the production Finance Assistant jar-capability hardening, native
  base-component and offline-contract work, feedback/report/wallet/jar flows,
  and the verified local regression suites described in
  `REL-1.1.0-native-ios.md`.
- Production rollout is API/worker only; no migration or web redeploy is part
  of this native-focused release. The final image digest and device-signing
  result are recorded in the release note after verification.

## Unreleased — native Stitch parity continuation

- Continue native Money Lover reference parity with populated transaction
  ledger, savings report and budget report compositions. Shared reference
  surfaces now own compact metrics, adaptive daily rows, masking and Dynamic
  Type/accessibility reading order; no API or financial calculation changes.
- Extend the same shared composition to populated Overview and Wallets
  surfaces, including compact income/expense metrics and the selection versus
  management title contract.
- Stabilize JSON request bytes with sorted keys so identical budget-series
  retries remain idempotency-safe and reproducible. Focused native tests and
  Release simulator build pass; physical-device owner UAT remains pending.

- Add local Savings Report with separate goal-wallet activity and account-wide
  surplus, month/wallet filters and paged transaction drilldown. Initial wallet
  balances enter history at creation and are shown separately from ledger rows.
- Add weekly/monthly/quarterly/yearly budget series with versioned next-period
  edits, stop, safe replay and unique worker/API catch-up. Optional positive
  carryover resets on wallet/category changes; past edits recalculate totals.
- Add deterministic budget forecast, daily spend and identical transaction
  scope; native screens reuse shared bases and preserve existing API decoding.
- Add migration035 locally and generated OpenAPI. Production rollout and
  owner/device acceptance are pending; no release announcement is implied.

- Harden native Reports owner scoping and fail closed when the month summary
  cannot be loaded (`089fd656`); signed simulator tests pass 307/307.
- Add offline transfer receipt replay assertions and document the remaining
  credential-gated production UAT (`54087c8`).
- Add deterministic AI/OCR fixtures for duplicate rows, missing dates, wallet
  inference and no-auto-approval, and explicitly mark savings reports and
  budget recurrence/forecast as partial until their API contracts exist
  (`982f8c48`). No production deployment or schema migration was performed
  for these local hardening changes.

- Native Reports now scopes `monthSummary` and Money Insider reads to the
  current owner through the existing `expectedOwnerID` guard. A month-summary
  failure now fails closed for the whole report instead of rendering insider
  totals with an empty breakdown/activity; no backend or API schema change.

- Production Finance Assistant prompt now states explicitly that the model may
  prepare a `propose_create_jar` proposal and must show the confirmation card;
  the direct-write boundary remains unchanged. API image
  `ai-jar-capability-20261015-215712` is deployed to API and worker with no
  migration or web update. Owner UAT is still required before resolving the
  related feedback.

- Finance Assistant now supports a prepared `propose_update_budget` action.
  The persisted proposal is owner/generation-bound, native reuses the shared
  budget confirmation card with update copy, and confirmation delegates to the
  existing budget PATCH validator. No production, migration or device rollout
  was performed.

- Finance Assistant now supports a prepared `propose_create_budget` action.
  Native renders the budget proposal with explicit confirmation; backend
  reuses the owner-scoped budget validator and `budget.create` idempotency
  ledger. No production, migration or device rollout was performed.

- Finance Assistant now supports a prepared `propose_update_jar` action. The
  model returns a validated, UUID-bound monthly replacement proposal; native
  renders the same confirmation card vocabulary with update-specific copy, and
  confirmation delegates to the owner-scoped jar PUT route. No provider,
  migration, production or device rollout was performed.

- Finance Assistant handbook wording now matches the implemented jar proposal:
  requests can use `propose_create_jar` and receive an explicit confirmation
  card, while direct edits/deletes and unconfirmed writes remain unavailable.
  This is a local prompt/contract correction; no API, database or deployment
  change was made.

- The native Feedback bubble now uses a stable lower-left shell dock instead of
  jumping around reported controls. Footer clearance keeps it above the visible
  composer/navigation without reserving a scroll lane. Local DesignSystem
  tests/build/guards pass; device and production rollout are unchanged.

- Finance Assistant now supports a prepared `propose_create_jar` action. The
  model only validates and returns a server-bound jar proposal; native cards
  show month/name/allocation and require explicit confirmation. Confirmation
  is owner/generation scoped and reuses the durable `transaction_mutations`
  ledger under `jar.create`, so retries cannot create a duplicate jar. Backend
  tests/vet and the native simulator build pass locally; the fresh unit-test
  runner was blocked by CoreSimulator diagnostics. No production rollout or
  migration was performed.

- Google mobile login now distinguishes invalid OAuth data (`401
  GOOGLE_MOBILE_LOGIN_INVALID`) from a local user-store outage (`503
  GOOGLE_USER_STORE_UNAVAILABLE`). The latter keeps database diagnostics out of
  the response while making the recurring “Google login invalid” failure
  actionable in logs and test fixtures. Covered by use-case and HTTP
  regressions; rollout remains pending.

- Prepared wallet actions now have a server-bound confirmation contract. Native
  cards with persisted action context call the owner/generation-bound advisor
  dispatcher; the backend ignores client financial fields, rejects stale
  generations and reuses the durable wallet mutation ledger. Legacy cards keep
  their ordinary endpoint fallback. Local backend/native regression tests pass;
  production rollout and device UAT remain open.

- Feedback resolution now invalidates in-flight list snapshots when a checkbox
  update begins, so a late refresh cannot move a just-resolved item back into
  the unresolved tab. The native concurrency regression passes locally.

- Backend now exposes unauthenticated `/docs` and `/docs/` entry points that
  redirect to the versioned Swagger UI at `/api/v1/docs/index.html`, keeping
  public documentation links stable without exposing user data. Locally
  verified; production rollout remains separate.

- Native wallet proposals now persist a server-generated `action_id` in the
  assistant message. Reopening a proposal after an app restart reuses the same
  idempotency boundary, so confirmation retries cannot create a duplicate
  wallet; legacy cards retain a safe mounted-card fallback.

- Finance Assistant adds a bounded read-only `get_product_help` tool and a
  shared `product_help` card on native and web. Results come from versioned
  handbook sections with stable source IDs and bounded excerpts; no owner
  financial data or external retrieval is exposed. No-match queries fail closed
  with `not_found`. This is a local implementation slice, not a production
  rollout or general live RAG.

- Native wallet confirmation cards now reuse one idempotency key for retries
  while the card is mounted, reducing duplicate-create risk after a transport
  timeout without adding a new endpoint or permission. Durable replay across
  app restarts remains out of scope for this local change.

- Native wallet confirmation cards now show all goal/credit fields before the
  user confirms (target/date, credit limit, statement balance and due dates),
  together with total-inclusion and description when present. This is a
  presentation-only correction; no API or backend behavior changed.

- Confirmed wallet creates can now send an `Idempotency-Key` through the
  existing PostgreSQL mutation ledger. Same-owner/same-payload retries replay
  the original wallet instead of creating a duplicate; changed payloads fail
  with conflict. Keys are validated as one ASCII 1–128 byte header; duplicate
  or invalid keys return `400`, and keyed requests fail closed when a storage
  adapter cannot provide the mutation ledger. Legacy no-key web writes remain
  unchanged.

- Advisor history now has an additive lifecycle `meta` contract containing the
  server-authoritative conversation ID/generation, last user activity, expiry
  and server time. The message array remains backward-compatible and reads do
  not extend TTL. Backend/native contract tests and the simulator warnings-as-
  errors build pass locally; this is not a production rollout.

- Finance Assistant now has an explicit, idempotent **Chat mới** reset locally.
  The owner-scoped reset purges only advisor content, rotates the generation and
  cancels queued/running provider contexts in the current process; replaying the
  same idempotency key cannot cancel newer work. Financial data is untouched.
  Backend migration `000034` and the native reset contract are tested locally;
  production rollout remains separate.

- Native Overview/Reports now reuse the encrypted owner-scoped financial snapshot
  for transient offline reads. Server-returned month summary and insider data
  are cached without local recomputation; Reports requires an exact matching
  month and shows a cached timestamp. Added round-trip, merge, fallback and
  month-scope tests; no API/schema change.

- Finance Assistant product guidance now matches the native shell: `Trợ lý` is
  documented as the primary tab instead of a destination under `More`. Added a
  provider regression and bumped the embedded handbook to v1.1; no API/schema
  or provider behavior changed.

- Native Reports now validates the month scope locally (`YYYY-MM`, months
  `01`–`12`) before any request and clears the previous report when a scope or
  load fails, preventing stale totals/trends from appearing under an invalid
  month. Added unit and authenticated UI regressions; no API/schema change.

- Overview's weekly income/expense card now applies the same report rules on
  web and native: account-local Monday–Sunday range, included income/expense
  rows only, excluding transfers, adjustments and rows opted out of reports.
  Web boundary/filter tests and 148 native unit/contract tests pass; no
  API/schema change.

- Native Reports now follows the current web hierarchy: the summary uses a
  safe income/expense comparison chart and Chi tiêu/Thu nhập/Chênh lệch metrics,
  the account-local trend is always rendered in its own card, category totals
  no longer duplicate a chart, and Money Insider shows the prior-period donut
  when available. API-backed iPhone 16 UI fixtures verify the visible trend,
  exact persisted trend values, donut reachability above the persistent shell,
  and excluded-row handling; no API/schema change.

- Native feedback now uses a shared fixed-height dock lane below a clipped,
  bounded shell route viewport, so scrolling More/category content and pinned
  editor actions cannot sit beneath the global trigger. Trigger hiding during
  screenshot capture preserves layout geometry. Duplicate shell route insets
  were removed while independent editor/sheet insets remain. Simulator UI
  regressions cover scroll reachability, capture, tab hit-testing and form
  clearance; no API/backend/schema behavior changed.

- Native AI transaction entry now mirrors web idempotent recovery: ambiguous
  transport/server failures keep the original request ID and draft, disable
  resubmission, and require GET result recovery. A confirmed 404 allows an
  explicit retry without discarding the user's input. 146 native unit tests
  and an isolated iPhone 16 recovery UI test pass; no API/backend/schema change.
- The native iOS app now packages the same wallet mark as the web PWA as a
  1024px AppIcon asset; the release build embeds it in the compiled asset
  catalog. No runtime/API behavior changed.
- Native transfer edit now marks source and destination wallets read-only,
  matching the existing pair-update API contract; changing amount/date/note
  continues to update both rows atomically. An API-backed iPhone 16 regression
  verifies pair state and deletion; no backend/API/schema change.
- Fix transaction creation so an explicit `included_in_reports=false` choice is
  persisted instead of being replaced by PostgreSQL's `true` default. Native
  Reports trend VoiceOver now reads daily income/expense totals in order and
  respects amount masking. PostgreSQL and API-backed iPhone 16 regression pass;
  no API or schema change.
- Native transaction/budget category pickers now render one shared card per
  root category, matching the web selector while retaining ancestry during
  search. API-backed iPhone 16 budget, recurring and transaction flows verify
  selection and persisted category IDs; no backend/API/schema change was needed.
- Picker rows now expose web-matching system, personal and ineligible subtitles;
  selection eligibility is independent of hierarchy context. The budget-wide
  clear control uses the shared transparent, leading-aligned ghost button.
  iPhone 16 API-backed picker, recurring and transaction regressions pass; no
  backend/API/schema/migration change.
- Native goal-wallet detail now owns the Transactions viewport scroll instead
  of nesting inside the Transactions outer scroll. This keeps its fixed
  “Thêm giao dịch” action above the shell with populated ledgers; an API-backed
  iPhone 16 regression verifies a real transaction row can scroll fully above
  the action. No backend/API contract changed.
- Personal categories now have an owner-scoped, reversible archive/restore
  contract (`is_active` plus `PATCH /api/v1/categories/{id}/status`). Existing
  transactions, summaries, budgets and recurring schedules are preserved;
  new transaction, credit, budget, recurring and AI-entry assignments reject
  inactive groups. Swagger and the public API reference are updated. The
  additive local migration is `000030_category_active`; production was not
  migrated or deployed.
- Category management now follows the web/Stitch action order, opens the editor
  by tapping a full category row, keeps delete in the editor, and shows
  applicability text from the selected wallet scope. API-backed iPhone UI
  verifies scope copy and create/edit/delete; no backend/API/schema change was
  needed.
- Native category management now uses a dedicated create/edit destination and
  visible shared checkbox rows for applicable wallets. An API-backed iPhone 16
  flow verifies create, selected-wallet persistence, filters, rename and delete.
  It exposed Swift's synthesized `walletIDs` wire key as `wallet_i_ds` instead
  of the Go API's `wallet_ids`; the request model now pins the correct wire
  name. The Go repository also creates a category and its owner-validated wallet
  links atomically. Public route/schema are unchanged.
- Native category management now has expense/income/debt segments, wallet
  scope, accent-insensitive search and a shared root/child tree. The
  owner-scoped category API supplies `wallet_ids`; category archive state is
  now also consumed by the web and native selectors and management screens.
- Category archive/restore responses now include the complete category shape,
  including `wallet_ids: []` when there are no explicit wallet links. The iOS
  client requires that field to decode the status response; previously the
  database-backed status route returned `null`, so archive/restore appeared to
  fail after the state had already changed. A PostgreSQL HTTP contract and an
  API-backed iPhone 16 flow now cover both transitions. The selected UI test
  still reports the separately tracked global feedback-bubble overlap.
- Native transaction and budget category selectors now use the searchable
  shared hierarchy. Transaction selection preserves wallet/type eligibility;
  budget “Tất cả danh mục” still sends the existing nil category scope. API-backed
  iPhone flows verify persisted category IDs; no backend/schema change was
  needed.
- Native recurring schedules now use the same category picker; the schedule
  and the normal transaction generated by “Chạy các giao dịch đến hạn” retain
  the selected category. API-backed iPhone proof passes without changing the
  recurring API contract.
- Native credit entry now uses that searchable category tree inside its
  existing action sheet. The picker preserves the entry draft and limits
  purchase/fee/interest to applicable expense groups and refunds to applicable
  income groups; the API-backed iPhone test confirms the chosen category ID
  persists. Credit payment remains unchanged; no backend/API/schema update was
  needed.
- Native Finance Assistant result cards now use the same compact row caps as
  web (5 transactions; 8 wallets/goals/budgets/jars), while retaining server
  totals and the existing read-only/action-confirmation contract. Focused
  Simulator tests and Release warnings-as-errors build pass; no API/schema
  contract changed.
- The authenticated More-route UI test now captures all 15 destinations using
  the isolated local API in `LOCAL_API_UI` mode. The iPhone 16 sweep passes
  1/1 and avoids offline/error placeholders; this is empty-fixture route
  evidence, not populated-data or pixel-parity sign-off. Screenshot review
  found a feedback-bubble collision with Account explanatory text; a focused
  API-backed RED/GREEN test now verifies the text wraps clear of the bubble.
- Revalidated the receipt OCR path with a synthetic fixture: private S3 put and
  signed-readback/cleanup passed, OCR recognized 45,000 VND, and DeepSeek scored
  24/24 expected fields in five synthetic cases. No proposal or ledger data was
  written; iOS picker/approval/download and owner-device UAT remain open.
- Native wallet management now mirrors the web's included/excluded groups and
  included-only “Tổng cộng”, with Sửa/Xong mode, row-to-edit and confirmed API
  deletion,
  add/service actions and a preserved Chi tiết route. The API-backed iPhone 16
  flow verifies aggregate routing and the total. It also exposed and fixed a
  backend persistence bug that changed explicit `is_in_total=false` to true;
  defaults and public API/schema remain unchanged. Backend suite, native
  130/130, UI 1/1 and Release warnings-as-errors pass.
- Native wallet-selection rows now follow the Stitch/web hierarchy: wallet
  title above balance, shared leading alignment and a selected-scope check on
  the aggregate or chosen wallet. Goal progress, credit availability and edit
  remain intact; native-only Chi tiết stays available as a compact trailing
  action. API-backed iPhone 16
  regression passes; no API or persistence contract changed.
- The wallet aggregate now uses a distinct globe badge in the shared teal tone
  instead of reusing the orange basic-wallet icon, matching the approved web
  selector contract. API-backed simulator screenshot and semantic UI assertion
  pass; no wallet data or public contract changed.
- Ordinary transaction ledgers now use the compact centered wallet-scope
  capsule directly under the header on iOS, matching the web hierarchy while
  preserving existing wallet selection and ledger filtering. API-backed
  simulator route/order and screenshot proof pass 1/1; no API or data contract
  changed.
- Native transaction amount now matches Stitch/web anatomy through an
  `MPAmountField` presentation variant: VND chip on the leading side, large
  unboxed amount, and Vietnamese thousands grouping for display while bound
  values remain raw integer digits. The regression checks position, grouping
  and keypad flow; native contracts pass 127/127 and the Release warnings-as-
  errors build passes.
- Native Finance Assistant wallet-confirmation cards now block repeat
  submission while the owner-scoped create request is running, show completion
  after success, and permit retry after an error, matching the web card. This
  is presentation-state parity only; the native contract suite passes 128/128
  and Release warnings-as-errors passes. No API/schema or backend behavior
  changed.
- Added API-backed native month-summary verification: a disposable ledger
  expense changes the server month aggregate and the iOS report renders that
  updated value. The stage wallet/transaction fixtures were removed after the
  simulator run; no reporting API or business rule changed.
- Native credit-wallet add-entry and payment actions now present in the shared
  `MPFormBottomSheet` instead of expanding inline in the ledger. The form stays
  pinned above scrolling fields, cancel/dismiss does not write, and saving
  prevents accidental dismissal; existing credit APIs and transaction rules
  are unchanged. iPhone 16 API-backed UI verifies both sheets open and cancel.
- Fixed native shell navigation from Account/More: Overview, Transactions and
  Assistant remain tappable while More is open, while the center add affordance
  stays disabled and the selected Account tab retains its narrow hit blocker.
  The iPhone 16 simulator regression covers long-press to AI Entry, return,
  switch to Transactions, normal add, and all 15 More destinations (3/3).
- Native sign-in now follows the password-only migration decision. The Google
  button, native PKCE client, client/redirect build settings and callback URL
  scheme are removed; email/password, refresh and Face ID remain. The iPhone 16
  launch UI regression passes with Google controls absent; the current unit
  suite is 127/127 after deleting two obsolete Google-auth tests.
- Native/web base-component parity: budget summary gauges now use real
  spent/limit data on both surfaces, recurring/budget statuses use the named
  `StatusBadge`/`MPStatusBadge`, and the component inventory no longer labels
  mounted selectors/charts as partial. The final native gate for this slice is
  127/127 after removal of obsolete Google-auth tests.
- Native portfolio trade API contract is now regression-covered for owner path,
  snake_case payload and idempotency; the append-only product behavior is
  unchanged.
- Fixed native portfolio trade history refresh: after recording a buy/sell,
  the position and immutable history now reload from their owner-scoped API
  sources. The iPhone 16 stage UI test verifies a fractional buy, updated
  position and persisted price/fee; no API or schema contract changed.
- Revalidated the native recurring schedule flow end to end against the
  disposable API fixture: create, update, delete and run-due pass 1/1. The UI
  harness targets the shared `recurring-add` ID and asserts the current `Thêm`
  label, removing the stale `Thêm lịch` selector.
- Added a real native Reports “Xu hướng” surface: `MPTrendChart` aggregates
  owner-scoped income/expense transactions by account-local day, excludes
  non-report and transfer/adjustment rows, preserves an empty state, and is
  covered by the API-backed Reports smoke test. Native unit/contract coverage
  was 126/126 at that historical checkpoint; the current auth parity gate is
  127/127 after removal of obsolete Google-auth tests.
- Mounted the matching web `BaseTrendChart` in Reports. It reuses the same
  account-timezone aggregation contract and tokenized SVG base, with SSR design
  tests for boundary/filter/empty behavior; no backend contract changed.
- Added the Money Insider comparison ring from the Stitch reference on both
  web and native. It is driven only by the owner-scoped `expense_change_bps`
  response, keeps the real percentage label for values above 100%, clamps only
  the visual stroke, and omits the ring when comparison data is unavailable.
  Native owns this through the shared `MPDonutChart` base; web reuses
  `BaseDonutChart`.
- Added the shared native VND amount calculator/keypad: every `MPAmountField`
  can open the Stitch/web keypad, evaluate bounded integer-safe arithmetic and
  commit only non-negative integer results. The transaction-editor flow is
  covered by a 1/1 UI test; the gate at that slice was 124/124.
- Added the shared native calendar picker: `MPDateStepper` now opens a
  tokenized Stitch-style Monday-first date grid from the transaction editor,
  and all standalone `MPDateField` inputs reuse the same base. It includes
  adjacent-month cells, month navigation, explicit confirm/cancel behavior and
  no feature-local picker styling. The focused UI flow passes 1/1 and the full
  native unit gate at that slice was 124/124.
- Updated native Budgets period editing to use the same account-timezone
  date-only picker for start/end boundaries. API payloads remain `yyyy-MM-dd`
  and the disposable API-backed create/update/delete flow passes 1/1 at
  `/tmp/mypocket-budget-date-ui2.xcresult`; the gate at that slice was 124/124
  at `/tmp/mypocket-insider-donut-final.xcresult` (that historical gate was
  126/126; the current gate is 129/129).
- Removed the remaining feature-level compact date picker from recurring and
  AI transaction-entry forms. `MPDateTimeField` now opens the shared calendar
  base with account-timezone hour/minute selection and preserves the clock when
  changing days; focused design-system coverage passes 43/43.
- Fixed transaction custom-period date fields to pass the account timezone into
  the shared range picker, preventing date drift when the device timezone
  differs from the user's account timezone.
- Matched native Feedback/changelog content parity: submitted descriptions,
  full release descriptions, status labels and linked `Fixed in <version>`
  output now render through shared `MPFeedbackStatus`/presentation helpers.
- Matched native month-detail metadata parity: account timezone and full local
  report-boundary timestamps now render with the exclusive next-month marker.
  Native unit coverage is 119/119 and the Release warnings-as-errors build
  passes; provider, production and physical-device gates remain open.
- Matched native ordinary transaction day-group anatomy with web/Stitch: each
  date card now renders a large local day number, localized weekday,
  month/year and signed group total through the account timezone; category
  grouping remains unchanged. Focused formatter test passes; device pixel
  approval remains open.
- Matched native transaction rows to the web category presentation registry:
  shared glyph/tone bases now render canonical category icons, green income,
  danger expense and neutral transfer/adjustment amounts, while legacy
  overview/report call sites retain their compatibility initializer.
- Matched ended-budget inspection behavior: native cards now expose `Xem /
  Xóa`, open a read-only scope/value preview, and keep edit/save controls
  disabled while preserving the existing owner-scoped delete confirmation.
- Matched wallet-card metadata/tone behavior: shared wallet rows now show the
  API description with canonical type-label fallback and use the danger tone
  for negative balances, including Overview and wallet management callers.
  Native unit coverage is 119/119; Release warnings-as-errors passes.
- Matched segmented-control styling to Stitch/web: active options now use the
  shared white surface pill on the muted track with action-colored text rather
  than the solid primary CTA style. This applies consistently to transaction,
  report and detail switches.
- Matched transaction-editor category identity to the web presentation
  registry: selected categories now reuse canonical `systemKey`/`iconKey`
  glyphs instead of a generic tag, including user-created icon fallbacks.
  The API-backed editor screenshot was refreshed; provider-specific wallet
  logos remain intentionally out of scope until an asset contract exists.
- Added the missing native category icon picker: editable category forms now
  use the shared `MPCategoryIconSelector` base, expose the same ten web keys,
  and persist the selected `icon_key` through the existing API payload.
- Revalidated the configured OpenAI-compatible provider for the native port:
  `DeepSeek-V4-Flash` text extraction, one strict transaction tool-call probe,
  and five strict JSON Schema cases pass without ledger writes. OCR image UAT
  still requires the server-side OCR credential; no production data changed.
- Added inspected simulator evidence for the ordinary transaction ledger and
  corrected secondary-route shell anatomy: shared bottom inset coverage now
  includes Categories, Portfolio, Credit and Notifications, and duplicate
  native back buttons are hidden behind the custom Stitch header. No API,
  schema or production data changed; authenticated device/provider visual UAT
  remains open.
- Fixed native bulk-delete request ownership: one confirmation now produces one
  sorted, de-duplicated API request, with local rows removed only after the
  atomic operation succeeds.
- Added a dedicated Notifications shell/header smoke and screenshot evidence;
  seeded inbox content and mark-read owner UAT remain pending because the local
  fixture has no notification-create route.
- Fixed the same shell-clearance gap on the scrollable Finance Assistant, AI
  Entry, Feedback and Account routes by composing the shared
  `MPBottomNavigationInset`; focused red/green route evidence is recorded in
  the validation matrix.
- Fixed the Wallets list/editor shell clearance so the form save footer also
  composes `MPBottomNavigationInset`; wallet CRUD/API behavior is unchanged.
- Fixed shell clearance for the transaction overflow's transfer and balance
  adjustment forms with the same shared inset; transfer/adjustment API
  contracts are unchanged.
- Fixed the AI entry proposal editor overlay to reserve its own shared shell
  inset and keep `Lưu đề xuất` above the persistent navigation. Added a scoped
  red/green simulator contract and screenshot evidence; provider/OCR behavior
  and production data are unchanged.
- Fixed the wallet transaction-detail overlay to reserve its own shared shell
  inset. The API-backed wallet mutation flow now proves the detail action card
  remains above the persistent navigation; transaction API behavior is
  unchanged.

- Ported the embedded goal-wallet ledger toward the supplied Stitch savings
  reference. Goal scope now uses shared compact wallet/search/overflow header
  controls, centered balance/remaining target, deadline/progress/advice,
  owner-scoped month grouping, report-link and add-transaction actions;
  ordinary period controls are hidden for goals. API-backed simulator capture
  passes 1/1;
  authenticated device/provider visual UAT remains open.

- Captured authenticated native More/account and transaction-editor visual
  evidence against the supplied Stitch reference. Transaction editors now
  present as full-screen reference forms above shell navigation, use the shared
  Stitch close capsule, and expose the existing AI receipt-entry route through
  a shared footer action. Focused and disposable API-backed editor UI checks
  pass 1/1 each; production/provider visual gates remain open.

- Centered the native More title independently of its support pill through the
  shared `MPCenteredScreenHeader`; the existing profile route smoke remains
  green and accessibility IDs are unchanged.

- Aligned native wallet badges with the web wallet contract through shared
  `MPWalletGlyph`: 48pt circular geometry and basic/goal/credit warning,
  success and teal tones. Wallet APIs/data and accessibility remain unchanged;
  native coverage is now 110/110 with Release/base/web guards passing.

- Matched native AuthGate to the web/Stitch reference: shared circular
  `MPBrandMark`, direct canvas title/subtitle, flat shared fields, full-width
  email sign-in. This historical visual entry included Google sign-in; the
  password-only migration decision below supersedes that control and removes
  native PKCE configuration.

- Verified the native Account API-key editor against the disposable local API:
  create and revoke pass 1/1 on the iPhone 16 Simulator. Production
  credential UAT remains open; no production key was created or revoked.
- Verified the native Reports route against the disposable local API: the
  Money Insider summary/detail flow passes 1/1 on the iPhone 16 Simulator.
  Report APIs and owner scope are unchanged; authenticated visual UAT remains
  open.
- Verified the native Finance Assistant overview route against the disposable
  local API (1/1). Provider-backed send/history/action behavior remains a
  separate external UAT gate; no model/provider call was made by this proof.
- Verified the native floating feedback bubble and submit flow against the
  disposable local API (1/1) with screenshot storage unavailable. Production
  credential/storage UAT remains open.
- Verified the native AI-entry attachment UI accepts the configured 20-file
  fixture, renders removable thumbnails and re-enables the picker after a
  removal (1/1). Live OCR/provider decision UAT remains open.
- Verified native AI-entry proposal review requires an explicit income/expense
  type before save (1/1); no proposal is auto-approved.
- Re-verified notification persistence boundaries with the local stage database:
  owner-scoped list/read, idempotent mark-read retries and cross-owner rejection
  pass. No notification-create API or production mutation was added.
- Release-wide verification remains green: backend race suite, native 108/108
  unit/contract target, native Release warnings-as-errors build, web typecheck/
  build/design guards and project/base/diff guards all pass. External provider,
  OCR, production credential and authenticated visual gates remain open.
- Matched native Portfolio summary anatomy to the web: added the supporting
  subtitle and grouped headline cost/value/P&L metrics under the shared
  `Tổng danh mục` reference card. Portfolio APIs, trade/price validation,
  timezone and owner scope are unchanged. Native unit coverage is 108/108 and
  Release/base/web guards pass; prior portfolio mutation proof remains recorded
  and authenticated visual UAT is open.
- Matched native credit-ledger summary anatomy to the web panel: a shared
  reference card now presents current debt and available credit, then limit and
  payment status/due details. `Thêm giao dịch` and `Thanh toán thẻ` use stable
  shared action IDs; credit APIs, payloads, empty state and owner-scope behavior
  are unchanged. Native unit coverage is 108/108 and Release/base/web guards
  pass; the bounded authenticated credit API UI smoke passes 1/1. The follow-up
  places actions before the statement and renders the empty message as plain
  text outside the list card; authenticated visual UAT remains open.
- Matched native Portfolio action anatomy to the web panel: the header now
  keeps only shared `Làm mới`, while loaded summaries expose `Thêm tài sản`
  and `Ghi giao dịch` cards built from shared native bases. Per-asset
  `Giao dịch` selects the shared trade card; portfolio API, validation,
  timezone and owner-scope behavior are unchanged. Full native unit coverage
  is 107/107 and Release/base guards pass.
- Matched native Travel Mode action anatomy to the web summary card: the header
  is navigation-only and the shared primary button `Thêm` now sits beside
  `Travel Mode` with stable ID `travel-add`; event APIs/data behavior are
  unchanged. Full native unit coverage is 106/106 and Release/base guards pass.
- Matched native Recurring action anatomy to the web summary card: the header is
  navigation-only and the shared primary button `Thêm` now sits beside
  `Giao dịch định kỳ` with stable ID `recurring-add`; schedule/run-due APIs and
  data behavior are unchanged. Full native unit coverage is 105/105 and
  Release/base guards pass.
- Matched native Jars action anatomy to the web monthly section: the header is
  navigation-only and the shared compact button `Thêm hũ` now sits beside
  `Hũ tháng` with stable ID `jar-add-month`; jar APIs/data behavior are
  unchanged. Full native unit coverage is 104/104 and Release/base guards pass.
- Matched the native Budgets header action to the web panel: the shared compact
  button now reads `Thêm` with stable accessibility ID `budget-add`; budget
  CRUD/API behavior is unchanged. Full native unit coverage is 103/103 and the
  Release/base-component guards pass.
- Matched native Finance Assistant structured cards to the web vocabulary:
  legacy summary/comparison/search/budget/wallet aliases now normalize to the
  canonical renderer, and `transaction_detail` shows wallet/category, amount,
  report inclusion and transfer markers with safe fallbacks. The backend
  response and tool contract are unchanged; focused card tests pass 15/15.
- Matched the web provider-capability boundary in native Finance Assistant:
  composer input/send now remain disabled until the capability is enabled and
  no run is active; the send method enforces the same policy independently.
- Matched the web `goal_progress` card in native Assistant: goal wallets now
  show a clamped shared progress bar and percentage, while ordinary wallet
  balance cards retain their simpler row layout.
- Matched the web `jar_progress` data contract: native now reads
  `allocation_amount` before budget fallback keys, preserves an explicit zero
  allocation, and avoids inventing a denominator when a jar is unconfigured.
- Added the remaining non-mutating Finance Assistant card context from web:
  source provenance, transaction result counts, comparison guidance and
  budget/jar period totals now render through shared native text/progress bases.
- Matched Assistant privacy-mask behavior with the web shell: the persisted
  mask state now reaches native Overview and Finance Assistant, redacting card
  amounts/summary counts and free-form assistant prose without changing API payloads.
  The follow-up passes 91/91 native unit tests, focused/API Assistant UI
  smokes, Release warnings-as-errors and project/base guards.
- Extended the same privacy-mask state through native `SecondaryScreen` into
  budgets, reports, jars and month detail. Their monetary labels and progress
  accessibility values now use the shared masked formatter; unrelated routes
  retain the web shell's intentional prop boundary. Native unit coverage is
  92/92 with focused UI, Release and base-component guard evidence.
- Matched the web Reports Money Insider detail surface with the shared native
  `MPReportsInsiderCard`: period change, daily average, spending/income ratio,
  top categories and top expenses now render from the existing owner-scoped
  insider response with privacy masking. Native unit coverage is 93/93 and the
  API-backed Reports UI smoke passes 1/1.
- Matched the web Budgets detail behavior with a shared native summary card:
  overall progress, budget/spent/remaining metrics, active daily-spend
  guidance, ended/future timing and 75%/over-budget statuses now render from
  the existing owner-scoped response. Native coverage is 94/94, the
  API-backed CRUD/status UI smoke is 1/1, and the Release warnings-as-errors
  build plus base-component guards pass; no API or migration changed.
- Matched the web AI Entry review header with a shared native `Lưu tất cả`
  action. It approves only pending proposals sequentially through the existing
  versioned owner-scoped endpoint, preserves partial statuses and surfaces the
  first failure without auto-retry. Native coverage is 95/95 and the pending
  review UI smoke passes 1/1; no backend or schema changed.
- Matched Overview recent transaction titles to the web category-first rule:
  native now resolves `category.name` and falls back to `Khoản thu`/`Khoản chi`
  while keeping notes as metadata. Category loading is optional and cannot
  block the screen; native coverage is 96/96 with Release and base guards
  passing, and no backend/API or migration changed.
- Matched Overview wallet subtitles to the web wallet-card rule: saved wallet
  descriptions now win, with the semantic wallet-type label as fallback. The
  shared row keeps an optional override for other consumers; native coverage
  is 97/97 with Release/base guards passing and no backend/API or migration
  change.
- Matched the Overview `Ví của tôi` action label to the web: native now uses
  the shared `Quản lý` action while preserving the existing wallet-management
  route. Native coverage is 99/99 with Release/base guards passing; no
  backend/API or migration changed.
- Matched the web credit-ledger empty state: a credit wallet with no statement
  rows now shows `Chưa có giao dịch tín dụng.` inside the shared ledger card,
  while populated rows and entry/payment actions remain unchanged. The native
  copy contract and API-backed UI route pass; no backend/API or migration
  changed.
- Matched the web Reports summary-card trailing action: native now displays the
  selected month beside `Tổng quan báo cáo`, while `Xem báo cáo` remains the
  explicit reload action. Native coverage is 100/100 with Release and
  base-component guards passing; no backend/API or migration changed.
- Matched the web Budgets empty branch: when no budget items exist, native now
  hides the overall summary card and renders the shared `MPEmptyView` with
  `Chưa có ngân sách.`. Native unit coverage is 101/101 and Release/base
  guards pass; no backend/API or migration changed.
- Matched the web Groups action anatomy on native Categories: the header is now
  navigation-only and the shared outline `MPButton` below it exposes `Nhóm mới`
  with stable ID `category-add-group`. The new outline treatment is a reusable
  base style; category CRUD/tree/API behavior is unchanged. Native coverage is
  102/102 and the Release warnings-as-errors/base guards pass.
- Matched the Stitch transaction editor form anatomy with shared
  `MPFormRow`/`MPDateStepper` bases: wallet, category, note and account-local
  date navigation now use the same icon/value/chevron rows, while the existing
  category/jar rules, API payload and safe-area save footer remain unchanged.
  Native unit coverage is 84/84 and the focused editor UI proof is 1/1.
- Aligned the shared native page inset with the Stitch/web container token:
  `MPMetrics.pagePadding` is now 16pt instead of a screen-local 20pt drift.
  All affected screens continue to use the shared base token and the semantic
  token regression asserts the value.
- Aligned filled native action buttons with the web/Stitch `rounded-full`
  contract through a shared pill-radius token; row and icon button geometry is
  unchanged.
- Removed the convenience-button system-font override so labels inherit the
  shared `MPFont.label` Manrope role.
- Completed the native base typography audit: Markdown, forms, selectors,
  async states, rows and UIKit-backed fields now use `MPFont`/Manrope roles;
  the base-component guard rejects future system-font drift.
- Completed the follow-up selector/date-field sweep: wallet, option,
  multi-option and compact date controls (including menu rows) now inherit
  the shared `MPFont.body` role instead of platform defaults. The final
  simulator unit/UI/build gates are recorded in `docs/work/VALIDATION_MATRIX.md`.
- Matched the Stitch overview header grouping: the visibility control now sits
  beside the balance amount while search and notifications remain in the
  trailing action cluster; existing accessibility IDs and routing are intact.
- Rebuilt the native More/profile landing surface from shared bases: centered
  identity/avatar, account subtitle, dedicated account-management row and a
  Stitch-style support pill in the header. Existing navigation contracts are
  unchanged.
- Historical implementation, later superseded: added the native Google sign-in
  entry point using the existing PKCE client and
  `/auth/mobile/google/exchange`; it stores the same access/refresh session
  as email login and reports missing iOS client configuration without making a
  network call. The focused unit/launch smoke passes 2/2; real Google Cloud
  redirect registration remains a device-UAT gate.
- Re-ran the complete native unit/contract target after the auth and base
  component audit: 83/83 passed on the iPhone 16 Simulator, with a Release
  warnings-as-errors build at `/tmp/mypocket-final-release`.
- Added the Stitch/web profile-card `Quản lý tài khoản` action to native More,
  reusing the shared menu-row base and existing owner-scoped account route.

- Fixed compact iPhone hit-testing in wallet/account secondary screens: wallet
  detail keeps its add-transaction CTA above the persistent shell, while the
  account destination list owns its bottom rows and disables the shell layer
  there. A narrow transparent guard now covers only the selected account-tab
  footprint, preventing repeat-tab taps from activating a nearby More row while
  keeping centered rows such as Hũ chi tiêu tappable. Wallet editor/mutation,
  jar mutation and the 15-destination route sweep pass in isolated simulator
  runs.
- Fixed the shared native UIKit text-field bridge so email and other text
  inputs remain at the 48px base-control height instead of stretching to the
  available form height. This restores the Stitch/web login layout without
  changing input bindings, accessibility identifiers or API behavior.
- Aligned the native bottom-navigation base with the Stitch reference: an
  inset rounded surface with soft elevation and a centered raised add action.
  Primary navigation and the short/long add gesture contracts remain intact.
- Release simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`, project
  guards and base-component tests remain green after the visual base changes.
- Reordered the native overview to match the supplied Stitch hierarchy:
  wallet summary, Money Insider, period report, month detail and recent
  activity. Added stable section identifiers for future authenticated visual
  comparison; all existing owner-scoped data and navigation callbacks remain
  unchanged.
- Native unit/contract coverage is now 76/76 after the overview section-order
  contract; the simulator Debug build and base-component guards remain green.
- Reworked native wallet detail composition around the Stitch ledger hierarchy:
  stable header/balance/summary/all-transactions regions, reference cards for
  goal/basic ledgers, and stable monthly ledger identifiers. Credit statement
  rendering and all wallet mutation contracts are unchanged.
- Native unit/contract coverage is now 77/77 after the wallet-ledger section
  contract; simulator build and base-component guards remain green.
- Reworked wallet create/edit around the Stitch form hierarchy: close/title
  header, grouped reference surface, shared conditional goal/credit fields and
  a persistent safe-area `Lưu` action. Existing validation and API payloads
  remain unchanged.
- Native unit/contract coverage is now 78/78 after the wallet-editor section
  contract.
- Reworked native Reports around the Stitch hierarchy: reference summary card,
  income/spending/net metrics, detail/trend segmentation, chart content and
  stable filter/summary/breakdown/activity identifiers. Existing month
  filtering, reload and owner-scoped report API behavior remain unchanged.
- Native unit/contract coverage is now 79/79 after the Reports section-order
  contract; the targeted Reports UI smoke passed 1/1 on the iPhone 16
  Simulator.
- Added owner-scoped account-action contract coverage for notification
  mark-read and API-key create/revoke. Tests assert the exact session header,
  snake_case create payload, idempotency key and 204 revoke handling without
  contacting a backend.
- Native unit/contract coverage is now 82/82 after the account-action
  contracts; Debug build and base-component guards remain green.
- Native category create/edit forms now render immediately below the screen
  header instead of after the full system category tree. The UI-test scroll
  helper now handles both directions and avoids relying on SwiftUI's transient
  `isHittable` value; category API payloads and owner scope are unchanged.
- The category tree is now flattened into stable depth-aware rows in a
  `LazyVStack`, reducing accessibility snapshot cost without changing tree
  order. Native unit/contract remains 82/82, and the real local API category
  create/update/delete UI flow passed 1/1 on the iPhone 16 Simulator.
- Revalidated the lazy category tree with an 82/82 native unit/contract run and
  a Release simulator build with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES`.
- Added a disposable local-API iPhone UI proof for Account API-key
  create/revoke; the flow uses a unique fixture name and teardown cleanup while
  preserving the existing session, idempotency and 204 contract tests.
- Verified the full native UI checkpoint on the iPhone 16 Simulator: 20/20
  owner-scoped API-backed cases plus 11/11 non-fixture smoke cases passed in
  bounded result bundles. The disposable database was migrated to version 29
  and teardown left no fixture wallets, transactions, budgets, categories,
  recurring schedules or travel events.
- Added the Stitch add-transaction form contract: centered close/title header,
  expense/income segmented control, wallet/category surface, VND amount
  context, optional detail section, account-local date and safe-area save bar.
- Reworked the native More/account screen into grouped Money Lover-style
  sections with a profile surface while retaining all 15 route IDs and
  owner-scoped navigation behavior.
- Added red-to-green UI coverage for the new form controls and made grouped
  More-route tests reveal scrollable entries before tapping them.
- Added deterministic iPhone UI coverage for Finance Assistant suggested
  questions, report month filtering/reload, and the account API-key editor.
- Added stable accessibility identifiers for report inputs, notification
  mark-read actions, and API-key create/revoke controls so later authenticated
  UAT can target the exact owner-scoped action.
- Native reports now show a shared overview card with income, spending, net and
  counted ledger activity, and Finance Assistant mirrors the web's first-use
  suggested prompts.

## Native Stitch account and transaction parity — 2026-10-07

- Introduced `MPReferenceScreenHeader` and extended `MPScreenHeader`/`MPSegmentedControl`
  with stable accessibility contracts so screens can adopt the Stitch hierarchy
  without changing existing route identifiers.
- Kept all transaction business rules intact: wallet/category applicability,
  active hũ validation, report inclusion, notes, edit/save errors and API
  payloads are unchanged; only the form composition and save placement moved.
- Reconciled iOS screen mapping and backlog docs with the actual
  `Features/Secondary/SecondaryFeatureViews.swift` layout and recorded the
  external physical-device/provider/pixel-signoff gates.

## Unified Money Lover visual port — 2026-10-07

- Ported the captured Money Lover hierarchy into the shared web and native
  renderers: large balance header, wallet icon rows, Money Insider insight,
  week/month report card, airy reference surfaces and persistent center add
  action.
- Kept MyPocket behavior/API as the source of truth. Existing transactions,
  wallets, jars, budgets, reports, AI entry, Finance Assistant, feedback and
  account routes remain mounted; the new cards only consume owner-scoped live
  data and link to existing routes.
- Added semantic reference-card tokens and native/web base components instead
  of screen-local styling. The attached Stitch archive is now recorded as the
  visual source contract in `docs/design/references/STITCH-MY-POCKET.md`; no
  source HTML or assets are copied into the runtime bundle.
- Hardened the shared native `MPTextField` with a UIKit-backed input bridge,
  stable accessibility identifiers and a deterministic clear action. This
  prevents SwiftUI focus updates from dropping characters or appending edits to
  stale text buffers in wallet, budget, category and recurring editors.
- Added explicit API-backed simulator coverage for recurring edit persistence:
  the test verifies the updated name through the local API before checking the
  rendered row, then runs the due schedule and removes its fixtures.
- Final simulator gate passed 102/102 tests with 0 failures and 0 skips on the
  iPhone 16 Simulator; Release static analysis, project guards and web
  design/type/build/auth checks are green. Production and physical-device
  gates remain intentionally separate.

## Production backend migration and rollout — 2026-10-06

- Created and validated a pre-migration PostgreSQL custom-format backup on
  `vm100-prod` before applying the forward migrations.
- Applied production migrations from version 27 through version 29 and
  verified the database is clean (`dirty=false`). No production data reset or
  destructive operation was used.
- Rolled `mypocket-api` to immutable release image `release-75f7f07` at digest
  `sha256:7718c93a141f91711f58b9dcb5f0b5d716eae738157de44daf23b4653f1abcdb`;
  the final API and worker deployment is `release-73c4c37` at digest
  `sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2`.
  The rollout completed with automatic rollback enabled and the public health
  and OpenAPI endpoints returned HTTP 200.
- Resolved the Redis network issue without restarting Redis: added the
  attachable overlay `mypocket-redis-net`, connected `redis-prod` with the
  stable `redis-mypocket` alias, verified `PONG` from `vm100-prod`, and
  switched the API service off the temporary container IP workaround.
- Source for the deployed image is commit `75f7f07` on
  `codex/native-ios-port`; this entry records the production rollout and its
  verification evidence.

## Recurring worker enablement — 2026-10-06

- Added `/app/worker` to the immutable API image and published
  `release-73c4c37` at digest
  `sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2`.
- Ran a one-shot production worker pass on `vm100-prod` with
  `WORKER_ONCE=true`; the active-schedule query completed without an error.
- Scaled `mypocket-worker` from `0/0` to `1/1`, using the stable Redis overlay
  and `/app/worker` entrypoint. The API stayed healthy throughout the update.
- Updated `mypocket-api` to the same `release-73c4c37` image so API and worker
  share one reproducible binary set.

## Native email login slice — 2026-10-06

- Native iOS now signs in through `POST /api/v1/login` using the account email
  as the username and a user-supplied password. Tokens continue to use the
  existing Keychain + rotating refresh-token contract; passwords are not stored.
- Removed the user-facing Google login control from the native screen. Google
  backend routes remain available during the migration window for browser
  compatibility and rollback; production Google shutdown is not claimed.
- Verified the password API contract and session persistence with targeted
  XCTest, the full 73-test native unit suite, base-component guards and the
  iPhone 16 Simulator UI target (6 passed, 20 intentional fixture skips). A
  synthetic invalid-credential probe confirmed production `/api/v1/login`
  returns `401 LOGIN_FAILED`; no production account was mutated.

## Simulator-first regression — 2026-10-06

- Full `LOCAL_API_UI` run passed 20/20 native UI cases on the booted iPhone 16
  Simulator, including feedback screenshot-storage fallback and owner-scoped
  mutation cleanup.
- Hardened the portfolio price-refresh UI test to wait for the refreshed
  formatted price instead of a pre-existing card title; the follow-up full run
  passed 20/20 with zero failures.
- Backend normal/race/vet checks, Release analysis, base-component guards and
  the web design/type/domain/build suites passed. Physical-device and
  production-provider gates remain separate.

## Field Ownership

- Human owns release approval.
- AI maintains change entries and links to release notes.

This file is the first-release baseline for the current MyPocket snapshot. Earlier
development notes remain available in Git history and are intentionally not presented
as active release entries here.

## [Unreleased] - 2026-10-07

### Fixed

- Native iOS design-system parity: primary CTAs use the Stitch `#4caf50`
  brand token, action/focus colors use `#006e1c`, and shared fields match the
  web/Stitch card, border, spacing and focus-ring treatment.
- Bundled the OFL-licensed Manrope variable font from the Google Fonts
  repository and switched `MPFont` to the registered font family, removing
  the previous system-face fallback from the native reference renderer.
- Centralized native typography in `MPFont` using the Stitch measured scale
  (32/20/17/15/13/14/12/24); the native app registers the bundled Manrope
  variable font through `UIAppFonts`.
- Ported the Stitch/Money Lover reference hierarchy into web/native overview
  surfaces, including the live Money Insider and week/month report cards.
- Verified the shared native add-action contract on the iPhone 16 Simulator:
  a short tap opens the manual editor and a 500ms hold opens AI entry without
  the follow-up tap action overwriting that route.
- Fixed native bottom navigation sizing so its safe-area surface and centered
  add-transaction action stay at the bottom of the viewport and never cover
  secondary menu rows; the complete More-route simulator sweep is green.
- Completed the native wallet-management handoff: the wallet list now has a
  shared back action when opened from More, and selecting a wallet routes to
  the account-scoped transaction ledger while the explicit “Chi tiết” action
  continues to open the specialized wallet detail surface.
- Added the credit-limit metric to both native credit entry points so the
  statement surface matches the web credit contract alongside debt, available
  credit, amount due and payment status.
- Ported the shared add-action long press: a 500ms hold opens AI transaction
  entry while a normal tap continues to open the manual transaction editor.
- Debug iOS builds explicitly enable SwiftUI Live Preview/hot reload settings;
  Release API configuration is unchanged.

- Fixed feedback submission when development runs without private S3 storage.
  The backend no longer panics on a typed-nil storage interface; it returns a
  controlled screenshot error, and both web and iOS preserve the feedback with
  a text-only fallback. Configured production storage still receives the
  captured PNG.

- Added real Chromium feedback E2E coverage through the Vite proxy, including
  capture, controlled storage failure, JSON fallback, persisted owner state and
  fixture cleanup.

- Replaced the disabled savings-wallet notification toggle with a persisted
  `goal_notifications_enabled` preference, owner-scoped `GET /notifications`
  inbox and idempotent mark-read route. Web and iOS use shared/base controls;
  notification timestamps use the authenticated account timezone; push/APNs
  and scheduled fan-out remain separate follow-up work.

- Native account timezone changes now update the persisted authenticated session
  context immediately, so notification and calendar surfaces use the new IANA
  timezone without requiring a logout or app restart.

- Native wallet management now matches the web wallet-scope contract: tapping
  a wallet row opens the transaction ledger with that wallet selected, while an
  explicit `Chi tiết` action remains available for balances, progress, credit
  statements and transaction editing. Clean iPhone 16 Simulator verification
  passed 69/69 native contracts and 19/19 API-backed UI tests with no leftover
  fixture wallets.

- Fixed native portfolio summary decoding for `asset_id`, `realized_pnl`,
  `unrealized_pnl` and total P/L fields. The API-backed iPhone 16 Simulator
  test now creates an owner-scoped asset, updates its manual price through the
  native UI, verifies the refreshed value and removes the unused fixture.

- Native Finance Assistant and AI-entry replies now use the shared
  `MPMarkdown` base renderer for headings, emphasis/code, lists, fenced code
  and pipe tables. Raw assistant Markdown is no longer shown as an unreadable
  delimiter string; native parser coverage is included in the 69/69 contract
  suite.

- Finance Assistant requests now prepend a bounded, versioned MyPocket product
  handbook. It grounds feature/navigation answers, routes financial claims to
  owner-scoped tools and allows only a validated wallet-creation proposal; the
  native/web confirmation cards call the existing wallet API explicitly, while
  category/budget/jar/transaction writes remain blocked. The orchestrator
  contract tests cover this system-context injection.

- Added `propose_create_wallet`, a strict no-write advisor tool that returns a
  confirmation card for basic, savings and credit wallets. Web and iOS use the
  existing owner-scoped wallet API only after an explicit confirmation.

- Wallet creation money fields now use the shared web `AmountField` base with
  formatted VND display and calculator input. Goal-wallet notification
  preferences are persisted and rendered through the shared/base controls;
  push/APNs delivery remains a documented follow-up.

- Fixed the native wallet-detail create/edit regression. The custom snake-case
  decoder now maps API `wallet_ids` to Swift `walletIDs`, category metadata no
  longer drops the detail screen into an error state, and the add action stays
  disabled until required metadata is ready. The iPhone 16 Simulator proof is
  green: 67/67 native contracts, 17/17 local API-backed UI tests and 0 failures
  in the default UI suite (11 intentional skips).

- Transaction category integrity is now enforced across the database, API and
  web/iOS editors. Ordinary income/expense, recurring and credit entries must
  use an applicable visible group; clients prefill the first valid group, while
  adjustment rows remain the explicit uncategorized exception. Migration
  `000028_transaction_category_required` backfills the canonical fallback
  groups before adding the database check.

- AI-entry live regression coverage now matches the approved review-first
  fallback: an unclear wallet is prefilled with the first supplied wallet and
  still carries a review question. DeepSeek-V4-Flash text, tool-call and
  strict-JSON-schema probes pass with synthetic data; no ledger writes occur.

- OCR/AI live pipeline verification now follows the production image contract:
  private S3 PUT/signed-readback, OCR `mode=wait` Base64 submission and
  DeepSeek-V4-Flash extraction pass 5/5 cases with 24/24 semantic fields
  (100.0%). Receipt images are classified as expenses unless the source
  explicitly describes income, refund or reimbursement received; no ledger
  writes occur in the probe.

- Tightened extraction classification for internal transfers: one explicitly
  stated transfer event keeps one draft and one amount; the live JSON Schema
  probe now passes the transfer-splitting regression as well as the other four
  synthetic cases.

- Native AI-entry attachment regression coverage now verifies twenty valid PNG
  thumbnails, the picker cap, close/remove behavior and picker re-enable on the
  iPhone 16 Simulator; the signed UI suite passes 17/17.

- Native base-component hardening moved remaining feature-level raw text/icon
  rendering into `MPText`/`MPIcon`; the guard now rejects raw `Text` and
  `Image` in feature code, and the full signed UI suite remains 17/17 after
  correcting decorative-icon accessibility labels.

- Native feedback capture now hides the floating feedback trigger for one
  rendered frame before taking the attached screenshot, preventing the
  feedback control from appearing in its own context. The signed simulator UI
  regression passes 16/16.

- Native feedback screenshots now use the backend's accepted bounded PNG
  contract; complex screen captures are downscaled to 1 MiB or less and sent
  as `screen.png`/`image/png` instead of rejected JPEG or oversized parts.

- Native recurring numeric drafts now validate thousands grouping strictly:
  valid values such as `5.000.000` remain supported, while mixed characters,
  whitespace inside a number and malformed separators are rejected instead of
  being silently stripped before a schedule write.

- Native iOS simulator stability: replaced the SwiftUI number-pad keyboard
  toolbar in the shared amount field with a UIKit input accessory. This keeps
  the `Xong` dismissal action and removes the invalid-frame runtime fault seen
  while scrolling mutation forms. The four affected API-backed UI tests pass
  4/4 on the iPhone 16 Simulator.

- Native wallet detail now preserves immutable ledger semantics: ordinary rows
  open the detail action flow for edit/delete, transfer rows use paired
  transfer actions, adjustment rows are read-only with direction-aware signed
  amounts, credit rows stay in the credit statement, and Travel Mode linking
  is limited to ordinary non-credit rows. The simulator contract suite is
  62/62 and the API-backed UI suite is 15/15.

- Native AI proposal editing now matches the ordinary transaction field
  contract: account-local date/time is preserved, stale existing hũ selections
  remain visible and require explicit resolution, and transfer-out categories
  cannot receive a hũ. Report inclusion remains editable and is sent in the
  proposal patch.

- Native ordinary transaction editing now matches the web form contract for
  account-local date/time, report inclusion and hũ assignment. Eligible basic
  and goal expense wallets show active account-month hũ choices; inactive
  existing assignments stay visible and are rejected explicitly until cleared
  or replaced. Type, wallet and category changes clear incompatible hũ values.
  Policy tests, the manual editor UI path and the complete signed iPhone 16
  `LOCAL_API_UI` suite pass (61/61 unit, 15/15 UI).

- The native local API UI harness no longer depends on Keychain entitlements
  when the simulator app is built unsigned. `-UITestAuthenticatedAPI` now uses
  a DEBUG-only in-memory session store, while normal and Release builds retain
  Keychain-backed sessions. This fixes the false login-gate failure after a
  successful disposable OAuth exchange (`errSecMissingEntitlement -34018`).
  The complete iPhone 16 `LOCAL_API_UI` suite passes 15/15.

- Native AI transaction-entry review now treats unknown/transfer model output
  as `Chưa xác định — chọn loại` until the user explicitly selects thu/chi.
  The editor exposes shared selectors for type, wallet, category and active
  month hũ, plus an explicit report-inclusion choice; invalid category/hũ
  combinations are cleared before the existing proposal PATCH/approval flow.
  The deterministic iPhone 16 review fixture and full signed `LOCAL_API_UI`
  suite pass (15/15); native XCTest is 59/59.

- Native Travel Mode now matches the web contract for active-event context,
  explicit date scope, safe delete confirmation, and the rule that only new
  ordinary income/expense transactions auto-link to the active trip.

- Native recurring schedules now show account-timezone next-run dates,
  localized type/frequency, wallet, note and active state; the screen explains
  ordinary-transaction semantics and confirms deletion without removing rows
  already created in the ledger.

- Native reports now include the month-summary category/type/count breakdown and
  the ledger activity explanation alongside Insider metrics, while keeping the
  Insider view usable when the optional month-summary request fails.

- Native credit statements now show positive current debt, localized payment
  status and credit-entry kinds, optional due-date context, and note-first row
  subtitles through shared base components.

- Native portfolio asset cards now show weighted average cost, realized and
  unrealized P/L, current value or explicit unknown pricing, plus account-timezone
  buy/sell history with quantity, price, fee and note context.

- Native user-facing labels now translate wallet, category, credit, feedback and
  Assistant status values through one shared vocabulary; API enum values remain
  unchanged on the wire.

- Native Tháng summary now decodes jar usage, uses bounded account-local month
  selection, shows period/status/category/jar details, protects dirty notes,
  and opens the selected month in Hũ chi tiêu.

- Native feedback submission now trims title/description and rejects
  unsupported feedback types before JSON or screenshot multipart requests.

- Native portfolio asset drafts now trim symbol/name values before submission,
  reject whitespace-only values, and expose a shared-header refresh action.

- Native goal-wallet detail now matches the savings summary contract: it shows
  reached/remaining amount, account-local target-day status, bounded progress,
  and the explicit target amount using shared base components.

- Native Hũ chi tiêu now uses a bounded account-local month selector and
  shared summary metrics for actual income, actual spending, allocated and
  unassigned spending. It warns (without blocking ledger writes) on
  over-allocation, keeps historical month configurations read-only, and adds a
  selected-jar cumulative report with stale-state clearing on month changes.

- Native category management now matches the backend/web scope contract:
  system categories cannot be deleted or have metadata rewritten, while
  applicable wallet IDs can be updated through the dedicated wallet-scope
  endpoint. Personal categories keep full editing controls.

- Native wallet editing now carries an explicit create/edit presentation payload,
  so an existing wallet reliably sends `PATCH /wallets/:id` instead of opening
  a create draft. The shared amount-field base adds a keyboard “Xong” action so
  credit notes remain reachable while the number pad is open.
- Added signed iPhone 16 simulator business-mutation coverage for wallet
  create/update, atomic transfer create/delete and credit entry/payment
  refresh. The complete `LOCAL_API_UI` suite passes 14/14 and leaves zero
  disposable wallets, transactions or feedback rows.
- Native transaction detail now selects Travel Mode events by stable ID rather
  than display name, and balance adjustments now preserve the account-local
  calendar day instead of the device timezone. Native XCTest is 48/48 and the
  signed simulator UI suite remains 14/14.
- Native Wallet Detail now has a signed simulator mutation smoke that creates a
  real expense through the API, verifies the persisted row, edits its note, and
  verifies the updated row. The complete API-backed iPhone 16 UI suite is now
  11/11; the disposable fixture wallet is removed in teardown.
- Native financial draft fields now reject malformed integer input instead of
  silently stripping non-digit characters. A shared strict parser covers
  transaction, transfer, adjustment, wallet, credit, jar, budget, assistant
  proposal and portfolio price/trade paths while preserving explicit recurring
  display normalization. The regression suite is now 47/47 unit tests and
  10/10 signed iPhone 16 API-backed UI tests.
- Re-ran the native iPhone 16 simulator regression after the goal target-date,
  recurring, jar-allocation and portfolio-trade validation fixes: 45/45 unit
  tests and 10/10 API-backed UI tests pass, with no
  disposable `Fixture*` wallets left behind. API-backed UI tests retain normal
  simulator ad-hoc signing so Keychain fixture sessions can be persisted;
  `CODE_SIGNING_ALLOWED=NO` remains for pure unit/build checks only.
- Native recurring editors now reject non-positive amount/interval values and
  unsupported type/frequency or reversed end-date drafts before sending an API
  request; the contract suite now passes 45/45 with the portfolio validator.
- Native jar allocation drafts now mirror the backend contract: fixed values
  stop at the JavaScript-safe integer ceiling and percentage values must be
  finite, positive, at most 100% and no more precise than two decimals.
- Native portfolio trade drafts now validate buy/sell, positive quantity
  precision and non-negative integer VND price/fee values locally, avoiding
  silent digit stripping before the API request.
- Native API requests now carry a UUID `X-Request-ID`; a JWT refresh retry
  keeps the same correlation ID so server errors can be traced end to end.
- Native transfer forms now exclude credit wallets and serialize the selected
  transfer day in the account timezone instead of the device timezone.
- Native wallet detail now refreshes the wallet summary after loading and keeps
  valid transaction rows visible when optional category metadata is unavailable.
- Native UI smoke tests now start from a deterministic unauthenticated state
  with the DEBUG-only `-UITestUnauthenticated` fixture, clearing only simulator
  Keychain/biometric state. The iPhone 16 suite passes 10 tests (5 passed, 5
  explicit API-fixture skips).
- Opt-in native API-backed UI tests now seed and tear down their own basic and
  credit wallet fixtures through the local API; the full `LOCAL_API_UI` suite
  passes 10/10 without relying on leftover simulator or database state.
- Native AI-entry attachment selection now stays disabled until private
  OCR/storage capability is configured; text-only input remains available.
- Native aggregate/basic ledgers now match the approved Week/Custom contract:
  Monday–Sunday account-local ranges, bounded previous/next week navigation,
  no all-time tab, and shared `Tiền vào`/`Tiền ra` cashflow summary.
- Native goal-wallet target dates now round-trip as account-local calendar dates,
  avoiding UTC/device timezone day shifts in the editor.
- Native session restore now refreshes an expired access token before enabling the
  Face ID lock gate. If the refresh token is expired or rejected, Keychain state
  and the biometric preference are cleared instead of opening an invalid session.
- Native category editing now preserves all applicable wallets. The shared
  multi-select base uses an empty `wallet_ids` selection for the documented
  all-wallet scope, so editing a multi-wallet category no longer silently keeps
  only the first wallet.
- Shared native buttons now apply a rectangular hit shape to their rendered
  control bounds, fixing compact “Thêm” rows that could appear visible but fail
  to open a destination in simulator/device UI automation.
- Native ledger period filters now exclude rows with malformed timestamps instead
  of accidentally including them, wallet detail no longer truncates the API's
  transaction history at 50 rows, and Finance Assistant keeps a still-running
  run cancellable when its local polling window expires.
- Debug simulator builds now expose `-UITestAuthenticatedAPI`, which exchanges
  the local fixture OAuth payload and persists a disposable session so an
  authenticated overview smoke uses the real URLSession/API path instead of a
  fake token.
- Native transactions now support the web parity bulk-delete flow with base
  selection controls, 100-row/immutable-row guards, atomic confirmation and a
  visible error that preserves the selection. Typed `204 No Content` responses
  are accepted for delete mutations instead of being decoded as JSON.
- Native Finance Assistant now restores one owner-keyed conversation, loads the
  latest 50 messages and read-only monthly overview on entry, pages older
  messages, and renders structured overview/source data without storing message
  text in UserDefaults. The API-backed simulator smoke covers the disabled
  capability response and the live overview contract.
- Native credit ledger entry/payment forms now match the web fields: entry kind,
  applicable category, payment source wallet, amount and optional note are all
  preserved in the credit API payload.
- Native basic/goal wallet detail now matches the web detail flow with
  account-timezone monthly groups, category/note labels, income/expense totals,
  and shared transaction add/edit actions.
- Native transaction ledger now uses the shared overflow menu for period,
  grouping, transfer, adjustment and bulk actions; rows show category/wallet
  context and signed amounts, while goal/credit scopes hide ordinary period
  controls.
- Selecting a credit wallet from the main native wallet list now opens the same
  credit statement, entry and payment actions as the dedicated credit route.
- Fixed native API query construction so credit statements, reports, jar views
  and Assistant history send query parameters as URL queries instead of
  accidentally escaping the delimiter into the path.
- Native transaction filtering now switches selected goal and credit wallets to
  their dedicated detail surfaces, so credit wallets no longer expose the
  ordinary transaction editor from the ledger tab.
- Native ledger loading now keeps valid rows visible when optional category or
  profile metadata fails; only the required transaction/wallet payload blocks
  the screen.
- Native portfolio now shows current/unrealized metrics, keeps price drafts
  independent per asset, allows clearing an unknown price, and preserves the
  selected trade date.
- Native recurring schedules now exclude credit wallets and filter categories
  by expense/income type, clearing stale category selections when the type
  changes.
- Native selectors now render localized semantic labels while preserving raw API
  enum values, and jar allocation mode changes clear stale drafts and reject
  zero-percent allocations to match the web validation contract.
- Native ordinary transaction and AI proposal editors now exclude credit
  wallets and clear incompatible categories on wallet/type changes; aggregate
  credit rows remain read-only and route mutations through the credit ledger.
- Native budget and balance-adjustment selectors now exclude credit wallets, and
  budgets only offer expense categories, matching the web scope rules.
- Native ended budgets are now view/delete-only; the editor no longer exposes a
  stale save path after the budget's end date.
- Native portfolio trade dates now use the account timezone and local-noon
  serialization, preventing device-timezone day shifts.
- Native wallet detail month headers now render real calendar labels, and Travel
  Mode rejects whitespace-only names, malformed dates and reversed date ranges
  before submission.

### Verification

- Native iOS simulator regression passed on iPhone 16: 45 unit/contract tests and
  5 standard authenticated UI smoke tests, with base-component and Xcode project
  guards. The Release simulator build also passed with warnings treated as
  errors, and the installed app launched successfully.
- An opt-in API-backed wallet-detail UI smoke passed with a seeded disposable
  wallet, opened the transaction editor fields, and deleted the fixture with a
  204 response afterward.
- The opt-in API-backed Assistant UI smoke passed on the same iPhone 16 using
  `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`, the disposable local
  API, and the `-UITestAuthenticatedAPI` fixture; no production credentials or
  data were used.
- The API-backed simulator fixture rendered the authenticated overview against
  local `:8080` with empty owner data; no production credentials or data were
  used.
- The authenticated UI sweep opened and returned from all 13 destinations in the
  “Thêm” menu; this caught and fixed the compact-row hit-target regression.
- The opt-in API-backed credit-wallet smoke passed on the iPhone 16 simulator,
  including the Giao dịch-tab and Ví-tab routes plus statement action buttons.
  Its disposable credit fixture was deleted with an HTTP 204 response after
  the run.
- Refresh rejection during an authenticated request now clears the native
  Keychain session and biometric preference through an auth invalidation hook;
  the regression suite covers the 401 → refresh 401 path.
- Backend race/integration verification passed against disposable PostgreSQL and
  Redis. Transfer browser/integration harnesses now cover linked update/delete
  semantics and idempotency-key cleanup.
- Feedback was verified through JWT, scoped user API key, service-token agent
  lifecycle and changelog publication; Redis feedback audit events were observed.
- AI-entry browser regression fixtures now generate ledger rows in the current
  account-local week, preventing calendar drift from hiding the expected grouped
  rows in future runs.
- Backend test fixtures were normalized with `gofmt`; `go vet ./...` and the
  warning-as-error Swift simulator build remain green.

### Release gates still open

- Provider-backed AI/OCR, real Google iOS OAuth credentials, authenticated
  staging/production smoke and physical-device validation remain required before
  a production release claim.

## [1.0.8] - 2026-10-01

### Fixed

- Finance Assistant now accepts a bounded batch of up to eight read-only tool
  calls from one model response. This prevents valid multi-query answers from
  failing when the provider returns five or more parallel finance lookups.

### Verification

- Added a regression test covering five tool calls in one provider response;
  backend tests and the frontend release gate remain required before rollout.

### Deployment

- API image `registry.dungxbuif.com/mypocket-api:release-46c48f91`
  (`sha256:2be9c9a189e7caee55f8b6fdd9e16a0a3dec1eaa4fff55f10501f65dcfded226`)
  is running in production; the service is 1/1 and `/api/v1/health` returns 200.

## [1.0.7] - 2026-10-01

### Added

- Travel Mode now has owner-scoped event CRUD, one active event per account,
  ordinary transaction auto-linking, and explicit transaction link/unlink.
  Deleting an event clears links without deleting ledger rows or changing
  report inclusion. Recurring, transfer, adjustment and AI approval paths do
  not auto-attach events.
- Credit entry categories now enforce applicable-wallet scope in both the
  handler and repository transaction.

### Verification

- `go test ./...`, `go generate ./cmd/api`, `npm run check:design`,
  `npm run typecheck`, `npm run test:design`, `npm run test:transactions`,
  `npm run test:calendar`, `npm run test:transaction-jars`, and `npm run build`
  pass locally. PostgreSQL integration tests remain environment-gated by
  `TEST_DATABASE_URL`.

### Deployment

- No production deployment is claimed by this local implementation commit.

## [1.0.6] - 2026-09-30

### Fixed

- Transaction ledger wallet scope now follows the supplied wallet-selector reference:
  the compact wallet capsule opens the shared `Chọn Ví` sheet with aggregate,
  included/excluded wallets, real balances, add-wallet and disabled-link states.
  Selecting a wallet keeps the existing account-local week/custom ledger behavior.

### Verification

- `npm run test:wallet-scope`, `npm run check:design`, `npm run typecheck`,
  `npm run test:design`, `npm run test:transactions`, `npm run test:calendar`,
  `npm run test:transaction-jars`, and `npm run build` pass locally.

### Deployment

- Production web service converged on
  `registry.dungxbuif.com/mypocket-web:feedback-ui-wallet-design-20260930-0700`
  (`sha256:933dd62630e3ee5ae7917c6ddaab36b3d858bc58f2fdb0abb87cc2dd5e36d80e`).
  Public HTML serves the new bundle and `/api/v1/health` returns 200.

## [1.0.4] - 2026-09-30

### Fixed

- AI transaction-entry proposals now normalize date-only model output (`dd/MM/yyyy`,
  `dd-MM-yyyy`, or `yyyy-MM-dd`) in the account timezone instead of silently replacing
  a recognized date with today; the missing-time assumption is shown for review.
- Exact duplicate proposals within one OCR/text submission are merged while preserving
  review questions from the duplicate rows.

### Deployment

- API image `registry.dungxbuif.com/mypocket-api:feedback-api-ai-date-20260930-0011`
  was rolled out to production and verified healthy.

## [1.0.5] - 2026-09-30

### Added

- Login and Google OAuth responses now include rotating refresh credentials;
  `POST /api/v1/auth/refresh` consumes a refresh token once and issues a new
  access/refresh pair.
- The web client retries one expired-session request through the refresh route
  and keeps API-key requests out of the refresh flow.

### Deployment

- API `feedback-api-refresh-20260930-0041` and web
  `feedback-ui-refresh-20260930-0032` were rolled out and their public health
  and invalid-refresh checks passed.

## [1.0.2] - 2026-09-30

### Fixed

- Increased the frontend API proxy read/send timeouts to 240 seconds so the
  bounded OCR/model processing window does not turn into an HTML 504 response.

## [1.0.3] - 2026-09-30

### Fixed

- Category search now promotes a parent group when the query matches a child
  category, so matching groups are visible at the top of the result.

## [1.0.1] - 2026-09-29

### Fixed

- Recent transactions on the overview now show the account-local date and time
  before wallet and note metadata.
- Feedback routes accept owner-scoped API keys with `feedback:read` and
  `feedback:write` scopes.

### Deployment

- Production API and web services were rolled out to the verified registry
  images after the backend and frontend checks passed.

## [1.0.0] - 2026-09-24

### Deployment note

- Homelab `money.dungxbuif.com` was redeployed on 2026-09-24 from the
  `feature/wallet-management` branch. Production PostgreSQL schema was reset
  and migrated to version 21; Redis DB 1 was flushed. A pre-reset PostgreSQL
  dump was kept on Pi5 under `/home/dungxbuif/backups/mypocket/`.
- Production now uses Docker images `homelab/mypocket-api:0ff84828-deploy-20260924015529`
  and `homelab/mypocket-web:0ff84828-deploy-20260924015529`. The old
  `mypocket-worker` Swarm service is scaled to zero because this code snapshot
  has no `cmd/worker` binary.
- Production AI entry, receipt OCR, and Finance Assistant are enabled through
  server-side environment variables. MyPocket reuses the shared homelab LLM and
  OCR credentials; no provider keys are exposed to the browser or committed docs.
- The production web proxy accepts the same 101 MiB AI-entry upload envelope as
  the backend so receipt images/PDFs are not rejected before OCR processing.
- The production client now normalizes account and AI-entry timezones before
  submission, falling back to `Asia/Ho_Chi_Minh` when a browser/webview reports
  an empty or non-IANA timezone such as `Local`.

### Added

- Authenticated wallet, category, transaction, budget, jar, account-timezone, feedback,
  and local Finance Assistant slices are present in the current Stage v1 codebase.
- AI transaction entry supports text and receipt review proposals with wallet/category
  suggestions, duplicate handling, explicit user review, and bulk save controls.
- Finance Assistant exposes owner-scoped read-only finance tools, conversation history,
  typed cards, local API-key scopes, and best-effort Redis access audit metadata.
- Transaction ledger supports aggregate/basic-wallet weekly and custom periods, savings
  wallet history, provisional credit history, wallet filtering, and internal transfer
  entry from the transaction options menu.
- Design documentation now follows `system / atoms / molecules / organisms / templates /
  pages / references`; owner PNG/HTML bundles are retained as non-runtime evidence.
- Native More profile help now opens the existing feedback/changelog surface;
  the visible help affordance no longer silently ignores taps.
- Wired the native overview header notification bell to the existing
  owner-scoped notifications screen; search remains presentation-only until a
  shared web/API contract is introduced.
- Added the shared native `MPBottomNavigationInset` base to Overview, More and
  secondary scroll screens so content reserves the persistent shell navigation
  area consistently on compact iPhones.
- Added a fixture-backed Overview reachability regression that scrolls to the
  final recent-transactions section and verifies it remains above the shell.
- Wired the native Google PKCE client into the shared auth screen; successful
  mobile exchange now uses the same persisted refresh-token session path as
  email/password login, while missing OAuth configuration fails clearly.
- Native unit/contract coverage is now 83/83 after the Google auth wiring.
- Removed the last screen-level native font override from the Overview metric;
  all native feature typography now resolves through `MPFont`.

### Changed

- The active release narrative is reset to `1.0.0`; old release prose is retained only
  in Git commits, not carried forward as a second active release history.
- Design contracts use page terminology for route-level specifications and a dedicated
  references tree for visual evidence.
- AI provider/model configuration remains environment-driven; credentials are not stored
  in documentation or source control.

### Fixed

- Corrected wallet-scoped weekly ledger filtering so totals and rows use the same
  account-local period predicate.
- Corrected AI-entry empty/503 failure paths for bounded model output and nullable usage
  initialization; incomplete proposals remain reviewable instead of being auto-saved.
- Corrected OCR attachment handling for validated image/PDF inputs and bounded wait/poll
  behavior without exposing provider secrets or financial payloads in logs.

### Documentation

- Added the approved release-baseline detail design and inventory.
- Reconciled design indexes, page contracts, system behavior links, work tickets, backlog,
  validation matrix, context, and agent-readable transaction guidance.

### Known release gates

- Production deployment and public release approval are not claimed by this entry.
- Finance Assistant SSE/reconnect, durable Redis audit delivery/alerting, configured-
  provider browser proof, and full owner UAT remain open where the validation matrix says
  pending or not run.
- Credit-card statement/payment accounting, recurring transactions, advanced reports,
  and Travel Mode runtime remain separate incomplete scopes.
- Frontend `npm run check:design`, lint, build, and tests must be rerun in an environment
  with Node installed; this cleanup environment reported `env: node: No such file or
  directory`.

### Native simulator deterministic-session regression — 2026-10-06

- Authenticated UI tests now use an in-memory session for both deterministic launch
  modes, preventing stale Keychain credentials from reaching real or fixture data.
- The More-route sweep reuses one authenticated process instead of terminating and
  relaunching between destinations, avoiding the Xcode 27 simulator runner kill
  while still exercising every destination and back action.
- Native contracts remain 65/65; all 9 API-backed flows pass as isolated local-
  fixture invocations.

Finance Assistant responses now render through a safe shared Markdown molecule:
headings, lists, inline emphasis/code and pipe tables are semantic UI instead of
raw model delimiters. Model HTML is never injected. The advisor history and
rendering contracts pass via `npm run test:ai-advisor`; design guard, typecheck
and production build remain green.

Wallet selection now has one ledger destination: choosing a wallet from account
management navigates to `/transactions?wallet_id=<id>` and applies the same
wallet-scoped weekly ledger used by the transaction selector. The old separate
wallet-detail route is no longer opened by that selection path. Coverage is in
`npm run test:wallet-scope`, with design, typecheck and build gates rerun.
