---
artifact_source: visual_reference
---

# Wallet selector reference

Runtime update: WalletSelectionList now implements grouped rows, total, selection check, header/actions through WalletManagementPanel. Edit selects a wallet into its edit sheet. Selecting a basic wallet routes to the wallet-scoped transaction ledger; the ledger selector supports `Tổng cộng` and individual wallets. Goal selection opens savings details, while credit selection opens the credit detail/ledger surface. Browser list/create navigation and the local API-backed scope flow are implemented; authenticated visual/device UAT and linked-service integration remain separate gates.

Owner-supplied screen.png and code.html are input evidence. Existing-wallet list shows aggregate, included wallets, selection check and actions. It is not a wallet-type specification. Its icon/title/subtitle/check row anatomy is reused for basic/goal/credit selection. Existing-wallet aggregate/filter behavior remains a separate scope. See [page contract](../../pages/wallets/README.md).

## Normalized composition

| Reference region | Required base | Intended behavior |
| --- | --- | --- |
| Đóng / Chọn Ví / Sửa | `MPSelectionScreenHeader` + shared button base | Center the title, keep symmetric capsule actions; Đóng returns to the previous screen, Sửa toggles wallet edit mode |
| Tổng cộng | SurfaceCard + row BaseButton + `IconBadge(Globe, categoryTeal)` + Text | Sum current balances of included wallets; teal globe badge; title and balance stack vertically; selected check |
| Tính vào tổng | Text/Heading | Semantic section label |
| Wallet list | SurfaceCard + row BaseButton + IconBadge + Text + Divider | Real wallet name/current balance stack vertically; selected check; no generic type subtitle |
| Thêm ví | BaseButton + IconBadge | Open Add Wallet |
| Liên kết dịch vụ | BaseButton disabled + Text | Explicitly unsupported until integration exists |

Do not copy sample bank names, balances, logos, blue selection or large corner radii into runtime. Use real data and shared tokens. Loading, empty plain status, retryable error and selected states must be implemented together. Excluded wallets need a separate labeled group if shown; aggregate excludes them.

Status: normalized reference plus implemented consuming scope contract. This image remains a visual reference, not a wallet-type specification. The wallet-type picker does not own scope filtering: management rows own selection, and the transaction/detail screens consume the selected wallet. The current native management route keeps the persistent MyPocket shell, but matches the centered title and capsule actions. No backend/schema changes are inferred from this image. Linked-service integration and authenticated visual/device UAT remain open.
