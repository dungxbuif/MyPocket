# Native completion reconciliation

Status: in_progress. Owner: “Đối soát và làm hết đi”. This is the current
cross-ticket index; older checkpoint paragraphs are historical evidence, not
current completion claims. [Full goal](NATIVE-FULL-COMPLETION-20261014.md).

## Before changes

Reconcile internal backlog/tickets/validation with code and actual artifacts;
correct public/agent guides only where behavior changed or a claim is stale.
Do not mark feedback resolved, publish release notes, delete production data or
install a queue-enabled client against an incompatible production backend.
Continue approved tests/fixes; preserve API/owner/auth/financial semantics unless
the applicable reviewed design explicitly changes them. No archived web work.

## Authoritative matrix

Fresh production queue contains48open records, not34/42. [Latest six assessment](NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md)
adds actual portfolio edit/delete and shared Back/spacing/fixed Feedback/Account
Add and AI jar-tool requirements. No production feedback status changed.

| Area | Local evidence/current implementation | Production/device gap | Next gate |
| --- | --- | --- | --- |
| Shared bases, compact rows, photos-first | UI plan Tasks2/4 locally verified; numeric/chat287/287PASS | full-screen/owner UAT incomplete | populated registry/visual sweep, independent review |
| Dedicated transaction note editor | TransactionsView note sheet; note RED→GREEN2/2 and API note readback in validation | owner UAT only, not missing implementation | correct stale feedback ticket, retain API/current CRUD regression |
| Remembered-device Face ID | protected device credential flow and local API/UI proof | auth backend already deployed12572c87; physical biometrics UAT pending | owner device acceptance, not auth reimplementation |
| API-key copy/revoke | actual local account key create/copy/revoke/reopen proof | updated account/full-owner policy not fully rolled out | full-owner rollout + owner UAT |
| Full-owner API keys | common RequireAuth, no scope gates, realPG/JWT/Redis security tests | old auth-only API image still live | scoped immutable backend release, isolation/security gates |
| Offline receipts/outbox/engine/foreground | encrypted store, acknowledgements, retry/lease/restart/fault E2E implemented locally | fresh production capability404; public OpenAPI lacks route | deploy matching backend BEFORE another new native install |
| Financial overview/report cache | encrypted ledger snapshot exists; overview/report injection/query snapshot absent | stale report error reproduced locally, not fixed | reviewed typed coherent read scope/cache implementation + realUI regression |
| Savings rejected draft |503same-key/pair/balance/report proofPASS |422menu hit-testing still unresolved | root hit-testing probe/design, no placement guesses |
| Category CRUD test | reviewed helper/API-first/last-drag recheck now280/280PASS (279unit+categoryUI), same names/singleton/archive/restore constraints | broader12business sweep live, delete API collection proof strengthened | final full-business gate, not product CRUD rewrite |
| AI numeric/privacy | exact amounts, no invented zero, exact count/JSON, masked fallback;287/287PASS | updated Release pending signing, phone unavailable | signed compatible-backend install + owner UAT |
| AI cancellation | approved conversation→run recheck, realPG4/4 + Go racePASS | patch not deployed; initial Stop unavailable until synchronousPOST returnsID | backend rollout, initial transport/action design |
| AI confirmed actions/help | wallet proposal parser and history local proof; static handbook exists | durable prepared-action replay, full goal/credit review summary, grounded help/RAG incomplete | reviewed action/help designs and real-provider eval |
| Idle chat expiry |30minapproved; written lifecycle spec9ab9248d | no expiry/deletion implemented or enabled | spec review→plan→TDD lifecycle/generation/replay/worker/native |
| Feedback placement/completeness | bubble/no reserved lane and owner resolve/reopen support local | production OpenAPI lacks owner-resolution route/migration32; protected-region overlap and new fixed-anchor request open | scoped backend rollout plus reviewed fixed-anchor/UI proof |
| Notifications | inbox/read API and native screen exist | no APNs provider/delivery proof; current Personal Team lacks entitlement | supported signing/delivery decision, not inbox-only completion |
| Release | earlier production-connected native candidate installed; no financial reset | latest build not installed; no final whole-branch review/UAT | compatible backend, strict signing, install/launch, owner acceptance |

Fresh production read-only evidence: `/api/v1/sync/capabilities`404; downloaded
public OpenAPI lists device/auth and advisor/feedback/key routes but no sync
capability. Known vm100 SSH identity works without agent; Docker reports API
auth-12572c87 healthy and workerrelease-73c4c37. vm100 ControlAvailable=false;
cannot update Swarm there or promote it. Manager pi rejected the explicit known
homelab identity; normal pi uses a1Password-backed public identity requiring
agent authorization. Do not chmod a public key or alter SSH/Keychain policy.

## Execution order

1. Reconcile stale ticket claims and finish bounded category proof.
2. Establish compatible scoped backend candidate/access; backup/test/rollout
   approved additive changes without importing unrelated dirty code.
3. Financial report/cache correctness and savings422 recovery; preserve server
   source of truth and visible failed/stale scopes.
4. AI idle lifecycle, cancellation/action replay/help; no unlimited raw history
   or unconfirmed financial mutation.
5. All-screen/feedback/physical tests, independent review, compatible install.
   Owner acceptance precedes feedback resolution/release publication.

Links: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[UI plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[offline plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md).
