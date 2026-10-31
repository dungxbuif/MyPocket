#!/usr/bin/env bash
set -euo pipefail

# Read-only release preflight. This script never applies migrations, updates a
# service, or writes to a database. It is safe to run from a checkout before an
# operator executes the commands printed at the end.

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
migrations_dir="${MIGRATIONS_DIR:-${repo_root}/backend/migrations}"
expected_latest="${EXPECTED_MIGRATION_VERSION:-000052}"

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

for version in 000035 000036 000037 000038 000039 000040 000041 000042 000043 000044 000045 000046 000047 000048 000049 000050 000051 000052; do
  [[ "$(printf '%s\n' "${migration_versions[@]}" | grep -Fx "${version}" || true)" == "${version}" ]] \
    || fail "required migration ${version} is missing"
done

for pair in \
  "000036_push_tokens.up.sql 000036_push_tokens.down.sql" \
  "000037_audit_outbox.up.sql 000037_audit_outbox.down.sql" \
  "000038_portfolio_trade_corrections.up.sql 000038_portfolio_trade_corrections.down.sql" \
  "000039_push_deliveries.up.sql 000039_push_deliveries.down.sql" \
  "000040_row_versions.up.sql 000040_row_versions.down.sql" \
  "000041_advisor_runtime.up.sql 000041_advisor_runtime.down.sql" \
  "000042_advisor_entry_tasks.up.sql 000042_advisor_entry_tasks.down.sql" \
  "000043_asset_units.up.sql 000043_asset_units.down.sql" \
  "000044_asset_market.up.sql 000044_asset_market.down.sql" \
  "000045_market_poll_claims.up.sql 000045_market_poll_claims.down.sql" \
  "000046_advisor_entry_input.up.sql 000046_advisor_entry_input.down.sql" \
  "000047_report_snapshots.up.sql 000047_report_snapshots.down.sql" \
  "000048_report_analyses.up.sql 000048_report_analyses.down.sql" \
  "000049_gold_market_catalog.up.sql 000049_gold_market_catalog.down.sql" \
  "000050_report_revision_delete_guard.up.sql 000050_report_revision_delete_guard.down.sql" \
  "000051_ai_entry_proposal_lifecycle.up.sql 000051_ai_entry_proposal_lifecycle.down.sql" \
  "000052_ai_entry_proposal_receipt_check.up.sql 000052_ai_entry_proposal_receipt_check.down.sql"; do
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

printf 'PASS: migration files are ordered through %s; 035→…→052 is the required forward path\n' "${expected_latest}"
printf '\nOperator commands (not executed by this script):\n'
printf '  cd <release-image-shell> && /app/migrate version\n'
printf '  cd <release-image-shell> && /app/migrate up\n'
printf '  curl -fsS https://money.dungxbuif.com/api/v1/health\n'
printf '  docker service ps mypocket-api --no-trunc\n'
printf '  docker service ps mypocket-worker --no-trunc\n'
printf '  docker service logs --since 10m mypocket-worker\n'
printf '  /app/worker with WORKER_ONCE=true (staging or approved window only)\n'
