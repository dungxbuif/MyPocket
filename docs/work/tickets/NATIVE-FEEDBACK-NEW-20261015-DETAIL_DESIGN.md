# Latest six native feedback items — reconciliation/design

Status: bounded implementation complete for unused-asset deletion and stable
feedback dock; review_required for the remaining changed UI/lifecycle/business
contracts.
Source: authenticated production Feedback API,48records fetched2026-10-15;
all status=open. This document contains engineering proposals, not production
status updates or instructions granting an LLM arbitrary HTTP/DB access.
Parent: [current reconciliation](NATIVE-RECONCILIATION-20261015.md).

Placement supersession: the earlier lower-left proposal and checkpoint below
are historical. The current approved shell contract is a stable lower-right
overlay; see [NATIVE-FEEDBACK-RIGHT-DOCK-20261015](../bugs/NATIVE-FEEDBACK-RIGHT-DOCK-20261015.md)
and the top section of the validation matrix.

## Actual source findings

| Feedback | Existing implementation | Remaining requirement/proposed change |
| --- | --- | --- |
| b017284a portfolio edit/delete and price | Native has unit-price input and quantity, create trade/price update/history; asset-delete API exists but no card action. Backend refuses deleting assets with trades; no rename/trade edit/delete contract | **Implemented bounded scope:** empty assets expose `portfolio-delete-{assetID}` with shared confirmation/error and reload through the existing owner-scoped `DELETE /portfolio/assets/{id}` route. Traded assets remain append-only and expose no delete action. Add reviewed asset metadata/trade correction contract before historic mutation; re-run position/realized-PnL invariants. Unit-price×quantity semantics remain unchanged and explicit; no investment history is silently deleted |
|1565a9e5 modal/close instead of unnecessary navigation | Shared mpFormBottomSheet/mpFullScreenCover and MPEditorScaffold exist | Short choices/notes/price/confirmation stay sheets with Close/Cancel; long wallet/budget/jar/category/trade editors full screen. Real task detail/history/report screens retain navigation, not converted indiscriminately |
|b209e80f edge-swipe Back | Secondary routes use custom onBack/state, not uniformly native NavigationStack | One shared route host/interactive-back contract; gesture only for child routes, no accidental financial save/discard. Unsaved long forms use existing cancel confirmation; root tabs do not swipe-back. Do not attach competing screen-local gestures |
|5ff74529 fixed Feedback and Account Add | Current bubble dynamically avoids obstacles; shell disables Add on account/assistant | Replace dynamic jumping with stable shell-owned lower-right anchor, above navbar/composer/footer; no reserved scroll lane. Bottom content scroll inset only lets last rows move clear of bubble. Retain minimum44hit target, keyboard-safe anchor, hide duplicate bubble while feedback dialog open. Enable global Add on root Account, preserving destination/draft; child editors must not acquire competing Add |
|f42c103d transaction items too close | Shared MPTransactionRow uses compact family tokens | Tune shared row separation rather than individual screens: propose vertical inset6 instead of4, retain minimum44hit area, thin divider and4pt internal text rhythm. Verify Overview/ledger/wallet/AI/portfolio together; do not enlarge all cards/fonts or revert compact metrics |
|824dea29 AI cannot create jar/full user API access | Advisor read tools plus prepared wallet/jar proposals; owner API keys are a separate full-access auth policy locally implemented | **Implemented locally for jar create/update:** typed schema, owner/run/generation binding, confirmation cards and owner-scoped create/update dispatch. `propose_update_jar` validates UUID/month/mode and PUT replacement is replay-safe. Expand tools across transactions/wallets/categories/budgets/goals/recurring/travel/portfolio only with matching domain validation. Never unrestricted URLs/SQL, service-token/admin access, unconfirmed financial writes or model-held credentials |

## Screen/component application for the changed feedback position

UI/UX Pro Max source check: SwiftUI stack search returned NavigationStack,
navigationDestination and environment dismiss, applicable to shared native route
hosting. Initial focus-not-obscured results were Web-only and are not claimed as
native conformance; narrower app-interface safe-area search supports native
insets/gesture-bar clearance. Native pro-rules supply minimum44pt targets and
keyboard/footer protection. No unverified palette or web CSS recommendation used.

| Screen group | Bases reused | Placement/interaction gate |
| --- | --- | --- |
| Overview, ledger, wallet details, savings/credit | MPTransactionRow/MPSurfaceCard/MPBottomNavigation + shell bubble | Stable lower-right above navbar; last row/menu remains reachable by scrolling, no permanent viewport lane |
| Assistant | MPChatComposer/MPAssistantCard/header menu + shell bubble | Lower-right above composer, respects keyboard; no hopping between message/text controls, does not cover Send or conversation action |
| Account/root More | MPMenuRow/MPBottomNavigation + shell bubble | Root Add usable, navigation shielding still blocks tap-through; bubble does not cover bottom menu action |
| Long editor routes | MPEditorScaffold/mpFullScreenCover | Save/Cancel pinned and unobscured; bubble anchor cannot overlap footer, no hidden draft/save on edge swipe |
| Note/price/picker/feedback dialogs | mpFormBottomSheet/MPTextEditor/MPOptionSelector/MPConfirmationDialog | Existing Close/Cancel, single active presentation; no secondary bubble over its own dialog |

The fixed-anchor requirement supersedes proposed obstacle-driven movement, but
does not waive protected-control reachability. A permanently fixed physical
screen coordinate conflicts with keyboard/footer movement; selected proposal is
a stable anchor relative to the currently visible navigation/composer/footer,
not arbitrary movement among free rectangles. The selected proposal was approved
for this bounded base change; protected-control and device/UAT review remains a
separate gate.

### Local implementation checkpoint — 2026-10-15

The shell now returns the lower-right dock position from the shared
`MPFeedbackBubbleLayout` regardless of obstacle reports. Footer clearance still
keeps the bubble above the visible composer/navigation, while obstacle geometry
no longer makes it jump sides or reserve a scroll lane. The regression test
`testFloatingFeedbackUsesStableLowerRightDockAndNeverReservesObstacleLane` was
red against the old obstacle-driven algorithm and is green after the bounded
base-component change. Full DesignSystem tests pass 90/90 and the warnings-as-
errors simulator build plus base/project guards pass. This proves the shared
layout contract only; populated-screen/device/UAT overlap proof remains open.

### Refresh/save race guard — 2026-10-15

The feedback list must not accept a list response that was started before a
resolution mutation and completed after that mutation. `FeedbackViewModel`
therefore invalidates the list generation both when a resolution starts and
when its request settles. A delayed GET is ignored after a successful or
failed PATCH; the mutation response remains the source of truth for the row.
The iOS unit test uses a deliberately stale delayed GET (the response is
captured before PATCH) so a server-side echo of the new value cannot mask the
race. The focused simulator proof passed 1/1 at
`/tmp/mypocket-feedback-race-green-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_21-14-02-+0700.xcresult`.
No API, storage, or production behavior changes.

## Verification required

TDD actual UI: root Account Add opens manual entry and cancellation returns;
Feedback anchor stable across scroll and layout load, all last-item controls
reachable, dialogs/keyboard/large text/landscape safe. Back swipe leaves child
route, preserves drafts or confirms discard and never writes. Portfolio deletion
checks exact API absence, traded asset refuses deletion without loss, quantity
1.5×unit-price100000 implies gross150000 beforefee, persisted position/history
match; historical edit/delete needs replay/order/negative-position tests.

AI actions require realPG owner isolation, stale generation/expired/revoked
credentials, confirm/reject/replay/concurrent duplicate and domain state readback.
LLM only prepares; deterministic backend executes after confirmation. Live
provider routing evaluation is separate from deterministic financial correctness.

Affected internal docs: shell/base/screen registry, portfolio domain/API, advisor
actions/lifecycle and validation. Public human+Markdown-agent guides/OpenAPI must
follow actual implemented contracts; do not advertise draft tool capabilities.
No release notes/feedback resolution before owner acceptance, no production or
financial writes performed during this design/reconciliation. The local UI proof
is currently blocked by the disposable PostgreSQL fixture being unavailable at
`127.0.0.1:55434`; the app target still compiles with warnings-as-errors.
