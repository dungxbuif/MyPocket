# Edit group — component contract

Source IMG-03, code.html and editable enable.html reference: [extraction](../../system/SOURCE_EXTRACTION.md). [Screen](../../screens/account-groups/README.md).

## Composition

PageBackHeader → metadata SurfaceCard → read-only explanation → ApplicableWalletsCard → actions/status.
CategoryEditForm owns name/kind/parent/icon/wallet selections; InlineControlRow slots own geometry, BaseTextInput title and BaseSelect inline own field styling. Icon picker opens BaseBottomSheet; selected IconButton has base-owned ring.
Wallet row uses IconBadge and BaseCheckbox; all colors come from shared catalog.

## Behavior

System metadata is read-only; explanation says only wallet applicability can change. Personal group exposes editable metadata and delete on screen header. Kind change resets parent selection. Parent choices are roots of matching kind excluding self. Required trimmed name is checked before save. Save forwards input to screen; errors and persistence are screen/API responsibilities.
Applicable wallet choices are multi-select, not single-select radio. No selected wallets is displayed as áp dụng tất cả ví by current contract. Do not infer backend semantics from a static empty list.
Current editing uses routes new/edit; icon selection alone uses a sheet. Source image's “Sửa” action is visual reference, not an instruction to create a second system-metadata editing mode.
