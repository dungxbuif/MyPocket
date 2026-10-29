---
artifact_type: detail_design
id: FEEDBACK-20261019-NATIVE
status: in_review
owner: shared
updated: 2026-10-19
---

# Native feedback closure

Authorization: owner requested all current feedback/backlog fixes with parallel
design/tests before implementation. Parent released the shared plan/test barrier
after expected RED evidence in all lanes. No production writes or physical-device
installation.

Trace: [backlog](../BACKLOG.md), [latest six](NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md),
[native reconciliation](NATIVE-RECONCILIATION-20261015.md),
[validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).
The follow-up visual/base audit is recorded in
[NATIVE-VISUAL-AUDIT-20261019](NATIVE-VISUAL-AUDIT-20261019.md).
Active phase is the native reconciliation queue, not a new numbered phase.

## Findings and design

### Refresh loading continuity (2026-10-19 follow-up)

The savings and budget report routes were still clearing their last successful
report before every refresh. That made a quick retry or foreground refresh
replace a populated screen with a blocking spinner, unlike the shared offline/
loading contract used by overview and wallets. The fix keeps the last report
while a request is in flight and renders only an inline refresh indicator when
content exists; the blocking loader remains for the first empty load. The
existing report transaction page already preserves its page during refresh.
Verification adds a deterministic policy regression and a Release warnings-as-
errors build; no API, persistence, or financial calculation changes are made.

| Feedback | Reproduction / root cause | Intended change and proof |
| --- | --- | --- |
| 5ff74529 Account Add | Select Account: `MainTabView` explicitly disables its global Add. | Enable on Account root only; retain Assistant/secondary suppression. Open ordinary transaction editor and return to Account on cancel/save. UI checks Add, open, close, caller restoration and secondary suppression. |
| b209e80f edge Back | More uses NavigationStack but hides native back controls; no interactive-pop restoration exists. | Shared navigation base restores native edge-pop for secondary routes. Existing editor scaffold disables pop while editing/busy; root has no pop. UI checks secondary-to-root gesture and editor protection. |
| f42c103d row spacing | `MPTransactionRow` has a 44pt minimum and no outer vertical padding. | Shared 6pt vertical inset around existing content; preserve typography, horizontal alignment and Dynamic Type expansion. Rendered-height regression and existing transaction presentation tests. |
| b017284a portfolio | Unit-price × quantity, price sheet and unused-asset confirmed deletion already exist. Asset rename and historic trade correction lack client/API contracts. | Existing contracts remain; backend agent coordinates any additive metadata/correction endpoints before client work. No speculative destructive history edits. |
| 1565a9e5 sheets | Short price/choice/note forms already use `mpFormBottomSheet`; long editors use `MPEditorScaffold`. | Reuse existing forms and Close/Cancel controls. Verify representative sheet and editor behavior. |
| 824dea29 AI cards | Jar create/update and budget create/update confirmation cards exist. | Preserve explicit owner confirmation; backend/provider behavior and new actions belong to backend slice. Run existing native card regressions. |

Older feedback already represented in source: shared money/date/quantity bases,
Photos-first attachment source selector, feedback image upload, independent
Face ID sign-in enrollment/forget, persistent wallet scope, masking, compact
metrics, duplicate Account/notification entry removal and Markdown chat rendering.
Historical open feedback is not evidence those implementations are absent.

## Bases and impact

Reuse `MPBottomNavigation`, `MPShellNavigation`, `MPTransactionRow`,
`MPEditorScaffold`, `mpFullScreenCover`, `mpFormBottomSheet`, and the More
NavigationStack. New gesture integration belongs in `DesignSystem/MPNavigation.swift`.
No new palette, raw screen control or competing per-screen gesture.
UI/UX Pro Max SwiftUI search confirms typed navigationDestination/NavigationStack;
native pro-rules guide 44pt targets, gesture conflict avoidance and safe-area checks.

Alternative rejected: globally enable Add everywhere (competes with editors),
or attach screen-local drag gestures (competes with scrolling and loses native
interactive transitions). No architecture/schema/auth/runtime changes. Existing
finance API and owner boundaries remain authoritative. Native UI interaction
changes do not require an ADR unless scope expands to new APIs.

## Verification and reconciliation

### Protected deep-link lifecycle

External routes are retained while the app is unauthenticated so a link opened
from a notification or browser can be consumed after a successful login. When
an authenticated or biometric-locked session transitions to unauthenticated,
the app root now discards that pending route. This prevents a URL opened under
one account from replaying after a different account signs in. Route parsing
and owner-scoped data loading remain unchanged; the guard is covered by the
coordinator regression test.

Before product edits, add failing UI tests for Account Add and edge Back, and a
rendered-height test for transaction separation. Use exclusive iPhone 16 simulator
slot EC80B092-CEF1-4A8E-A5A3-7420298EFD4D and unique `/tmp` artifacts.
After implementation, run focused tests, full native units, project/base guards,
and Release warnings-as-errors. No web changes or web build is appropriate to
this native-only slice (web is archived).

Root owns shared context/backlog/validation/release reconciliation. Internal
screen/base docs and public human/Markdown guides must describe actual Account
Add, swipe and row behavior; no API/OpenAPI/ERD change for these native fixes.
Owner/device visual acceptance remains separate from simulator regression proof.

RED evidence: `xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj
-scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D'
-derivedDataPath /tmp/mypocket-feedback-native-20261019
-resultBundlePath /tmp/mypocket-feedback-native-red-20261019.xcresult`
with three `-only-testing` selectors for the new tests and signed simulator
settings (`CODE_SIGN_IDENTITY=- CODE_SIGNING_ALLOWED=YES DEVELOPMENT_TEAM=AU442S3MN3`).
All three failed for the intended assertions: row height 44 < 56, Account Add
absent, and secondary edge drag stays on Account settings. UI setup succeeded;
no fixture/backend failure. Xcode later stalled in test-session cleanup and was
interrupted after the assertions; RED is observed test output, not a completed
result-bundle claim.

Focused GREEN: `/tmp/mypocket-feedback-native-green-20261019.xcresult` reports
96 passed, zero failed/skipped/runtime warnings (93 DesignSystem tests + three
UI tests including category-draft protection). Command uses the same project,
scheme/destination/derived data, `-enableCodeCoverage NO -parallel-testing-enabled NO`
and selects DesignSystemTests plus the three new UI tests. Base/project guards
pass. Native `MPTransactionRow` now adds6pt top/bottom; Account root Add restores
caller on editor close; the scoped bridge protects editor/draft states.

Peer review identified overlapping route delegate replacement. The first test
attempt failed during harness setup (navigation view not loaded), which is not
product RED. After loading the UIKit view, `/tmp/mypocket-feedback-bridge-red2-20261019.xcresult`
records the expected original-delegate restoration failure. The fix captures
the original UIKit delegate directly when replacing a bridge and retains it
only until restore; it never retains a disappearing route controller.
Focused lifecycle/policy GREEN reports2/2, zero skipped/warnings at
`/tmp/mypocket-feedback-bridge-green-20261019.xcresult`. Lifecycle proof covers
overlap, prior-route reattachment, root restore and repeated cleanup.
Shared docs, public transaction guide and AI-agent table are reconciled.

## Final verification — 2026-10-19

- Native full `MyPocketIOSTests` plus six selected navigation UI regressions:
  **354 native unit tests passed,0 failed,0 skipped,0 runtime warnings**, exit0;
  the selected signed UI artifact remains 6/6.
  Artifact `/tmp/mypocket-feedback-native-final-20261019.xcresult`. Same test
  command settings as focused GREEN; selectors are the full unit target plus
  Account Add/cancel, edge Back, category draft protection, selected-Account tap
  shielding, last-More-action reachability and no-reserved-Feedback-lane UI tests.
- Release simulator build: `xcodebuild build -quiet -project mobile-ios/MyPocketIOS.xcodeproj
  -scheme MyPocketIOS -configuration Release -destination 'generic/platform=iOS Simulator'
  -derivedDataPath /tmp/mypocket-feedback-release-20261019 CODE_SIGNING_ALLOWED=NO
  SWIFT_TREAT_WARNINGS_AS_ERRORS=YES GCC_TREAT_WARNINGS_AS_ERRORS=YES` passed exit0.
  Build log `/tmp/mypocket-feedback-release-20261019/Logs/Build/3E8A632A-688A-471D-9814-654F2100D042.xcactivitylog`.
- Exported and inspected the Account root render at
  `/tmp/mypocket-feedback-native-final-attachments-20261019/3832C9F7-2AFF-48AE-AECF-0EBCD4E35716.png`:
  center Add is visible; Feedback remains lower-right over scroll content with
  no reserved lane. The separate final-action reachability UI test passes.
- `bash mobile-ios/scripts/check-base-components.sh`,
  `bash mobile-ios/scripts/verify-project.sh` and `git diff --check`: PASS.
  All shell invocations used `rtk proxy` as required.

### Refresh continuity follow-up verification — 2026-10-19

- The shared `MPLoadingPresentationPolicy` now distinguishes an initial
  blocking loader from an inline refresh indicator. Savings reports, budget
  reports, report transaction pages, jars, and wallet details retain populated
  content while refreshing; only an empty first load uses the blocking loader.
- Focused policy test passed in
  `/tmp/mypocket-loading-redgreen-20261019.xcresult`.
- Full native unit target passed **359 tests, 0 failures, 0 skipped** in
  `/tmp/mypocket-native-loading-final-20261019.xcresult`.
- Release simulator build passed with warnings-as-errors in
  `/tmp/mypocket-native-loading-release-20261019`.

This closes the local refresh-flash regression. Physical-device visual and
production/provider acceptance remain separate gates owned by the integrator.

Internal design/token/base/More/transaction docs and public transaction guide
(including AI-agent table) updated. No API, schema, security, deployment or ADR
change required for these three native interactions. Shared lifecycle docs are
the parent integrator's responsibility. No commit, deployment or physical
installation performed by this lane.

### Feedback submit/refresh race follow-up — 2026-10-20

Reproduction: start a feedback list refresh, hold its GET response, submit a new
feedback item, then release the stale GET. Before this fix the refresh generation
was unchanged by `FeedbackViewModel.submit`, so the stale list replaced the
locally inserted item and made a successful submission disappear from the inbox.
The same generation invalidation already used by resolution mutations is now
applied before and after the submit request. The fix is native-only and does not
change the feedback API, persistence, or production status.

Regression: `APIClientTests.testFeedbackRefreshCannotEraseNewlySubmittedItem`
uses a held stale GET and a real POST-shaped response; it asserts the newly
created row remains after the refresh completes. The focused simulator test and
full native unit target must be rerun before this item is marked verified.

Acceptance limits: native simulator behavior is verified; no whole-app physical
device, small-phone/landscape/Dynamic Type visual sweep or owner acceptance is
claimed. Edge Back applies to More's secondary NavigationStack routes; embedded
subforms deliberately retain explicit Cancel/Back instead of silently discarding
drafts. Portfolio historical correction/new API support and live-provider AI
tool execution remain coordinated backend/product concerns, not native test claims.
