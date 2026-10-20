#!/usr/bin/env bash
set -euo pipefail

# Read-only release preflight. This script never applies migrations, updates a
# service, or writes to a database. It is safe to run from a checkout before an
# operator executes the commands printed at the end.

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
migrations_dir="${MIGRATIONS_DIR:-${repo_root}/backend/migrations}"
expected_latest="${EXPECTED_MIGRATION_VERSION:-000041}"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

[[ -d "${migrations_dir}" ]] || fail "migration directory not found: ${migrations_dir}"
command -v awk >/dev/null 2>&1 || fail "awk is required"
command -v sort >/dev/null 2>&1 || fail "sort is required"

migration_versions=()
while IFS= read -r version; do
  migration_versions+=("${version}")
done < <(
  find "${migrations_dir}" -maxdepth 1 -type f -name '*.up.sql' -print \
    | sed -nE 's#^.*/([0-9]{6})_[a-z0-9_]+\.up\.sql$#\1#p' \
    | sort -n
)
[[ "${#migration_versions[@]}" -gt 0 ]] || fail "no forward migrations found"

previous=''
for version in "${migration_versions[@]}"; do
  if [[ -n "${previous}" && "${version}" < "${previous}" || "${version}" == "${previous}" ]]; then
    fail "migration versions are not strictly increasing"
  fi
  previous="${version}"
done

latest="${migration_versions[${#migration_versions[@]}-1]}"
[[ "${latest}" == "${expected_latest}" ]] || fail "latest migration ${latest}, expected ${expected_latest}"

for version in 000035 000036 000037 000038 000039 000040 000041; do
  [[ "$(printf '%s\n' "${migration_versions[@]}" | grep -Fx "${version}" || true)" == "${version}" ]] \
    || fail "required migration ${version} is missing"
done

for pair in \
  "000036_push_tokens.up.sql 000036_push_tokens.down.sql" \
  "000037_audit_outbox.up.sql 000037_audit_outbox.down.sql" \
  "000038_portfolio_trade_corrections.up.sql 000038_portfolio_trade_corrections.down.sql" \
  "000039_push_deliveries.up.sql 000039_push_deliveries.down.sql" \
  "000040_row_versions.up.sql 000040_row_versions.down.sql" \
  "000041_advisor_runtime.up.sql 000041_advisor_runtime.down.sql"; do
  read -r up down <<<"${pair}"
  [[ -f "${migrations_dir}/${up}" ]] || fail "missing ${up}"
  [[ -f "${migrations_dir}/${down}" ]] || fail "missing ${down}"
done

[[ -f "${repo_root}/backend/Dockerfile" ]] || fail 'backend/Dockerfile is missing'
grep -Fq 'COPY --from=build /out/migrate /app/migrate' "${repo_root}/backend/Dockerfile" \
  || fail 'API image does not contain /app/migrate'
grep -Fq 'COPY --from=build /out/worker /app/worker' "${repo_root}/backend/Dockerfile" \
  || fail 'API image does not contain /app/worker'

if [[ -n "${BACKUP_FILE:-}" ]]; then
  [[ -f "${BACKUP_FILE}" ]] || fail "backup file not found: ${BACKUP_FILE}"
  command -v pg_restore >/dev/null 2>&1 || fail 'pg_restore is required to validate BACKUP_FILE'
  pg_restore --list "${BACKUP_FILE}" >/dev/null || fail 'pg_restore could not read BACKUP_FILE'
  printf 'PASS: backup archive is readable (contents not printed)\n'
else
  printf 'WARN: BACKUP_FILE is not set; production rollout remains blocked\n'
fi

if [[ "${CHECK_DATABASE_VERSION:-0}" == "1" ]]; then
  [[ -n "${DATABASE_URL:-}" ]] || fail 'CHECK_DATABASE_VERSION=1 requires DATABASE_URL'
  command -v go >/dev/null 2>&1 || fail 'go is required for database version check'
  printf 'Read-only migration version check:\n'
  (cd "${repo_root}/backend" && MIGRATIONS_PATH="${migrations_dir}" go run ./cmd/migrate version)
else
  printf 'SKIP: database version check (set CHECK_DATABASE_VERSION=1 explicitly)\n'
fi

printf 'PASS: migration files are ordered through %s; 035→036→037→038→039→040→041 is the required forward path\n' "${expected_latest}"
printf '\nOperator commands (not executed by this script):\n'
printf '  cd <release-image-shell> && /app/migrate version\n'
printf '  cd <release-image-shell> && /app/migrate up\n'
printf '  curl -fsS https://money.dungxbuif.com/api/v1/health\n'
printf '  docker service ps mypocket-api --no-trunc\n'
printf '  docker service ps mypocket-worker --no-trunc\n'
printf '  docker service logs --since 10m mypocket-worker\n'
printf '  /app/worker with WORKER_ONCE=true (staging or approved window only)\n'
