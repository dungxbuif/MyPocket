# Offline fault proxy and real native E2E

Status: in_review. Local fault/restart/retry gate passed; full native goal open.

Before code: standalone Node built-in HTTP proxy binds127.0.0.1:18081 only and
forwards solely to existing disposable API127.0.0.1:8081. Reject remote upstreams,
never change backend/prod config. Control requires local-test header, wallet-scoped
modes pass/reject503/reject422/commit-then-drop. Only matching POST transaction
creates affected; other login/read/capability/fixture calls pass through. Stats
contain counts/statuses and SHA256key references, never tokens/payload/key values.
CLI/tests include no secrets. Temporary proxy is test infrastructure, not a product
limit/production endpoint. App build targets proxy; UITest fixture HTTP stays8081.

TDD proxy boundary before implementation, then actual nativeUI tests: create via
wallet form, assert pending/not counted in serverstate, terminate/relaunch same
app (no uninstall), same intent still visible; restore pass, explicitRetryNow,
assert one server row and one unique key reference. Commit-then-drop proves
ambiguous response after real commit, replay200 and exactlyone ledger row.
Capture actual pending screenshots. Synthetic fixture wallet/rows cleanup, keep
receipts as audit. Production/iPhone untouched. Full units/guards/project/diff and
human+agent/internal docs reconcile; do not claim full offline from two cases.

ActualUIRED64301: queue exists after503 but lacks identifying note. Beforefix:
show existing frozen note as wrapped secondary MPText, group each item and section
with accessibility children contain so actions associate with the correct draft.
No new styling/financial data source; note remains encrypted at rest. Target only
the synthetic note in E2E, not whichever queue item happens to appear first.

Visualreview after503GREEN5341: exported pending4AB4A202 under
/tmp/mypocket-offline-pending-visual-20261014 shows correct draft note/status and
recovery actions, but two drafts occupy most viewport; Feedback covers firstamount
(known blockedplacement, no fourthguess). Bounded compactdesign presented via
brainstorming: retain amount/note/status, move peritem actions to shared3dot menu,
no palette/font/logic change. Approval was pending at that checkpoint and was
subsequently received below; no code preceded approval.

Owner approved “compact + menu ba chấm” on 2026-10-14. Before implementation:
retain type, amount, wrapped note and status; replace full-width per-item recovery
buttons with existing MPOverflowMenu (44pt target, descriptive accessible label).
Only eligible actions appear; disable menu during sync/action. Keep discard
confirmation, retry key, replacement semantics, global sync and palette/font/base
unchanged. No backend contract change; Feedback placement remains separately blocked.
TDD: real restart/retry E2E must open that item's menu before retry; observe RED
against old layout, then GREEN for both503 and lost committed response. Inspect
375pt pending screenshot, run full native units/base guards and reconcile guides.

Pre-compact evidence: unit run53964 passed252/252, no skips/runtime warnings;
fault E2E22418 passed2/2, no skips. These do not verify the compact interaction.

Compact RED34418: missing overflow control, expected. First GREEN attempt23703:
1/2 passed; lost-commit retry verified,503 failed because the known Feedback
bubble intercepted the menu tap and opened Feedback instead. Exported screenshot
0D571EA1 and hierarchy26F65E74 under /tmp/mypocket-compact-visual-20261014 confirm
three73pt rows/44pt menu targets and overlay on third menu. No fourth placement
fix: that design remains blocked. For recovery behavior verification only, scroll
the real row clear of the bubble before opening the menu. This is not evidence
that overlay placement is accepted or accessible in all positions.

Final compact E2E73398 exit0,2/2PASS0skip/no runtime warnings:
/tmp/mypocket-compact-green2-20261014.xcresult. Both503 and real commit/drop
exercise the menu after app restart, verify same intent/key and exactlyone server
row with correct amount/note/category. Final commit/drop stats201,200,200;
twodrops,onehashedkey. Proxy tests3/3PASS; native base guard/guardtests/project/
diff checks pass. Exported pending3CF00F85 inspected on375pt SE: short-note rows
compact, all four fields present; old failed synthetic test drafts persist by
design. Feedback now overlays wallet selector in this capture: still blocked.

Final normal-base unit29293exit0,252/252PASS0skip/no runtime warnings:
/tmp/mypocket-compact-unit-20261014.xcresult. Build uses8081 again; temporary
owned proxy49662 stopped,18081 no listener; API8081 health200. No test data wipe.

Docs review: transactions screen/registry, public guide including structured
AI-agent reference updated; current/backlog/validation/ledger reconciled. No new
API/config/schema/auth/dependency/architecture decision from compact UI, so no
Swagger/ERD/newADR changes. No release-note publication before owner acceptance.
Remaining: discard/rejectededit/transfer error UI, scheduler/background/auth,
landscape/DynamicType/darkmode/accessibility and physical-device UAT. No prod
rollout/install and no full offline/native goal completion.
