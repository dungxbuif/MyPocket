#!/bin/sh
set -eu
test "$#" -eq 1
test -n "${RESTORE_DATABASE_URL:-}"
test -n "${RESTORE_DATA_ROOT:-}"
snapshot="$1"
(cd "${snapshot}" && sha256sum -c SHA256SUMS)
pg_restore --clean --if-exists --no-owner --dbname="${RESTORE_DATABASE_URL}" "${snapshot}/finance-assistant.pgdump"
mkdir -p "${RESTORE_DATA_ROOT}"
tar -C "${RESTORE_DATA_ROOT}" -xzf "${snapshot}/workspaces.tar.gz"
