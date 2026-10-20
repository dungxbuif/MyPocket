# Disposable `LOCAL_API_UI` harness — 2026-10-20

## Problem

The native transaction/report visual gate is opt-in and expects a local API at
`http://127.0.0.1:8081/api/v1`. Recent focused simulator runs skipped the
populated readback cases because that listener was not running. Starting the
normal development API on port 8080 is not sufficient: the UI build must use a
separate database and Redis instance so the gate cannot touch developer or
production data.

## Decision

`mobile-ios/scripts/local-api-ui-fixture.sh` is the single disposable runner.
It creates a uniquely named local PostgreSQL database, applies all checked-in
migrations, starts a loopback-only Redis process on a free port, then starts
the API on port 8081 with `OAUTH_FIXTURE_MODE=true` and the fixture seed user.
The generated state is kept under `/tmp/mypocket-local-api-ui` and `stop`
validates the generated database name before dropping it. No production URL,
port 8080, shared Redis logical database, or committed credential is used.

## Verification

1. `start` must wait for `GET /health` and the fixture OAuth exchange to pass.
2. `status` reports the API/Redis PIDs, ports, database and base URL without
   printing secret values.
3. `stop` must terminate child processes and drop only the generated database.
4. Run focused `LOCAL_API_UI` tests with
   `MYPOCKET_TEST_API_BASE_URL=http://127.0.0.1:8081/api/v1` and
   `MP_API_BASE_URL` set to the same URL, then stop the fixture.

The savings report fixture treats a goal wallet created inside the selected
month as opening balance `0` plus `initial_balance_added`; this is intentional
and prevents a newly-created wallet's seed from leaking into the prior-period
opening balance. Native assertions must use that contract.

The harness is local-only and does not alter deployment configuration.
