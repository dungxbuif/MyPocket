# Offline ordinary category validation

Status: in_review. Local regression verified; full native goal remains open.
Scope: existing approved offline contract; no new feature or backend change.

Before implementation: transaction_handler.go validateInput normalizes category
and rejects nil; native OutboxPayload.validate instead uses optional map default
true. Thus an invalid ordinary draft can become durable and later rejected.
Require a nonempty trimmed category identifier for income/expense enqueue and
changed replacement. Transfer payload has no ordinary category input and must
remain valid. Do not delete/rewrite existing encrypted drafts or change replay
keys/receipts. Validation belongs at mutation creation, not decrypt/load.

TDD real SQLite: missing category for each type must fail before insertion;
invalid replacement must leave original rejected and no child/cancellation.
Also test empty/whitespace/NUL identifier and valid categorized income/expense,
existing transfer regression/full native units. Run base guards/project/diff.
Reconcile public human+agent guide, transaction spec, context/backlog/validation
and offline ledger. API/schema/security/runtime unchanged; no Swagger/ERD/newADR.
No release publication before owner acceptance; no prod/install in this slice.

Links: [full goal](NATIVE-FULL-COMPLETION-20261014.md),
[offline plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[validation](../VALIDATION_MATRIX.md), [guide](../../guides/offline-sync-api.md).

RED67864 exit65,2/2failed with expected durable-invalid-intent and invalid-
replacement assertions (/tmp/mypocket-category-validation-red-20261014.xcresult).
Cause confirmed at ordinary optional-map fallback true; minimal fix false.
Decrypt/load path does not revalidate payload, so old encrypted drafts remain
readable for explicit user correction/discard. Transfer path unchanged.
Full units + existing API-backed native transfer lifecycle currently running.

Initial GREEN50398exit0,255/255PASS0skip (254units + transferUI). Existing
transferUI checks pair/amount/update and disappearance, but not wallet balances,
report flags or server deletion. Before strengthening proof: explicitly select
source/destination in native controls, read actual wallet balances100000→87500/
112500→77500/122500→100000 each, assert paired report flags false and final GET
has no pair. Add fixture-only balance decoding; no product code changes for this
verification extension. Keep synthetic exact-wallet cleanup, no prod writes.

Compatibility review found the shared validator is also used before replay in
FinanceAPI.applyOutbox. Tightening that replay validator would strand a legacy
uncategorized pending draft at a sending lease (invalidPayload is treated as
storage error). Keep the old frozen-payload replay validation; add category
requirement only at new enqueue/replacement via validateNewIntent. Existing
intent keeps its exact key/body, server returns its original receipt or typed
business rejection; no silent payload repair. Add actual APIClient/adapter RED
with a legacy nil-category leased intent and HTTP400 boundary: must expose
server rejection, not local invalidPayload. This is required compatibility,
not an exception for creating new uncategorized drafts.

Compatibility RED69099exit65: actual adapter fails with local invalidPayload
before HTTP. Fix: validateNewIntent used only at enqueue/replaceRejected;
existing replay validate unchanged. Final40360exit0,256/256PASS0skip/no runtime
warnings (255units + strengthened native transferUI),
/tmp/mypocket-category-compat-final-20261014.xcresult. Both creation/replacement
and legacy HTTP rejection regressions pass. Initial transfer-state5833 also
passed255/255 before compatibility test was added. Native transferUI now proves
explicit wallet direction, balances87500/112500 then77500/122500, report flags
false, server pair deletion and both opening balances restored100000. Report
aggregate UI is not inferred from flags; offline transfer fault recovery remains
a separate gate. Local API8081 health200; no fixture/prod reset or app uninstall.

Base guard/guardtests/project/diff all pass. Human+agent guide and transaction
spec reconciled, context/backlog/validation/ledger updated. No public API/DB/
security/runtime change, so Swagger/ERD/newADR unchanged. Release publication
still withheld for owner acceptance. Foreground/reachability/timer/background
pause wiring NOT implemented in this prerequisite slice; full goal remains open.
