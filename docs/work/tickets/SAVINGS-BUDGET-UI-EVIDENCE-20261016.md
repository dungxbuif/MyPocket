---
artifact_type: ticket
id: SAVINGS-BUDGET-UI-EVIDENCE-20261016
status: verified
owner: ai
updated: 2026-10-16
---

# Savings and recurring budget simulator evidence

## Scope and decision before implementation

Add targeted opt-in XCTest UI coverage in `mobile-ios/MyPocketIOSUITests/LaunchTests.swift` for the approved [design](../../superpowers/specs/2026-10-16-savings-budget-design.md) and [plan](../../superpowers/plans/2026-10-16-savings-budget-implementation.md). Verify Reports → savings, distinct goal and account surplus sections, real drilldown rows, native recurring-budget creation, next-boundary edit, stop, forecast and matching expense rows. Read back actual API state after UI mutations. Use deterministic synthetic integer amounts and uniquely named resources.

Small task exemption: yes. Reason: verification infrastructure only; approved product implementation is owned separately. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. No production writes, physical iPhone navigation, app resets, or product-source edits. Existing debug-only `-UITestAuthenticatedAPI` and `LOCAL_API_UI` behavior is reused.

Known constraint: existing fixture OAuth hardcodes `fixture@example.com` and `fixture-ios-user`; the API has no configurable fixture-owner suffix. Use only disposable local PostgreSQL `127.0.0.1:55436/mypocket_savings_budget`, local API port `18086`, and UUID-suffixed synthetic rows. Confirm account contents before setup. Exact synthetic wallet/transaction cleanup is required; series stop preserves generated historical periods by contract.

## Verification plan

Build and execute only the two new UI scenarios on the booted MyPocket iPhone 16 simulator with signing enabled, `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`, API base `http://127.0.0.1:18086/api/v1`, and a dedicated derived-data/result path. Savings: compare UI sections and drilldown IDs to owner-scoped API totals. Budget: create monthly/carry-enabled series through UI, assert base/effective/spent/forecast against actual API, edit next config through UI, stop through UI and verify version/active/generated-period preservation. Keep any failure artifacts and distinguish harness failures from product failures.

## Documentation reconciliation

This note records test scope, setup and evidence. Parent owns [backlog](../BACKLOG.md), [context](../../CONTEXT.md), [validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md), requirements, architecture/API/ERD/ADR and public human/agent references for the product feature. This verification addition introduces no public contract or deployment change. Physical-device and owner UAT remain separate.

## Results

First run: `/tmp/mypocket-savings-budget-ui-first-20261016.xcresult` failed 0/2, zero skips/runtime warnings. Both cases failed during fixture login before app launch with `NSURLErrorDomain -1004` at `127.0.0.1:8081`. Xcode did not forward either host `MYPOCKET_TEST_API_BASE_URL` or `SIMCTL_CHILD_MYPOCKET_TEST_API_BASE_URL` into the runner. No product behavior was reached. Correction: build-for-testing and apply `MYPOCKET_TEST_API_BASE_URL` explicitly to the generated local `.xctestrun` runner environment, preserving the existing configurable helper and product sources.

Second run: `/tmp/mypocket-savings-budget-ui-second-20261016.xcresult` repeated the same pre-launch fixture error. Inspection showed the temporary patch had set the unit-test environment rather than UI-runner environment. Corrected the UI-specific environment and verified with `/usr/libexec/PlistBuddy -c 'Print :MyPocketIOSUITests:EnvironmentVariables:MYPOCKET_TEST_API_BASE_URL'`; exact result `http://127.0.0.1:18086/api/v1`. This was a setup correction, with no product-source change.

Third run: `/tmp/mypocket-savings-budget-ui-third-20261016.xcresult` reached native series creation and report, then failed the shared reveal helper while the simulator inherited landscape and privacy masking from previous work. Interrupted after identifying those states; no pass claimed. Exact interrupted wallet `1fc76a2f-438e-457a-8d65-2b8f693a494c` and its three synthetic transactions were removed through local authenticated API calls. Its series remains stopped with `scope_deleted`, consistent with dependent-wallet deletion. Test methods now select portrait and use Foundation's existing argument domain `-mypocket.privacy.masked NO`, without persistent preference writes or app resets.

Fourth run: `/tmp/mypocket-savings-budget-ui-fourth-20261016.xcresult` failed before UI with the same `8081` connection error because build-for-testing regenerated the temporary `.xctestrun`. The final invocation verified the UI target's exact environment key before running. No further temporary environment edits or reruns were made afterward.

## Authoritative passing evidence

`/tmp/mypocket-savings-budget-ui-fifth-20261016.xcresult`: **2/2 passed, 0 failed, 0 skipped, 0 runtime warnings**, 103.5 seconds. Device: signed (normal simulator ad-hoc signature, no `CODE_SIGNING_ALLOWED=NO`) iPhone 16 / iOS 18.5, `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. Build uses iOS Simulator 27.0 SDK. Log: `/tmp/mypocket-savings-budget-ui-fifth-20261016.log`. This is focused local API/simulator verification, not physical-device or production UAT.

Verified by the two test methods:

- Savings Reports navigation; current month; goal opening 100,000 and closing 5,100,000; paired incoming transfer 5,000,000 separated from ordinary deposits; income 20,000,000, expense 12,000,000, surplus 8,000,000. Goal drilldown contains one transfer row; account surplus drilldown contains exactly the two ordinary API transaction IDs and excludes the transfer note. Populated UI displays both distinct sections and matching amounts.
- Native monthly recurring creation with carryover enabled and exact wallet scope; first generated budget has base/effective limit 1,000,000, carry 0, spent 300,000 and remaining 700,000. API daily sum and two-row drilldown sum both equal spent; the 90,000 report-excluded expense is absent. Forecast is `estimated`, `ceil(300000 * 31 / 16) = 581250`, and the UI displays that server value.
- Native next-period edit produces pending name/limit 2,000,000 effective 2026-11-01 while current config and period remain 1,000,000. Version increments once.
- Actual native stop confirmation sends `POST /api/v1/budget-series/7993cfac-89f3-48bd-8e68-bc02adcb45fd/stop`, returning 200 at 21:20:57 +0700 before teardown. Immediate API readback asserts inactive/version 3, preserves the existing budget and matching two transaction IDs. The suspected alert-binding no-op was not reproduced.

Final owner-scoped API readback: `/wallets` returns `[]`; `/transactions` returns `[]`. Two stopped synthetic series histories remain in this disposable DB (one interrupted, one successful), as series deletion is intentionally not a product route. The inherited transfer cleanup helper's no-key DELETE was rejected with 400; subsequent exact-wallet deletion cascaded the pair, and final empty transaction readback confirms cleanup. No financial user data or unrelated fixture owner was altered.

### Exact commands

From repository root, build:

```sh
rtk proxy xcodebuild build-for-testing -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -derivedDataPath /tmp/mypocket-savings-budget-ui-20261016 -only-testing:MyPocketIOSUITests -parallel-testing-enabled NO 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI' MP_API_BASE_URL=http://127.0.0.1:18086/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet
```

The generated local `MyPocketIOS_iphonesimulator27.0-arm64.xctestrun` UI target `EnvironmentVariables` contained `MYPOCKET_TEST_API_BASE_URL=http://127.0.0.1:18086/api/v1`, independently verified with PlistBuddy before the passing command:

```sh
rtk proxy xcodebuild test-without-building -xctestrun /tmp/mypocket-savings-budget-ui-20261016/Build/Products/MyPocketIOS_iphonesimulator27.0-arm64.xctestrun -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -resultBundlePath /tmp/mypocket-savings-budget-ui-fifth-20261016.xcresult -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedSavingsReportSeparatesGoalTransfersAndAccountSurplus -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedRecurringBudgetCreateForecastEditNextAndStop -parallel-testing-enabled NO -collect-test-diagnostics never -quiet
rtk proxy xcrun xcresulttool get test-results summary --path /tmp/mypocket-savings-budget-ui-fifth-20261016.xcresult --format json
rtk proxy git diff --check -- mobile-ios/MyPocketIOSUITests/LaunchTests.swift docs/work/tickets/SAVINGS-BUDGET-UI-EVIDENCE-20261016.md
```

The first two return success with the counts above; diff check passes. Future ordinary runs can use the existing helper's default loopback 8081 API fixture and corresponding build URL; host-shell environment forwarding alone is insufficient. No new app endpoint override was introduced.

### Screenshots and visual review

Eight retained screenshots exported to `/tmp/mypocket-savings-budget-ui-evidence-20261016`; `manifest.json` maps test/step names to image filenames. Inspected goal, surplus, forecast, pending-config and stopped-period captures. Goal metrics and account surplus are distinct; the current budget remains visible after stop; next-period pending config and inactive state are visible. This does not claim all-screen visual acceptance, dynamic-type coverage or all feedback bubble interactions.

- Goal section: `63CAB352-F030-4723-937A-FC98100E3F63.png`.
- Account surplus: `23D20CF9-BAFD-446A-9499-434527AB35BC.png`.
- Forecast: `C2477785-C13E-422A-9A83-D4253070AA97.png`.
- Next-period edit: `F398631E-6A6A-4DB4-8BD7-F2C29B090C89.png`.
- Stop/preserved generated period: `F0F98942-B777-442C-9ED3-E57BC371879C.png`.

Docs review complete for this test-only addition. Public docs/API/ERD/security/runtime references require no changes from these tests; parent owns the feature's broader reconciliation. Owner/physical-device UAT remains open. No commits or staging performed.
