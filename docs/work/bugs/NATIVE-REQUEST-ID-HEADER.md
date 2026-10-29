---
artifact_type: bug
id: NATIVE-REQUEST-ID-HEADER
status: regression_verified
owner: shared
---

# Native request correlation header

## Finding

The Go API returns and logs a stable `request_id` for diagnostics, and the
native port plan requires every request to send `X-Request-ID`. The native
`APIClient` did not attach that header, making simulator/provider failures
harder to correlate. A JWT refresh retry must reuse the same ID so one logical
request has one trace.

## Intended behavior

- Generate one UUID per `send` or `sendMultipart` operation.
- Attach it as `X-Request-ID` to session, API-key and unauthenticated requests.
- Reuse the same ID for the single access-token refresh retry and multipart
  retry.
- Do not log or persist credentials in the correlation value.

## Verification plan

- Capture normal and multipart requests in APIClient tests, including refresh
  retry, and assert the header is present and stable.
- Run native unit tests, release build/guards and the signed iPhone 16 UI suite.

## Verification result — 2026-10-06

- APIClient tests assert `X-Request-ID` is present on normal requests and is
  reused across a multipart 401 refresh retry.
- Native XCTest: 45/45 passed; Release build with warnings-as-errors,
  base-component guard, project verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 200.267s.
