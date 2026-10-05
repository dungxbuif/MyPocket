---
artifact_type: detail_design
id: NATIVE-IOS-01
status: approved_for_scaffold
owner: shared
updated: 2026-10-05
human_fields:
  - scope_approval
  - acceptance_approval
ai_fields:
  - context_loaded
  - implementation_decisions
  - verification_plan
shared_fields:
  - status
  - trace
trace:
  backlog: ../../work/BACKLOG.md
  architecture: ../../architecture/ARCHITECTURE.md
  api: ../../architecture/API.md
  auth: ../../architecture/AUTH_API.md
  design: ../../design/README.md
  adr: ../../decisions/ADR-010-native-ios-swiftui-client.md
  plan: ../plans/2026-10-05-native-ios-swiftui-port.md
---

# Native iOS SwiftUI client — detail design

## Decision summary

Create a first-party SwiftUI iOS client in `mobile-ios/` on the isolated
`codex/native-ios-port` branch. The current Vite web client and the existing
Expo migration branch remain intact. The native client talks to the existing
owner-scoped Go API; it never reads the database and never contains provider
secrets.

The request's “100% port” means behavioral/API parity for every currently
mounted web feature and its loading, empty, error, confirmation and review
states. It does not mean copying DOM/CSS pixels. SwiftUI gets a native layout
with the same semantic tokens, copy, business rules and interaction outcomes.

## Why this is needed

The current mobile work is Expo/React Native, while the owner now wants a
native iOS build that can be tested directly in Xcode Simulator first. The web
design contract already requires base-first composition, but the native branch
must enforce the same rule with SwiftUI views rather than reusing DOM/RN code.

## Scope and non-goals

### In scope

- A buildable SwiftUI app under `mobile-ios/` with an iOS Simulator target.
- Native auth gate, Google Authorization Code + PKCE exchange, JWT access
  token, rotating refresh token, logout and Face ID unlock after a valid
  session. Tokens live in Keychain; no token is placed in a deep-link URL.
- Parity for the web feature matrix below, using the same API semantics.
- A reusable `DesignSystem` module. Feature screens may compose design-system
  views and platform-neutral view models only.
- Async/await API client with one owner-scoped retry after access-token expiry,
  request cancellation, typed problem responses and offline-safe read errors.
- XCTest unit/contract tests, SwiftUI preview fixtures and simulator UI smoke
  tests. The first platform gate is Simulator; physical-device signing comes
  later and does not block this scaffold.
- Documentation, ADR, validation-matrix evidence and a release-note entry.

### Out of scope for this port

- Removing or rewriting `app/` or `mobile/` Expo code.
- App Store/TestFlight distribution, push notifications, background sync or
  local ledger writes while offline. Offline cache policy is a follow-up
  decision; the first build must fail safely without inventing finance data.
- New financial business rules, database migrations or provider/model changes.
- Shipping Google client secrets, OCR keys, LLM keys, API keys or service
  tokens in the app bundle.

## Feature parity matrix

| Web capability | Native route/flow | Required behavior |
| --- | --- | --- |
| Auth gate | `Auth/RootGate` | loading, Google sign-in, OAuth failure, logout, refresh rotation |
| Face ID | `Auth/BiometricUnlock` | opt-in after first login; prompt only when session is valid; fallback to passcode/login |
| Overview | `Overview/OverviewScreen` | owner home totals, mask/unmask, refresh, empty/error |
| Transactions | `Transactions/TransactionsScreen` | `Tổng cộng`/wallet scope, account-timezone week/custom period, local-date grouped rows, loading/empty/error, and atomic bulk-delete selection capped at 100 IDs while excluding transfer/adjustment rows |
| Manual entry | `Transactions/TransactionEditor` | income/expense, category, wallet, date, note, validation, create/edit save/cancel without losing the selected record |
| Transfer | `Transactions/TransferEditor` | source/destination wallet, atomic API, confirmation and replay-safe error |
| Transaction actions | `Transactions/TransactionDetail` | edit/delete, immutable adjustment/transfer rules, travel link |
| Wallets | `Wallets/WalletsScreen` + `Wallets/WalletDetailScreen` | basic/goal/credit list and detail, derived balance, add/edit/delete, immutable type/opening balance after create, goal target date and credit statement metadata |
| Credit ledger | `Wallets/CreditStatementScreen` | purchase/refund/fee/interest, payment pair, statement and due status |
| Budgets | `Budgets/BudgetsScreen` | list/progress, create/edit/delete, date interval, wallet/category scope |
| Reports | `Reports/ReportsScreen` | account-local month, category/top-expense cards, previous-period delta and spending/income ratio, zero-safe states |
| Jars/months | `Jars/JarsScreen`, `Months/MonthDetailScreen` | month allocation, live totals, note CRUD, history |
| Recurring | `Recurring/RecurringScreen` | schedule CRUD and run-due result; no silent background execution |
| Categories/groups | `Categories/CategoryList`, `CategoryEditor` | tree, wallet applicability, create/edit/delete and validation |
| Travel Mode | `Travel/TravelScreen` | one active event, activate/deactivate, transaction linkage |
| Portfolio | `Portfolio/PortfolioScreen` | assets, prices, buy/sell validation, unknown valuation |
| AI entry | `AIEntry/AIEntryScreen` | text/photo/PDF picker, up to 20 files, processing, proposal edit, approve/reject |
| Finance Assistant | `Assistant/AssistantScreen` | one persistent conversation, read-only tools, structured cards, errors/cancel |
| Feedback | `Feedback/FeedbackBubble` + `FeedbackScreen` | floating access, silent screen snapshot, submit/list/status/changelog |
| Account | `Account/AccountScreen` | profile/timezone, API-key links, feature links, logout |

## Architecture

```text
+----------------------+     HTTPS/JSON      +-----------------------------+
| SwiftUI feature views | ------------------> | Go API / owner auth         |
| (mobile-ios)          | <------------------ | JWT + rotating refresh      |
+-----------+----------+                      +--------------+--------------+
            |                                                |
            v                                                v
+----------------------+                      +-----------------------------+
| DesignSystem          |                      | PostgreSQL / Redis / S3    |
| base controls/cards   |                      | source of truth             |
+----------------------+                      +-----------------------------+
            |
            v
+----------------------+                      +-----------------------------+
| Core                  | -- Keychain ------> | iOS Secure Enclave/Keychain |
| APIClient/Auth/Cache  |                      | Face ID policy              |
+----------------------+
```

Dependency direction is one-way: `Features -> Core + DesignSystem`; feature
views cannot import web/Expo source or call repositories directly. `Core` owns
HTTP, auth, decoding and cache boundaries. Backend authorization remains the
source of truth for owner isolation.

## Base-component contract

All reusable visual behavior belongs in `mobile-ios/MyPocketIOS/DesignSystem`.
Feature code may not instantiate raw `Button`, `TextField`, `SecureField`,
`DatePicker`, `ProgressView`, `Form`, `NavigationStack`, or `.sheet` for a
user-facing control unless that usage is inside a named design-system base.

Required bases for the first port:

| Base | Responsibility |
| --- | --- |
| `MPButton` / `MPIconButton` | semantic action, loading/disabled state, 44pt target |
| `MPTextField` / `MPAmountField` | labels, validation, keyboard and money formatting |
| `MPDateField` / `MPDateRangePicker` | account-local date labels and calendar selection |
| `MPSurfaceCard` / `MPMetricCard` | surface, radius, padding, masking and metric semantics |
| `MPBottomSheet` / `MPConfirmDialog` | safe-area sheet, focus, cancel/confirm and destructive copy |
| `MPSegmentedControl` / `MPWalletSelector` | period and wallet scope selection |
| `MPTransactionRow` / `MPWalletRow` | domain row states and tap affordance |
| `MPProgressBar` / `MPChart` | budget/goal and report visualizations |
| `MPLoadingView` / `MPEmptyView` / `MPErrorView` | explicit async states and retry |
| `MPAttachmentPicker` / `MPAttachmentThumb` | 20-file selection, thumbnail/close, 5 MiB validation |
| `MPAssistantCard` / `MPActionCard` | structured AI response and explicit confirmation |
| `MPFeedbackBubble` | global floating feedback entry and snapshot exclusion |
| `MPBottomNavigation` / `MPScreenHeader` | route shell, safe-area and header actions |

The design system consumes semantic tokens (`surface`, `canvas`, `ink`,
`secondary`, `brand`, `action`, `danger`, spacing, radius, type scale) mapped
from the web contract. Feature screens only choose layout and data; they do
not redefine colors, shape, shadow or typography.

## Auth/API contract

- Add a mobile-only backend route `POST /api/v1/auth/mobile/google/exchange`
  accepting `{code, code_verifier, redirect_uri, platform}`. The server
  allowlists the exact iOS redirect URI/client, exchanges the code with Google,
  validates issuer/audience/verified email, and returns the existing login
  shape (`token`, `expires_at`, `refresh_token`, `refresh_expires_at`, `user`).
- Web Google callback stays unchanged. The iOS app uses
  `ASWebAuthenticationSession` with S256 PKCE and a custom scheme/Universal
  Link callback. The callback carries only the authorization code and state.
- `APIClient` injects the access token, retries one 401 through a single-flight
  refresh actor, and never refreshes API-key requests. A failed refresh clears
  Keychain session and routes to auth gate.
- All finance requests include the authenticated owner context implicitly via
  the JWT; no owner ID is accepted from feature UI as an authority.
- Feedback uses `POST /api/v1/feedback` with the user JWT or an owner API key
  with `feedback:write`. The production agent token is never shipped in iOS.

## Failure and security model

- A missing/expired provider capability is shown as a typed unavailable state;
  the app never fabricates proposals or totals.
- AI entry keeps proposals in review. No file upload or assistant action writes
  a transaction without explicit user confirmation.
- Network loss during a write displays an indeterminate result and offers
  read-only reload; it does not auto-retry financial writes without an
  idempotency key.
- Logs contain route, status, latency and request ID only. Tokens, OCR text,
  screenshots and financial payloads are redacted.

## Verification plan and acceptance

### Automated

- `swift test` for token/session codecs, date/money formatters, API envelope and
  problem decoding, refresh single-flight, base-component accessibility and
  AI/feedback DTOs.
- Backend `go test ./...` plus mobile OAuth negative cases (redirect mismatch,
  invalid verifier, provider error, unverified email, code reuse).
- `xcodebuild -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=<available iPhone>' test`.
- A static base-component guard fails when feature source contains raw control
  declarations outside `DesignSystem`.

### Simulator smoke flow (first manual gate)

1. Launch the app in an iPhone Simulator.
2. Confirm auth loading and Google sign-in/cancel/error states with a fixture
   or staging OAuth client; no production secret is entered in the simulator.
3. Load overview, switch wallet scope, open a transaction, and verify masked
   money and account-local dates.
4. Create a draft transaction, transfer between two non-credit wallets, and
   verify the confirmation/result states against the API.
5. Open AI entry, select a fixture image/PDF, review proposals, and confirm no
   write occurs before approval.
6. Open Finance Assistant and Feedback from their global controls; verify the
   screenshot excludes the feedback overlay.
7. Enable Face ID simulation, lock/unlock, expire the access token fixture and
   verify one refresh then logout on refresh rejection.

### Definition of done

- Every row in the feature matrix has a route/view model and an API contract
  test; no feature is marked complete from a placeholder screen.
- Every feature view composes only named design-system bases; guard passes.
- Simulator build/tests pass and the smoke flow is recorded in the validation
  matrix. Physical-device signing remains a follow-up, not a hidden failure.
- Web/Expo files are untouched except for docs links, and no production deploy
  or production database mutation occurs in this phase.

## Alternatives considered

1. **Continue Expo/React Native:** rejected for this request because the owner
   explicitly wants a native Swift test target and direct Xcode Simulator path.
2. **Share web DOM or RN visual components:** rejected; renderer primitives,
   focus, accessibility and safe-area behavior differ.
3. **Use a large iOS UI framework:** rejected for the first port; a small
   project-owned DesignSystem keeps base ownership auditable and avoids a
   second visual system.
4. **Reuse browser Google callback:** rejected because cookies and web redirect
   semantics do not provide a safe native session boundary.

## Reconciliation targets

- Add ADR-010 for the SwiftUI boundary and supersession of the Expo-only
  implementation choice.
- Add `docs/architecture/MOBILE_IOS.md` for setup, API/auth and simulator
  operations; update `docs/architecture/API.md` and `AUTH_API.md` for the
  mobile exchange.
- Add native screen references to `docs/design/README.md` and keep per-screen
  web behavior contracts authoritative for semantics.
- Update `docs/work/VALIDATION_MATRIX.md`, `docs/CONTEXT.md`,
  `docs/work/BACKLOG.md` and `docs/releases/CHANGELOG.md` only after proof.
