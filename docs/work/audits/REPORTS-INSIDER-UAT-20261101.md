# Reports / Money Insider — local verification record

Implemented paths include the query-backed report model, chart data, grounded
insight candidates, snapshot revision checks, async deep analysis lifecycle and
bounded retention worker. The deterministic insight and checked-money tests
are now part of the entity suite.

Verified locally:

- `backend/go test ./...` — PASS.
- Native full `MyPocketIOSTests` — PASS, 442 tests.
- Base/project guards and diff check — PASS.

Populated local API UAT passed for Money Insider rendering, the income/expense
comparison ring, persisted daily income/expense trend values, and exclusion
of a transaction marked `included_in_reports=false`. The native client sends
the supported weekly alias contract and uses a compact daily trend by default
for the readable report screen.

Not yet owner-accepted: production latency/provider-call measurements and
visual sign-off. Ordinary report reads do not claim an AI call.
