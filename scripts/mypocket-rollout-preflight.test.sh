#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
output="$("${repo_root}/scripts/mypocket-rollout-preflight.sh" 2>&1)"
printf '%s\n' "${output}" | grep -Fq 'PASS: migration files are ordered through 000041'
printf '%s\n' "${output}" | grep -Fq '035→036→037→038→039→040→041 is the required forward path'
printf 'PASS: rollout preflight targets migration 000041\n'
