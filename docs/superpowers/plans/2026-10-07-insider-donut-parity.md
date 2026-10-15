# Money Insider comparison ring parity

## Intent

The supplied Stitch reference shows the Money Insider card with a compact
comparison ring beside the current-period average. The web already exposes a
reusable `BaseDonutChart`, but the mounted Insider card did not render it and
the native card had no equivalent base. This slice adds the ring only when the
owner-scoped report response contains `expense_change_bps`; it never invents a
financial percentage or makes a provider call.

## Contract

- `expense_change_bps` is basis points, converted through the existing shared
  percentage formatter.
- The visible center keeps the actual percentage (including values above
  100%); the stroke is clamped to the ring's 0–100% geometry.
- Positive change uses the danger semantic, negative change uses the action
  semantic, and zero uses the neutral line semantic.
- Missing comparison data renders no ring and keeps the existing fallback copy.
- Native uses a reusable `MPDonutChart` base; the feature card supplies only
  the value, label and semantic tone. The web uses the existing
  `BaseDonutChart` base.
- No API, database, schema, ownership or masking behavior changes.

## Verification plan

1. Add pure native ring presentation tests for positive, negative, zero and
   missing comparison values.
2. Run the focused native unit target, the full native unit target, Release
   warnings-as-errors and project/base/diff guards.
3. Run the web design checks and production build.
4. Reconcile the design inventory, iOS renderer contract, validation matrix,
   audit and changelog with exact evidence paths.

## Result

- Focused native ring contract: PASS.
- Full native unit/contract target: PASS, 124/124 at
  `/tmp/mypocket-insider-donut-final.xcresult`.
- Release warnings-as-errors: PASS at `/tmp/mypocket-insider-donut-release`.
- Reports API-backed UI smoke: PASS, 1/1 at
  `/tmp/mypocket-insider-donut-api-ui.xcresult`.
- Native project/base-component/diff guards and web design/build checks: PASS.
- Physical-device pixel approval and production/provider UAT remain external.
