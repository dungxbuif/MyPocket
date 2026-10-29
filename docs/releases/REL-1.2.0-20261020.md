---
artifact_type: release_note
id: REL-1.2.0-20261020
status: deployed_to_production_and_physical_uat
owner: shared
release_approval: approved_by_owner_2026-10-20
---

# MyPocket 1.2.0 — native, AI and integration hardening

## Release intent

Integrate the complete `codex/feedback-backlog-20261019` candidate into
`main` without rewriting either branch. This release groups the verified
native feedback work, Finance Assistant reliability work and authenticated
MCP/API parity surface completed after 1.1.0.

## Intended behavior

- Preserve the native-only product direction and shared iOS base components.
- Ship the owner-scoped Finance Assistant fixes and quality harness currently
  present in the candidate, without weakening confirmation, CAS, replay or
  owner-isolation boundaries.
- Publish `/api/v1/mcp` through the official Go SDK with the same same-owner
  capability available to JWT and full-owner API-key callers.
- Preserve production financial data and existing feedback resolution state.
- Re-audit every production feedback item that the owner has not marked
  resolved. Production resolution remains owner-controlled and is not changed
  by this release preparation.

## Feedback/UAT audit plan

The authenticated production snapshot on 2026-10-20 contains 50 records:
32 owner-resolved and 18 unresolved. Sixteen unresolved records predate the
existing 48-record engineering classification. Two new records require fresh
assessment:

- `baca2415`: multi-image AI entry diagnostics and private attachment cleanup.
- `c86b669d`: render AI-extracted transactions as a compact list of the same
  editable transaction-card contract used for one transaction.

Audit with MobileNext `mobilecli` on the dedicated iPhone 16 simulator, using
the current candidate build and stable accessibility identifiers. Physical
iPhone-only Face ID, APNs and offline interruption remain explicit external
gates if the paired phone is unavailable. Do not treat an old installed build
or a synthetic fixture as release acceptance.

## Technical integration plan

1. Reconcile and commit every intentional dirty/untracked candidate change;
   exclude only generated or secret material.
2. Run backend normal/race/vet, native unit/base/project/build gates and the
   feedback traceability checks.
3. Build/install the candidate on the dedicated simulator and audit unresolved
   feedback with `mobilecli`; record pass, partial or external-only evidence.
4. Reconcile `CHANGELOG.md`, `CONTEXT.md`, `BACKLOG.md`,
   `VALIDATION_MATRIX.md` and this note with actual results.
5. Merge the named candidate into `main` with a merge commit, rerun release
   gates on the merged tree, then remove only the completed temporary
   worktree/branch when Git proves no unique uncommitted files remain.

## Verification required before completion

- Backend: `go test ./...`, `go test -race ./...`, `go vet ./...`.
- Native: full `MyPocketIOSTests`, shared-base/project guards and Release
  simulator build with warnings as errors.
- MCP: exact SDK/catalog/route/documentation tests already in the candidate,
  plus merged-tree backend verification.
- Feedback: authenticated queue count, current unresolved classification and
  Mobile MCP accessibility evidence.
- Git: clean candidate commit, non-destructive merge, merged-tree checks and no
  lost commits from either side.

## Documentation impact

- Human/public: this release note and `CHANGELOG.md`.
- Agent-readable: `CONTEXT.md`, `BACKLOG.md`, `VALIDATION_MATRIX.md`, feedback
  audit note and existing OpenAPI/MCP contracts.
- No ERD or new ADR is planned unless implementation reveals a schema or
  durable architecture change.

## Reconciled candidate result

- Native metadata: `1.2.0 (3)`.
- Backend normal/race/vet, native 374-unit gate, base/project guards and
  warnings-as-errors simulator build passed on the candidate tree.
- MobileNext inspected the 20-attachment and proposal-card fixtures. It proved
  current behavior but also confirmed two residuals: cleanup is an unscheduled
  operator command and proposal cards are not the canonical transaction card.
- The initial focused UI run exposed one stale assertion that contradicted the
  approved Account-root Add contract and the existing dedicated regression.
  After correction, the merged-tree focused run passed: 16 selected cases,
  11 executed successfully and five fixture-dependent cases skipped by their
  declared environment guards.
- Production API/worker deployment completed on 2026-10-20 as immutable image
  `release-20261020-a2de756` (Swarm digest
  `sha256:5b54465e743f230f416172e87646548110d8b5e0470858c94167b5aaca1f3b45`).
  The API-only web service remains scaled to `0/0`.
- Migration advanced from version 37 to version 40 with `dirty=false`. The
  owner explicitly approved skipping the usual backup; no financial data was
  reset or deleted.
- A dedicated production UAT account was created outside the primary account.
  Credentials are not stored in Git. Authenticated GET-only regression passed;
  no feedback status or financial record was mutated.
- Release `1.2.0` is installed on the physical iPhone. MobileNext logged into
  the isolated production UAT account and traversed Overview, Transactions,
  Account and the floating feedback entry point. Face ID-specific acceptance
  and offline fault injection remain separate gates; password login and the
  fixed field layout passed.

## Main integration

- Merge commit: `c9ea660089ca24ec5ffff6ee68719f3628568bd1`.
- Both branch histories were preserved with a non-fast-forward merge; no reset,
  rebase or history rewrite was used.
- The clean legacy worktree branch `codex/web-gap-implementation` had 11 unique
  pre-native migration commits. It was joined through archival merge
  `ef8a08db26496635c0e21011180f65e2ee2c218a` using the current `main` tree,
  because replaying its obsolete `app → web` move and Google-auth backend would
  regress the native-only release. Its history remains reachable without
  reintroducing superseded runtime code.
- Merged-tree backend normal/race/vet/module gates passed.
- Merged-tree native 374/374 unit tests, base/project guards, focused UI suite,
  warnings-as-errors Release build and artifact preflight passed.

