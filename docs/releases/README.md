# Changelogs

This folder stores the active release baseline and future release notes. The current
repository snapshot is documented as the first release, `1.0.0` dated 2026-09-24.
Earlier development history remains available in Git commits and is not duplicated in
the active changelog.

## Files

- `CHANGELOG.md`: version-level changes.
- `RELEASE_NOTES_TEMPLATE.md`: template for phase, epic, or release notes.

## Release documentation cadence

- Every externally visible fix or feature gets a `CHANGELOG.md` entry before it
  is deployed.
- Every production release gets a versioned release note in this folder. The note
  records the shipped commit/image digest, verification evidence, known issues,
  and links back to the changelog and work artifacts.
- Release notes move from `draft` to `verified` only after tests and deployment
  checks have fresh evidence. `done` remains a human release-approval state.
- Planned work, experiments, and unverified local behavior stay in plans or
  tickets; they must not be presented as shipped release changes.
- AI-agent-readable surfaces are kept stable: Markdown front matter, the
  versioned changelog, and explicit verification/deployment sections.
