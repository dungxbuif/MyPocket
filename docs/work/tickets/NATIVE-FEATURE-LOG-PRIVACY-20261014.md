# Shared native feature error diagnostics

Status: in_review. Existing full-native/privacy contract correction, independent
of blocked goalqueue/Feedback paths. No new auth/security boundary or API behavior.

Before code: Assistant capabilities/overview have4 reflected-error DEBUG prints,
Budgets save1, local auth fixture1 plus reflected fixture error message. HTTP
redaction does not protect these logs. Arbitrary associated ProblemDetails,
transport/decoding descriptions and NSError.userInfo may contain private data.

Bounded approach: APIError owns a shared diagnostic description for Error, allowlist
only known failure class and numeric HTTP status, unknown Error returns unexpected.
Never reflect/stringify arbitrary errors or problem fields. Feature DEBUG logs
consume this formatter; typed userMessage stays unchanged. Local-only simulator
fixture debug label may use the same safe summary. No new telemetry storage/sink,
dependencies, routes, schemas, keys or UI/base styling.

TDD: actual formatter stdout with canary in every problem field, transport/decode
associated strings and unknown NSError domain/userInfo. Verify no canary; status
and failure category remain; typed userMessage not redacted. Initial formatter
scaffold preserves old reflection so tests can compile and demonstrate existing
leak before safe implementation. Not wired into features until GREEN change.
Full native units plus Assistant navigation/composer/draft API UI smoke; guards/
project/diff. No DB reset, production mutation, signed install or release.

Docs: reconcile native architecture, public human+agent finance assistant guide,
context/backlog/validation/full goal. No Swagger/ERD/ADR because wire/data/config
unchanged. Broad log/physical-device privacy acceptance is not implied by units.

Links: [HTTP prerequisite](NATIVE-HTTP-LOG-PRIVACY-20261014.md),
[goal](NATIVE-FULL-COMPLETION-20261014.md), [validation](../VALIDATION_MATRIX.md),
[guide](../../guides/finance-assistant-ios.md).

RED93397exit65,3/3 failed0skip, valid runtime leak reproduction at the shared
formatter using old reflection. Bundle:
`/tmp/mypocket-feature-privacy-red-20261014.xcresult`. Includes missing-status
problem, all reflected problem fields, transport/decoding associated strings and
unknown NSError. No compiler/test infrastructure failure. One safe implementation
attempt: allowlisted class/numericstatus only, six feature prints and fixture
debug label switched to formatter. Typed userMessage unchanged.

GREEN66267exit0:272/272PASS0skip,268 native units +4UI (actual local API advisor
overview/details, sole primary tab/draft retention, compact keyboard composer,
suggested questions). Bundle `/tmp/mypocket-feature-privacy-green-20261014.xcresult`.
No runtime warnings. Shared formatter stdout proves no private problem/transport/
decoding/NSError fields, preserves categories/status. Feature call sites inspected
and all current native print/debugPrint/NSLog matches use bounded metadata or
this formatter. This is not per-screen injected-error UI proof or live LLM proof.
Base guard/guardtests/project/diffpass,APIhealth200.

Inspected97F7C806 keyboard capture: composer above keyboard, disabled send and
suggestions still accessible; tab bar consumes vertical space. Existing plan
requires navigation remain usable, and tab-switch draft test uses it while
keyboard open. Do not hide it speculatively or call this a proven layout defect;
retain as potential UX review, not an implementation change/acceptance claim.

Docs nativearchitecture + publichuman/agent guides reconciled; no schema/API/
configuration or visual-contract change, no Swagger/ERD/ADR/UI spec update.
Release withheld, no prod/iPhone install/reset, fullgoal remains in_progress.
