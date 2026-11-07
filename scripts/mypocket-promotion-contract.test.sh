#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
ci_file="$repo_root/.gitlab-ci.yml"

grep -Fq 'https://gitlab.dungxbuif.com/dungxbuif/homelab-ops.git' "$ci_file"
grep -Fq 'git push -o ci.skip origin HEAD:main' "$ci_file"
grep -Fq 'git commit -m "chore(release): promote mypocket ${CI_COMMIT_TAG#prod-}"' "$ci_file"
grep -Fq 'git push origin "${CI_COMMIT_TAG/prod-/mypocket-prod-}"' "$ci_file"

if grep -F 'git commit -m "chore(release): promote mypocket' "$ci_file" | grep -Fq '[skip ci]'; then
  echo 'promotion commit must remain eligible for its receipt-tag pipeline' >&2
  exit 1
fi

echo 'MyPocket promotion contract: ok'
