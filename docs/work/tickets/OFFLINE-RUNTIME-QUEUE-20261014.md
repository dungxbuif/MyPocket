# Native financial queue runtime

Status: in_progress; approved offline Task5/6, full native goal remains open.

Before code: MainActor coordinator publishes only active owner queue, errors and
confirmed-read revision. Durable enqueue precedes all network sends and repeats
reuse form UUID/frozen payload. submit returns applied or queued, never invents
a server transaction. Existing create forms will use coordinator; edit/delete/
credit/adjustment remain online. New creates must not fall back to legacy POST
after an intent is enqueued. Matching backend rollout required before installation:
old/unsupported capability keeps draft queued with explicit unsupported notice.

Coordinator verifies current SessionStore owner before enqueue, before publish and
after await; switching account clears published rows and blocks old-owner queue.
Explicit same-owner activation resumes blocked rows. Stop cancels drain and hides
rows, not deletes them. UI is a distinct pending section made from existing
MPSurfaceCard/MPText/MPStatusBadge/MPButton, never merged into confirmed totals.
Retry reuses key/payload and is allowed only retryWait; conflicts/rejections need
explicit edit/new intent or discard. Discard changes local status only, refuses
active sending/applied; uncertain retry discard warns server may already contain
the transaction and must refetch, never sends a compensating financial write.

TDD core first: real SQLite persistence-before-send, stable key/idempotent submit,
unsupported queue not success, owner switching/no cross-owner publish, auth stop,
retry-only controls and safe discard. Then wire Auth/root/form/ledger, read revision
subscriptions, foreground/network triggers/timer and background cancellation.
Run native units+actual local API/UI/normal/stale/pending flows, guards/project/diff;
publichuman+agent/internaldesign/currentdocs updates. No production mutation,
install/release/wholegoal completion until coordinated gates pass.

Composition ruling: default Auth creates matching-environment queue; custom API
clients need explicit injection, no accidental production namespace. AppRoot
activates/stops it with authenticated owner and passes optional environment value
to shared create editors. Ordinary create and transfer create use one State UUID
per editor; edits keep existing online path. Queue section observes coordinator
above ledger, distinct from confirmed transactions/totals. Revision subscriptions
reload ledger/overview only after receipt+CAS. Initial UI offers sync/status; safe
retry/discard/conflict editors and foreground/network scheduler remain next and
must not be represented as complete. No install before matching backend rollout.

Concurrency review before fix: coordinator has one global isSyncing guard. A
cancelled old-account HTTP operation may ignore cancellation; while awaiting it,
new account submit is durably queued but its drain suppressed. Add real-store
paused-boundary regression before replacing guard with per-activation drain epoch.
Late old drain must not clear new owner loading/error/publication state. Engine
generation/lease safety remains required; this is coordinator liveness, not a
reason to weaken ownership or duplicate an uncertain old request.

## Local checkpoint

Coordinator epoch-bound owner/publish, persist-before-send/noonlinefallback and
confirmedrevision implemented. DefaultAuth/root→sharedcreateforms StateUUID;
edits online unchanged. Shared distinctqueue section; ledger/overview subscribe
to receipt+CAS revision. Unsupportedcapability holdsdraft; no background/network
timer/peritemrecovery yet. GenericSync respects persistedretrydate.
RED3560→20848GREEN242units; composition69046GREEN242. Actualcreate/editUI45495
exit0,1/1PASS0skip,/tmp/mypocket-queue-create-ui-20261014.xcresult:25,000/category
persisted+editednote APIreadback. Readonly testappSQLite1row/sealed563bytes,
container449392A1 after testinstallation containerchange; no decryption/keyexport.
Screenshot4685B100detail inspected under /tmp/mypocket-queue-create-visual-20261014,
sharedbuttons/navbar visible, not pendingUIproof. APItestwalletcleanup, no prod/
iPhone mutation. ConcurrencyRED41241→peractivation drainEpoch; final23701exit0,
243/243units0skip,/tmp/mypocket-queue-final-20261014.xcresult. Oldownerblocked and
newownerapplied/revisiononlynew proven with realstore/pausedexternalboundary.
Guards/project/diffpass,XcodeGennewfiles. No backendwire/schema/packagechange.
Setupfailure keeps online-only path, never fallbackafterpersistedintent. Next
safe retry/discard/conflictedit, pendingUI/restart/replay/networkE2E, foreground/
network/background lifecycle/auth/remaincache/allnativegaps. Matchingbackend
rollout required; no prod/install/release/fullgoalclaim.
