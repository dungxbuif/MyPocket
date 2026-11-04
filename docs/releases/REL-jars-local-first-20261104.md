---
artifact_type: release_note
id: REL-jars-local-first-20261104
status: ready
date: 2026-11-04
deployment: none_merge_only
human_fields:
  - release_approval
  - known_issues_acceptance
---

# Jars as wallet-style plan cards and local-first native app

Status: ready. The owner approved merging to the main branch and pushing to GitHub and GitLab on 2026-11-04.
Nothing was deployed: no production API/worker rollout, migration or device install was performed.

## Included

See the changelog entry "Jars as wallet-style plan cards and local-first native app, 2026-11-04". In short: wallet-style
jar list and detail with a cumulative plan figure and reconciliation, `jar_id` transaction filter, `PUT` month upsert,
Assistant jar-assignment confirmation card, local-first native architecture, and the fix for the raw context card.

## Verification

- Backend `go test ./...`: PASS. Native unit: 478/478. Simulator `LocalFirstRuntimeTests`: 14/14.
- Synthetic jar-suggester benchmark: 32 notes, 23 proposals, 0 wrong, 9/9 no-evidence rows left to the user.

## Known limits (owner to accept)

- No physical-device timing: the owner declined installing a debug build; simulator numbers only.
- Jar-suggestion quality is not measured on the owner's real notes.
- Jars have no delete API (only per-month removal), so test accounts accumulate jars.

## Trace

[Changelog](CHANGELOG.md) · [Validation matrix](../work/VALIDATION_MATRIX.md) · [ADR-029](../decisions/ADR-029-jar-cumulative-plan-presentation.md) ·
[Backlog](../work/BACKLOG.md)
