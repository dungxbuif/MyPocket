# Native iOS SwiftUI port Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (or superpowers:subagent-driven-development) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the current web frontend's mounted finance behavior to a native SwiftUI iOS app in `mobile-ios/`, using project-owned base components and proving the first build/run on iOS Simulator.

**Architecture:** Keep the existing web and Expo clients unchanged. Add a native SwiftUI renderer with a `DesignSystem` boundary, an async URLSession/API core, Keychain-backed auth and feature-specific view models. Add only the mobile Google PKCE exchange required by native auth; all financial rules and owner isolation remain in the Go API.

**Tech Stack:** Swift 6.4, SwiftUI, XCTest, URLSession async/await, AuthenticationServices (`ASWebAuthenticationSession`), LocalAuthentication (Face ID), Security/Keychain, Xcode 27, iOS Simulator.

**Spec:** `docs/superpowers/specs/2026-10-05-native-ios-swiftui-design.md`

## Global Constraints

- The web client in `app/` and the Expo client in the historical migration branch remain unchanged during this port.
- Native code lives under `mobile-ios/`; feature views may not import DOM, React, React Native or web CSS.
- Every user-facing control/card/loading state is implemented in `mobile-ios/MyPocketIOS/DesignSystem` and consumed through a base component.
- The iOS bundle contains no Google secret, OCR key, LLM key, API key or service token.
- Finance data is read/written only through authenticated owner-scoped API calls; the client never accepts `ownerID` as an authorization authority.
- Financial writes require explicit user intent and an idempotency key where the API contract requires one; network retries never duplicate a write.
- Do not claim physical-device, TestFlight, App Store or production OAuth proof from Simulator-only evidence.
- Every code/API/config change updates the affected docs and records exact verification commands.

## Review Focus

- Feature screens bypassing `DesignSystem`: the static guard and a fixture fail on raw SwiftUI controls outside the base module.
- Refresh rotation races: concurrent 401 responses result in one refresh request and all callers use the newest session.
- Account-local dates/timezones: transaction periods and month labels match the backend contract at UTC boundaries.
- Financial write ambiguity: timeout/offline responses do not auto-repeat mutations and show a recoverable state.
- AI/feedback privacy: screenshots exclude the feedback overlay and logs/requests never contain provider credentials or another owner's data.

---

### Task 1: Create the native Xcode target and simulator build gate

**Files:**
- Create: `mobile-ios/MyPocketIOS.xcodeproj/project.pbxproj`
- Create: `mobile-ios/MyPocketIOS/App/MyPocketIOSApp.swift`
- Create: `mobile-ios/MyPocketIOS/App/AppRootView.swift`
- Create: `mobile-ios/MyPocketIOS/Config/Environment.swift`
- Create: `mobile-ios/MyPocketIOS/Resources/Info.plist`
- Create: `mobile-ios/Config/Debug.xcconfig`, `mobile-ios/Config/Release.xcconfig`
- Create: `mobile-ios/.gitignore`, `mobile-ios/README.md`
- Test: `mobile-ios/scripts/verify-project.sh`
- Docs: `docs/architecture/MOBILE_IOS.md`, `docs/design/platforms/ios.md`

**Interfaces:**
- Produces scheme `MyPocketIOS`, bundle identifier `com.mypocket.ios`, deployment target iOS 17, and `Environment.apiBaseURL` selected by build configuration.
- `AppRootView` initially renders a deterministic `MPText`/loading placeholder from `DesignSystem` so the target can build before feature work.

- [ ] **Step 1: Write the failing project verification script**

  Assert the project, scheme, source path, bundle identifier, no tracked secret files, and a Simulator SDK destination are present.

- [ ] **Step 2: Run it to verify the scaffold is absent**

  Run: `bash mobile-ios/scripts/verify-project.sh`
  Expected: FAIL because `mobile-ios/` and the Xcode target do not exist.

- [ ] **Step 3: Create the minimal SwiftUI Xcode target**

  Generate the project with a reproducible project file (use Xcode project generation tooling only at authoring time, then commit the generated project). Keep environment values as placeholders; do not copy production credentials.

- [ ] **Step 4: Build the empty app for a generic Simulator destination**

  Run: `xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -sdk iphonesimulator -configuration Debug build CODE_SIGNING_ALLOWED=NO`
  Expected: PASS once the installed Xcode includes the iPhoneSimulator SDK. If no Simulator runtime is installed, record that as a platform prerequisite rather than claiming a booted app.

- [ ] **Step 5: Commit the isolated scaffold**

  ```bash
  git add mobile-ios docs/architecture/MOBILE_IOS.md docs/design/platforms/ios.md
  git commit -m "chore: scaffold native ios simulator target"
  ```

### Task 2: Implement the SwiftUI DesignSystem and base-component guard

**Files:**
- Create: `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPText.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPButton.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPFields.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPAsyncStates.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Create: `mobile-ios/MyPocketIOS/DesignSystem/MPAttachmentViews.swift`
- Create: `mobile-ios/scripts/check-base-components.sh`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`, `mobile-ios/scripts/check-base-components.test.sh`
- Docs: `docs/design/system/BASE_COMPONENTS.md`, `docs/design/platforms/ios.md`

**Interfaces:**
- `MPButton(style:loading:isDisabled:action:content:)` owns all action buttons and 44pt hit targets.
- `MPTextField`, `MPAmountField`, `MPDateField`, `MPWalletSelector` expose controlled bindings and validation messages.
- `MPSurfaceCard`, `MPMetricCard`, `MPTransactionRow`, `MPWalletRow`, `MPProgressBar`, `MPChart`, `MPBottomSheet`, `MPConfirmDialog`, `MPLoadingView`, `MPEmptyView`, `MPErrorView`, `MPAttachmentPicker`, `MPAssistantCard`, and `MPFeedbackBubble` are the required domain bases from the spec.

- [ ] **Step 1: Write failing design-system tests**

  Verify semantic token values, minimum 44pt controls, masking behavior, attachment limit/error rendering, and accessibility labels for loading/empty/error states.

- [ ] **Step 2: Run tests to verify they fail**

  Run: `xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=<available-id>'`
  Expected: FAIL because the bases/tests are not present. If no runtime/device exists, run the test target with `-sdk iphonesimulator` only after the platform prerequisite is installed and record the missing runtime.

- [ ] **Step 3: Implement bases without feature-specific visual overrides**

  Wrap native SwiftUI controls inside the base files. Use semantic tokens and accessibility identifiers. Feature modules must receive data/actions through parameters, not reach into global singleton state.

- [ ] **Step 4: Add the static raw-control guard**

  Scan `mobile-ios/MyPocketIOS/Features` and `App` for raw user-facing `Button`, `TextField`, `SecureField`, `DatePicker`, `ProgressView`, `.sheet` and custom clickable wrappers. Allow only imports/uses inside `DesignSystem` and explicitly documented navigation composition.

- [ ] **Step 5: Run guard and tests, then commit**

  Run: `bash mobile-ios/scripts/check-base-components.test.sh` and the DesignSystem XCTest target.
  Expected: PASS with zero feature violations.

  ```bash
  git add mobile-ios docs/design
  git commit -m "feat: add native ios base component system"
  ```

### Task 3: Build typed API, session, Keychain and refresh core

**Files:**
- Create: `mobile-ios/MyPocketIOS/Core/API/APIClient.swift`
- Create: `mobile-ios/MyPocketIOS/Core/API/APIError.swift`
- Create: `mobile-ios/MyPocketIOS/Core/API/Envelope.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Auth/Session.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Auth/KeychainSessionStore.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Auth/RefreshCoordinator.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Models/AuthModels.swift`, `FinanceModels.swift`, `FeedbackModels.swift`, `AIModels.swift`
- Test: `mobile-ios/MyPocketIOSTests/APIClientTests.swift`, `SessionStoreTests.swift`, `RefreshCoordinatorTests.swift`, `ModelDecodingTests.swift`
- Docs: `docs/architecture/MOBILE_IOS.md`, `docs/architecture/AUTH_API.md`

**Interfaces:**
- `APIClient.send<T:Decodable>(_ request: APIRequest<T>) async throws -> T` decodes `{data,meta}` and Problem Details.
- `SessionStore` saves/loads/clears `AccessSession` through Keychain; raw tokens never enter `UserDefaults` or logs.
- `RefreshCoordinator.refreshIfNeeded()` is an actor-backed single-flight operation; callers receive the same replacement session.
- `APIClient` retries exactly once on a JWT 401, never retries API-key requests, and attaches `X-Request-ID` for diagnostics.

- [ ] **Step 1: Write failing URLProtocol/session tests**

  Cover success envelope, RFC problem decoding, malformed JSON, Keychain round trip, missing session, two concurrent 401s producing one refresh, rotated-token replay, and no retry for API-key auth.

- [ ] **Step 2: Run tests to verify failure**

  Run: `xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=<available-id>' -only-testing:MyPocketIOSTests`
  Expected: FAIL with missing types.

- [ ] **Step 3: Implement the core**

  Keep request construction typed and endpoint-specific. Map server error `code` and `request_id` into `APIError.problem`; do not show raw provider output in user-facing copy.

- [ ] **Step 4: Run core tests and `swift-format`/compiler checks**

  Expected: all core tests PASS and no token appears in test logs.

- [ ] **Step 5: Commit**

  ```bash
  git add mobile-ios docs/architecture/MOBILE_IOS.md docs/architecture/AUTH_API.md
  git commit -m "feat: add native api and rotating session core"
  ```

### Task 4: Add the mobile Google PKCE exchange to Go

**Files:**
- Modify: `backend/internal/config/config.go`, `.env.example` or deployment env template
- Create: `backend/internal/infrastructure/auth/google_mobile_oauth.go`
- Create: `backend/internal/usecase/mobile_auth.go`
- Create: `backend/internal/controller/http/mobile_auth_handler.go`
- Modify: route registration and generated Swagger
- Test: backend config/usecase/provider/handler tests for the mobile route
- Docs: `docs/architecture/API.md`, `docs/architecture/AUTH_API.md`, `docs/architecture/MOBILE_IOS.md`, `docs/decisions/ADR-010-native-ios-swiftui-client.md`

**Interfaces:**
- `POST /api/v1/auth/mobile/google/exchange` accepts `{code, code_verifier, redirect_uri, platform}` and returns existing `LoginOutput`.
- Configuration contains allowlisted iOS client ID and redirect URI; secrets remain server-side and are never serialized to the app.

- [ ] **Step 1: Write failing backend contract tests**

  Cover valid iOS exchange, redirect mismatch, unknown platform, invalid PKCE verifier, provider failure, unverified email, code reuse, and refresh rotation.

- [ ] **Step 2: Run targeted Go tests to confirm failure**

  Run: `go test ./internal/controller/http ./internal/usecase ./internal/infrastructure/auth`
  Expected: FAIL because the route/config/usecase do not exist.

- [ ] **Step 3: Implement allowlisted code exchange**

  Reuse the existing user/session issuance path after Google identity validation. Do not alter browser callback behavior. Compare redirect/client configuration exactly and reject unknown values before provider calls.

- [ ] **Step 4: Run backend verification**

  Run: `go test ./...` and targeted `go test -race` for refresh/session paths.
  Expected: PASS with existing web auth tests unchanged.

- [ ] **Step 5: Regenerate Swagger and commit**

  ```bash
  go generate ./cmd/api
  git add backend docs/architecture docs/decisions
  git commit -m "feat: add mobile google pkce exchange"
  ```

### Task 5: Implement native auth gate, Google callback and Face ID

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/Auth/AuthView.swift`
- Create: `mobile-ios/MyPocketIOS/Features/Auth/AuthViewModel.swift`
- Create: `mobile-ios/MyPocketIOS/Features/Auth/BiometricUnlockView.swift`
- Create: `mobile-ios/MyPocketIOS/Core/Auth/GooglePKCEClient.swift`
- Modify: `App/AppRootView.swift`, Info.plist URL schemes/associated-domain placeholders
- Test: `AuthViewModelTests.swift`, `GooglePKCEClientTests.swift`, `BiometricPolicyTests.swift`
- Docs: `docs/architecture/MOBILE_IOS.md`, `docs/architecture/AUTH_API.md`

**Interfaces:**
- `AuthViewModel.signIn()` opens `ASWebAuthenticationSession`, validates state/PKCE callback, exchanges code, then loads profile/home.
- `BiometricPolicy` returns `disabled`, `available`, `enabled`, `unavailable`; enabling is explicit after first login and never bypasses server expiry.

- [ ] **Step 1: Write failing auth tests**

  Cover cancellation, state mismatch, callback without code, redirect mismatch, successful session persistence, expired access token refresh, refresh rejection logout, and Face ID unavailable fallback.

- [ ] **Step 2: Implement auth gate and secure session flow**

  Use only fixture/staging URLs in Simulator configuration; no production secret. Route errors through `MPErrorView` and expose request ID for support without exposing tokens.

- [ ] **Step 3: Run auth tests and generic Simulator build**

  Expected: PASS; UI smoke can wait until a runtime is installed.

- [ ] **Step 4: Commit**

  ```bash
  git add mobile-ios docs/architecture
  git commit -m "feat: add native oauth and face id gate"
  ```

### Task 6: Implement overview, wallets and transaction parity

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/Overview/*`
- Create: `mobile-ios/MyPocketIOS/Features/Wallets/*`
- Create: `mobile-ios/MyPocketIOS/Features/Transactions/*`
- Modify: `Core/Models/FinanceModels.swift`, `Core/API/FinanceAPI.swift`
- Test: view-model/API contract tests for overview, wallet types, ledger scopes, transfer and CRUD
- Docs: corresponding `docs/design/pages/*` platform notes, `docs/work/VALIDATION_MATRIX.md`

**Interfaces:**
- `OverviewViewModel.load()` uses home/profile endpoints and exposes masked totals.
- `WalletsViewModel` supports basic, goal and credit detail; credit mutations use the dedicated credit API only.
- `TransactionsViewModel` owns wallet scope, Week/Custom range, grouped rows and refresh propagation.
- `TransactionEditorViewModel` and `TransferEditorViewModel` validate locally, send typed requests, and require confirmation for destructive/atomic actions.

- [ ] **Step 1: Add failing API fixture tests**

  Pin wallet scope filtering, account-local week ranges, transfer pair payloads/idempotency, credit route separation, immutable adjustment behavior, and derived balance refresh.

- [ ] **Step 2: Implement view models using only `Core` + `DesignSystem`**

  No raw controls in feature files. Use fixture repositories for previews and URLProtocol for tests; production view models use APIClient.

- [ ] **Step 3: Run unit/contract tests and base guard**

  Expected: PASS; all transaction writes have explicit pending/success/error states.

- [ ] **Step 4: Commit**

  ```bash
  git add mobile-ios docs/design docs/work/VALIDATION_MATRIX.md
  git commit -m "feat: port native overview wallet and ledger flows"
  ```

### Task 7: Implement secondary finance features

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/Budgets/*`, `Reports/*`, `Jars/*`, `Months/*`, `Categories/*`, `Recurring/*`, `Travel/*`, `Portfolio/*`, `Account/*`
- Modify: `Core/Models/FinanceModels.swift`, `Core/API/FinanceAPI.swift`, `App/AppRootView.swift`
- Test: feature view-model/API contract tests for every route in the parity matrix
- Docs: `docs/design/pages/{budgets,reports,savings,account,travel,portfolio}/*`, `docs/architecture/MOBILE_IOS.md`

**Interfaces:**
- Each feature exposes `load()`, typed mutation methods, and explicit `loading/empty/error/ready` state.
- Calendar-date APIs use `YYYY-MM-DD` labels and the account timezone; no feature converts date-only fields through the device timezone.
- Recurring `run-due` is an explicit user action; no background scheduler is added.

- [ ] **Step 1: Write fixture tests for each endpoint family**

  Include budget overlap/ended validation, jar month configuration and note CRUD, category wallet applicability, travel active-event constraints, portfolio insufficient quantity/unknown price, and recurring idempotency.

- [ ] **Step 2: Implement feature view models/screens**

  Compose `MPMetricCard`, `MPProgressBar`, `MPWalletSelector`, `MPDateRangePicker`, `MPBottomSheet`, and async-state bases. Keep feature-specific layout in feature files only.

- [ ] **Step 3: Run tests, base guard and generic build**

  Expected: PASS with no placeholder data imported from web `mockFinance`.

- [ ] **Step 4: Commit**

  ```bash
  git add mobile-ios docs/design docs/architecture docs/work/VALIDATION_MATRIX.md
  git commit -m "feat: port native budgets reports and account flows"
  ```

### Task 8: Implement AI entry, Finance Assistant and feedback parity

**Files:**
- Create: `mobile-ios/MyPocketIOS/Features/AIEntry/*`
- Create: `mobile-ios/MyPocketIOS/Features/Assistant/*`
- Create: `mobile-ios/MyPocketIOS/Features/Feedback/*`
- Modify: `Core/API/AIAPI.swift`, `Core/API/FeedbackAPI.swift`, models and app navigation
- Test: `AIEntryViewModelTests.swift`, `AssistantViewModelTests.swift`, `FeedbackViewModelTests.swift`, screenshot-redaction fixture
- Docs: `docs/architecture/FEEDBACK_API.md`, `docs/architecture/MOBILE_IOS.md`, AI/assistant screen contracts, `docs/work/VALIDATION_MATRIX.md`

**Interfaces:**
- `AIEntryViewModel.process(text:attachments:timezone:)` sends multipart to `/api/v1/ai/entry/process`, accepts at most 20 JPEG/PNG/PDF files at 5 MiB each, and never approves automatically.
- `AssistantViewModel` persists one conversation, handles structured response parts and explicit action confirmation; read-only tools stay backend-owned.
- `FeedbackCapture` snapshots the current screen behind a redaction/exclusion layer; `FeedbackViewModel.submit` sends optional PNG plus text through the user JWT. The global bubble is hidden from the captured image and never uses the production agent token.

- [ ] **Step 1: Write failing AI/feedback contract tests**

  Cover attachment count/size/type validation, processing/unavailable states, empty proposals, explicit proposal approval, assistant tool/error cards, feedback screenshot exclusion and owner-scoped status/changelog decoding.

- [ ] **Step 2: Implement typed clients, view models and base cards**

  Use `MPAttachmentPicker`, `MPAttachmentThumb`, `MPAssistantCard`, `MPActionCard`, `MPFeedbackBubble`, `MPBottomSheet` and `MPStatusView`; no screen-specific duplicated controls.

- [ ] **Step 3: Run tests and base guard**

  Expected: PASS; tests prove no transaction is written by processing or read-only assistant flows.

- [ ] **Step 4: Commit**

  ```bash
  git add mobile-ios docs/architecture docs/design docs/work/VALIDATION_MATRIX.md
  git commit -m "feat: port native ai assistant and feedback flows"
  ```

### Task 9: Integrate navigation, simulator smoke test and reconcile docs

**Files:**
- Modify: `mobile-ios/MyPocketIOS/App/AppRootView.swift`, `DesignSystem/MPNavigation.swift`
- Create: `mobile-ios/MyPocketIOSUITests/MyPocketIOSUITests.swift`
- Modify: `docs/work/VALIDATION_MATRIX.md`, `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, `docs/releases/CHANGELOG.md`, `README.md`
- Verify: `docs/architecture/MOBILE_IOS.md`, `docs/design/platforms/ios.md`, ADR-010, spec and plan links

**Interfaces:**
- `AppRootView` exposes the six primary destinations (overview, transactions, budgets, reports, assistant, account) and routes secondary flows through base navigation/sheets.
- Every screen has stable accessibility identifiers for UI tests and Simulator smoke capture.

- [ ] **Step 1: Add failing route/UI smoke assertions**

  Assert auth gate, primary tab labels, global feedback bubble, add/AI entry actions, and deep links to wallet/transaction/assistant/feedback routes.

- [ ] **Step 2: Wire the complete navigation tree**

  Preserve state refresh propagation and safe-area behavior. Verify modal dismissal, back navigation and keyboard focus with the base components.

- [ ] **Step 3: Install/select an iOS Simulator runtime and run the full gate**

  Run:

  ```bash
  xcrun simctl list runtimes
  xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=<available iPhone>'
  xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=<available iPhone>' build
  ```

  Then manually execute the Simulator smoke flow in the spec and record pass/fail, screenshots and any environment blocker. Do not claim a booted Simulator if Xcode has no installed runtime.

- [ ] **Step 4: Run repository regression gates**

  Run `go test ./...`, `npm run check:design`, `npm run test:design`, `npm run typecheck`, `npm run build` in `app/`, `git diff --check`, and the iOS base-component guard.

- [ ] **Step 5: Reconcile docs and release note**

  Update the validation matrix with exact commands/results, mark the phase status accurately, update context/backlog, add a first native iOS release-note entry, and document remaining physical-device/OAuth runtime gaps.

- [ ] **Step 6: Commit the verified port**

  ```bash
  git add mobile-ios docs README.md
  git commit -m "feat: add native ios finance client parity"
  ```

## Execution reconciliation — 2026-10-05

The implementation is now in the native iOS worktree and the Simulator gate is
green. The generated Xcode project, SwiftUI base-component guard, typed API
client, auth/session flow, mounted finance routes, AI/feedback surfaces and
release configuration are present. The current verification evidence is:

```bash
bash mobile-ios/scripts/check-base-components.test.sh
bash mobile-ios/scripts/verify-project.sh
bash -lc 'set -o pipefail; xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -sdk iphonesimulator -destination "platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D" test'
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Release -sdk iphonesimulator -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' build
```

The latest completed simulator gate passes 65 native unit/contract tests and
17 signed `LOCAL_API_UI` tests (82/82 combined); the Release warnings-as-errors build passed,
base/project guards passed, the no-invalid-frame log check passed, and the
simulator app was relaunched after the test run. The feedback-bubble regression
and the AI-entry twenty-file thumbnail/cap/remove flow are covered by the
17-test UI suite.
The latest parity fixes also cover stable-ID selectors for duplicate wallet or
category names and account-timezone boundary formatting.
The plan's original red/green checkbox history is retained as authoring
history; release readiness is not claimed until an iOS OAuth client ID is
configured, provider-backed AI/OCR is exercised, and authenticated staging or
production API smoke is completed. No production mutation was performed.

## Self-review

- **Spec coverage:** Auth/API is Tasks 3–5; all parity rows are Tasks 6–8; base ownership is Task 2; simulator proof and docs are Task 9.
- **Step scan:** Each task has a failing proof, implementation, verification and commit boundary. No task relies on a hidden feature-specific control or backend database access.
- **Type consistency:** `APIClient`, `SessionStore`, `RefreshCoordinator`, `FinanceAPI`, `AIEntryViewModel`, `AssistantViewModel` and `FeedbackViewModel` are named once and consumed by later tasks.
- **Review focus coverage:** base guard (Task 2), refresh race (Task 3), timezone (Task 6/7), write timeout (Task 6), privacy (Task 8).
- **Proportion:** The plan intentionally splits the full parity port into independently testable vertical slices; no web files are moved or deleted.

## Execution handoff

The owner has already selected the execution outcome (native SwiftUI and
Simulator-first validation). Implement this plan in the isolated worktree
`/Users/dungxbuif/worktrees/mypocket-ios-native` and stop at the platform
prerequisite if no iOS Simulator runtime is installed; do not substitute an
unverified device or production deployment.
