# Budget progress cards

Sources IMG-04 and IMG-05: [extraction](../../system/SOURCE_EXTRACTION.md). [Gauge](../budget-meter/README.md).

## Anatomy

SurfaceCard; leading category IconBadge; name and remaining/help; right limit amount; Progress; optional today marker and warning badge. Shared base owns rounded track, fill and palette; parent supplies values and labels.
Observed examples: Mua sắm limit 8.000.000đ / remaining 5.810.000đ; Ăn uống 5.000.000đ / remaining 2.005.000đ; Di chuyển 3.000.000đ / over 450.000đ with VƯỢT 15%.

## Behavior and conflicts

Fill = spent/limit ×100, visual clamp 0–100; numeric label may exceed 100. Negative/invalid denominator behavior must follow product contract before real data integration.
Current preview caller still uses ratioPercent, which clamps its displayed percentage; preserving raw over-limit numeric percentage is part of the real budget slice. Progress itself accepts and clamps independently.
IMG-04 says safe <80%, warning 80–99%, danger ≥100%; IMG-05 warns when ahead of today marker and danger >100. These conflict. Do not choose silently. Current BudgetProgressItem supports over-limit danger only, matching spent > limit; today marker/warning are planned.
Persisted budget data, empty budget, unavailable forecast and accessibility value text are screen-slice responsibilities. Budget and Jar remain different products.

## Current bases

BudgetProgressItem and GoalCard compose SurfaceCard/IconBadge/Text/Progress. BudgetGauge owns arc. No screen-local progress div, colors or badge geometry.
