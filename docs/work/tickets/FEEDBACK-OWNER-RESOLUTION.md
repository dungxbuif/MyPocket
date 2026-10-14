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
`{"resolved":true|false}`, JWT or user API key feedback:write. Scope owner from
authentication, never request data; cross-owner/missing ID404. Repeating the same
value is idempotent, preserving resolved_at. Reopening clears resolved_at. Agent
status/changelog operations cannot change owner confirmation. Redis audit records
the separate owner-resolution event; no public raw feedback route.

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
