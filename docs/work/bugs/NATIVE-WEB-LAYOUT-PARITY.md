# Native iOS ↔ Web layout parity audit

Date opened: 2026-10-06
Status: implementation gaps closed; authenticated pixel/UAT sign-off remains open

## Intent

The iOS renderer already uses shared base components and the web semantic
tokens, but its primary shell does not yet preserve the web layout contract.
This work audits the authenticated shell against the current web composition
and closes the highest-impact differences before adding more native features.

## Parity contract

- Keep the web primary navigation order: Tổng quan, Giao dịch, Trợ lý, Tài khoản.
- Keep one persistent bottom navigation with the centered global `+` action for
  ordinary transaction creation. Long-press/secondary entry remains reserved
  for AI entry; the native surface must expose the same manual add affordance
  without requiring the user to discover the ledger header.
- Keep the web overview hierarchy: sticky balance header, month summary card,
  wallet card and recent-transactions card. Native screens compose the same
  semantic bases rather than creating screen-local styling.
- Preserve the transactions contract: wallet selector, weekly/custom period
  controls, three-dot actions, and the same global add flow.

## Initial gaps found (now resolved)

1. Native bottom navigation exposes Ví and Thêm instead of Trợ lý and Tài
   khoản, so the primary shell is not isomorphic with the web shell.
2. Native bottom navigation has no centered add-transaction action.
3. Native overview renders a four-metric grid as its first content block and
   omits the web month-summary card and its detail/jar actions.
4. Native overview has no web-equivalent wallet management affordance and the
   recent-transaction card lacks the web card header action.
5. The native shell uses `MPScreenHeader` as the overview header, while the
   web shell owns a sticky total-balance header shared by primary pages.

The five items above record the initial audit, not current implementation
status. They were closed as follows:

- Primary destinations now match the web order in
  `MPShellNavigation.primaryItems`: Tổng quan, Giao dịch, Trợ lý, Tài khoản.
- `MPBottomNavigation` owns the persistent centered global add action; long
  press routes to AI transaction entry while ordinary tap opens manual entry.
- `OverviewView` renders the balance header, wallet card, Money Insider,
  monthly report, month summary and recent transactions in the approved order.
- Wallet management is reachable from the wallet card; recent-transaction
  section action and month/jar navigation are wired to existing native routes.
- The reusable `MPBalanceHeader` owns the web-aligned balance, visibility,
  search and notification anatomy instead of a generic screen header.

Automated route/add/overview evidence is recorded in
`docs/work/VALIDATION_MATRIX.md` (`testAuthenticatedShellExposesPrimaryRoutes`,
`testGlobalAddLongPressOpensAIEntry`, `testAPIBackedOverviewFinalSectionStaysAboveShellNavigation`).
This closes the implementation checklist only. It does not claim pixel-level
owner acceptance on a physical device or populated production data.

## Implementation and verification plan

- Extend the shared native navigation base with the centered add action and
  stable accessibility identifiers.
- Route the add action into the existing `TransactionEditorView`, preserving
  the existing API and validation contract.
- Align `MainTabView` primary items with the web shell and keep secondary
  finance features under the account surface.
- Recompose `OverviewView` in the web order using existing metric/surface,
  action and async-state bases.
- Add focused unit/UI coverage for the shell action contract, then run the
  native project verifier, all native XCTest targets, a Release simulator
  build and a production-backed visual smoke on the simulator.
## New visual reference

On 2026-10-06 the owner opened Money Lover on the paired iPhone 15 Plus. A
device screenshot was captured as `/tmp/moneylover-iphone.png` for this local
review. The reference shows a large balance header, wallet card with icon
rows, Money Insider card, monthly report card, pale canvas background and a
center floating add button. It is visual evidence only; the implementation
continues to use MyPocket's web API and extended features as the behavior
source of truth.

The native shell will therefore use Money Lover's visual hierarchy while
retaining MyPocket's four web primary destinations (Tổng quan, Giao dịch, Trợ
lý, Tài khoản) and global add action. No Money Lover source code or assets are
copied.

## Scope expansion: unified web and native port

The requested port covers both clients. The production web client remains the
behavior and API source of truth (wallets, transactions, budgets, jars,
assistant, OCR, feedback and transfers); the native client must expose the same
capabilities through the same semantic base components. The Money Lover
reference controls visual hierarchy only: an airy balance header, wallet rows
with icon badges, a promotional/insight surface, a monthly report surface and
a persistent center add action. The web client will receive the same visual
language so users do not see two unrelated products.

The owner-provided Stitch archive is available at
`/Users/dungxbuif/Downloads/stitch_my_pocket.zip`. Its design tokens and screen
inventory are captured in
`docs/design/references/STITCH-MY-POCKET.md`; the archive is visual evidence,
not runtime source. The web API and existing MyPocket contracts remain the
behavior source of truth.

### Verification completed / remaining

- The shared controls and shell contracts are implemented through native base
  components. Project/base-component guards and XCTest evidence are recorded in
  the validation matrix.
- Remaining: compare authenticated, populated web and iOS device captures
  screen-by-screen and obtain owner pixel/UAT sign-off. This is an acceptance
  gate, not an unimplemented primary-shell capability.
