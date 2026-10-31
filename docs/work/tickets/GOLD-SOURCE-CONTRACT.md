---
artifact_type: ticket
id: GOLD-SOURCE-CONTRACT
status: implemented
updated: 2026-11-01
lane: active
---

# Gold price-source contract gate

## Scope

This ticket records the source contract shared by MyPocket's market poller and
the production n8n Hanoi gold workflow.

## Captured contract

The configured source keys are `btmc` and `btmh`. Both are official HTML GET
endpoints, with no user-provided URL and no credential-bearing request in the
MyPocket default configuration:

- BTMC: `https://btmc.vn/`, `GET`; the table exposes rows identified by
  `VÀNG MIẾNG VRTL BẢO TÍN MINH CHÂU` and `NHẪN TRÒN TRƠN BẢO TÍN MINH CHÂU`.
  The source labels values as `1 = 1.000 VNĐ`; the adapter stores the first
  (buy) field as VND per `chỉ`.
- BTMH: `https://baotinmanhhai.vn/bang-gia-vang`, `GET`; embedded product
  records expose `Trang sức - Nhẫn tròn BTMH` and `Vàng trang sức 24K (999.9)`.
  The first numeric price after each stable row marker is the buy field in VND
  per `chỉ`.
- Source timestamps are parsed in `Asia/Ho_Chi_Minh`; a missing or malformed
  timestamp, missing product, invalid response, or partial source fails that
  provider poll closed. Existing quotes are retained on failure.
- Stable external product IDs are `btmc-vrtl-buy`, `btmc-ring-buy`,
  `btmh-jewelry-ring-buy`, and `btmh-jewelry-24k-buy`. All are `VND/chỉ` with
  `valuation_side=buyback`.

## Current implementation state

Both adapters are enabled in the backend provider registry. The shared HTTP
transport bounds timeout to 10 seconds and response body to 1 MiB. Migration
`000049` enables the four catalog rows. The worker polls each provider once per
10-minute slot using `market_poll_claims`; source endpoint overrides remain
deployment-only through `GOLD_BTMC_ENDPOINT` and `GOLD_BTMH_ENDPOINT`.

## Verification criteria

Sanitized fixture tests cover buyback mapping, identity, unit/currency,
timestamp parsing, partial responses, and bounded transport. The n8n production
deployment record is the read-only source of truth for the two official
endpoints and their 10-minute schedule. The adapter never averages sources or
falls back across providers.

## Links

- [Accumulated-assets design](../../superpowers/specs/2026-10-31-accumulated-assets-design.md)
- [Implementation plan](../../superpowers/plans/2026-10-31-accumulated-assets.md)
