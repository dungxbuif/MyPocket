---
artifact_type: acceptance_harness
id: FEEDBACK-20261019-UAT-HARNESS
status: active
updated: 2026-10-19
---

# Synthetic acceptance harness for remaining UAT feedback

This harness keeps the 13 records classified as `needs_uat` tied to deterministic
local evidence without pretending that provider, deployment, owner or physical
device acceptance has happened. The machine-readable contract is
[FEEDBACK-20261019-UAT-HARNESS.json](FEEDBACK-20261019-UAT-HARNESS.json).

## How to run

Backend route/integration evidence (disposable PostgreSQL/Redis only):

```bash
cd backend
TEST_DATABASE_URL='postgres://mypocket_test:local-fixture-only@127.0.0.1:55436/mypocket_savings_budget?sslmode=disable' \
TEST_REDIS_URL='redis://127.0.0.1:58651/0' \
go test ./internal/controller/http ./internal/usecase ./internal/infrastructure/repository -count=1
```

Native deterministic evidence (simulator or local XCTest destination):

```bash
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -only-testing:MyPocketIOSTests test -enableCodeCoverage NO \
  -parallel-testing-enabled NO CODE_SIGN_IDENTITY=- CODE_SIGNING_ALLOWED=YES
```

The Go manifest test checks that every case has a source file, named local test
symbol and at least one explicit external-only gate. It also rejects production
URLs or credentials in the manifest. It is a traceability guard, not a claim
that the external gate passed.

The manifest guard passed locally with:

```text
go test ./internal/controller/http -run TestFeedbackAcceptanceHarnessManifestIsSyntheticAndTraceable -count=1
ok
```

The broader HTTP package requires the disposable Redis/PostgreSQL services
above. If Redis is not running, the existing real-JWT/Redis boundary test is
expected to fail; that is an environment failure, not evidence of a production
acceptance result.

## Exact external-only gates

The remaining gates are intentionally separated:

- configured AI/OCR/APNs provider calls and response quality;
- production deployment/version/migration readback;
- owner-authenticated populated-data acceptance;
- physical iPhone Face ID, Keychain, notification permission and offline
  foreground lifecycle;
- visual approval of compact/accessibility layouts.

No test in this harness mutates production data, resolves a production feedback
record, or treats a synthetic fixture as a real receipt/provider result.
