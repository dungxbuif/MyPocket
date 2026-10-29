# First Release Atomic Design Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Turn the current working snapshot into a documented first-release `1.0.0` baseline and reorganize design documentation to mirror the runtime Atomic Design hierarchy without changing product behavior.

**Architecture:** Keep shared design contracts in `system`, primitives in `atoms`, reusable compositions in `molecules`, stateful sections in `organisms`, route shells in `templates`, and route contracts in `pages`. Preserve owner-provided screenshots/HTML as non-runtime evidence in `references`, with every moved artifact linked from an index. Rewrite the active changelog to describe only this baseline while leaving Git history untouched.

**Tech Stack:** Markdown/YAML front matter, Git, React/Vite design checks, existing Go/Node test suites, `npm run check:design`, `git diff --check`.

**Spec:** `docs/work/tickets/RELEASE-BASELINE-DETAIL_DESIGN.md`

## Global Constraints

- Do not rewrite Git history, delete tags, alter API/database/authentication, or modify secrets.
- Preserve `screen.png` and `code.html` visual references; they are evidence, never runtime assets.
- Delete only exact duplicate/orphaned documentation artifacts proven safe by inventory and link scans.
- Treat the current snapshot as release `1.0.0` dated `2026-09-24`; list unverified gates explicitly.
- All Markdown links must remain local and pass `npm run check:design`.
- Use `apply_patch` for file content edits and preserve unrelated user changes in the dirty worktree.

## Review Focus

- A relative link from a ticket or design README still points to the former `screens` or bundle path; add a link scan and run `npm run check:design` in Task 3.
- A visual-reference bundle has no stable destination or its README loses `artifact_source: visual_reference`; verify the complete bundle inventory in Task 2.
- The first-release changelog accidentally claims pending AI, deployment, or production gates are complete; compare every entry against `docs/work/VALIDATION_MATRIX.md` in Task 4.
- Staging current runtime work accidentally includes an API key, token, `.env`, or generated artifact; run a staged-file secret scan in Task 5.
- A stale root path (`docs/design/screens`, `design/`, or `refereces/`) survives in master docs; run the repository-wide reference scan in Task 3 and Task 4.

## File Map

- Create: `docs/design/organisms/README.md`, `docs/design/templates/README.md`, `docs/design/pages/README.md`, `docs/design/references/README.md`.
- Move: component specs from `docs/design/atoms` and `docs/design/molecules` only when their level changes; route specs from `docs/design/screens/**` to `docs/design/pages/**`.
- Move: visual-reference bundles currently at `docs/design/ch_n_v_wallet_selector`, `docs/design/th_m_v_add_wallet`, `docs/design/stitch_my_pocket`, and `docs/design/stitch_my_pocket 2` into named subdirectories under `docs/design/references/`.
- Modify: `docs/design/README.md`, `docs/design/INDEX.md`, `docs/design/system/SOURCE_EXTRACTION.md`, affected ticket links, `docs/CONTEXT.md`, `docs/README.md`, `docs/work/VALIDATION_MATRIX.md`, and `docs/releases/CHANGELOG.md`.
- Preserve: `app/src/atomic/**`, backend source/migrations, tests, and unrelated user changes unless verification finds a concrete release-blocking defect.

### Task 1: Capture the approved baseline and establish an inventory

**Files:**
- Modify: `docs/work/tickets/RELEASE-BASELINE-DETAIL_DESIGN.md` (approval/status metadata only).
- Create: `docs/work/RELEASE-BASELINE-INVENTORY.md`.
- Test: command output captured in the inventory note.

**Interfaces:**
- Consumes: current `git status --short`, `docs/design` file list, Markdown link list, and validation matrix.
- Produces: a frozen list of files in scope and explicit exclusions for all later tasks.

- [ ] **Step 1: Record the approved design state**

  Change `approval: pending` to `approval: approved`, set `status: in_progress`, and add the approval date to the detail design front matter. Do not change the design scope.

- [ ] **Step 2: Generate the inventory note**

  Create `docs/work/RELEASE-BASELINE-INVENTORY.md` containing the exact counts and paths from:

  ```bash
  rtk proxy git status --short
  rtk proxy rg --files docs/design | sort
  rtk proxy rg -n '\]\([^)]*(screens|stitch_my_pocket|ch_n_v_wallet_selector|th_m_v_add_wallet|refereces)' docs --glob '*.md'
  ```

  Mark every modified/untracked runtime file as “preserve unless independently verified”; mark each design artifact as “move/update/delete candidate”.

- [ ] **Step 3: Verify the inventory is non-destructive**

  Run `rtk proxy git diff --check` and confirm the inventory itself contains no secrets, API keys, binary contents, or generated output.

- [ ] **Step 4: Commit the inventory checkpoint**

  ```bash
  git add docs/work/tickets/RELEASE-BASELINE-DETAIL_DESIGN.md docs/work/RELEASE-BASELINE-INVENTORY.md
  git commit -m "docs: approve first release baseline scope"
  ```

### Task 2: Rebuild the Atomic Design documentation tree and preserve visual evidence

**Files:**
- Create: `docs/design/organisms/README.md`, `docs/design/templates/README.md`, `docs/design/pages/README.md`, `docs/design/references/README.md`.
- Move: `docs/design/screens/**` to `docs/design/pages/**`.
- Move: the four top-level visual-reference bundles into `docs/design/references/` with stable kebab-case directory names.
- Modify: moved README front matter and `docs/design/system/SOURCE_EXTRACTION.md` source paths.

**Interfaces:**
- Consumes: Task 1 inventory and `app/src/atomic/{atoms,molecules,organisms,templates,pages}`.
- Produces: a one-to-one documentation taxonomy with stable paths and preserved visual evidence.

- [ ] **Step 1: Create destination directories and landing pages**

  Add landing READMEs that state the ownership boundary and link to the existing system gateway. `pages/README.md` retains the existing screen-spec template content but names route contracts as pages.

- [ ] **Step 2: Move route contracts**

  Move `docs/design/screens/` to `docs/design/pages/` as a directory move, preserving each page README and front matter. Do not rename page slugs inside this task; only the taxonomy changes.

- [ ] **Step 3: Move visual references into evidence folders**

  Use these destinations, preserving each bundle’s `README.md`, `code.html`, and `screen.png`:

  ```text
  docs/design/references/wallet-selector/
  docs/design/references/add-wallet/
  docs/design/references/transaction-entry/
  docs/design/references/transaction-actions/
  docs/design/references/category-and-goal-flows/
  ```

  Preserve `artifact_source: visual_reference` in each README. If a bundle contains multiple source screens, keep its internal subdirectories intact below the stable destination.

- [ ] **Step 4: Update the extraction register**

  Replace old source paths in `docs/design/system/SOURCE_EXTRACTION.md` with the new `references/` paths and state that these files are evidence only, not runtime code.

- [ ] **Step 5: Check the tree and duplicates**

  Run `rtk proxy find docs/design -type f -print | sort` and `rtk proxy rg -n 'artifact_source: visual_reference' docs/design`. Confirm every visual bundle has exactly one destination and no empty legacy directory remains.

- [ ] **Step 6: Commit the tree move**

  ```bash
  git add docs/design
  git commit -m "docs: align design references with atomic hierarchy"
  ```

### Task 3: Repair design gateway, inventory, and repository links

**Files:**
- Modify: `docs/design/README.md`, `docs/design/INDEX.md`, `docs/design/pages/README.md`, affected page READMEs and tickets under `docs/work/tickets/`.
- Test: design link checker and repository-wide path scan.

**Interfaces:**
- Consumes: new paths from Task 2 and actual runtime component inventory.
- Produces: authoritative docs that no longer claim missing/stale bases as current or point to deleted paths.

- [ ] **Step 1: Update the gateway rules**

  Replace the Markdown-only statement with the approved policy: Markdown is the contract; declared `visual_reference` PNG/HTML bundles are retained evidence under `references/`.

- [ ] **Step 2: Reconcile the component inventory**

  Update `docs/design/INDEX.md` entries using `app/src/atomic` as the source of truth. Keep “partial/planned” only when the implementation or validation matrix proves that state. Add links to `organisms`, `templates`, and `pages` landing pages.

- [ ] **Step 3: Rewrite moved relative links**

  Replace every `screens/` and old bundle path in `docs`, `app`, and backend Markdown with its new path. Do not rewrite historical prose that explicitly describes Git history unless it contains a live link.

- [ ] **Step 4: Run the design guardrail**

  ```bash
  cd app
  npm run check:design
  ```

  Expected: exit 0 with no missing local links or stale design path errors.

- [ ] **Step 5: Run the stale-path scan**

  ```bash
  rtk proxy rg -n 'docs/design/screens|docs/design/(stitch_my_pocket|ch_n_v_wallet_selector|th_m_v_add_wallet)|docs/design/INDEX\.md.*screens|refereces/' . --glob '!node_modules/**' --glob '!app/dist/**' --glob '!docs/superpowers/plans/**' --glob '!docs/work/RELEASE-BASELINE-INVENTORY.md'
  ```

  Expected: no live link remains; any historical mention is documented in the inventory note.

- [ ] **Step 6: Commit documentation reconciliation**

  ```bash
  git add docs/design docs/work/tickets docs/architecture docs/README.md docs/CONTEXT.md
  git commit -m "docs: reconcile atomic design contracts and links"
  ```

### Task 4: Reset the active changelog to the first release baseline

**Files:**
- Modify: `docs/releases/CHANGELOG.md`, `docs/releases/README.md`.
- Modify: `docs/CONTEXT.md`, `docs/work/VALIDATION_MATRIX.md` only for release-baseline links/status.
- Test: release-entry consistency scan.

**Interfaces:**
- Consumes: verified current behavior from the working tree and validation matrix.
- Produces: one active `1.0.0` release entry with no stale prior version sections presented as current history.

- [ ] **Step 1: Build the release entry from evidence**

  Replace the active changelog body with `## [1.0.0] - 2026-09-24` and sections `Added`, `Changed`, `Fixed`, `Documentation`, and `Known release gates`. Include only behavior supported by tests/docs; explicitly mark production deployment, configured-provider browser proof, and any pending UAT as gates.

- [ ] **Step 2: Remove historical active sections without rewriting Git**

  Delete the old `[Unreleased]`, `[1.1.1]`, `[1.1.0]`, and `[1.0.0]` prose from the working changelog. Add one sentence to the README that prior development history remains available in Git commits and this file is the first-release baseline.

- [ ] **Step 3: Reconcile context and validation links**

  Update only release/version wording and moved design paths in `docs/CONTEXT.md` and `docs/work/VALIDATION_MATRIX.md`; do not mark incomplete workflows complete.

- [ ] **Step 4: Verify claims**

  Run:

  ```bash
  rtk proxy rg -n '^## \[|Known release gates|production|pending|not run' docs/releases/CHANGELOG.md docs/work/VALIDATION_MATRIX.md
  rtk proxy git diff --check
  ```

  Confirm every known gate in the changelog appears as pending/not run in the validation matrix where applicable.

- [ ] **Step 5: Commit the release docs**

  ```bash
  git add docs/releases docs/CONTEXT.md docs/work/VALIDATION_MATRIX.md
  git commit -m "docs: establish first release baseline"
  ```

### Task 5: Verify the full snapshot and create the final release commit

**Files:**
- Modify: only files proven necessary by verification failures.
- Test: frontend, backend, docs, and staged secret scans.

**Interfaces:**
- Consumes: commits from Tasks 1–4 and the remaining user-owned runtime changes.
- Produces: a clean, reviewable first-release snapshot with no accidental secrets.

- [ ] **Step 1: Run frontend verification**

  ```bash
  cd app
  npm run lint
  npm run build
  npm run check:design
  npm test -- --runInBand
  ```

  Record failures in the detail design; fix only regressions caused by the cleanup.

- [ ] **Step 2: Run backend verification**

  ```bash
  cd backend
  go test ./...
  go vet ./...
  ```

  Confirm migration files remain untouched by documentation moves and no environment secret is tracked.

- [ ] **Step 3: Audit the final diff**

  ```bash
  rtk proxy git status --short
  rtk proxy git diff --check
  rtk proxy git diff --stat
  rtk proxy git diff --name-only --cached
  rtk proxy rg -n 'sk-[A-Za-z0-9]|AIza|BEGIN PRIVATE KEY|Bearer [A-Za-z0-9._-]{20,}' --glob '!*.lock' --glob '!docs/work/RELEASE-BASELINE-INVENTORY.md' .
  ```

  Exclude local `.env` files and any unrelated work from staging. If unrelated changes cannot be separated safely, stop before commit and report the exact paths.

- [ ] **Step 4: Update verification results**

  Fill the verification section of `docs/work/tickets/RELEASE-BASELINE-DETAIL_DESIGN.md` with exact commands and outcomes; update the validation matrix only with evidence from this run.

- [ ] **Step 5: Create the release-baseline commit**

  Stage only the approved cleanup/release files and any current runtime files explicitly confirmed as part of this first-release snapshot, then commit:

  ```bash
  git commit -m "chore: prepare first release baseline"
  ```

- [ ] **Step 6: Final status check**

  ```bash
  rtk proxy git status --short
  rtk proxy git log -3 --oneline
  ```

  Expected: no unintended staged files, the release baseline commit is visible, and any remaining dirty files are listed explicitly for the user.

## Self-review

- Spec coverage: every scope item in `RELEASE-BASELINE-DETAIL_DESIGN.md` maps to Tasks 1–5.
- Placeholder scan: no `TBD`, `TODO`, or unspecified implementation step remains in this plan.
- Type/path consistency: all later tasks consume the `pages` and `references` destinations created in Task 2.
- Review focus: each identified failure mode has a concrete inventory, link, claim, or staged-secret check.
