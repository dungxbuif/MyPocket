---
artifact_type: detail_design
id: RELEASE-BASELINE-01
status: in_progress
owner: ai
approval: approved
human_fields:
  - approval
  - constraints
  - scope_decisions
ai_fields:
  - problem
  - context_loaded
  - brownfield_scope
  - proposed_approach
  - design_tradeoffs
  - architecture_overview
  - execution_flow
  - test_plan
  - reconciliation_plan
shared_fields:
  - status
  - trace
trace:
  backlog_item: docs/work/BACKLOG.md
  requirement: release baseline and design contract hygiene
  phase: release-baseline
  ticket_or_bug: RELEASE-BASELINE-01
  test_verification: docs/work/VALIDATION_MATRIX.md
  validation_matrix: docs/work/VALIDATION_MATRIX.md
  docs_review: docs/design/README.md
  adrs: []
  master_docs_touched:
    - docs/design/README.md
    - docs/design/INDEX.md
    - docs/releases/CHANGELOG.md
    - docs/CONTEXT.md
---

# DETAIL DESIGN: First release baseline and Atomic Design documentation cleanup

## Status

- ID: RELEASE-BASELINE-01
- Status: in_progress
- Approval: approved
- Author: Codex
- Updated: 2026-09-24; owner approval recorded

## 1. Context & Scope

### Problem statement

The repository contains an active, uncommitted Stage v1 implementation together with
historical release notes and a mixed `docs/design` layout. The current design index
contains stale implementation status, while screen specifications and visual reference
bundles are not organized by the Atomic Design levels used by `app/src/atomic`.

### Goal

Create a clean first-release baseline from the current working snapshot, without
rewriting Git history, and make the design documentation mirror the runtime hierarchy:

```text
docs/design/
  system/
  atoms/
  molecules/
  organisms/
  templates/
  pages/
  references/
```

The release documentation will describe this snapshot as `1.0.0` dated 2026-09-24.
Old release sections are not carried forward as active release history, but prior Git
commits remain available for archaeology. No tag, deploy, migration reset, or database
destructive action is part of this change.

### Scope

- Reorganize `docs/design` Markdown specifications to the Atomic levels.
- Move declared owner visual references (`screen.png`, `code.html`) under
  `docs/design/references/` with stable source metadata and preserve them as evidence.
- Update relative links, indexes, design contracts, and validation references.
- Replace the active changelog with a single first-release `1.0.0` entry that records
  verified behavior and explicitly lists open release gates.
- Remove only exact duplicate, orphaned, or superseded documentation artifacts found
  by link/inventory checks. Do not delete source code, migrations, screenshots, or work
  tickets merely because they are historical.
- Commit the resulting repository snapshot using conventional commits after tests pass.

### Out of scope

- Rewriting Git history or deleting prior commits/tags.
- Changing product behavior, API contracts, database schema, secrets, deployment, or
  authentication.
- Claiming production readiness for features whose existing validation matrix marks them
  pending.

## 2. Design decisions and trade-offs

| Alternative | Benefit | Risk | Decision |
| --- | --- | --- | --- |
| Delete all HTML/PNG references after Markdown extraction | Smaller repository | Loses owner evidence and breaks traceability for linked specs | Reject |
| Keep bundles at the current top level | No link churn | Does not mirror Atomic Design and leaves the taxonomy ambiguous | Reject |
| Move specs to `atoms/molecules/organisms/templates/pages` and references to a dedicated evidence tree | Mirrors runtime, keeps evidence, makes links explicit | Requires link migration and verification | Choose |
| Squash or delete database migrations as part of the release cleanup | Shorter migration directory | Changes runtime/data safety and exceeds documentation scope | Reject |
| Preserve all historical changelog sections | Complete chronology in one file | Conflicts with the requested first-release baseline | Reject; history remains in Git |

## 3. Architecture and ownership

| Area | Responsibility |
| --- | --- |
| `docs/design/system` | Shared tokens, base contracts, behavior, and extraction register |
| `docs/design/atoms` | Single-purpose visual primitives and controls |
| `docs/design/molecules` | Reusable compositions of atoms |
| `docs/design/organisms` | Screen-level reusable sections with state/fetching contracts |
| `docs/design/templates` | Page layout shells and composition templates |
| `docs/design/pages` | Route/page contracts formerly stored under `screens` |
| `docs/design/references` | Owner-provided visual evidence; never runtime code |
| `docs/releases/CHANGELOG.md` | Current first-release facts and open gates |

Runtime code remains under `app/src/atomic/{atoms,molecules,organisms,templates,pages}`.
The docs use `pages` for route contracts; the old `screens` directory is removed only
after every link and validation rule points to the new location.

## 4. Execution flow

1. Inventory all Markdown links and every visual-reference bundle.
2. Create the target directories and move each spec/bundle according to its declared
   Atomic level; preserve README content and add source metadata where missing.
3. Update relative links, `docs/design/INDEX.md`, the design gateway, and affected
   tickets/master docs.
4. Rebuild the first-release changelog from verified current behavior and open gates.
5. Run design-link checks, frontend tests/lint/build, backend tests, and a final secret
   and diff audit.
6. Commit the cleanup and release baseline atomically (or as two clearly linked commits
   if the final diff contains unrelated user work that cannot be safely separated).

## 5. API, data, security

No API, database, authentication, secret, or runtime data model changes are intended.
The release note must not include API keys, tokens, OCR payloads, user financial data,
or provider secrets. Existing uncommitted runtime changes are reviewed for accidental
secret staging before commit.

## 6. Verification plan

- Documentation: `npm run check:design` and a link scan from the repository root.
- Frontend: existing lint, typecheck, build, and focused design/ledger tests.
- Backend: existing Go unit/integration test commands and migration consistency checks.
- Release hygiene: `git diff --check`, secret-pattern scan on staged files, and a final
  `git status`/`git diff --cached` review.
- Reconciliation: update `docs/CONTEXT.md`, `docs/README.md`, design index/gateway,
  changelog, and validation matrix only where the moved paths or release status require it.

## 7. Verification results

- `git diff --check`: PASS.
- Static repository link scan excluding this plan/inventory and the rewritten changelog:
  PASS; no stale live `design/screens`, old visual bundle, or legacy reference-tree path
  remains.
- Visual evidence inventory: PASS; 12 `screen.png` and 12 `code.html` files are under
  `docs/design/references/`.
- Backend `go test ./...`: PASS.
- Backend `go vet ./...`: PASS.
- Frontend `npm run lint`, `npm run build`, `npm run check:design`, and `npm test`:
  NOT RUN successfully because this environment has no Node executable (`env: node: No
  such file or directory`). These commands remain a release gate in the changelog.
- Secret-pattern scan over the current diff: no matches.

## 8. Reconciliation

- Design gateway, inventory, system extraction register, page contracts, work tickets,
  backlog, validation references, context, release README, and changelog were updated.
- No API, database, authentication, deployment, or secret behavior was changed by this
  cleanup. Existing runtime/migration changes remain in the worktree for separate review.

## 9. Approval gate

The owner approved this design on 2026-09-24. The key scope decision was to preserve
owner-provided visual references under `docs/design/references/` and remove only
duplicate/orphaned artifacts proven safe by link and inventory checks.
