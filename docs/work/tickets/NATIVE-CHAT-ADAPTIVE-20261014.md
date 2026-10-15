# Chat adaptive accessibility verification

Status: in_review — owner approved adaptive chat interaction on 2026-10-14. Existing owner-approved
UI plan Task5/6, full-native goal.
No new palette/font/base family/auth/API/provider/streaming decisions.

Before test: normal-size SE chat keyboard/composer gates pass, but largest Dynamic
Type and landscape remain explicitly unverified. MPFont is already relativeTo
system styles, MPChatComposer caps1–4 lines with44pt action, shell reserves normal
bottom navigation. Do not infer large-text/landscape correctness from normal UI.

Verification-only first: add actual native UI tests. Launch largest accessibility
text category through standard iOS launch override and prove it took effect by
larger composer field versus normal baseline. Enter multiline draft, verify full
draft retained, entire focused composer/action above keyboard, >=44pt targets,
remaining conversation>=44pt. Separately rotate real simulator to landscape and
assert orientation actually changed before checking keyboard/composer/viewport.
Reset orientation on teardown; no app data/DB reset, private provider call or
global text setting. Existing fallback fake-session fixture intentionally cannot
submit, so no claim of live AI/backend conversation behavior.

If runtime RED, investigate exact rendered geometry before proposing a bounded
base/shell fix; no guessed Feedback placement changes. Existing1–4line maximum
permits a smaller adaptive line cap if evidence requires, preserving scalable
fonts and internal draft scrolling. Orientation config changes require explicit
design review rather than silently adding new runtime support.

Bases: MPChatComposer, MPText/MPFont, MPScreenHeader, MPBottomNavigation,
MPChatOverview, existing scroll/message primitives. UI/UX Pro Max SwiftUI verified
Dynamic Type guidance supports scalable fonts/layout; no exact native UX-domain
match after one narrower retry, so native layout checks use skill's defaults/
quick-reference dynamic-type/focus-not-obscured/44pt rules (not web CSS advice).

Required gates: new tests and normal keyboard/draft/suggestions/APIoverview,
full native units, base guards/project/diff; inspect captures. Reconcile native
screen/registry/public human+agent guide/context/backlog/validation. Do not mark
entire accessibility/landscape/all-screen goal complete from chat tests.

Links: [UI plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[screen](../../design/pages/finance-assistant/README.md),
[goal](NATIVE-FULL-COMPLETION-20261014.md), [validation](../VALIDATION_MATRIX.md).
No prod/install/reset/release, goalqueue/Feedback/cache review gates unchanged.

2026-10-15 current-helper regression: previous full chat gate predates semantic
reveal correction. Before any further product change, rerun all8existing chat UI
cases on SE/local8081, sequential LOCAL_API_UI, warnings-as-errors, fresh bundle
`/tmp/mypocket-chat-current-helper-20261015.xcresult`. Includes draft roundtrip,
overview, normal/AX/landscape keyboards, suggestions and disabled-send privacy
control. No new test/product code, model calls, data reset or deployment. Numeric
and business refinement approvals pending independently. Inspect actual failures
and captures; earlier green results cannot certify the current helper revision.

Current helper gate88653terminal0:8/8PASS0skip/no runtime warnings,
`/tmp/mypocket-chat-current-helper-20261015.xcresult`. API overview and actual
landscape, normal/AX composer, draft tab roundtrip, suggestions and disabled-send
cases pass with current helper. No product/test source changed this turn.
GeometryA80B6F81: app667×375, conversation63pt, composer53pt/send44×44 above
keyboard y213. Focused landscapeFD268710 retains portrait-sized cropped buffer;
cannot claim full pixel landscape acceptance. AX restoredF165D9F3 inspected:
full-label menu present, Feedback glyph oversized and overlays explanatory text.
Existing blocked placement/style issue remains; no fourth speculative fix.

Coverage correction: these8cases do NOT exercise populated persisted history
pagination or cancellation of a live run. Native unit source search finds no
advisorMessages/loadOlder/cancel state test. Adapter sends conversation_id and
optional before_seq; backend handler/repository limits50, native hasOlder==50
matches that bound, so no page-size defect established by source inspection.
Source-only cancellation risk: cancel clears activeRunID but earlier send polling
task still runs and may reload/clear state after a newer send. No live race
reproduction/provider call or fix. Existing plan requires those behavioral gates;
do not substitute this layout subset. Internal docs/ledger reconciled; public
contract unchanged this verification-only pass. No release/prod/device action.

20261015 next approved-plan verification slice: characterize the existing native
history/cancel wire adapter with transport-boundary fixtures (no product changes).
History test must fail if cursor is omitted, owner query injected, returned order
not chronological or structured parts lost. Cancellation test must fail if route/
method/auth changes or it submits another chat message. Exercise real FinanceAPI,
APIClient and decoders; fake only network bytes, using full server message shape.
This is adapter coverage, not UI merge/scroll position, actual DB pagination,
provider cancellation or native send/cancel race proof. Run targeted tests, all
native units and guards; reconcile evidence without calling this a product fix.

Adapter characterization24082exit0:2/2PASS0skip/no runtime warnings,
`/tmp/mypocket-advisor-wire-20261015.xcresult`. Existing implementation already
satisfies these assertions, so this is added coverage, not a RED→GREEN bug fix.
Tests use real APIClient/FinanceAPI/decoders and a sequential actor fake only at
HTTPTransport. Complete server message fields retained; latest54/53/52 becomes
52/53/54, older51/50 becomes50/51, text/card amount125000 retained. Request path,
GET/session Bearer auth/cursor52/no owner override/no body checked. Cancel uses
exact POST run/cancel, one request, session auth and a fresh UUID idempotency key;
no implicit message resubmission. No fake-provider or fake-UI success claimed.

Final72133exit0:273/273native units PASS0skip/no runtime warnings,
`/tmp/mypocket-advisor-wire-units-20261015.xcresult`. Guards/guardtests/project/
diff pass. No product files changed, no live API/DB/provider writes or UI action
in these unit tests. Does not close native history scroll/merge/race/livecancel
or full goal. Public human/agent docs, API/OpenAPI/ERD/runtime unchanged because
wire behavior unchanged; release publication/feedback resolution still withheld.

Exact final command:
```sh
rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E' -derivedDataPath /tmp/mypocket-reveal-approved-build-20261015 -resultBundlePath /tmp/mypocket-advisor-wire-units-20261015.xcresult -only-testing:MyPocketIOSTests -parallel-testing-enabled NO -collect-test-diagnostics never 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI' MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet
```

RED26992exit65,1/2PASS0skip: landscape actually rotated, composer/action inside
viewport and above keyboard, draft retained, but conversationheight7pt<44.
Largest-text case passes its focused-control checks; inspectedC202E4F9 nevertheless
shows tab glyph/label overflow at largest size. That is NOT full accessibility
acceptance. Landscape captureAB56E211 inspected; allscreenchrome still needs audit.
Bundle `/tmp/mypocket-chat-adaptive-red-20261014.xcresult`.

Root evidence: MPChatComposer always permits4lines, even compact vertical size;
header/navigation/keyboard consume fixed parts of the375pt landscape viewport.
One bounded base fix: verticalSizeClass compact caps input to1line with internal
scroll, regular retains1–4lines; fonts, draft, actions/navigation unchanged. This
is within existing1–4line maximum, not a new screen/architecture/config. No
orientation setting change needed (actual app already rotated despite no plist
orientation list). Still separately verify API pinnedoverview and largest-type
navigation before declaring landscape/fulla11y complete. No Feedback placement fix.

First fix gate10130exit65,272PASS/1landscapeFAIL0skip; conversation still7pt.
Bundle `/tmp/mypocket-chat-adaptive-green-20261014.xcresult`. One-line cap does
not address root cause; reverted only this turn's MPChatComposer environment/cap
change, preserving all preexisting edits. No successful product fix claimed.
Next diagnostic gate attaches numeric app/header/composer/input/action/viewport/
keyboard/tab/overview frames, adds actual local API pinnedoverview landscape
variant. Do not infer enabled/realoverview behavior from fake-session-only tests.

Diagnostic82726exit65: API overview variant confirms0pt conversation. Fake variant
also hit an instrumentation bug (querying absent overview.frame); fixed that
query to report.zero when absent, NOT a product change or weakened assertion.
Final88928exit65,272PASS/2landscapeFAIL0skip (268unit+4otherUI pass; both landscape
UI fail). `/tmp/mypocket-chat-adaptive-final-diagnostic-20261014.xcresult`.
No runtime warnings. ActualAPI frame document9E8C4D1E: app667×375, header44pt slot,
overview44pt+y44, conversation0 at y96; composer y103.5/h53, tab y129/h37,
keyboard begins213. Composer overlaps tab region; all fixed sections cannot fit
the keyboard-reduced shell viewport. Fake114B7397: no overview, conversation7pt.
Capture843995F5 inspected. One-line cap unchanged53pt composer and therefore was
not root fix. Largest-text control gate passes, but visual tab overflow remains
unaccepted; do not call this full Dynamic Type support. Normal chat/APIoverview/
draft/keyboard regressions and268units pass. Guards/project/diffpass.

Approach sent for human review (bounded existing flows, not an approved change):
while composing, hide bottom tabs and restore on explicit keyboard dismissal;
move overview access into header action, retain same details sheet/source, add
explicit keyboard-dismiss action. At largest accessibility sizes use a full-label
tab-selection menu through existing selector base rather than clipping4labels in
fixed-height bar. Preserve palette/Manrope/draft/security/provider/read-only
semantics, min44pt and normal navigation. No new UI code until design approval.
Then update screen/base contract and prove normal/AX/landscape/actualoverview,
keyboard dismissal/tab switching/draft/error/history/privacy behavior.

No loop guard reset:4verification invocations,1product attempt reverted; pause now
for interaction decision, not another geometry guess. No production/style change
shipped this turn; tests + evidence/docs only. Main goal remains in_progress.

Approval received: implement the proposed keyboard-dependent navigation/overview
and full-label accessibility menu. Preserve the diagnostic history above. Extend
UI tests first for hidden/restored tabs, explicit dismissal, retained draft,
overview sheet access while typing and full-label menu navigation. Composer base
owns focus dismissal via an explicit request; shell owns its available viewport.
No API/schema/public API contract changes; human/agent assistant guide and screen
contract need reconciliation after verified implementation.

Approved implementation evidence (not a loop-history reset): RED72363exit65
confirmed normal keyboard still showed tabs and lacked explicit dismiss action.
Implementation then uses shell available height, composer-owned focus request,
header overview/dismiss actions and shared MPMenuSelector accessibility variant.
GREEN9200exit0:274/274PASS0skip (268units+6UI), no runtime warnings. Bundle
`/tmp/mypocket-chat-approved-green-20261014.xcresult`. ActualAPI/fake landscape
focused conversation both63pt; action44×44, composer above keyboard, overview
sheet reachable, dismissal restores tabs and draft retained. Numeric attachments
7966D25E/4DE4EC5B; exported captures inspected under
`/tmp/mypocket-chat-approved-visual-20261014`. Largest-text keyboard and restored
full-label menu inspected; existing oversized Feedback at AX size remains a
separate unaccepted placement/style issue, not fixed by this slice.

Final validation extends AX menu switching to Account and back with draft intact,
captures landscape BEFORE opening overview (opening sheet dismisses keyboard),
and includes suggestions/disabled-send/multiline tests. Running; record terminal
result before final claim. Base guard, guard tests, project verification and
git diff --check pass. Human+agent guide, screen/registry, native architecture
reconciled. Swagger/ERD unchanged because API/schema unchanged. No release note
published or owner-feedback marked resolved; owner acceptance pending.

Final20967exit0:276/276PASS0skip (268units+8UI), no runtime warnings.
`/tmp/mypocket-chat-approved-final-20261014.xcresult`. AX menu Account→Assistant
preserves full draft; normal tab roundtrip, suggestions, disabled-send multiline,
actual API overview and landscape header sheet/dismiss gates pass. Focused
landscape capture E24A758E and numeric14FD5740 inspected:63pt conversation,
53pt composer,44pt send; exported screenshot retains known portrait-buffer crop,
so this is NOT pixel-perfect full-landscape render acceptance. Compact screenshot
B603C54E caught transition (keyboard AX exists before image renders); do not use
that image as stable keyboard visual proof. Largest-text captures inspected from
first green run. Base/guardtests/project/diff pass. No live LLM, physical-device,
dark-mode/reduced-motion/VoiceOver/whole-app acceptance inferred.
