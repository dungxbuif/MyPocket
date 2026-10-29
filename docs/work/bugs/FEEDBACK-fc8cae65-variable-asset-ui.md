---
artifact_type: bug
id: FEEDBACK-fc8cae65
status: blocked
updated: 2026-10-26
---

# Variable-value asset UI

Source: owner-supplied, read-only transcription of unresolved production
feedback `fc8cae65` (status open, `resolved=false`):

> Title: Tài sản
>
> Description: Chưa có UI để nhập các tìa khoản biến động

The typo `tìa` is retained from the source. This note does not change the
production feedback record or its status.

## Source and scope

The original title/body is now available from the owner. Earlier retained
ticket files paraphrased it as "Variable-value assets" and classified it as
needing owner UAT. That paraphrase was narrower than the actual description.
The owner has since chosen a separate aggregate balance account for changing
total value, with transfers for deposits/withdrawals and valuation-only value
changes. Gold remains an itemized portfolio asset. The approval-gated
[detail design](../tickets/FEEDBACK-fc8cae65-INVESTMENT-ACCOUNT-DETAIL_DESIGN.md)
records implementation approaches, open choices, and proof requirements.

Scope is portfolio-specific native/API behavior, tests, and documentation.
Do not edit feedback, notification, AI import, or production data.

## Intended behavior and decision gate

Current portfolio contracts allow an asset with fixed-point quantity, integer
VND unit price, trades, and explicit current-price updates. Valuation is
unknown until a current price exists; portfolio total market value is unknown
while any held asset lacks a price. Price movement must not change a wallet
balance. The native entry point is `Ví của tôi` -> `Danh mục đầu tư`, with
`Thêm tài sản`, `Ghi giao dịch`, and per-asset `Cập nhật giá` actions.

The owner selected a new aggregate account, distinct from a portfolio
holding. Current portfolio tests therefore cannot close this feedback; they
only prove the itemized path. Proposed schema/API/reporting and gold
type/unit changes need owner review of the linked detail design before code.
No migration, account creation, production data change, or implementation is
authorized by this note alone.

## Verification plan

1. Compare the owner-provided verbatim feedback with portfolio/account
   contracts; preserve the production record unchanged.
2. Audit the native portfolio asset, trade, and price UI against the existing
   API/model contract. Run focused backend and native tests using local data.
3. Obtain approval for the linked detail design's aggregate-account boundary,
   overlap handling, transfer semantics, and gold unit/type choices. Only then
   plan implementation, reproduce missing behavior with failing tests, and
   implement the approved slice.
4. Reconcile affected requirements, API/ERD, portfolio user/agent docs, and
   validation evidence after code. Physical owner UAT and deployed valuation
   readback remain separate gates, never inferred from local tests.

## Local result

The existing native screen exposes asset creation, trade entry/history,
explicit current-price editing, and position/valuation readback. The current
API and model preserve an unknown valuation until a price is supplied. These
cover the itemized-asset interpretation in local source/tests, but not the
newly requested balance-account workflow. That new workflow is absent from
current API/schema/native UI by design, so this is a feature gap rather than
a regression in existing portfolio math. Attempt count: zero fixes; no
implementation in this design-only task.

Verification on 2026-10-26:

- `cd backend && go test ./internal/entity ./internal/controller/http
  ./internal/infrastructure/repository -run Portfolio -count=1`: passed.
- Focused native model tests for portfolio summary nullable valuation and
  trade details sorting/filtering: 2 passed, 0 failed/skipped. Result bundle:
  `/tmp/mypocket-fc8cae65-model-20261026.xcresult`.
- API-backed native UI test
  `testAPIBackedPortfolioAssetCreateAndPriceRefresh`: 1 passed, 0
  failed/skipped. It used a disposable local API database, which the fixture
  wrapper dropped after the test. Result bundle:
  `/tmp/mypocket-fc8cae65-price-20261026.xcresult`.

This work note and its linked draft detail design are the only files edited
for the clarified request. The portfolio guide, screen spec, API reference,
and AI-agent-readable Markdown still describe implemented behavior; they must
not claim the proposed aggregate account as shipped. No production deployment
or data change was made. Design approval, implementation, physical owner UAT,
and production-connected valuation readback remain open.

Trace: [backlog](../BACKLOG.md), [native phase](../phases/PHASE-002-native-ios-swiftui.md),
[portfolio requirements](../../requirements/SPEC.md), [portfolio rules](../../requirements/BUSINESS_RULES.md),
[screen](../../design/pages/portfolio/README.md), [guide](../../guides/portfolio-ios.md),
[validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).
