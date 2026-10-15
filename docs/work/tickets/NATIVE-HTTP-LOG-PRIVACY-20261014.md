# Native HTTP diagnostic privacy

Status: in_review. Correctness/security gap within approved full-native goal
and existing offline contract's no-raw-financial-payload logging requirement.

Before implementation: APIClient.executeEnvelope DEBUG logs request.path, including
query values, and converts every non2xx response body to a raw printed string.
Server ProblemDetails can echo private notes/prompts/credentials. Multipart has
no corresponding print. No change to HTTP semantics or UI error presentation.

Bounded correction: retain method/status/generated requestID diagnostics, omit
raw path/query/body. Do not introduce a telemetry dependency, persisted log store,
remote sink, token logging, or suppress typed errors returned to the caller.
Regression test exercises actual APIClient with synthetic private canaries in
URL query and non2xx response; capture process stdout in sequential XCTest and
prove no canary leaks while correlation/status remains and typed error survives.
Also exercise successful JSON request query logging. Tests own capture helper;
no test hooks in production. Full native units and normal wallet/transfer API UI
regressions; base guards/project/diff. No production/data/signing/reset actions.

Reconcile internal MOBILE_IOS architecture and human+agent offline guide, context,
backlog/validation/full-goal ledger. No Swagger/schema/ERD change: wire unchanged.
No release publication until owner acceptance. Goalqueue/Feedback blocked paths
remain stopped; this is an independent privacy defect, not another attempt.

Links: [goal](NATIVE-FULL-COMPLETION-20261014.md),
[contract](../../superpowers/specs/2026-10-06-offline-financial-sync-contract.md),
[validation](../VALIDATION_MATRIX.md), [guide](../../guides/offline-sync-api.md).

RED39064exit65,2/2 failed as expected,0skip; actual stdout exposes query canaries
on both200 and422 and error-body canary on422. Bundle:
`/tmp/mypocket-http-privacy-red-20261014.xcresult`. Keep response decoding and typed
ProblemDetails intact; fix only DEBUG metadata interpolation/body print.

Horizontal investigation: AssistantViews has4 DEBUG prints interpolating reflected
errors (capabilities/overview), Budgets has1 on save, AuthViewModel's local-only
fixture has1. These may expose typed ProblemDetails too; APIClient redaction does
not sanitize independently printed errors. Separate view-level repro/gate needed
before claiming whole-app log privacy. No unverified patch to those flows here.
Final gate92498exit0:267/267PASS0skip,265 native units plus actual local API-backed
ordinary create/update and transfer create/edit/delete UI. Bundle:
`/tmp/mypocket-http-privacy-green-20261014.xcresult`; no runtime warnings.
Both stdout regressions pass:200 query and422 body/query omitted; method/status/
generated requestID remain, typed422 detail unchanged for caller. Base guard,
guard tests, project verification, git diff --check pass; APIhealth200.
Internal architecture and public human/agent Markdown guide reconciled. No API/
schema/ERD/config/dependency/visual change, so no Swagger/ADR/UI spec update.
No production/install/release or whole-app privacy/goal completion claim.
