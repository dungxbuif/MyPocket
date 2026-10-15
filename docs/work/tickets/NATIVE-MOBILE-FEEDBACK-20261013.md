---
artifact_type: ticket
id: NATIVE-MOBILE-FEEDBACK-20261013
status: in_progress
owner: shared
source: owner request to continue production feedback fixes on iOS
platform: ios
---

# Native iOS feedback follow-up — 2026-10-13

## Current reconciliation — 2026-10-14

The intake and scope statements below are historical. The latest authenticated
production API fetch returned 48 open feedback records; [current
classification](NATIVE-FEEDBACK-PROD-AUDIT-20261014.md) lists the current queue
and genuine remaining work. Owner explicitly approved the separate additive auth
backend rollout and device installation, preserving financial data.
[Rollout](NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md) is deployed; after owner unlock,
strict-signature verification, device installation and launch succeeded.
Release publication still awaits owner
acceptance, and no production feedback status has changed.

Current shared cards are flat with 12pt padding (superseding earlier elevated
Assistant cards); feedback uses overlay/no reserved dock; compact profile, shared
row/typography and optional image picker are implemented. Final focused proof:
169 unit + 4 compact UI, and 8 API-backed UI; [validation](../VALIDATION_MATRIX.md).
Reconciled2026-10-15: dedicated note editor is implemented locally, not an open
implementation gap (TransactionsView note sheet, UI RED→GREEN2/2 and API note
readback in validation). Disabled Account add-circle/other placement and older
feature requests still require their own evidence. Hardware biometrics and owner
UAT are not claimed complete. See [current cross-ticket matrix](NATIVE-RECONCILIATION-20261015.md).

## Intent and source

Continue the owner-requested feedback → fix loop for the native iOS app. The
last authenticated production feedback-page review surfaced reports about
Face ID after logout, transaction form/list usability, asset balance entry,
category-tree connectors, shared modal/selector styles, screenshot attachment,
font and content density, report-period labels, cache behavior, remembered
wallet/week selection, amount masking, month summary, bottom navigation,
icons and Finance Assistant card presentation. This follow-up targets iOS only;
the web client, backend deployment, production data, and release publication are
out of scope.

The latest queue audit used the authenticated agent API on the production node;
the agent token was not copied into this local checkout. Direct browser access
to the route is not evidence that the authenticated queue is unavailable. Keep
the queue count and per-item dispositions in the audit document, and re-fetch
before claiming the queue is exhausted.

## Contract and constraints

- Preserve existing Go API and financial business rules. Use API-backed native
  tests for persistence-sensitive flows; do not introduce web or backend edits
  unless the actual iOS contract proves a backend gap and the owner approves a
  separate design.
- Reuse the named SwiftUI bases in `docs/design/system/BASE_COMPONENTS.md`.
  Extend a shared base first when the required behavior is missing.
- User-facing money, month, week, and privacy behavior must remain owner- and
  account-timezone-scoped. No financial totals may be inferred from mock data.
- Keep the one-time API-key copy improvement in
  `API-KEY-COPY-DETAIL_DESIGN.md`; this ticket covers the additional feedback.
- Do not update changelog/release notes or install a device build until the
  owner has reviewed the working build. The current signing prerequisite is
  tracked separately in `PORT-100-AUDIT.md`.

## Audit and implementation batches

Audit the current native source and tests first; “reported” does not mean “still
broken.” Record each item as `open`, `already covered`, `fixed`, or `blocked` in
the validation matrix.

1. **Financial correctness and access state:** transaction report inclusion
   control/default/persistence; note-entry reachability and detail copy;
   amount masking across account/wallet/report/assistant; month-summary and
   report-period clarity; biometric app-lock/refresh flow after user-initiated
   lock and session expiry. Authentication changes must preserve the approved
   password/refresh-token contract and include negative cases.
2. **Ledger workflows:** remove unnecessary wallet name from rows when the
   current wallet scope already identifies it; keep note as optional row
   context; persist selected wallet safely; open the newest week containing
   transactions when entering a scoped ledger; expose asset quantity/value
   changes through the native portfolio workflow if that existing API supports
   them.
3. **Native usability and visual system:** category-tree connector geometry,
   shared selector/sheet composition, typography legibility, compact card
   density, income/expense transition feedback, bottom navigation height,
   missing semantic icons and Assistant card elevation. Keep all styling in
   DesignSystem bases and verify small-device/safe-area/accessibility behavior.
4. **Feedback capture:** attach a privacy-safe native screenshot to feedback
   using the existing multipart contract and shared capture/redaction bases.

Each issue is implemented as a narrow RED → GREEN regression with a focused
test, then a relevant suite. Do not bundle unrelated behavior changes to make a
visual checklist appear complete. If intended behavior is ambiguous (notably
what “logout” means while retaining refresh credentials, or asset adjustment
accounting), mark that item blocked with evidence instead of changing security
or ledger semantics.

## Progress — 2026-10-13

- Fixed: transaction ledger subtitles no longer repeat the wallet name; they
  show the optional note or a short local date when the note is empty. The
  focused regression passed with 1/1 tests.
- Fixed: global privacy masking now flows through the shared SwiftUI
  environment to wallet selection/details, transaction list/details,
  recurring, portfolio, credit, savings bases and AI-entry proposals. It masks
  visible money and keeps accessibility amounts masked; editable form fields
  remain usable.
- Fixed: report inclusion is now a shared checkbox (selected by default for a
  new transaction and restored from the saved value when editing).
- Fixed: month-detail period copy now shows inclusive local dates and a short
  account-timezone disclosure rather than midnight timestamps, an exclusive
  boundary note, and a raw IANA timezone identifier. API interval remains
  half-open.
- Fixed: ledger week selection starts with the current account-local week when
  it contains rows, otherwise the newest past week with transactions in the
  selected wallet scope; empty ledgers remain on the current week.
- Fixed: iOS remembers the selected wallet scope locally per authenticated
  owner and clears the preference when the wallet becomes unavailable.
- Fixed: the shared segmented-control base animates selected-state changes and
  turns the animation off when iOS Reduce Motion is enabled.
- Fixed: Finance Assistant structured cards now use the shared elevated
  reference-surface style rather than a flat standard card.
- Already covered in the existing native implementation: API-key copy uses the
  shared clipboard action; feedback submission supports a screenshot
  attachment; the portfolio already exposes current-price updates and persisted
  buy/sell history. These are not reimplemented in this follow-up.
- Pass so far: mask-policy and transaction-subtitle tests passed 2/2 on iPhone
  16 Simulator, iOS 18.5. Broader suite and UI verification remain pending.
- The broad unit+UI scheme run did not complete: after over ten minutes the
  simulator UI-test runner remained on the More screen without publishing
  progress. It was interrupted; this is not a test pass. Run unit/contract and
  targeted UI tests separately.
- Fresh final verification: unit/contract 156/156 with zero skips/failures at
  `/tmp/mypocket-ios-feedback-unit-final7-20261013.xcresult`; Release
  warnings-as-errors build at
  `/tmp/mypocket-ios-feedback-release-final4-20261013`; project/base guards and
  `git diff --check` pass. The earlier broad UI scheme remains incomplete; the
  focused checkbox and privacy UI tests pass separately.
- Remaining: category-tree connector geometry, typography/content density,
  bottom-navigation height and missing semantic icons lack a current native
  feedback screenshot or a reproducible failing test. The simulator was
  booted to the Home Screen during this review, so no authenticated app screen
  could be visually assessed. Do not tune these from guesswork. Clarify whether
  Face ID should ever survive an explicit logout; silently retaining refresh
  credentials is not an acceptable assumption. The production feedback queue
  also remains unverified because its direct route returned 404 and no agent
  service token is configured locally.

## Impact

| Area | Impact |
| --- | --- |
| iOS source | Yes: feature views, auth state, and shared DesignSystem bases where the audit confirms a gap |
| API/public contract | No planned change; existing authenticated endpoints only |
| Database/migrations | No |
| Security/auth | Possible biometric state transition; no token lifetime, refresh, or revocation contract may change silently |
| Web/backend/runtime | No |
| Public docs | No endpoint/setup changes expected; update user-facing native screen docs and AI-readable design contracts |
| Internal docs | Update this ticket, affected screen specs, backlog, context, and validation matrix |
| Release notes | Deferred until owner approval |

## Verification

- Add a failing native unit or API-backed UI regression for each reproduced
  issue before product code changes.
- Run focused XCTest, full native unit/contract suite, relevant
  `LOCAL_API_UI` route tests, Release warnings-as-errors build,
  `scripts/verify-project.sh`, base-component guard/test, and `git diff --check`.
- Record exact command/result and remaining owner UAT in
  `docs/work/VALIDATION_MATRIX.md`.
- Physical-device build/install is a separate gate; do not report it as done
  unless signing and installation succeed.

## Docs review

- Transactions: `docs/design/pages/transactions/README.md`.
- Reports/month/overview: corresponding `docs/design/pages/*/README.md`.
- Auth/API-key/security: `docs/architecture/AUTH_API.md` and account settings
  design only if the actual session contract changes.
- Shared visual behavior: `docs/design/system/BASE_COMPONENTS.md` and the
  relevant base/screen specification.
- API reference and AI-agent `llms` surfaces: no change unless an API contract
  is added or altered.
