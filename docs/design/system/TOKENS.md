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

## Money Lover reference foundation (Task 1)

The native reference foundation uses adaptive semantic roles rather than
screen-owned colors. `MPColor.pageCanvas` is the neutral/lavender page field,
`groupedSurface` is the flat grouped-row background, `separator` is the subtle
row divider, `primaryAccent` is the MyPocket selection/action accent, and
`income` / `expense` are restrained financial-direction colors. `primaryText`,
`secondaryText`, `mutedText`, and `rowFill` preserve readable contrast on those
adaptive backgrounds. Each role has an explicit light and dark appearance.
Existing `canvas`, `surface`, `line`, `ink`, `heading`, `secondary`, `muted`,
`row`, `accent`, `brand`, `action`, and `danger` names remain source-compatible
aliases so feature migration can happen independently without creating a
second palette.

Grouped native geometry is owned by `MPMetrics`: 24pt grouped-surface corners,
44pt minimum rows, 16pt row horizontal padding, 10pt row vertical padding, and
a 56pt leading inset for separators aligned after a leading icon. Grouped rows
must grow vertically at accessibility Dynamic Type sizes; clipping or a fixed
maximum height is forbidden.

Verification: `DesignSystemTests` resolves every adaptive role in light and
dark traits, checks compatibility aliases and shared metrics, measures grouped
surface/row layout at normal and accessibility text sizes, and checks inset
divider alignment. The focused native test target plus
`mobile-ios/scripts/check-base-components.sh` is the Task 1 gate.

Native owner refinement (2026-10-13): `MPMetrics.pagePadding` remains 16pt;
`MPMetrics.buttonCornerRadius` is **8pt**, replacing the native pill-action
shape. Filled, outline, secondary and destructive `MPButton` variants reuse
this token; screens must not override their shape. Control height is now 44pt
with a minimum 44pt hit target. Card rounding is unchanged by this refinement;
circular icon and feedback actions remain circular. The web table above is a
legacy reference, not the authority for this native override. These values
have native DesignSystem token assertions; the urgent 2026-10-14 density pass
updates assertions but skips test execution at the owner's explicit request.

Native typography is centralized in `MPFont`: display/title, balance/largeTitle,
headline, subheadline body/chat, footnote metadata, caption labels, compact
tabular `amount`, and title3 currency.
Each role uses an explicit semantic iOS text style and weight, so Dynamic Type
scales the hierarchy instead of every role silently inheriting a larger style.
UIKit form input text stays 16pt. SwiftUI text follows Dynamic Type;
transaction amounts move below copy at accessibility sizes. Shared cards use
12pt horizontal/8pt vertical padding, 12pt corners and no shadow; default transaction/wallet rows use a
32pt glyph and 44pt minimum content height. Transaction rows add the shared
`transactionRowVerticalPadding` of 6pt above/below for readable separation;
wallet row geometry is unchanged.
Page section gaps are 8pt and dense row gaps 4pt. Menu/category rows keep a 44pt
minimum and grow for content; profile is one tappable identity row, no duplicate
manage-account row. Typography remains unchanged; do not achieve density by
clipping accessibility text. Account root retains center Add; Account child
destinations and Assistant hide it.
The shell never reserves layout space for Feedback. The 52pt bubble is a
trailing overlay inside the route viewport, 8pt above the visible navbar;
there is no dock, spacer, inset or content-height subtraction.
The selected primary tab uses the shared `surfacePill` style (`MPColor.row`,
22pt radius); shell geometry is centralized in `MPShellNavigation` so the add
button and feedback bubble cannot drift from the reference layout. The web
target keeps the OFL-licensed Manrope font, while native iOS uses the system SF
Pro metrics through `MPFont`. Shared prose uses medium weight without changing
the approved text sizes; feature
screens do not choose fonts directly.

Text dùng Text/Heading với size/weight/tone. Số tiền dùng numeric; scale text 10/11/12/14/16/18/20/24/30/36px. Kích thước 10–11px chỉ cho metadata phụ; không dùng cho input/hành động.
Layout spacing theo Tailwind scale; kích thước đặc thù tree/chart phải được base sở hữu và đặc tả. Không khai báo thêm màu literal trong TS/TSX/CSS ngoài theme. Không thêm font/theme song song.
