#!/bin/sh
set -eu
test -n "${DATABASE_URL:-}"
test -n "${DATA_ROOT:-}"
test -n "${BACKUP_ROOT:-}"
snapshot="${BACKUP_ROOT}/$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "${snapshot}"
pg_dump --format=custom --file="${snapshot}/finance-assistant.pgdump" "${DATABASE_URL}"
tar -C "${DATA_ROOT}" -czf "${snapshot}/workspaces.tar.gz" workspaces
sha256sum "${snapshot}/finance-assistant.pgdump" "${snapshot}/workspaces.tar.gz" > "${snapshot}/SHA256SUMS"
chmod -R go-rwx "${snapshot}"
printf '%s\n' "${snapshot}"
