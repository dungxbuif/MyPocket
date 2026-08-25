# MyPocket Mobile Component Design Contract

Source assets: `design/IMG_7696.PNG` through `design/IMG_7731.PNG`.

This contract is authoritative for the PHASE-001 PWA/mobile shell. Components must match the screenshots closely: rounded iOS-style finance UI, Vietnamese copy, green primary actions, red expense/destructive emphasis, gray app canvas, white grouped panels, and a persistent bottom navigation with a raised center add action.

## Visual Tokens

| Token | Value | Usage |
| --- | --- | --- |
| App canvas | `#f2f3f8` | Body background |
| Surface | `#ffffff` | Cards, grouped rows, sheet panels |
| Text primary | `#111111` | Headings, important labels |
| Text secondary | `#8e8e93` | Supporting labels, helper text |
| Divider | `#e8e8ec` | Row separators and chart grid |
| Primary green | `#2dbd4f` | Primary CTA, active links, positive values |
| Soft green | `#e8f7ed` | Secondary CTA backgrounds |
| Expense red | `#ff5a66` | Spend totals and destructive emphasis |
| Income blue | `#32a9df` | Income values |
| Icon navy | `#29495a` | Category/wallet icon circles |
| Warning orange | `#ff8800` | Premium ribbon/accent |

Typography uses the system font stack. Mobile shell headings are bold and compact; dashboard balance text is the only hero-scale type. Letter spacing remains `0`.

## Layout Rules

- Target mobile width first, matching iPhone screenshot proportions. Desktop may center the same shell in a constrained column or use a left rail, but it must keep the same components.
- Page padding is `18px` to `20px`; vertical rhythm is spacious with grouped sections separated by `28px` to `36px`.
- Cards use white surfaces, subtle borders or shadows, and large radii matching the screenshots: page cards and grouped panels use `24px` to `28px`; inner chips/buttons use fully rounded pills.
- Do not nest visual cards inside other decorative cards. Grouped row panels may contain rows separated by dividers.
- Bottom navigation is fixed, translucent/white, pill-shaped, and safe-area aware. The center add button is a raised green circle.

## Core Components

### App Top Bar

- Overview: large total balance at top-left, label `Tổng số dư`, eye visibility button, search icon, notification icon with red badge.
- Subpages: circular/pill back action left, centered bold title, optional circular icon action right.
- Filter chips use rounded white pills with icon + label, e.g. `Tổng cộng`, `Tháng: 08/2026`.

### Balance And Wallet Card

- White rounded panel titled `Ví của tôi`, right action `Xem tất cả` in green.
- Wallet rows: circular icon, bold wallet name, right-aligned VND amount; dividers between rows.
- Negative values use dark/navy text unless specifically expense/destructive.

### Report And Analytics Cards

- Section headers are gray bold labels with optional green action link.
- Charts live in white rounded panels; use red for spending, blue for income, green for budgets, pale gray for comparison data.
- Segmented controls are light gray tracks with white selected segment.

### Transaction Add Sheet

- Presented as a rounded white sheet over a dimmed/blurred background.
- Header: left pill `Hủy`, centered `Thêm Giao Dịch`, optional right action.
- Amount row: small `VND` pill and large numeric amount.
- Grouped rows use leading black line icons, muted placeholder text, and chevrons.
- Bottom action bar has a wide disabled/enabled `Lưu` pill plus a green circular receipt/image action.
- Image source dialog is centered, translucent, rounded, with stacked large gray pill buttons.

### Account Screen

- Profile card has centered avatar circle, optional orange premium ribbon, username, email, and Google mark.
- Account actions use full-width white pill rows. Destructive actions are red.
- Device rows use circular platform icon and two-line text.

### Budget Screen

- Header uses bold title, account filter chip, menu/help icon pills.
- Budget summary includes large green amount, gauge arc, and three compact stats.
- Budget rows use category icon, amount, remaining amount, progress bar, and `Hôm nay` marker.

### Offline And Auth States

- Offline status should be visible but not noisy: a compact gray/green pill near top content or nav.
- Auth loading/unauthenticated states must keep the same app-shell frame. Login action is a green full-width pill; no marketing hero.
- No PHASE-001 shell may show real finance editing behavior; values are sample/placeholder until finance phases land.

## Navigation Contract

Bottom tabs:

- `Tổng quan`
- `Sổ giao dịch`
- Center `+` quick add
- `Ngân sách`
- `Tài khoản`

Tabs use icon over label. Active tab sits on a soft gray oval. The plus action opens the transaction add sheet.

## Verification

- Component/unit tests must assert the five destinations, quick add sheet, offline indicator, and auth state.
- Browser verification must include mobile viewport screenshots and PWA manifest/service worker presence.
- Text must not overflow pill buttons, cards, tabs, or the bottom navigation at mobile width.
