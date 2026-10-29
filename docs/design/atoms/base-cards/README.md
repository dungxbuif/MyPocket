# Base cards — design & behavior

Source IMG-01: [extraction register](../../system/SOURCE_EXTRACTION.md). [Shared tokens](../../system/TOKENS.md).

## Anatomy

White SurfaceCard; heading/action row; optional divider; repeated leading icon + flexible title/subtitle + trailing amount. Cards with settings replace amount with checkbox/switch/navigation slot. Pair stats use equal columns and a shared metric base.
Card corners current 12px; source 24px superseded by owner-directed radius reduction. Padding and elevation use named props. No individual wallet/transaction row becomes another card.

## Variants

- Wallet: IconBadge from global wallet kind → tone mapping. Name, subtitle, right-aligned balance; negative danger, nonnegative ink; masked value renders bullets.
- Transaction: category icon, title, note/wallet metadata, right-aligned signed money. Positive action, transfer secondary, expense danger. Keep sign/copy alongside tone.
- Settings: leading icon, title, help, trailing control. Tap model belongs to one control; don't nest checkbox inside button.
- Metric: label/value/footnote and optional status; theme type scale, numeric figures.
- Feedback: StatusMessage danger uses alert; neutral status uses status role.

## Events / states

Clickable list rows use BaseButton row or a base link; passive rows should not imply actions that do not exist. Current preview WalletCard/TransactionItem have no action callback; retain preview classification.
Masking changes visible amount, not stored value. Loading/empty/error/retry are parent screen states, with text/feedback bases.
Switch illustrations show two states but do not prove persistence timing; implement dedicated toggle contract before using that interaction.

## Verification

Base loading/disabled/single-elevation tests; browser fixture validates primary color and shared cards. Per-screen data/UAT remains in screen ticket.
