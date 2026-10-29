# Native Wallets — populated reference state

Status: `in_progress`  
Reference: [Money Lover iPhone survey](../reference-survey/2026-10-16-money-lover-iphone.md), ML-11–ML-21  
Owner: native iOS

## Composition

The Wallets route has two explicit modes. The selector uses the centered
`Chọn Ví` title and aggregate scope. Management uses `Ví Của Tôi`, grouped
included/excluded wallets, and the existing add/edit affordances. Both modes
use the same data and privacy mask; management never changes which wallets are
included in totals.

| Region | Shared base | Behavior |
| --- | --- | --- |
| Selector/management header | `MPSelectionScreenHeader`, `MPButton` | Back and edit actions remain stable; title changes only with mode. |
| Aggregate scope | `MPSurfaceCard(.reference)`, `MPWalletAggregateGlyph` | `Tổng cộng` sums only `isInTotal` wallets and is never a synthetic wallet. |
| Wallet groups | `MPWalletSelectionList`, `MPWalletSelectionRow`, `MPWalletManagementRow`, `MPDivider` | Populated rows show name above balance, 40pt wallet glyph, selected state, and type-specific metadata. |
| Goal/credit details | `MPProgressBar`, `MPText`, `WalletDetailView` | Goal target and credit availability remain visible without changing ledger semantics. |
| Add/edit | `WalletEditorView`, `MPEditorScaffold`, `MPWalletSelector` | Existing wallet API, validation, and owner authorization remain unchanged. |

## States and proof

- `WalletsView` loads the authenticated wallet list and account timezone. It
  does not seed or mutate sample data.
- Empty, loading, error, selector, management, and populated included/excluded
  group states remain represented by shared async/list bases.
- `WalletsReferenceTests` covers the mode titles, aggregate total, and group
  separation. Existing wallet CRUD tests remain the behavior authority.

Known gaps: reference type-help dialogs, sort UI, currency search, linked-wallet
landing, archive/member actions, and full existing basic/credit detail parity
remain in the survey backlog.
