# Native simulator auth fixture

## Scope

The local API UI suite must authenticate through the disposable backend fixture when launched with `-UITestAuthenticatedAPI`. A simulator run currently receives HTTP 200 from `/auth/mobile/google/exchange` but remains on the login screen, so the failure is inside response decoding or session persistence rather than route rendering.

## Intended behavior

- The fixture exchange response is decoded using the same API envelope as production.
- The resulting access and refresh tokens are persisted for the test process.
- A DEBUG-only diagnostic identifies the failing stage without logging token values.
- No production behavior or secret is changed.

## Verification

- Run the isolated API-backed assistant UI test on the booted iPhone 16 Simulator.
- Run the complete `MyPocketIOSUITests` suite with `LOCAL_API_UI`.
- Run native unit tests and release build after the fix.

## Resolution — 2026-10-06

The failure was `errSecMissingEntitlement (-34018)` from the unsigned
simulator build, not an OAuth or backend error. `AuthViewModel` now injects an
in-memory session store only for the DEBUG `-UITestAuthenticatedAPI` fixture;
normal and Release builds still construct `KeychainSessionStore`. The full
iPhone 16 `LOCAL_API_UI` suite passed 15/15 after the fix.
