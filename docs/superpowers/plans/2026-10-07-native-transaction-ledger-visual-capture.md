# Native transaction-ledger visual capture

## Intent

Add authenticated simulator evidence for the ordinary transaction ledger so
the native port can be compared with the supplied Stitch/web ledger contract,
not only validated through semantic controls.

## Contract

- Use the disposable local API fixture and the existing owner-scoped ledger
  route; do not mutate production data.
- Preserve the web/native contract: wallet scope, week/custom controls,
  cashflow totals and date-grouped transaction rows.
- Capture the rendered ledger after the controls and cashflow card load.
- Any correction must use existing shared base components and retain the
  current transaction filtering semantics.

## Verification

- Run the focused API-backed ledger UI test and export the PNG attachment.
- Inspect the capture against the supplied Stitch transaction references.
- PASS: the ordinary ledger capture is exported and inspected at
  `/tmp/mypocket-ledger-visual-final/Logs/Test/` with PNG attachment
  `/tmp/mypocket-ledger-visual-final-attachments/7F6A03B6-627B-4664-949F-B70A4E3AD26D.png`.
  The rows remain reachable above the persistent shell after scrolling; no
  feature-local safe-area workaround was required.
- If code changes are needed, use a red assertion before the fix, then rerun
  the focused UI test, full native unit target, Release warnings-as-errors and
  project/base-component/diff guards.
