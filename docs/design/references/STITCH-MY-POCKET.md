# Stitch MyPocket reference

Status: integrated reference input  
Reviewed: 2026-10-07  
Source: `/Users/dungxbuif/Downloads/stitch_my_pocket.zip`

This archive is a visual and component reference, not runtime business logic.
The web API and existing MyPocket domain rules remain the source of truth for
wallets, transactions, budgets, jars, reports, AI/OCR, feedback, transfers,
credit and account isolation. HTML in the archive is treated as design
evidence; it is not copied into the production bundle.

## Source inventory

The archive contains 31 PNG screen captures at approximately 738×1600 (29
product states plus the two component-example captures), two HTML component
examples, and one token contract:

- `financial_clarity_system/DESIGN.md`: color, typography, spacing, shape,
  elevation and component guidance.
- `base_component_c_c_lo_i_bi_u_th_ng_k/`: chart/summary component example.
- `s_a_nh_m_edit_category/`: category edit screen example.
- `image.png_1`, `image.png_2` … `image.png_11`, `img_7696` … `img_7712`,
  `img_7793`: screen-state evidence used for visual comparison.

The archive is intentionally kept outside Git under Downloads. This document
records the stable contract so an agent can reproduce the mapping without
depending on the ZIP being present in the repository.

## Canonical tokens

The exact primitive values come from `DESIGN.md` and are mirrored in
`app/src/ui/theme.css` and
`mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`:

| Role | Value |
| --- | --- |
| Canvas/background | `#fbf9f9` |
| Card | `#ffffff` |
| Primary CTA/container | `#4caf50` |
| Primary/action | `#006e1c` |
| Action hover/fixed variant | `#005313` |
| Body/heading text | `#1b1c1c` |
| Secondary text | `#556158` |
| Muted/outline | `#6f7a6b` |
| Outline variant | `#becab9` |
| Divider/surface container | `#efeded` |
| Muted row surface | `#f5f3f3` |
| Negative/error | `#ba1a1a` |
| Positive soft surface | `#d9e6da` |
| Main card radius | `24px` |
| Control radius | `12px` |
| Page margin | `16px` |
| Filled action shape | full pill (`rounded-full`) |
| Card padding | `20px` |
| Grid gutter | `12px` |
| Font | Manrope (web-loaded; native bundles `Manrope-Variable.ttf` under the OFL license) |

`rounded-reference-card` / `MPSurfaceCard.Style.reference` owns the 24px
reference-card shape and the soft `0 4px 20px` shadow. Existing 12px control
corners remain valid for fields and compact controls.

## Screen-to-base mapping

| Reference region | Web base | Native base |
| --- | --- | --- |
| Balance/header | `AppHeader`, `Heading`, `IconButton` | `MPBalanceHeader`, `MPIconButton` |
| Wallet rows | `WalletCard`, `SurfaceCard`, `IconBadge` | `MPWalletRow`, `MPSurfaceCard(.reference)` |
| Wallet-ledger header/summary | centered wallet selector immediately below header, followed by period controls and summary surface | `MPScreenHeader`, `MPWalletScopeSelector` + `MPWalletScopeGlyph`, `MPSegmentedControl`, `MPMetricCard`, `MPSurfaceCard(.reference)` |
| Wallet-ledger transaction groups | `TransactionRow`, reference card | `MPTransactionRow`, `MPSurfaceCard(.reference)` |
| Wallet create/edit form | `ReferenceScreenHeader`, `ReferenceCard`, `BaseTextInput`, `BaseSelect`, `BaseButton` | `MPReferenceScreenHeader`, `MPSurfaceCard(.reference)`, shared wallet fields/selectors, safe-area `MPButton` |
| Money Insider/insight | `OverviewMoneyInsiderCard`, `SurfaceCard` | `MPMoneyInsiderCard`, `MPReportsInsiderCard` |
| Week/month report | `OverviewReportCard`, `SegmentedControl`, `BaseBarChart` | `MPOverviewReportCard`, `MPSegmentedControl`, `MPBarChart` |
| Category management (`img_7793`) | `GroupManagementPanel`, `SegmentedControl`, `BaseCategoryTree` | `CategoriesView` + planned shared category-tree base; Stitch also shows a wallet-scope filter and search |
| Category edit | `CategoryEditForm`, `BaseTextInput`, `BaseSelect`, `BaseButton` | shared `MPTextField`, `MPOptionSelector`, `MPButton` |
| Global add | `BaseFab` / shared shell action | `MPBottomAddButton` |
| Primary navigation | `BottomNavigation` | `MPBottomNavigation` (inset rounded surface, blur/shadow, centered add) |

Segmented controls use a muted line track with a white active pill and
action-colored label. Native owns this in `MPSegmentedControl` plus the shared
`MPButtonStyle.segmentedSelected` variant; feature screens must not use the
solid primary button style for an active segment.

All screen-level controls must compose these bases. A new visual variant needs
an entry in the design system and a consumer test; it must not introduce
screen-local colors, radii, or raw interactive elements.

For `img_7793`, the visible expense/income/debt segment and root/child tree
match the web category-management contract. The Stitch wallet filter can be
derived from the authenticated category `wallet_ids` response: an empty list
means all wallets, otherwise the category applies only to the listed owner
wallets. The inactive-category toggle is now supported: personal categories
can be soft-archived and restored through the owner-scoped
`PATCH /api/v1/categories/{id}/status` contract. Existing ledger, report,
budget, and recurring references remain intact; new assignments reject an
inactive category. The management screen hides inactive categories by default
and exposes them through the toggle. See the current [category-management
contract](../pages/account-groups/README.md), [API
reference](../../architecture/API.md), and [archive design](../../superpowers/specs/2026-10-08-category-archive-design.md).
Native parity evidence and any remaining visual gaps are tracked in
[`PORT-100-AUDIT`](../../work/PORT-100-AUDIT.md) and
[`BACKLOG`](../../work/BACKLOG.md).

The shared native text-field bridge is explicitly bounded to the 48px control
height from this contract. This is important for UIKit-backed inputs because a
missing intrinsic height can stretch an email or text field across the form and
break the reference rhythm.

Two reference screenshots are intentionally reconciled against the current
web/API contract rather than copied literally: the native shell keeps the
dedicated `Trợ lý` tab requested by the product (some Stitch captures show a
`Ngân sách` tab in that slot), and the transaction editor exposes the supported
expense/income types only. The current backend has no debt-transaction write
contract; credit/debt behavior is provided through the specialized Credit
surface and debt category tree. Adding a disabled or fake third segment would
make the native UI look closer while misrepresenting available behavior.

## Behavior and feature coverage

Visual parity does not change behavior. The current port preserves the
implemented feature matrix:

- overview totals, wallet scopes, Money Insider and report links;
- ordinary income/expense transactions, edit/delete, date/time, category and
  jar assignment;
- transfer and adjustment actions from the transaction overflow menu;
- wallet CRUD and basic/goal/credit specialized detail contracts;
- budgets, category hierarchy, jars, month notes, recurring schedules and
  Travel Mode;
- reports, portfolio, notifications and account settings;
- AI entry with attachment/review flow, Finance Assistant read-only tools,
  feedback/changelog and owner-scoped API calls.

Remaining release gates are proof gates, not visual placeholders: live
provider/OCR credentials, authenticated production feedback, physical-device
signing and owner visual UAT. See `docs/work/VALIDATION_MATRIX.md` for the
current status and exact commands.
