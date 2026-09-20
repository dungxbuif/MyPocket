---
artifact_source: visual_reference
---

# Wallet selector reference

Runtime update: WalletSelectionList now implements grouped rows, total, selection check, header/actions through WalletManagementPanel. Edit selects a wallet into its edit sheet. Goal selection opens savings details; basic/credit currently retain local selection only, not global report/transaction filtering. This supersedes the reference-only status below. Browser list/create navigation inspected; complete downstream scope flow remains pending.

Owner-supplied screen.png and code.html are input evidence. Existing-wallet list shows aggregate, included wallets, selection check and actions. It is not a wallet-type specification. Its icon/title/subtitle/check row anatomy is reused for basic/goal/credit selection. Existing-wallet aggregate/filter behavior remains a separate scope. See [screen contract](../screens/wallets/README.md).

## Normalized composition

| Reference region | Required base | Intended behavior |
| --- | --- | --- |
| Đóng / Chọn Ví / Sửa | BaseBottomSheet form + BaseButton | Close picker; edit navigates to wallet management |
| Tổng cộng | SurfaceCard + row BaseButton + IconBadge + Text | Sum current balances of included wallets; selected check |
| Tính vào tổng | Text/Heading | Semantic section label |
| Wallet list | SurfaceCard + row BaseButton + IconBadge + Text + Divider | Real wallet name/current balance; selected check |
| Thêm ví | BaseButton + IconBadge | Open Add Wallet |
| Liên kết dịch vụ | BaseButton disabled + Text | Explicitly unsupported until integration exists |

Do not copy sample bank names, balances, logos, blue selection or large corner radii into runtime. Use real data and shared tokens. Loading, empty plain status, retryable error and selected states must be implemented together. Excluded wallets need a separate labeled group if shown; aggregate excludes them.

Status: normalized reference only; existing-wallet picker and its consuming scope/filter contract are not implemented by the wallet-type picker. Product decision still needed for where wallet scope applies. No backend/schema changes are inferred from this image.
