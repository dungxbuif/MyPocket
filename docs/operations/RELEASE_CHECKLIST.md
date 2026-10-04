---
artifact_type: release_checklist
id: RELEASE_CHECKLIST
status: draft
owner: shared
human_fields: [release_approval, rollback_acceptance]
ai_fields: [pre_release_checks, post_release_checks, verification_notes]
shared_fields: [status]
---

# Release Checklist

## Field Ownership

- Human owns release approval and rollback acceptance.
- AI records pre-release checks, post-release checks, and verification notes.

## Pre-Release

- [ ] Tests passed
- [ ] Local web gate passed (`docs/operations/LOCAL_RELEASE_GATE.md`)
- [ ] Master docs reconciled
- [ ] ADRs updated
- [ ] Release notes prepared
- [ ] Release note links to the matching changelog entry, commit, image digest,
      tests, and known issues
- [ ] Rollback plan confirmed

## Post-Release

- [ ] Deployment verified
- [ ] Monitoring checked
- [ ] Incidents recorded if any
- [ ] `docs/CONTEXT.md` updated
- [ ] Release note status is `verified`; human approval is recorded before
      changing it to `done`
