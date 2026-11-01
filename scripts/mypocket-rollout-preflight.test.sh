#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
output="$(EXPECTED_MIGRATION_VERSION=000053 "${repo_root}/scripts/mypocket-rollout-preflight.sh" 2>&1)"
printf '%s\n' "${output}" | grep -Fq 'PASS: migration files are ordered through 000053'
printf '%s\n' "${output}" | grep -Fq '035→…→053 is the required forward path'
printf 'PASS: rollout preflight targets migration 000053\n'
