# Finance Assistant product-help retrieval

Status: local implementation complete; production/device acceptance remains open. This closes the code-only gap where
the assistant had a static handbook but no grounded, source-identifiable help
retrieval. It does not add financial writes, external browsing, owner data
access, migrations, provider configuration or production rollout.

## Contract

`get_product_help` is a read-only advisor tool. It accepts a bounded natural
language `query` (1–240 bytes) and optional `max_results` (1–5). The registry
matches against a versioned, embedded handbook section catalog. Results contain
the handbook version, stable source IDs, titles and short excerpts. No raw
prompt, account data, URL or arbitrary instruction is returned.

No-match results are explicit (`status: not_found`, empty sources); the model
must say the feature is not documented instead of inventing navigation. The
tool is owner-independent, but it is still executed inside the authenticated
advisor loop so the existing principal/credential checks and audit boundary
remain in force.

## UI contract

The result is a `product_help` structured card. Web and native render a compact
“Hướng dẫn trong ứng dụng” card with the answer sources and version. Unknown
old clients continue to use their existing textual fallback.

## Verification

- Go unit tests prove matching, bounded arguments, no-match behavior and tool
  registry exposure.
- Native model/card tests prove the stable type/title and source rendering path
  does not coerce arbitrary values.
- Web production build, design/type checks and existing unit contracts remain
  green.
- No environment, database, production, device or release action is part of
  this slice.

## Local evidence

- `go test ./...` and `go vet ./...` pass after registry/handler changes.
- Native focused card proof: `/tmp/mypocket-product-help-native-green-20261015.xcresult`.
- Native warnings-as-errors simulator build, base-component guards and project
  verification pass.
- Web `npm run build`, `test:design`, `test:feedback`, `test:auth` and
  `test:transaction-jars` pass. Browser suites requiring a live fixture remain
  separately environment-gated and were not used as acceptance for this slice.
