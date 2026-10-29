# Native transaction-editor icon parity

## Decision

The Stitch/web transaction editor uses the selected wallet/category identity in
the form rows. Native currently uses a generic document wallet symbol and a
generic tag for every selected category. The existing category presentation
registry already maps the canonical backend `system_key` families, so the
editor will derive its category row icon from that shared registry. Wallet
rows keep the shared wallet symbol until a provider-specific wallet-logo asset
contract exists; no fake bank logos will be introduced.

## Scope

- Extend `TransactionsPresentation.glyph` to recognize the canonical icon-key
  fallbacks used by user-created categories.
- Feed the selected category glyph into the shared `MPOptionSelector` row in
  `TransactionEditorView`.
- Preserve API payloads, selection behavior and base-component ownership.

## Verification

- Add a pure glyph fallback assertion.
- Run focused/full native tests and Release warnings-as-errors.
- Re-run the API-backed editor screenshot and record the result.

## Result

- Implemented `iconKey` fallback resolution in the shared native glyph policy
  and wired the selected category glyph into the base editor selector row.
- Focused glyph contract and API-backed editor UI test pass; full native unit
  target is 118/118 and the Release warnings-as-errors build passes.
- Screenshot evidence:
  `/tmp/mypocket-visual-wallet-editor-icon2-attachments/0BFB1ECA-9040-426B-ABE8-1F95A365B1DC.png`.
- Provider-specific wallet logos remain out of scope because the current API
  contract exposes no logo asset. The shared wallet symbol is intentionally
  retained until that contract exists.
