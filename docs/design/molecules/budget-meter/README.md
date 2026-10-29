# Budget semi-circle gauge

Source IMG-05, with IMG-04 alternative: [extraction](../../system/SOURCE_EXTRACTION.md).

180° arc with rounded caps, neutral track and accent fill. Label “Số tiền bạn có thể chi”; sample 45.075.000đ. Supporting metrics: total budget, total spent, days remaining; CTA Tạo Ngân sách. Sample numbers are not calculation requirements.

Current BudgetGauge accepts value percent and label, renders SVG arc with normalized path length 100 and clamped fill; value >100 gets danger. Accessible title includes label and value. Parent composes SurfaceCard, Text and MetricBox.

Planned today marker = currentDay / daysInPeriod ×100 in account timezone; it belongs to a shared progress component and must receive explicit period data. No hard-coded “12 days” or “14 days” in production forecasts.
One export draws remaining and another spent; caller must name the metric, not silently reverse arc meaning. Current preview gauge label is Đã dùng (spent).
Verification: zero/100/over/invalid values, SVG geometry and named tone; real budget source and warning thresholds require budget slice proof.
