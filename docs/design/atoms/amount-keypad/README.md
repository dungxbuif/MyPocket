# Amount input & numeric keypad

Source IMG-02: [extraction](../../system/SOURCE_EXTRACTION.md). Status: partial preview; dedicated amount/keypad bases must be created before real transaction use.

## Observed layout / copy

Header Hủy / Thêm Giao Dịch; segment Khoản chi, Khoản thu, Vay/Nợ. Currency badge VND, amount 0, helper Không đồng, mode Chờ nhập. Save disabled at zero; image action alongside. Horizontal presets 150.000, 50.000, 500.000, 40.000. Key rows: C/÷/×/backspace, 7/8/9/−, 4/5/6/+, 1/2/3/XONG, 0/000/comma; XONG spans two rows.
Use base key/chip/button + Text; complete keypad needs one reusable molecule owning grid and input events.

## Source behavior

HTML stores primary operand, secondary operand, operator and calculated flag. Digits edit active operand; 000 appends zeros; comma is a decimal separator. Setting another operator computes the prior pair when available. Backspace removes digit or clears operator; C resets; preset replaces value and clears operator; XONG evaluates pending expression then formats vi-VN result.
Display uses thousands separators, separate operator indicator and approximate amount-in-words. Save in export is a mock action, not proof of persistence.

## Production contract required before use

Record allowed precision/negative values/max length, divide-by-zero error, rounding and VND integer policy against transaction requirements. Avoid copying JavaScript demo number parsing/zero-division fallback as business rules. XONG confirms amount editing; Lưu submits the form only when all required fields validate. Hủy must follow screen's dirty-form rule.
States: zero, entering, operator pending, calculated, invalid expression, disabled/saving. Accessible key labels for symbol keys; min 44px controls; keyboard support must be specified/tested.
Current QuickAddSheet has incomplete arithmetic and is not mounted. Do not claim the exported calculator is implemented.
