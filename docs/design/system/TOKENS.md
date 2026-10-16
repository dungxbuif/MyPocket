# Token contract

Nguồn primitive duy nhất: app/src/ui/theme.css, được import từ app/src/styles.css. Named component classes nằm ở app/src/ui/variants.ts; app/src/atomic/atoms/tokens.ts chỉ re-export để tương thích import cũ.

| Role | Token | Giá trị | Dùng cho |
| --- | --- | --- | --- |
| CTA | brand / brand-hover | #4caf50 / #006e1c | primary button, FAB |
| Action | action / action-hover | #006e1c / #005313 | link, positive amount, focus |
| Selected/progress | accent | #4caf50 | checked, ring, progress |
| Canvas/card | canvas / card | #fbf9f9 / #ffffff | nền màn / card |
| Text | heading / ink | #1b1c1c / #1b1c1c | heading / body |
| Metadata | secondary / muted | #556158 / #6f7a6b | nội dung phụ / chú thích |
| Divider/row | line / row | #efeded / #f5f3f3 | border / hover và muted card |
| Danger | danger / danger-hover | #ba1a1a / #93000a | âm, lỗi, xóa |
| Danger background | danger-soft / danger-line | #fff1f2 / #ffe4e6 | feedback |
| Success background | success-soft / success-line | #d9e6da / #becab9 | positive badge |
| Warning | warning / warning-soft / warning-line | #d97706 / #fffbeb / #fef3c7 | cảnh báo |
| Card/control radius | radius-card / radius-control | 12px / 12px | control corners; reference cards use 24px |
| Container margin | page-margin | 16px | screen content inset on web and native |
| Action shape | button-radius | rounded-full / 999pt | filled CTA, secondary and destructive buttons; row controls keep control radius |
| Badge radius | radius-badge | 16px | compact visual badge khi dùng |
| Card shadow | shadow-card | 0 4px 20px, black 3% | subtle elevation |

Category palette dùng các named tones trong IconBadge; business mapping chọn tone, không cung cấp màu. Các màu orange/teal/red còn có token riêng trong theme; extended category palette lấy từ Tailwind tại variant registry, không ở screen.

Native source: `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift` owns current
iOS tokens. The table above is the archived web reference; native UX follows
the owner's current design rather than a third-party app. `MPColor.brand`
(`#4caf50`) owns brand/icon accents; `MPColor.action` (`#006e1c`) owns readable
primary CTA backgrounds, links, positive values and focus/selection. Native
form controls use
`MPColor.surface`, `MPColor.line`, `MPColor.accent` and `MPColor.successLine`
through `MPFormControlModifier`; feature screens must not provide ad-hoc
border or focus colors.

Native owner refinement (2026-10-13): `MPMetrics.pagePadding` remains 16pt;
`MPMetrics.buttonCornerRadius` is **8pt**, replacing the native pill-action
shape. Filled, outline, secondary and destructive `MPButton` variants reuse
this token; screens must not override their shape. Control height is now 44pt
with a minimum 44pt hit target. Card rounding is unchanged by this refinement;
circular icon and feedback actions remain circular. The web table above is a
legacy reference, not the authority for this native override. These values
have native DesignSystem token assertions; the urgent 2026-10-14 density pass
updates assertions but skips test execution at the owner's explicit request.

Native typography is centralized in `MPFont`: display 28, headline 18,
subheadline 16, body 15 (medium), metadata 13 (medium), labels 14/12 and currency
20 points. UIKit form input text stays 16pt. SwiftUI text follows Dynamic Type;
transaction amounts move below copy at accessibility sizes. Shared cards use
12pt horizontal/8pt vertical padding, 12pt corners and no shadow; default transaction/wallet rows use a
32pt glyph with no outer vertical padding and 44pt minimum content height.
Page section gaps are 8pt and dense row gaps 4pt. Menu/category rows keep a 44pt
minimum and grow for content; profile is one tappable identity row, no duplicate
manage-account row. Typography remains unchanged; do not achieve density by
clipping accessibility text. Disabled Account-tab center add action is hidden.
The shell reserves only the 56pt navbar; its 52pt feedback bubble is an overlay,
not a full-width lane or spacer. The web target keeps the OFL-licensed Manrope
The selected primary tab uses the shared `surfacePill` style (`MPColor.row`,
22pt radius); shell geometry is centralized in `MPShellNavigation` so the add
button and feedback bubble cannot drift from the reference layout. The web
target keeps the OFL-licensed Manrope font, while native iOS uses the system SF
Pro metrics through `MPFont`; feature
screens do not choose fonts directly.

Text dùng Text/Heading với size/weight/tone. Số tiền dùng numeric; scale text 10/11/12/14/16/18/20/24/30/36px. Kích thước 10–11px chỉ cho metadata phụ; không dùng cho input/hành động.
Layout spacing theo Tailwind scale; kích thước đặc thù tree/chart phải được base sở hữu và đặc tả. Không khai báo thêm màu literal trong TS/TSX/CSS ngoài theme. Không thêm font/theme song song.
