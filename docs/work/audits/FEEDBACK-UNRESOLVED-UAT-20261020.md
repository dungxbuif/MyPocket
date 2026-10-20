---
artifact_type: feedback_uat_audit
id: FEEDBACK-UNRESOLVED-UAT-20261020
status: complete_with_residuals_production_deployed
owner: shared
updated: 2026-10-20
---

# Unresolved production feedback audit — 2026-10-20

## Source and safety boundary

Authenticated read-only `GET /api/v1/agent/feedback?limit=100` was executed on
the production node using its runtime `FEEDBACK_AGENT_TOKEN`. The token was not
printed, copied or stored locally. The response contains 50 owner-scoped
records: 32 with `resolved=true` and 18 with `resolved=false`. No production
feedback status, screenshot, changelog or financial record was changed.

## Audit method

- Cross-reference the 16 historical unresolved rows with
  `FEEDBACK-20261019-CLASSIFICATION.json`, current source and named tests.
- Investigate the two new rows against the native AI-entry UI, attachment
  lifecycle and backend cleanup command.
- Build the current candidate for the dedicated simulator and use MobileNext
  `mobilecli` for accessibility-tree/navigation evidence.
- Record physical-device/provider-only gaps as external acceptance, not as
  locally verified behavior.

## Current queue summary

| Group | Count | Current treatment |
| --- | ---: | --- |
| Owner resolved | 32 | Excluded from this audit; do not reopen automatically. |
| Historical unresolved | 16 | Reconcile prior fixed/needs-UAT/contract evidence with current code. |
| New unresolved | 2 | Fresh source, test and Mobile MCP assessment required. |

## Historical unresolved reconciliation

| Classification | IDs | Evidence / remaining gate |
| --- | --- | --- |
| Covered locally; owner has not marked resolved | `a738ad4a`, `8eeb848f`, `75dee199`, `e5428521`, `aa6ca3cf`, `f4597cd2`, `f7887563`, `1565a9e5`, `b209e80f`, `5ff74529`, `f42c103d` | Shared category/transaction/card/sheet/navigation bases and focused native regressions cover the reports. The Account-root Add assertion found during this audit was stale and was corrected to match the already-approved contract. |
| Needs external or populated UAT | `1db32af3`, `d098ff72`, `2d0811c0`, `26fcad87`, `fc8cae65`, `824dea29` | Requires some combination of physical Face ID, feedback-sheet device geometry, populated portfolio data, deployed provider response or owner acceptance. Local code/tests are not promoted to device acceptance. |
| Remaining product/accounting contract | `b017284a` | Portfolio history remains append-only. Broader edit/delete semantics require a reviewed accounting contract before implementation. |

## New records

| ID prefix | Report | Initial evidence | Status |
| --- | --- | --- | --- |
| `baca2415` | Multi-image failures are hard to diagnose; confirm unused attachment cleanup | MobileNext accessibility audit proves 20 thumbnails and independent remove actions. The regular worker now invokes the bounded `AIEntryAttachmentCleanup` pass when private storage is configured; the one-shot `cleanup-attachments` command remains available for diagnostics. Per-file provider/source diagnostics remain an explicit UAT concern. | implemented locally — pending production worker/UAT confirmation for scheduling and source-file diagnostics |
| `c86b669d` | AI transaction suggestions should be a list of cards matching single-transaction entry | MobileNext audit proves proposals render as multiple `MPSurfaceCard` rows with type, amount/date, note and edit/skip/save actions. The rows now compose the canonical `MPTransactionRow` contract for category glyph/title, date-or-note subtitle, signed amount, masking, and Dynamic Type. | implemented locally — pending visual UAT |

## Mobile MCP evidence

- Dedicated simulator: `MyPocket-iPhone-16`, iOS 18.5,
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.
- Current-source Debug candidate built and installed before inspection.
- MobileNext accessibility traversal verified 20 attachment rows
  (`IMG_01.PNG` … `IMG_20.PNG`) with individual `Bỏ tệp` actions and the pinned
  analyze action. The floating feedback bubble can overlap a remove-control
  region while the long list is mid-scroll, so this is not recorded as full
  visual acceptance.
- Proposal fixture verified a card with unresolved type, amount/date, note and
  `Sửa` / `Bỏ qua` / `Lưu`; this supports the partial classification above.
- Focused native UI suite result: 9 passed, 5 fixture-dependent skipped and one
  stale Account-root Add assertion failed. The assertion contradicted the
  approved Account contract; after correction the isolated regression passed.
- Merged-tree confirmation: 16 selected feedback/AI-entry/navigation cases,
  11 passed and five environment-guarded cases skipped; zero failures.

## Safety conclusion

No feedback status was changed. The owner remains the authority that moves a
report between unresolved and resolved. The two new residuals are included in
the production image and authenticated GET-only UAT passed, but they remain
open for owner verification because physical-device visual/provider evidence is
not substituted by server checks. The six external UAT rows likewise remain
open.

## Production follow-up — 2026-10-20

- API/worker image `release-20261020-a2de756` is live; migrations are at
  version 40 and clean.
- Dedicated UAT login succeeded against `https://money.dungxbuif.com`.
- MobileNext detected the paired physical iPhone online and MyPocket `1.1.0`
  installed. The app is behind its lock screen; do not claim Face ID or touch
  regression until the owner unlocks it.

## Reconciliation targets

- [Feedback remaining map](../tickets/FEEDBACK-20261019-REMAINING.md)
- [Validation matrix](../VALIDATION_MATRIX.md)
- [Physical iOS UAT](../../operations/IOS-PHYSICAL-UAT-20261019.md)
- [Release note](../../releases/REL-1.2.0-20261020.md)

