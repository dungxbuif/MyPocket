# Native draft recovery and atomic-transfer fault E2E

Status: in_review. Existing approved offline Task6/7 verification; full goal open.

Latest result supersedes ordinary rejected-transfer residual below:52081exit0,
269/269PASS0skip (268units+1actualAPI UI), no runtime warnings.
`/tmp/mypocket-rejected-transfer-20261014.xcresult`. Restart preserves draft;
unchanged confirmation creates no new key; changed22500 uses second key, yields
one excluded pair, source77500/destination122500, monthly income/expense/count
unchanged. Confirmed server-ID detail scoped note verified; capture0134D5FF
inspected under `/tmp/mypocket-rejected-transfer-visual-20261014`:22500, pair
edit/delete and report exclusion visible. No product fix needed. Savings-specific
recovery remains blocked/unverified. Guards/guardtests/project/diff pass.
Owned proxy8206 cwd/command validated then stopped; normal8081 simulator build
15493exit0. No production/physical-device/release or destructive reset.

2026-10-14 continuation: previous goal turn made verified adaptive-chat progress.
Next independent gate is rejected transfer editing in ordinary/aggregate scope,
without touching the blocked savings selector helper. Reuse existing recovery
scenario with verifyGoalScope=false and editRejectedTransfer=true. Prove restart,
unchanged-confirm refusal/no new key, edited22500/new key, exactly one paired
transfer, correct two balances and unchanged monthly report; open confirmed row
detail and inspect capture. Test-only addition first; no product/API/DB/style
change unless observed failure justifies it. Run all native units plus this UI
against disposable8081 and owned loopback18081; stop proxy and restore normal
8081 simulator build afterwards. Savings-specific gate remains blocked, not
substituted by this independent proof. Reconcile human/agent recovery guide and
validation after terminal evidence; no production/install/release/feedback status.

Before changes: local store/coordinator recovery and ordinary503/commit-drop/
foreground tests pass, but rejected editor/date/discard and transfer failure UI
have no real API proof. Extend only loopback fault infrastructure to scope POST
/transactions/transfer by source_wallet_id as well as ordinary wallet_id.
No production fault route or raw body/key logging. TDD proxy scope before code.

Real native UI scenarios, disposable8081+18081:

- 422 ordinary: preserve a deliberately previous-day date across restart/editor;
  unchanged confirmation must stay in editor/no new key/request. Change amount,
  confirm prepared replacement, one server row with changed amount/same date,
  exactlytwo keys (original/new). Never silently rewrite old intent.
- 409 conflict: use the same changed-confirmation/date scenario, but verify the
  conflict state also offers edit rather than transient retry or silent overwrite.
  Add loopback reject409 mode via proxy TDD before running that UI scenario.
- Discard rejected: warning cancel retains row; confirm removes local row, restart
  does not send and server balance remains unchanged.
- Discard uncertain committed: same warning; confirm never deletes/reverses the
  already committed server row. Read actual balance and row after restart.
- Transfer503 and commit-drop: explicit source/destination in native editor,
  queued note/status, restart and same-key retry, exactlytwo paired rows/onepair,
  expense/income correct direction, reportfalse, balances changed once.
  Read monthly report before/after too: income, expense and report count unchanged,
  not just checking flags on the pair.

Use existing MPOverflowMenu/MPConfirmDialog/prefilled editors; no feature-local
controls, new palette or style. Scroll menu clear of separately blocked Feedback
overlay for recovery verification; do not claim placement/accessibility fixed.
Test-only helpers/fixtures, no product code unless evidence exposes a defect;
record root cause and RED→GREEN before fixing. Full native units/base guards/
project/diff and human+agent/screen/current/backlog/validation/ledger reconcile.
No API/schema/env/auth/newADR changes expected; no Swagger/ERD work if unchanged.
No release publication/prod/device install or destructive reset in this gate.

Initial proof: ordinary recovery28248exit0,3/3PASS0skip;
/tmp/mypocket-recovery-actions-ui-20261014.xcresult. Transferfault4135exit0,
2/2PASS0skip;/tmp/mypocket-transfer-fault-ui-20261014.xcresult includes unchanged
monthly income/expense/reportcount, actual balances and original committed IDs.
Screenshots2A5DCF82/B4D298CC inspected: expected target leaves queue, but applied
ledger lies below old synthetic drafts; Feedback still overlays content. Before
strengthening proof: scroll to exact confirmed server-ID row, open native detail,
assert note in detail and capture it. This verifies acknowledged read/refetch/
navigation too, not just queue disappearance and API state. Overlay clearance
remains test navigation workaround, not a fourth placement fix/acceptance.

Detail assertion review: TransactionDetailView displays `Ghi chú: <note>`, not
the bare ledger subtitle. Scope the assertion to transaction-detail-overlay and
that prefixed text; otherwise an underlying ledger label could falsely satisfy
the detail check. This is test correctness, not a product defect/style change.

Scoped gate37055 failed2UI/263unitPASS. Root cause from hierarchyEA3C4887 under
/tmp/mypocket-recovery-scoped-debug-20261014: transaction-detail-overlay inherited
by five sibling accessibility groups/buttons; firstMatch is header only while
note is correctly rendered in third group (screenshotAE56F03C). Before fix:
use the same `.accessibilityElement(children: .contain)` convention already used
by shared queue/group bases on TransactionDetailView root. One named container,
children remain individually accessible; no combine/hide or visual/style change.
Assert unique container+scoped note, rerun ordinary/transfer recovery detail and
normal create/edit/transfer mutation UI. This is a bounded existing-contract
accessibility fix, no new base/control, API/data/security/runtime change.

UI/UX Pro Max search note: SwiftUI grouping search returned an unrelated settings
Form recommendation; narrower reader/focus search returned generic semantics,
not an exact native grouping match. Neither result was persisted as design advice.
Fix uses the existing repository's contain-group pattern and observed AX evidence,
with built-in semantic-control/accessibility defaults; colors/fonts/geometry
unchanged. Darkmode/landscape/DynamicType/physical VoiceOver remain open, not
inferred from the automation identity fix.

## Verification checkpoint

- Proxy TDD: transfer scoped fault initially201 instead of503;409 control initially
  refused400. `rtk proxy node --test mobile-ios/scripts/offline-fault-proxy.test.mjs`
  final5/5PASS; no raw keys/payload logging or remote upstreams.
- Recovery28248:3/3PASS0skip (422 changed-only/date, rejected discard,
  uncertain-committed discard). Transfer4135:2/2PASS0skip, same paired IDs,
  correct balances and unchanged monthly income/expense/reportcount.
- Regression32155:267/267PASS0skip (263unit+409 and existing three fault/lifecycle
  UI), /tmp/mypocket-recovery-regression-final-20261014.xcresult.
- Detail59472:4/4PASS0skip opens confirmed ordinary422/409 and transfer503/drop
  rows by serverID. Its original global note assertion was weaker than intended;
  do not count it as scoped detail-content proof.
- Scoped37055:263unitPASS/2UIFAIL exposes duplicate screen identity. Contain-group
  fix35239:265/265PASS0skip/no runtime warnings (263unit+ordinary409/transferdrop),
  /tmp/mypocket-recovery-contained-final-20261014.xcresult. Unique detail identity
  and prefixed note inside it verified, not underlying ledger text. Final B8F61E91
  inspected under /tmp/mypocket-recovery-contained-visual-20261014:22500, previous
  day13/10, correct synthetic note; palette/layout unchanged.
- Normal44956exit0:265/265PASS0skip (263unit+ordinary create/edit+atomic transfer
  create/edit/delete), /tmp/mypocket-recovery-normal-final-20261014.xcresult.

Fault UI builds use `DEBUG LOCAL_API_UI OFFLINE_FAULT_UI`, app base18081 and
fixtures8081; normal builds use `DEBUG LOCAL_API_UI`, base8081. All use
SE182D41A2-4FC6-445B-A592-70FDCB5C662E, warnings-as-errors, sequential UI and
diagnosticsnever; select named LaunchTests methods. Fresh xcresult path on rerun.
Guards/guardtests/project/diff pass. Proxy67789 stopped after validated cwd;
no database/app reset or production/iPhone writes.

Residual gates: rejected transfer editor UI, goal-scoped queue visibility,
cache/overview/report and auth/device/radio coverage, landscape/dark/DynamicType/
VoiceOver/full native acceptance. Covered underlying ledger remains in AX while
detail overlay is open (observed in failed hierarchy): separate content-visibility
audit, not solved by unique grouping. Feedback placement remains separately
blocked; no fourth speculative fix. No whole offline/Task5–7/goal completion.
Human+agent guide/screen/registry and context/backlog/validation/ledger reconciled.
API/schema/config/security behavior unchanged, so Swagger/ERD/newADR unchanged.
Release notes/publication remain withheld pending owner acceptance.

Links: [recovery contract](OFFLINE-QUEUE-RECOVERY-20261014.md),
[full goal](NATIVE-FULL-COMPLETION-20261014.md),
[plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[validation](../VALIDATION_MATRIX.md), [public guide](../../guides/offline-sync-api.md).
