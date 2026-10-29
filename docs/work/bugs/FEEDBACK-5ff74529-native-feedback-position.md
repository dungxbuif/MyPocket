# Feedback 5ff74529: native bubble position and Account Add

Status: in_review. Owner authorization: the 2026-10-26 follow-up request.
Scope: shared native shell feedback placement/clearance, focused UI and base
tests, and matching internal docs. AI extraction, proposal card content,
ordinary transaction entry, finance behavior, backend, production feedback
status and deployment are outside this work.

## Intent and current evidence

The owner wants the Feedback trigger to stay in a predictable position without
covering AI review actions, while Account root retains the global Add action.
The earlier `c86b669d` two-card iPhone 16 Simulator capture shows the lower
card's approval button partly behind the bubble. Current `MainTabView` places
the 52pt trigger over the bounded route viewport using
`MPFeedbackBubbleLayout`; `MPEditorScaffold` reports footer height but the
route still extends behind the shell-level overlay. Its final proposal can
therefore occupy the bubble's fixed region. The focused red simulator test
confirmed intersection: bubble `(329, 646, 44, 44)` and approval
`(258.08, 634, 107, 44)` in `/tmp/mypocket-feedback-5ff74529-red.xcresult`.
The report's
earlier jumping and missing Account Add have existing fixes and unit/UI checks;
they need fresh regression proof, not speculative rewrites.

## Decision and affected contracts

Use the previously approved [dedicated shell dock](../tickets/NATIVE-FEEDBACK-BUBBLE-DOCK-DETAIL_DESIGN.md)
to resolve the confirmed overlap. A fixed overlay cannot avoid arbitrary
scrolling controls; extra bottom padding only changes the final scroll stop.
`MainTabView` will give the existing `MPFeedbackBubble` a stable trailing 60pt
dock below the clipped route viewport and above the persistent navigation.
The dock stays in place while feedback capture/presentation hides the trigger;
keyboard input hides the trigger and can reclaim the dock's space. Reuse
`MPFeedbackBubbleLayout`, `MPBottomNavigation`, and shell viewport identifiers.
The bubble remains at least 44pt and never moves in response to row obstacles.
Verify the AI approval action, More/category reachability, bubble stability,
and Account root Add/child suppression. No API, config, schema, or data flow
changes are intended. This intentionally supersedes the later overlay-only
placement text in the native shell/base docs; those docs must be reconciled.

## Verification plan

1. Run the focused iPhone 16 UI reproduction against the deterministic
   two-proposal AI review fixture before implementation (RED), recording card,
   approval, bubble and viewport frames. This is recorded above.
2. Run the same test after the shared fix (GREEN), inspect a screenshot, and
   run stable-position, More/category reachability, Account Add, and feedback
   capture/entry regressions. Run native base/project guards and the required
   design checks/build.
3. Reconcile the native shell/base contract and AI-entry screen spec. Keep
   physical-device, populated production state, and owner visual acceptance
   explicitly open. Do not change production feedback status.

## Local result and docs review

- RED: `testAIEntryImportShowsDistinctTransactionReviewCards` failed on the
  iPhone 16 Simulator with the measured bubble/approval intersection above
  (`/tmp/mypocket-feedback-5ff74529-red.xcresult`).
- GREEN: the same test passed after `MPFeedbackDock` was placed between the
  bounded route viewport and navbar; its bubble stays within the dock, outside
  the viewport and last approval action, and at the same frame after scrolling
  (`/tmp/mypocket-feedback-5ff74529-green1.xcresult`). The captured two-card
  screen was visually inspected.
- PASS: four focused iPhone 16 shell/UI checks covered AI cards, More/category
  reachability, Account root Add with cancel return, and hiding Feedback before
  screenshot capture (`/tmp/mypocket-feedback-5ff74529-neighbors.xcresult`).
  Three additional secondary-route viewport checks passed
  (`/tmp/mypocket-feedback-5ff74529-routes.xcresult`).
- PASS: the two-card overlap and Account Add checks also passed on iPhone SE
  (3rd generation); its captured review screen was visually inspected
  (`/tmp/mypocket-feedback-5ff74529-se.xcresult`).
- PASS: the dock-layout unit test passed, and the Account bubble lower-edge tap
  passed on a dedicated iPhone 16 Simulator
  (`/tmp/mypocket-feedback-5ff74529-edge-isolated.xcresult`). An earlier
  lower-edge run on the shared simulator was interrupted by another build's
  installation of `com.mypocket.ios`: `runningboardd` logged an
  `installcoordinationd` termination at 21:28:33, after the tap but before
  the feedback sheet assertion (`/tmp/mypocket-feedback-5ff74529-edge.xcresult`).
  The isolated pass resolves that run as simulator interference, not a
  reproduced hit-area defect.
- PASS: native base-component and project guards; `app/` design check, seven
  design tests, and production build. The web build retained its existing
  chunk-size warning.
- Internal native shell/base/token/AI screen docs now describe the 60pt dock.
  No API, schema, finance flow, setup or configuration changed. No public guide
  or agent API reference needs revision for this visual-only native placement;
  these Markdown contracts and the focused UI proof are the agent-readable
  record. Historical overlay-only decisions remain traceable as history.

Still open: affected deployed/physical-device visual and touch acceptance,
varied populated routes/Dynamic Type, and owner signoff. The
focused simulator runs do not close the production feedback. Production status,
deployment and financial data were not touched.

## Owner follow-up: floating overlay (2026-10-29)

The owner clarified that Feedback must float above the route and must not reserve
a content lane. This supersedes the 60pt shell-dock allocation above while
preserving the same trailing bubble geometry and keyboard/capture visibility
rules. The route viewport now subtracts only the bottom navigation height; the
transparent Feedback host is layered over that viewport. The expected regression
is a route height equal to container height minus navigation height, with no
additional Feedback subtraction.

Verification: focused `ReferenceDisplayTests` route-height regression, the
native base/project guards, simulator test suite, and a signed Release device
build. No API, config, schema, finance flow, or production feedback status
changes are intended.

## Follow-up result

- RED: the new route-height test failed to compile before the production layout
  helper existed.
- GREEN: focused route-height test passed; the combined reference/feedback unit
  slice passed 19/19; and the shell/editor UI regression passed 1/1 on the iPhone
  16 Simulator.
- PASS: base-component guard, project verification, `git diff --check`, signed
  Release build with Swift warnings treated as errors, release-artifact preflight,
  and strict deep code-sign verification.
- Device artifact: `/tmp/mypocket-1.3.0-build6-device-20261029/Build/Products/Release-iphoneos/MyPocketIOS.app`.

Physical-device installation and owner visual acceptance remain open because
the connected iPhone was offline during this run.

Trace: [feedback audit](../audits/FEEDBACK-UNRESOLVED-UAT-20261020.md),
[prior dock decision](../tickets/NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md),
[earlier collision investigation](NATIVE-LIST-FEEDBACK-BUBBLE-COLLISION.md),
[AI card evidence](FEEDBACK-c86b669d-native-ai-review-cards.md),
[native shell contract](../../design/platforms/ios.md),
[base contract](../../design/system/BASE_COMPONENTS.md),
[AI screen spec](../../design/pages/assistant/README.md),
[validation matrix](../VALIDATION_MATRIX.md).
