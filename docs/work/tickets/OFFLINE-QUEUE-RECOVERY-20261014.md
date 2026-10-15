# Offline per-intent recovery

Status: in_progress; approved runtime/offline Task6, full native goal open.

Before code: retryNow changes only retryWait.nextAttemptAt, preserving immutable
intent/key/attempt count; other states refused. Discard is a local cancelled
transition only for pending/retryWait/conflict/rejected/blockedAuth; sending and
applied cannot be discarded by this command. Repeated cancelled discard idempotent.
UI confirmation warns unknown attempts may already exist on server; discard never
deletes/compensates server records and requests authoritative read refresh.

Conflict/rejected edit uses a new key only after changed draft and confirmation.
Atomically cancel old row and insert frozen new intent linked by supersedesID,
same owner/operation. Unchanged payload refused. Retrying same replacement key is
idempotent; changed replacement payload/key collision refuses. Pending/sending/
retryWait/authblocked cannot be replaced because outcome is uncertain. Existing
encrypted record schema gets optional supersedesID, backwards-compatible nil;
no plaintext/SQL/schema or backend migration.

TDD realSQLite: allowed/refused transitions, owner isolation, exactkey/payload on
retry, local-only cancellation, atomicreplacement/idempotency/concurrent collision
and rollback preservation. Coordinator checks session/epoch before publication;
shared queue UI uses MPButton/MPConfirmDialog and reused existing editors, no
new palette/card/input styles. Full units/APIUI/pendingvisual gates and docs human+
agent/registry/currentstate after evidence. No prod/install/release claim.

Editor integration comparison rule before fix: formatting the same occurred_at
instant in account timezone is not a changed draft. Compare validated logical
payload with timestamps normalized only for equality, never alter stored wire
payload or idempotency hash/key. Add timezone-equivalent replacement RED first.

Editor date rule: preserve queued timestamp exactly when date control was not
changed. Transfer date-only editor preserves original timestamp while selected
account-local calendar day is unchanged, instead of reassigning local noon just
because a rejected draft was opened. Ordinary preserves original fractional
instant when Date equals parsed original. New day selection keeps established
editor date behavior. Full actual rejected/pending editor UI proof remains a gate.

## Local checkpoint

Store retryNow/discard/atomic replaceRejected and optional encrypted supersedesID
implemented; no SQL/backend migration. Retry preserves key/body/attempt; local
discard excludes sending/applied; replacement only changed conflict/rejected,
sameowner/operation, idempotent childkey, oldcancel+newinsert oneSQLtransaction.
Semantic timestamp equality is comparison-only, persistedwire/hash unchanged.
Coordinator owner/epoch wrappers and same-key result path implemented. Sharedqueue
UI has RetryNow/Edit/Discard with MPButton/MPConfirmDialog, ambiguous-result warning,
and postdiscard readreload. Existing transaction/transfer editors reused/prefilled,
explicit confirmnewkey, preserve original queuedtime when date unchanged.

RuntimeRED93734 threecases, expanded6944 rollback/concurrent→13434GREEN248units.
CoordinatorRED88919three→10197GREEN251units. Date-equivalenceRED46389→11635GREEN
252units. Finaleditorbuild5403exit0,252/252unitsPASS0skip,
/tmp/mypocket-recovery-editor-20261014.xcresult. Actual SQLite trigger insertfailure
proves oldcancellation rollback; twostores/same replacementkey onechild; wrongowner/
state/key/content denied; account-switch baseline staysgreen. Guards/project/diff
pass. Normalcreate/editUI58897exit0,1/1PASS0skip,
/tmp/mypocket-recovery-normal-ui-20261014.xcresult; pending/conflict
editor/retry/discard visual/accessibility/networkE2E remain unverified. Newdate
editor logic is compiled, not direct datepicker UI proof. No fourthbubblefix.

Next failure-proxy/pending/restart/replay E2E, transfer/date editor checks, queue
visibility in goal-scoped ledger (current card is in ordinary aggregate/basic
scroller), foreground/network/background scheduler/auth/allremainingnativegaps.
No prod/install/release/fullgoal completion.
