---
artifact_type: bug
id: FEEDBACK-b017284a-940c-4533-8950-7c7995f956ea
status: in_progress
updated: 2026-10-26
---

# Portfolio item edit/delete and unit price

Source: unresolved feedback `b017284a-940c-4533-8950-7c7995f956ea`.
Scope: portfolio backend/native files, their tests, and portfolio documentation.
No production mutation, deployment, or feedback status change.

## Reproduction and current contract

- Open the native portfolio and inspect an asset card. Create, trade, price,
  and unused-asset delete actions exist, but there is no asset metadata edit
  action or native PATCH adapter. The public iOS guide already claims asset
  symbol/name editing; this is a real native gap.
- A trade row exposes replace/void actions and the native correction editor
  calls preview then confirmed commit. The backend stores immutable revisions
  and replays positions. The production rollout of correction routes is a
  separate gate; this local task will not deploy them.
- A buy uses integer VND unit price multiplied by fixed-point quantity, plus
  fee. The native trade editor labels the input `Giá đơn vị`; no total/quantity
  conversion appears in that path. A focused fractional-quantity assertion
  will guard against reversing this meaning.
- Deleting a portfolio asset is available only before any trade. The backend
  rejects deletion of traded assets, including concurrent trade creation.

## Root cause and decision

The backend metadata PATCH was added after the native asset editor was built,
but the native client and card were not connected to it. Reuse the existing
asset editor scaffold and validator for renaming symbol/name. Keep asset ID,
trade rows, quantity, and position unchanged. A failed update keeps the editor
and draft visible; success reloads portfolio data. Do not add direct trade
row mutation or broaden deletion: correction revisions already define that
accounting behavior.

Small task exemption: yes.
Reason: wire an already documented owner-scoped API route into the existing
native editor, with no new business rule or API/schema/runtime contract.
Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

## Verification plan

1. Add a failing native adapter test for `PATCH /portfolio/assets/{id}` and
   an API-backed native UI regression for edit, save, and persisted readback.
   Use the PostgreSQL repository test to prove trade-history preservation.
   Run the focused adapter test red, then green.
2. Add an explicit backend unit test for fractional quantity × unit price
   plus fee, unless current tests already prove that exact invariant.
3. Run surrounding backend portfolio tests, native focused tests/build, and
   repository design/project guards where applicable. Record skipped fixture
   gates with reason.
4. Reconcile the portfolio screen spec, human iOS guide and its AI-agent
   reference, validation matrix, and this note. No architecture/ADR/ERD or
   public API reference change is expected because the existing contract is
   reused. UAT on a production-connected device remains open.

## Local result and evidence

- Native asset cards now open a prefilled metadata editor. The existing PATCH
  route changes symbol/name under the same asset ID; success reloads, while
  an API failure leaves the editor and draft visible. No trade or position
  mutation was added.
- The backend already multiplies scaled quantity by unit price. New regression
  `TestComputePortfolioPositionMultipliesUnitPriceByFractionalQuantity` passed
  before implementation: `1.5 × 100000 + 2500 = 152500` VND cost basis. The
  native create/correction editors already label the field `Giá đơn vị` and
  submit quantity and unit price separately.
- Native adapter RED: focused `xcodebuild` failed to compile because
  `FinanceAPI.updatePortfolioAsset` did not exist. GREEN: focused adapter
  XCTest exited 0 after the method was added.
- API-backed UI: the first run was terminated by Xcode while another test used
  the same simulator. A dedicated simulator run then failed fixture login
  because the test runner used default port 8081 while the fixture was on
  18082. The final dedicated run used port 8081 and passed 1/1, zero skips:
  `/tmp/mypocket-portfolio-rename-20261026-2.xcresult`. It checked fields,
  save, same ID, persisted metadata and refreshed card. Its disposable
  database was dropped by the fixture wrapper.
- Backend: `cd backend && go test ./internal/entity ./internal/controller/http
  ./internal/infrastructure/repository -run Portfolio -count=1` passed. With a
  migrated disposable PostgreSQL database, the four focused portfolio
  repository tests for metadata history preservation, duplicate/owner scope,
  concurrent delete and correction version guard passed. Native base guard,
  project verification and `git diff --check` passed.

Passing native API-backed UI command from the repository root:

```sh
rtk proxy env MP_LOCAL_API_UI_STATE_DIR=/tmp/mypocket-portfolio-api-ui-20261026 bash mobile-ios/scripts/local-api-ui-fixture.sh run -- rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'id=A8F59B50-E67F-42E9-89F9-53CBD756F5B2' -derivedDataPath /tmp/mypocket-portfolio-derived-20261026 -resultBundlePath /tmp/mypocket-portfolio-rename-20261026-2.xcresult -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedPortfolioAssetRenamePersistsWithoutCreatingAnotherAsset -parallel-testing-enabled NO -collect-test-diagnostics never 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI' MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 test -quiet
```

Focused PostgreSQL command used inside the disposable fixture wrapper:

```sh
TEST_DATABASE_URL="$DB_DSN" go test ./internal/infrastructure/repository -run "TestPortfolio(UpdateAssetMetadataDoesNotRewriteTrades|UpdateAssetRejectsDuplicateSymbolAndForeignOwner|DeleteCannotCascadeConcurrentTrade|CorrectionVersionGuard)$" -count=1
```

Docs review: the portfolio screen spec now describes the actual correction
revision model and asset rename action. The public iOS guide and its Markdown
AI-agent reference explain unit price × quantity and the rename flow. Existing
API reference, architecture, requirements, ERD and ADR need no update because
no contract, schema or boundary changed. Shared context/backlog and release
notes are outside this portfolio-only ownership; the portfolio validation row
holds local proof. Production correction routes, deployment and physical UAT
remain open, so this note does not change the production feedback status.

Trace: [backlog](../BACKLOG.md), [native phase](../phases/PHASE-002-native-ios-swiftui.md),
[portfolio correction contract](../tickets/FEEDBACK-20261019-PORTFOLIO-CORRECTION-CONTRACT.md),
[API](../../architecture/API.md), [screen](../../design/pages/portfolio/README.md),
[guide](../../guides/portfolio-ios.md), [validation](../VALIDATION_MATRIX.md),
[release notes](../../releases/CHANGELOG.md).
