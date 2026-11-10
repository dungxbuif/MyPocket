#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
output="$(EXPECTED_MIGRATION_VERSION=000054 "${repo_root}/scripts/mypocket-rollout-preflight.sh" 2>&1)"
printf '%s\n' "${output}" | grep -Fq 'PASS: migration files are ordered through 000054'
printf '%s\n' "${output}" | grep -Fq '035→…→054 is the required forward path'
printf 'PASS: rollout preflight targets migration 000054\n'
