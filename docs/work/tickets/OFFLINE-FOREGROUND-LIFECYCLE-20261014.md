# Foreground-only automatic financial sync

Status: in_review. Implements existing approved offline Task5/6 lifecycle,
not background financial sync. Full native goal remains open.

Before code: AppRoot currently activates only by owner, never scene activity;
FinancialMutationQueue drains on create/login/manual only, no network or timer.
Reuse this coordinator (no new agent/service framework). AppRoot sends typed
owner+foreground context. Background/inactive keeps encrypted rows/identity but
pauses new sends and cancels wake timers; logout remains a distinct auth block.
An in-flight response cannot acknowledge a cancelled generation or start another
write. Its uncertain sending lease remains until expiry, then same-key replay.

SDK NWPathMonitor feeds availability into the default coordinator, not custom
test clients. Only unavailable→available wakes sync; reachability is not proof
server is healthy. One cancellable foreground timer follows the FIFO head's
retryAt/lease expiry. Pending unsupported capability retries no faster than30s;
network unavailable/background/authblocked has no timer. Fresh capability and
owner/session checks remain mandatory. No force-reset of backoff/key/payload.
Use injected clock/sleep for deterministic tests, actual defaults in production.
Owner epoch plus unique drain/wake IDs suppress late callbacks after lifecycle
and account changes; do not use background pause to resume authblocked rows.

TDD: real encrypted SQLite + coordinator + external API boundary: background
create stays pending/no network, foreground/network wake sends original key,
due timer applies once, cancellation/logout/wrong-owner cannot wake old work,
paused in-flight late response cannot apply/drain next, resume keeps lease until
due. Run all native units + existing real API create/transfer regressions and
proxy restart/replay gates as appropriate; inspect actual lifecycle UI if added.
No UI styling/base change, public API/schema/auth/deployment dependency change.
Reconcile human+agent sync guide, native runtime docs, context/backlog/validation
and ledger. No Swagger/ERD/newADR needed; release withheld until owner acceptance.

Links: [plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[contract](../../superpowers/specs/2026-10-06-offline-financial-sync-contract.md),
[full goal](NATIVE-FULL-COMPLETION-20261014.md),
[validation](../VALIDATION_MATRIX.md), [guide](../../guides/offline-sync-api.md).

RED72468: all three startup/network/timer behavior tests failed against compiling
old-behavior delegates (not merely missing symbols). First coordinator gate33609
passed258units0skip. PauseRED18554 proves late held response acknowledges and
drains second row in background; cancelled/logout timer and authblocked wake
cases already passed. Add engine suspension generation and unique coordinator
drain token; preserve unexpired sending lease. Default composition starts gated,
SDK monitor + AppRoot owner/scene identity open it only when active+reachable.

Actual UI gate before adding test: real503 create, Home→app background, switch
wallet-scoped proxy to pass, observe no server rows over2seconds (covers initial
1–1.25s retry), activate app with no Retry tap, assert one exact server row and
one key. Existing restart/manual retry + commit/drop regressions retained.
Only disposable8081/18081 fixtures; no production configuration change.

Final evidence:
- Pause/runtime gate85017exit0,263/263PASS0skip (261units + real API native
  transfer lifecycle + ordinary create/edit); /tmp/mypocket-lifecycle-pause-green-20261014.xcresult.
- Actual lifecycle/fault UI13577exit0,3/3PASS0skip/no runtime warnings;
  /tmp/mypocket-lifecycle-ui-final-20261014.xcresult. Home/server pass/no writes
  during2s background/foreground auto-apply, same-key restart/menu retry, real
  committed-lost-response replay each leave one correct server row. Final proxy
  commitdrop statuses201,200,200,200,200;4drops,1hashedkey.
- Final units94868exit0,263/263PASS0skip/no runtime warnings;
  /tmp/mypocket-lifecycle-unit-final-20261014.xcresult. Includes early foreground
  no lease steal, wake at60s lease expiry; FIFO100s retry head blocks nextpending,
  early99s wake cannot send; unsupported30s cooldown; stale noncooperative timer
  cannot wake logged-out owner; foreground/network cannot resume authblocked.
- Guards/guardtests/project/diff pass. Applied screenshotD63C86B9 inspected under
  /tmp/mypocket-lifecycle-visual-20261014: target intent removed, old synthetic
  rejected drafts preserved. Feedback still overlays third row; no placement fix
  or visual/a11y acceptance claimed.

Normal8081 build restored (Info.plist checked), ownedproxy63813 stopped; local
API health200. No app uninstall/database reset/production/device changes.
Human+agent guide, architecture runtime, contract checkpoint, transaction screen/
registry, context/backlog/validation/ledger reconciled. No API/schema/env/package
contract change, so Swagger/ERD/newADR not changed. Release remains unpublished
pending owner acceptance. Actual radio/offline switch and physical-device UAT,
rejectededit/discard/transferfault UI, goalqueue/cache/report/auth/allnative gates
remain; this is not full Task5/6/7 or full goal completion.
