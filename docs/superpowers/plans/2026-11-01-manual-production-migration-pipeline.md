# Manual Production Migration Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a protected, manual MyPocket production migration job that pulls an immutable image from `registry.dungxbuif.com`, runs the existing migration binary, and verifies the expected schema without changing the normal CI/CD release flow.

**Architecture:** Keep the existing application delivery path intact and add a small included GitLab CI job for an explicit production migration action. The job runs on the homelab deploy runner, receives production-scoped variables from the MyPocket project, uses the exact image reference selected for the release, and performs pre/post migration version checks. The backend keeps a single local command wrapper for the same migration entry point.

**Tech Stack:** Go, golang-migrate, Docker CLI, GitLab CI, Bash contract tests, iOS/Xcode source already present in the repository.

**Spec:** User request in conversation: create production-scoped MyPocket GitLab variables, add a manual migration pipeline while preserving normal CI/CD, verify the combined code, tag the release, and push mobile source to GitHub.

## Global Constraints

- All production images use `registry.dungxbuif.com`.
- The migration job is manual, protected, and scoped to the `production` environment.
- The migration job must use an immutable image reference, never a mutable `latest` tag.
- Normal CI/CD remains enabled through the existing GitLab Auto DevOps include.
- Runtime credentials are never committed or printed in logs.
- The expected latest migration for this release is `000053`, with `dirty=false` after migration.
- The final verification is performed once after all implementation changes, as requested by the owner.

## Review Focus

- A manual job accidentally runs on a feature branch or unprotected ref.
- A mutable image tag causes migration to run against a different image than the release.
- A dirty migration state is allowed to continue.
- The project variables are created with the wrong environment scope or without protection/masking.
- The normal CI/CD flow is replaced instead of preserved.

### Task 1: Backend migration command and migration guard

**Files:**
- Create: `backend/Makefile`
- Modify: `scripts/mypocket-rollout-preflight.sh`
- Modify: `scripts/mypocket-rollout-preflight.test.sh`
- Modify: `backend/migrations/migrations_test.go`

**Interfaces:**
- Produces `make -C backend migrate` and `make -C backend migrate-version` wrappers around the existing `cmd/migrate` binary.
- Updates the repository migration contract from `000052` to `000053`.

- [x] Add the Make targets and update all preflight/test expectations to include migration `000053_report_analysis_restore`.
- [x] Keep the existing `cmd/migrate` implementation as the single migration engine.

### Task 2: Manual GitLab migration pipeline

**Files:**
- Create: `.gitlab-ci.yml`
- Create: `.gitlab/ci/migrate-production.yml`
- Create: `scripts/mypocket-migrate-pipeline.test.sh`

**Interfaces:**
- Consumes production-scoped variables from the `dungxbuif/MyPocket` project: image reference, registry credentials when required, migration network, and `DATABASE_URL`.
- Produces manual job `migrate:production` on tag `homelab-vm100-deploy` with `resource_group: mypocket-production-db`.

- [x] Include Auto DevOps so the existing normal CI/CD remains available.
- [x] Add a protected manual production job that pulls the selected immutable image, checks the current migration version, runs `/app/migrate up`, and verifies version `53 dirty: false`.
- [x] Fail closed on missing image reference, missing database URL, dirty pre-state, pull failure, or wrong post-state.
- [x] Add a shell contract test for the job's manual/protected/ref/image/version guards.

### Task 3: GitLab project variables and release verification

**Files:**
- No secret values committed.
- Update release metadata only if the final tag/version decision requires it.

- [x] Create the MyPocket project variables with environment scope `production`; protect and mask secret values.
- [x] Exclude seed/test/local-only values from production variables.
- [x] Run the combined backend, migration contract, pipeline contract, and iOS/source verification requested by the owner; the touched migration package and command pass, while the full suite has unrelated pre-existing failures recorded in the validation matrix.
- [x] Create the agreed release tag and push the mobile source/tag to GitHub; push the repository release to GitLab as already requested for this project.

## Execution evidence — 2026-11-02

- Pipeline `#764` built and published a multi-platform image from commit
  `d2e08bfe7247cf0078eff9e93b610e646044729d`.
- Manual job `migrate:production #2546` ran on `homelab-vm100-deploy` and
  pulled immutable digest
  `sha256:2709d78bab378ff9931381e92d89d168cf9b0da906b6d0ff675163c43e8767ec`.
- The migration readback changed from `version: 52 dirty: false` to
  `version: 53 dirty: false`; the job ended with `Job succeeded`.
- This execution record covers image publication and database migration. A
  Swarm API/worker service re-rollout is intentionally not inferred from the
  migration job and remains a separate operational action.

