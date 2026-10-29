# Transaction form rows

Source IMG-06: [extraction](../../system/SOURCE_EXTRACTION.md).

## Anatomy

Leading slot 36–40px for icon/bank/currency; flexible center for label/value/input; trailing slot for chevron, date stepper or toggle. Row touch area minimum 44px. Group related rows in SurfaceCard with shared divider.
Typography Text/Heading; inputs BaseTextInput/BaseSelect; leading icon IconBadge; controls BaseButton/IconButton/checkbox or planned switch. InlineControlRow is layout-only, not another field style.

## Observed rows

Wallet Techcombank, amount VND/0, Chọn nhóm, Ghi chú, date with previous/next controls, Với, Đặt vị trí, Chọn sự kiện, Đặt nhắc nhở, Thêm Hình Ảnh, Không tính vào báo cáo, Lưu plus image button.

## States / events

Default placeholder muted; filled value ink; pressed uses shared feedback; disabled noninteractive. Selector activates its corresponding picker; calendar arrows step date; toggle changes a boolean; image action opens source selection. These are design intentions, not proof that dialogs/APIs exist.
Before a screen uses a row, specify onActivate, value, disabled, required/error, field label and picker result. Existing FormSelectorRow is a preview with no activation callback; it must be extended at base before real screen wiring.
Do not infer required fields, attachment uploads, geolocation permission or recurring semantics from image alone; product ticket controls these.
