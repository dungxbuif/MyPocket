---
artifact_type: release_baseline_inventory
id: RELEASE-BASELINE-INVENTORY
status: active
owner: shared
source_design: tickets/RELEASE-BASELINE-DETAIL_DESIGN.md
captured: 2026-09-24
---

# Release baseline inventory

This inventory freezes the scope before the Atomic Design documentation move. It is
evidence for the cleanup, not a claim that every dirty file belongs in the release.

## Worktree snapshot

- Branch: `feature/wallet-management`.
- The worktree already contains user-owned runtime, test, migration, AI, and docs
  changes. These remain preserved unless a separate verification identifies a concrete
  defect caused by the cleanup.
- The release cleanup itself is limited to documentation taxonomy, links, release notes,
  and verification records.
- Secrets and local environment files are never staged by this work.

## Design tree snapshot

- Design files before the move: 67.
- Visual evidence before the move: 12 `screen.png` files and 12 `code.html` files.
- Existing Atomic levels: `system`, `atoms`, `molecules`, and `screens`.
- Target route level: `pages` (renamed from `screens`).
- New levels: `organisms`, `templates`, and `references`.

### Existing specifications to preserve

- `system/`: `DESIGN.md`, `TOKENS.md`, `BASE_COMPONENTS.md`, `BEHAVIOR.md`,
  `SOURCE_EXTRACTION.md`.
- `atoms/`: `amount-keypad`, `base-cards`, `file-upload`, `long-press`.
- `molecules/`: `budget-meter`, `budget-progress-cards`, `category-tree`, `edit-group`,
  `transaction-form-rows`.
- `screens/`: route contracts, `current-ui.md`, and the screen template; move to
  `pages/` without changing page slugs.

### Visual evidence to preserve

- `ch_n_v_wallet_selector/` → `references/wallet-selector/`.
- `th_m_v_add_wallet/` → `references/add-wallet/`.
- `stitch_my_pocket/` → `references/transaction-flows/`.
- `stitch_my_pocket 2/` → `references/category-goal-flows/`.

Each bundle keeps its README, `screen.png`, and `code.html` and retains
`artifact_source: visual_reference` where present.

## Link classes requiring rewrite

- Direct route links under `docs/design/screens/...`.
- Relative references from system/atom/molecule READMEs to `../screens/...`.
- Ticket and validation links under `docs/work/` and `docs/decisions/`.
- Visual bundle links to `ch_n_v_wallet_selector`, `th_m_v_add_wallet`, and the two
  `stitch_my_pocket` directories.
- Historical prose that explicitly refers to an old Git commit is retained unless the
  prose contains a live local link.

## Exclusions

- No source code or migration is deleted or renamed.
- No API, database, deployment, auth, or secret changes are introduced.
- No old Git commit, tag, or release object is removed.
- No visual evidence is deleted merely because Markdown extraction exists.

## Verification commands captured

```bash
git status --short
find docs/design -type f | sort
rg -n '\]\([^)]*(screens|stitch_my_pocket|ch_n_v_wallet_selector|th_m_v_add_wallet|refereces)' docs --glob '*.md'
find docs/design -type f -name 'screen.png' | wc -l
find docs/design -type f -name 'code.html' | wc -l
```

The commands reported 49 tracked modifications, 10 untracked groups/files, 12 PNG
references, and 12 HTML references at capture time. The untracked plan/design files
created by this task are intentionally included in the documentation work and are
reviewed separately before the final release commit.
