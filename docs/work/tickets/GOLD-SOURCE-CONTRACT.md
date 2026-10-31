---
artifact_type: ticket
id: GOLD-SOURCE-CONTRACT
status: waiting_external_contract
updated: 2026-11-01
lane: blocked
---

# Gold price-source contract gate

## Scope

This ticket records the external information required before MyPocket can
enable automatic gold pricing. It does not authorize a live request, a
catalog seed, or production configuration.

## Required owner input

For each of `gold_source_a` and `gold_source_b`, provide a sanitized contract
containing:

- endpoint and HTTP method;
- authentication method and the names of required secret settings (never the
  secret values);
- a sanitized example response and its error/partial-response behavior;
- stable external product IDs and the product-to-unit mapping;
- currency and buy/sell meaning, including which field is the store buyback
  price;
- source timestamp field and timezone/format, including behavior when missing.

## Current implementation state

Both adapters exist only as disabled provider shells and report
`waiting_external_contract`. They do not contain an invented URL, schema,
credential, product mapping, fixture claimed as live, or enabled catalog row.
The shared HTTP transport already bounds timeout to 10 seconds and response
body to 1 MiB once a real adapter contract is supplied.

## Unblock criteria

After the owner supplies both contracts, add sanitized fixture tests for
buyback mapping, identity, unit/currency validation, source timestamp parsing,
malformed/partial/error responses, and one bounded read-only integration check
per source. Enable catalog entries only after those checks identify the exact
products. Never convert units, average sources, or fall back across providers.

## Links

- [Accumulated-assets design](../../superpowers/specs/2026-10-31-accumulated-assets-design.md)
- [Implementation plan](../../superpowers/plans/2026-10-31-accumulated-assets.md)
