---
artifact_type: detail_design
status: in_progress
approval: owner approved two tabs and reversible owner confirmation in chat on 2026-10-14
---

# Owner-confirmed feedback resolution

Links: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[public feedback guide](../../guides/feedback-ios.md).

Add independent `resolved` boolean default false and nullable `resolved_at` to
feedback, additive migration32 only. Preserve dev status/fixed_at/changelog and
all existing reports. PATCH /api/v1/feedback/:id/resolution accepts explicit
`{"resolved":true|false}`, JWT or a full-owner user API key. Legacy key scope
metadata is ignored; owner comes from authentication, never request data;
cross-owner/missing ID404. Repeating the same value is idempotent, preserving
resolved_at. Reopening clears resolved_at. Agent status/changelog operations
cannot change owner confirmation. Redis audit records the separate
owner-resolution event; no public raw feedback route.

iOS uses existing MPSegmentedControl, MPCheckboxRow, MPSurfaceCard/MPText. Tabs
Chưa giải quyết/Đã giải quyết with counts, flat compact items. Move item only
after successful API response; per-item pending disables duplicate taps. Failure
leaves item/tab count unchanged with error. Existing fixed reports still start
unresolved until owner confirms. Keep create form in bubble/overlay opened from
header, so viewing reports is not below a long form. Changelog stays available
through a shared expandable disclosure, separate from resolution tabs.

Verification: failing route regression first, backend owner/repeat/unresolve/
API-key scope and migrated PG persistence tests, native decode/filter/update
tests, Release compile/signature/install. Deploy isolated feedback-only candidate
on current deployed auth baseline, backup first, apply32 forward and preserve
financial/feedback content counts. No web maintenance or release publication.

Docs: public guide/API/OpenAPI, internal migration/auth boundary, native screen
contract, context/backlog/validation. No financial business/API changes.

Independent review caught stale GET/PATCH ordering: pull refresh can complete
after a saved checkbox and replace it with old state. Reproduced by held GET
transport regression (actual PATCH count1 while expected0). Serialize load/write
at viewmodel boundary and disable checkboxes during refresh; block refresh while
any save is pending. Test permits write again after refresh completion. No
backend change needed; the earlier sequential UI proof does not cover this race.

## Verification and rollout checkpoint

- Backend route RED: missing PATCH resolution route; GREEN full race suite on
  migrated disposable PostgreSQL, both current source and isolated candidate.
  Integration tests cover owner isolation, timestamps/retries, reopening and
  unchanged fixed/changelog history; HTTP tests cover true/false, missing/null/
  invalid boolean, read-only API key403 and cross-owner404.
- Native wire RED2 has explicit nil-vs-true/resolved_at failures in diagnostics;
  initial unsigned run interrupted while collecting diagnostics. Native GREEN
  `/tmp/mypocket-feedback-resolution-green-20261014.xcresult`:170/170 pass.
- UI `/tmp/mypocket-feedback-resolution-ui-20261014.xcresult`:1/1 pass, actual
  local API resolve→move tab→uncheck→move back with server re-read. Screenshot:
  `/tmp/mypocket-feedback-resolution-ui-attachments-20261014/35EC109B-E073-4298-A304-E49C05A03E85.png`.
- Review race RED2 `/tmp/mypocket-feedback-race-red2-20261014.xcresult` fails
  PATCHcount1 vs0; full native GREEN171/171 at
  `/tmp/mypocket-feedback-race-green-20261014.xcresult`. Initial race test helper
  type-inference compile error corrected before behavioral RED; not counted as
  proof of the production race.
- Isolated backend commit `1f6a7f7d4a07ac537e34d1a599f4f3f083c7637a`, pushed on
  `codex/feedback-owner-resolution-20261014`, parent deployed auth12572c87.
  Unrelated dirty finance/web source excluded; candidate Swagger regenerated.
- Built/pushed `mypocket-api:feedback-1f6a7f7d` digest
  `sha256:fec2c747fb06158c5da54ae60cbe9faef4ff48f6ade5d918d84915a8c0551acd`.
- Prod validated backup `/var/backups/mypocket/feedback-resolution-1791939806/pre-resolution.dump`
  113107bytes/214catalog lines. Migration31→32 dirty=false; before/after counts
  1user/3wallets/22transactions/42feedback, all wallet/ledger/feedback-content
  hashes identical. All old confirmations remain false. No prod feedback ticks.
- **Activation blocked:** manager pi SSH agent communication/auth failure;
  worker vm100 still runs healthy auth12572c87. Owner asked to unlock/approve
  1Password SSH. Do not claim new route active in production.
- Final device build reached codesign, pending owner Keychain approval. A stale
  earlier pending signing process was stopped before final source build. New
  app not installed while production route remains unavailable. Existing native
  density build stays installed. Release and owner UAT pending.
