# Native persisted Assistant history verification

Status: in_review — local history proof complete, broader chat/goal acceptance open. Existing approved UI plan Task5/6 history verification,
not a new product flow/design. Previous goal turn added adapter coverage and
273/273unit PASS. Numeric/business/Feedback/cache bounded reviews stay pending.

Before implementation: read AuthViewModel local fixture login, native conversation
UserDefaults key/history/scroll behavior, FinanceAPI adapter and backend history
handler/repository/schema. Dedicated local container
`mypocket-native-feedback-pg-20261013`, database mypocket_feedback, role fixture.
Read-only query finds fixture@example.com owner e4aedf4a-1aa7-4918-ba09-0287c95b154c
has0conversations/0messages. No production or provider operation.

## Verification approach

Seed one exact synthetic conversation with51messages/26completed synthetic runs;
do not use queued runs or call a model. Tooling must require a local Docker Unix
endpoint, exact container/database/role and fixture owner/email. Fail create if
owner has any conversation. Cleanup only exact conversation, after checking
owner/count/prefix; never reset an app/database or delete by broad prefix.
No financial rows touched. Seed IDs/content remain reproducible.

First gate40442exit65:273unitPASS/1historyUIFAIL0skip/no runtime warnings,
`/tmp/mypocket-advisor-history-e2e-20261015.xcresult`. Actual latest50 and older1
API readbacks, native old01 loading and load-more disappearance pass. Test then
expects offscreen51 in AX tree; hierarchy076D2CE6 lists only visible-nearby01…17,
so count0/missing-frame errors are a test visibility assumption, not proven data
loss. Correction1: assert old01 visibility/no jump, scroll the identified chat to
51 before checking its UI uniqueness, then restart and independently re-read API
rows unchanged. Preserve literals/data assertions, no product patch. Keep prior
failure evidence. Fixture create51 and duplicate-create refusal (still51) verified.

XCTest uses normal authenticated API launch and Foundation argument-domain
override of the existing owner-scoped conversation key (not a new runtime flag
or UserDefaults write). Read actual local API latest50 (seq51…2), scroll the
identified assistant-conversation to Tải tin cũ, load seq1, verify oldest and latest
once, no jump to latest, and restart/re-read unchanged persisted rows. No chat
send/cancel/provider call. Capture populated history. This does not prove live
provider cancellation, send/cancel races, full card safety/privacy or owner UAT.

Characterize existing behavior; if failure, inspect root before product changes.
Keep all assertions; no new product code under this verification scope. Run all
units+this actual API/native UI test, guards/project/diff. Retain failing evidence
and loop guard if any fixture/scroll correction is needed.

## Documentation/reconciliation

Internal: this ticket, chat ticket, context/backlog/validation/fullgoal/planledger.
Public human/agent API guide/catalog/OpenAPI/ERD/config/auth unchanged (verification
only, no product/wire behavior change). No release/feedback resolution before
owner acceptance. No physical install/deployment/reset. Cleanup synthetic fixture
and verify0remaining before ending the slice.

Links: [goal](NATIVE-FULL-COMPLETION-20261014.md),
[chat](NATIVE-CHAT-ADAPTIVE-20261014.md), [validation](../VALIDATION_MATRIX.md),
[plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[screen](../../design/pages/finance-assistant/README.md).

## Final local evidence

27531exit0:274/274PASS0skip/no runtime warnings (273native units+1actual API/UI),
`/tmp/mypocket-advisor-history-final-20261015.xcresult`. LatestAPI50seq51…2 and
olderAPIseq1; UI scroll/load final old page, one visible01, no jump to51, scroll
back to one visible51, restart/latest and independently identical API messages
all pass. Characterization of existing product behavior, not a product bug fix.
Capture76D78B35 inspected: pinned overview, trailing user bubbles/flat assistant
messages and composer/navigation visible. No universal Feedback/a11y/card-safety
acceptance inferred from that one populated capture.

Fixture script `mobile-ios/scripts/manage-advisor-history-fixture.sh` create
exits0/51messages. Duplicate create exits3/no overwrite/still51. Remote Docker
override probe `rtk proxy env DOCKER_HOST=tcp://127.0.0.1:9 bash
mobile-ios/scripts/manage-advisor-history-fixture.sh create` exits1 with
Refusing non-local Docker endpoint before DB access. Cleanup exits0/deletes only
the exact synthetic conversation and cascades26runs/51messages; subsequent
exact-ID SELECT counts0|0|0. Regenerable synthetic data only, no user finance.
Default normal8081 simulator launch47423 removes argument-domain override;
no app/UserDefaults/DB reset, no persistent conversation-key write in test.
Syntax/guards/guardtests/project/diff gates pass. No tests currently running.

Command (from repo root, after script create):
```sh
rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E' -derivedDataPath /tmp/mypocket-reveal-approved-build-20261015 -resultBundlePath /tmp/mypocket-advisor-history-final-20261015.xcresult -only-testing:MyPocketIOSTests -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedAssistantLoadsOlderPersistedHistoryWithoutJumpingToLatest -parallel-testing-enabled NO -collect-test-diagnostics never 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI' MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet
```
Then `rtk proxy bash mobile-ios/scripts/manage-advisor-history-fixture.sh cleanup`.
Docs reconciled: ticket/chat/context/backlog/validation/fullgoal/ledger, screen,
public human+agent native guide. No API/ERD/auth/config/ADR changes because no
product contract changed. UAT not required for added test infrastructure; overall
chat/owner/device UAT remains open. No release note/feedback resolution/publication,
production deployment or physical installation. Live cancellation/race and all
other full-goal gaps remain explicitly open.
