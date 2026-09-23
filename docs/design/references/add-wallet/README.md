---
artifact_source: visual_reference
---

# Add Wallet reference

Owner-supplied screen.png and code.html are input evidence, not runtime implementations. Sheet with cancel/title/save; primary grouped card for name, VND, balance; service action; exclude-total switch plus help. No note field. App typography, colors and 12px card geometry come from shared bases. See [page contract](../../pages/wallets/README.md).

## Normalized composition

| Reference region | Shared base | Runtime behavior |
| --- | --- | --- |
| Header Hủy / Thêm Ví / Lưu | BaseBottomSheet form + BaseButton | Cancel draft; submit valid form; saving disables actions |
| Wallet icon and name | IconBadge + BaseTextInput title | Icon follows type; name required; no independent icon picker |
| Currency row | IconBadge + Text | Fixed VND; no unsupported selector |
| Current balance | FormField + BaseTextInput title | Integer opening balance; defaults to zero |
| Row grouping | SurfaceCard + Divider | Canonical radius/colors; no copied export CSS |
| Wallet type extension | FormField + BaseSelect | Owner chốt select khi tạo; khóa loại sau lưu. Runtime picker cũ cần thay |
| Service action | BaseButton disabled + Text | Explicitly unavailable |
| Exclude total | BaseSwitch + SurfaceCard + Text | checked means is_in_total=false |

No decorative device frame, status bar or home indicator is implemented. Empty/invalid/saving/error behavior follows the screen contract. Reference is normalized, implementation checks pass; browser visual acceptance remains pending.
