---
artifact_type: release_notes
id: REL-1.0.8
status: verified
owner: shared
human_fields:
  - release_approval
  - known_issues_acceptance
ai_fields:
  - summary
  - changes
  - verification
  - linked_work
shared_fields:
  - status
  - trace
trace:
  backlog_items: []
  requirements: []
  phases: []
  tickets_or_bugs: []
  test_verification: backend/go-test-all
  validation_matrix: TBD
  docs_review: docs/releases/README.md
  adrs: []
  changelog: CHANGELOG.md#108---2026-10-01
---

# Release Notes: Finance Assistant batched tool calls

## Status

- ID: `REL-1.0.8`
- Status: `verified`
- Release date: `2026-10-01`

## Summary

Finance Assistant no longer rejects a valid model response containing five
parallel read-only finance tool calls. The bounded orchestrator budget is eight
tool calls per assistant turn; this is a runtime safety guard, not a user usage
limit.

## Changes

- Replaced the previous hard-coded four-call guard with the bounded eight-call
  budget.
- Added a regression test for five tool calls returned in one provider response.
- No database migration or user-data reset was part of this release.

## Verification

- `go test ./...` passed in `backend`.
- Focused advisor regression tests passed.
- `npm run check:design` passed in `app`.
- Production health endpoint returned 200.
- Swarm API service is running `1/1` on image
  `registry.dungxbuif.com/mypocket-api:release-46c48f91` with digest
  `sha256:2be9c9a189e7caee55f8b6fdd9e16a0a3dec1eaa4fff55f10501f65dcfded226`.

## Known Issues

- The Finance Assistant remains synchronous JSON in this release; streaming,
  durable audit delivery, and the separate AI quality harness are not included.
- Human release approval and known-issue acceptance remain pending until the
  owner explicitly promotes this note from `verified` to `done`.

## Linked Work

- Changelog: [CHANGELOG.md](CHANGELOG.md#108---2026-10-01)
- Code fix: commit `46c48f91`
- Deployment documentation: [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
