---
artifact_type: detail_design
id: NATIVE-FEEDBACK-CONSISTENCY-AUTH
status: in_progress
platform: ios
approval: owner instructed implementation and delegated remaining UX decisions on 2026-10-13
---

# Native shared presentation and remembered Face ID sign-in

Links: [feedback intake](NATIVE-MOBILE-FEEDBACK-20261013.md),
[backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[native architecture](../../architecture/MOBILE_IOS.md),
[base contracts](../../design/system/BASE_COMPONENTS.md).
Release notes remain deferred until owner review, per explicit owner instruction.

## Owner refinement: button corners (2026-10-13)

The owner clarified that reduced rounding applies to buttons, not cards.
Change the shared native `MPMetrics.buttonCornerRadius` from 999pt to 8pt;
all MPButton surfaces/borders consume this token. Preserve the 48pt control
height and 44pt minimum hit target, card radius and circular feedback/icon
actions. No web, API, database or authentication changes in this refinement.
UI/UX Pro Max touch guidance supports preserving the iOS 44pt hit target;
8pt rounding is the owner's visual direction, not a library requirement.
Verification: token regression RED/GREEN, native unit suite and Simulator
render inspection. Update public design tokens and internal validation/context;
release publication remains deferred.

## Intent and diagnosis

The owner authorized fixing the remaining iOS feedback, choosing routine UX
details autonomously and checking rendered screens on Simulator. Cards and
transaction items must look consistent across screens. Existing consumers use
MPTransactionRow but independently derive titles, notes, icons, signed amounts
and transfer tones. Wallet history in particular loses the category glyph and
renders transfers as income/expense. SurfaceCard's standard/reference styles
also differ in padding, radius and elevation without a semantic reason.

Face ID currently unlocks a persisted session only. Logout clears that session
and the biometric preference. The approved requirement is remembered sign-in
after logout, with an explicit option to forget this device. Keeping the same
refresh token after logout would not provide independent session revocation.

## Execution slices

### Owner-approved finish pass (2026-10-13)

Latest instruction supersedes the dedicated dock: scroll viewport reaches the
bottom navigation; feedback is only a trailing 52pt overlay with 8pt bottom
clearance, no full-width lane, spacer or reserved height. Outside the bubble
hit area touches pass to the content. Hide the bubble before screenshot capture
and while the feedback panel is open, without relayout. Keep primary action
centers reachable by scrolling and check actual taps on compact phones; visual
overlap at the trailing edge is inherent to an overlay, not a dedicated lane.
Retire MPFeedbackDock, preserve MPFeedbackBubble, and bound the selected-account
tab hit shield to the navigation's real height if regression proves interception.
Simulator interaction revealed a second viewport problem: NavigationStack could
keep a larger scroll region than the root's clipped visible area when relying
on safeAreaInset alone. Last settings actions appeared only during overscroll,
then bounced out of reach; the logout tap never reached the API. Use an explicit
GeometryReader-sized route viewport and a sibling 56pt navbar, with bubble
overlay inside that viewport. Test both viewport edges and actual last-action
tap, not accessibility existence alone.

Compact shared presentation: card padding 12pt (was 16); transaction glyph
36pt and vertical padding 4pt (was 40/6); body 15pt, secondary 13pt, heading
18pt, summary currency 20pt, display 28pt. Labels and form inputs retain their
existing readable scale (UIKit inputs 16pt); controls retain 44pt hit targets
and 48pt standard control height. Card radius remains 12pt, button 8pt.
Keep Dynamic Type expansion and semantic colors; no screen-local compact
components. Verify actual rendered card/row size plus large-text expansion.
Profile identity becomes a compact horizontal avatar/name/email row above the
existing shared manage-account action. Card surfaces are flat per production
feedback (559242fe), replacing the previously assumed elevation. Header action
buttons also use the 8pt button radius rather than wrapping the base in capsules.

Feedback image attachment (ecf02904): reuse MPAttachmentThumb and add one shared
MPImageAttachmentPicker for photo selection/removal/replacement. Convert the
chosen image to PNG within the existing 1 MiB backend contract using the shared
ScreenSnapshot encoder; reject unreadable/oversized input with visible error.
Automatic capture pre-fills the same optional attachment; both bubble form and
feedback route expose it. The form fields scroll while the send action stays
visible. No endpoint, storage/schema or permission broadening is needed.

Fix Face ID test's missing tap before typing the enrollment confirmation
password, then test actual Keychain enrollment/logout/biometric login on the
disposable API. Investigate any product failure before modifying auth.
Read production feedback only with existing valid credentials; never infer
queue exhaustion from local tests or rotate user credentials silently.
Build/install native on the connected owner's phone after local gates pass;
do not publish release notes before owner approval. New backend auth rollout
must be additive and preserve all financial data.

1. Extend the existing MPTransactionRow with a single domain presentation
   adapter for TransactionModel. Use it on Overview, ledger, wallet history,
   credit history and Assistant transaction results. Keep portfolio trade rows
   on the same visual primitive with trade-specific semantics. Centralize
   category/icon, note/date, signed amount and neutral transfer/adjustment rules.
   Standardize SurfaceCard padding/radius/shadow through one shared token set;
   retain source-compatible styles, with no screen-local overrides.
2. Run populated native screens on iPhone 16 and compact iPhone SE using local
   synthetic API fixtures. Check transaction consistency, category connectors,
   text wrapping, icon semantics, tab height, last-row/action reachability and
   screenshot capture. Fix reproduced layout problems in bases.
3. Add independent, revocable device sign-in credentials to the Go API. Issue a
   256-bit random secret only after an authenticated session plus password
   verification; persist only its SHA-256 digest, account ID, expiry, password
   fingerprint and device identity in PostgreSQL. Default validity 180 days,
   independent of 30-day refresh tokens. Re-enrollment replaces the credential
   for that owner/device. Password changes invalidate it. It is a device bearer
   credential protected by iOS, not server verification of biometric data.
4. Store the device secret in an iOS Keychain item with ThisDeviceOnly and
   biometryCurrentSet access control. Face ID reads that item and exchanges it
   for a new normal session. Cancellation/network errors preserve enrollment;
   an expired/revoked credential returns the user to password sign-in. Logout
   revokes the current access session and refresh token while retaining device
   enrollment. Forget-device revokes enrollment and removes local credentials.
   Active app access continues using the existing refresh rotation contract.

## API and schema

- POST /api/v1/auth/devices: session-only, password + device_id + device_name;
  returns device credential secret once, ID and expiry. API keys cannot enroll.
- POST /api/v1/auth/device-login: device ID + secret; returns LoginOutput.
- DELETE /api/v1/auth/devices/:id: session-only, owner-scoped revoke.
- POST /api/v1/auth/logout: session-only; revoke current access session and
  submitted refresh token belonging to the same owner; no device revocation.
- Additive migration for device_credentials; no financial table/reset changes.
- Existing per-user local selected-wallet preference stays in UserDefaults;
  only the UI selection ID is local, financial data stays on the backend.

## Reused bases and scope

MPTransactionRow, MPTransactionGlyph, MPSurfaceCard, MPButton, MPText,
MPStatusMessage, MPCategoryTree and the existing shell/navigation bases.
No competing screen-specific card or transaction component. Changes target
mobile-ios plus the backend auth module needed by the accepted Face ID flow.
Web maintenance, production deployment, commit, device installation and release
publication are separate from this local implementation/verification scope.

## Verification and documentation

### Urgent owner density refinement — 2026-10-14

Owner reports the installed cards/items still waste vertical scroll space and
explicitly requests immediate correction/install without running tests. Reduce
shared card padding12→8, control48→44, transaction row52→44 normal text (no
outer vertical padding), glyph36→32, metric gaps8→4 and page section gaps→8.
Menu/category rows keep44pt minimum with smaller decorative icons. Profile is
one tappable identity row instead of identity plus duplicate manage row.
Do not compress financial text or freeze accessibility heights; Dynamic Type
can grow. Keep fonts/colors/navbar/API/accounting unchanged. Hide disabled
Account FAB so it does not consume visual room above navigation. Reuse existing
MP bases; no screen-specific card/item variants.

Tests explicitly skipped by owner; update intended geometry assertions but do
not report them passing. Mandatory Release compilation, strict signature/API URL
checks and non-uninstall device installation remain. Reconcile public tokens/
bases and internal validation; no release publication without owner acceptance.

Current authority (2026-10-14): later owner direction supersedes the dock/16pt
padding/16pt body proposal below. Shared cards are flat, padded12/radius12;
body15 medium/secondary13, profile horizontal. Navbar56 is a sibling of an
explicitly sized route viewport; bubble52 overlays it without a reserved lane.
Button/header corners8, circular icon actions retained. Optional feedback photo
uses shared MPImageAttachmentPicker/MPAttachmentThumb and bounded PNG. Current
tokens/bases/platform docs describe actual source. Auth-only prod rollout was
subsequently authorized; see [execution](NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md).

Production feedback 196110ec: ordinary/aggregate ledger removes the redundant
Giao dịch title and places the existing wallet-scope base in the header's
leading slot. Header actions remain available for basic wallets as well as
aggregate scope. MPScreenHeader supports an empty visual title; the shared
wallet selector supports leading alignment. No new screen-only control.

Earlier proposal, superseded above: use one 16pt/12pt-radius
card surface; plain row actions add no horizontal inset beyond their card;
body/secondary text uses 16pt/14pt; primary button uses the existing dark action
green for readable white labels. Consolidate category management and selection
into MPCategoryTree with exact 40pt/32pt glyph bounds and connectors drawn in
the measured row background. Keep the approved dock architecture but reduce its
trigger/lane from 60pt to 52pt; navigation stays at least 44pt interactive with
8pt total interior padding and 4pt bottom margin. Re-run compact-device collision
and reachability tests after these geometry changes.

Session families are recorded in Redis so logout also rejects earlier access
tokens minted before refresh; device enrollment remains independent. Enrollment
password rejection uses 403 to avoid treating a bad confirmation password as an
expired session and triggering automatic refresh. Existing legacy Redis session
values remain readable during an additive deployment.

Pre-deploy independent review identified legacy-family migration and delayed
refresh issuance races. Ruling: pre-upgrade access/refresh records use a stable
per-owner legacy family; a logout of such a session may sign out the owner's
other pre-upgrade sessions once (old records do not carry enough information
to distinguish original families). Fresh password/device logins use independent
random families and are unaffected. Recheck revocation after both credentials
are stored and before any response; deny/clean up a delayed issuance if logout
won the race. If logout occurs after this check, its marker expires later than
all already-issued refresh credentials. Tests use a deterministic interleaving
and TTL-aware fake cache to verify no post-marker resurrection. No financial
data or device enrollment is revoked by legacy session migration.

- Presentation regressions cover custom/system categories, transfers,
  adjustment direction, credit refund/payment, privacy masking and long notes.
- Auth tests cover enrollment password, digest-only persistence, owner
  isolation, expiry, password change, revocation, logout/refresh invalidation,
  Face ID cancellation, transport errors and forgotten devices.
- Native UI tests verify consistent row content, login actions, privacy,
  compact-screen wrapping and shell reachability using synthetic data.
- Run full native unit suite, relevant UI tests, Release warnings-as-errors,
  project/base guards and Go auth/integration regression suite.
- Reconcile AUTH_API, native architecture, API reference/OpenAPI and public
  human/agent docs for the new auth endpoints; update native screen/base specs,
  validation, backlog and context. Do not publish release notes before review.
