#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ci_file="${repo_root}/.gitlab-ci.yml"
job_file="${repo_root}/.gitlab/ci/migrate-production.yml"

[[ -f "${ci_file}" ]] || { echo 'FAIL: .gitlab-ci.yml missing' >&2; exit 1; }
[[ -f "${job_file}" ]] || { echo 'FAIL: migration job file missing' >&2; exit 1; }

grep -Fq 'CI_PIPELINE_SOURCE == "web"' "${ci_file}"
grep -Fq 'Auto-DevOps.gitlab-ci.yml' "${ci_file}"
grep -Fq '.gitlab/ci/migrate-production.yml' "${ci_file}"
grep -Fq 'homelab-mac' "${ci_file}"
grep -Eq '^build:' "${ci_file}"
grep -Fq 'backend/Dockerfile' "${ci_file}"
grep -Fq 'docker buildx build' "${ci_file}"
grep -Fq -- '--platform linux/amd64,linux/arm64' "${ci_file}"
grep -Fq -- '--push' "${ci_file}"
grep -Fq 'build-metadata.json' "${ci_file}"
grep -Fq 'BUILT_IMAGE_REF=' "${ci_file}"
grep -Fq 'migrate:production:' "${job_file}"
grep -Fq 'when: manual' "${job_file}"
grep -Fq 'allow_failure: true' "${job_file}"
grep -Fq '$CI_COMMIT_REF_PROTECTED == "true"' "${job_file}"
grep -Fq '    - job: build' "${job_file}"
grep -Fq '      artifacts: true' "${job_file}"
grep -Fq 'name: production' "${job_file}"
grep -Fq 'homelab-vm100-deploy' "${job_file}"
grep -Fq 'resource_group: mypocket-production-db' "${job_file}"
grep -Fq 'registry.dungxbuif.com' "${job_file}"
grep -Fq 'BUILT_IMAGE_REF' "${job_file}"
grep -Fq -- '--entrypoint /app/migrate' "${job_file}"
grep -Fq '"${image_ref}" up' "${job_file}"
grep -Fq 'version: ${MYPOCKET_EXPECTED_MIGRATION_VERSION:-53} dirty: false' "${job_file}"

echo 'PASS: MyPocket manual production migration pipeline contract'
