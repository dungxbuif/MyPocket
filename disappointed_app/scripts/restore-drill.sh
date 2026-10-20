#!/usr/bin/env bash
set -euo pipefail

: "${MYPOCKET_BACKUP_DIR:?Set the backup directory to verify}"
: "${MYPOCKET_RESTORE_DATABASE_URL:?Set an explicit isolated restore database URL}"
: "${MYPOCKET_RESTORE_S3_PREFIX:?Set an explicit isolated S3 prefix}"

SOURCE_DATABASE_URL="${MYPOCKET_DATABASE_URL:-${DATABASE_URL:-}}"
RESTORE_S3_ENDPOINT="${MYPOCKET_S3_ENDPOINT:-${S3_ENDPOINT:-}}"
RESTORE_S3_BUCKET="${MYPOCKET_S3_BUCKET:-${S3_BUCKET:-}}"
RESTORE_S3_REGION="${MYPOCKET_S3_REGION:-${S3_REGION:-us-east-1}}"
RESTORE_S3_ACCESS_KEY="${MYPOCKET_S3_ACCESS_KEY_ID:-${S3_ACCESS_KEY:-}}"
RESTORE_S3_SECRET_KEY="${MYPOCKET_S3_SECRET_ACCESS_KEY:-${S3_SECRET_KEY:-}}"

: "${RESTORE_S3_ENDPOINT:?Set MYPOCKET_S3_ENDPOINT or S3_ENDPOINT}"
: "${RESTORE_S3_BUCKET:?Set MYPOCKET_S3_BUCKET or S3_BUCKET}"
: "${RESTORE_S3_ACCESS_KEY:?Set MYPOCKET_S3_ACCESS_KEY_ID or S3_ACCESS_KEY}"
: "${RESTORE_S3_SECRET_KEY:?Set MYPOCKET_S3_SECRET_ACCESS_KEY or S3_SECRET_KEY}"

if [[ -n "$SOURCE_DATABASE_URL" && "$MYPOCKET_RESTORE_DATABASE_URL" == "$SOURCE_DATABASE_URL" ]]; then
  echo "Refusing to restore into the source database" >&2
  exit 1
fi
case "$MYPOCKET_RESTORE_DATABASE_URL" in
  *"/mypocket_restore_"*) ;;
  *) echo "Restore database name must start with mypocket_restore_" >&2; exit 1 ;;
esac
case "$MYPOCKET_RESTORE_S3_PREFIX" in
  restore-drill/*) ;;
  *) echo "Restore S3 prefix must start with restore-drill/" >&2; exit 1 ;;
esac

for command_name in pg_restore psql aws shasum diff mktemp awk; do
  command -v "$command_name" >/dev/null || { echo "Missing required command: $command_name" >&2; exit 1; }
done
for artifact in database.dump database-row-counts.tsv database-tool-versions.txt objects.sha256 backup-files.sha256; do
  [[ -f "$MYPOCKET_BACKUP_DIR/$artifact" ]] || { echo "Missing backup artifact: $artifact" >&2; exit 1; }
done

dump_version="$(awk -F= '$1 == "pg_dump" { print $2 }' "$MYPOCKET_BACKUP_DIR/database-tool-versions.txt")"
dump_major="${dump_version%%.*}"
restore_version="$(pg_restore --version | awk '{print $NF}')"
restore_major="${restore_version%%.*}"
target_version="$(psql "$MYPOCKET_RESTORE_DATABASE_URL" -At -c 'SHOW server_version')"
target_major="${target_version%%.*}"
if [[ -z "$dump_major" || "$restore_major" != "$dump_major" || "$target_major" != "$dump_major" ]]; then
  echo "Backup, pg_restore, and target PostgreSQL majors must match" >&2
  exit 1
fi

umask 077
(
  cd "$MYPOCKET_BACKUP_DIR"
  shasum -a 256 -c backup-files.sha256
  if [[ -s objects.sha256 ]]; then
    shasum -a 256 -c objects.sha256
  fi
)

pg_restore --clean --if-exists --no-owner --no-privileges --dbname "$MYPOCKET_RESTORE_DATABASE_URL" "$MYPOCKET_BACKUP_DIR/database.dump"

actual_rows="$(mktemp)"
restored_objects="$(mktemp -d)"
actual_hashes="$(mktemp)"
cleanup() {
  rm -f "$actual_rows" "$actual_hashes"
  rm -rf "$restored_objects"
}
trap cleanup EXIT

psql "$MYPOCKET_RESTORE_DATABASE_URL" -At -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name" |
while IFS= read -r table_name; do
  [[ "$table_name" =~ ^[a-zA-Z0-9_]+$ ]] || { echo "Unsafe table name in restored catalog" >&2; exit 1; }
  psql "$MYPOCKET_RESTORE_DATABASE_URL" -At -F $'\t' -c "SELECT '$table_name', count(*) FROM \"$table_name\""
done > "$actual_rows"
diff -u "$MYPOCKET_BACKUP_DIR/database-row-counts.tsv" "$actual_rows"

AWS_ACCESS_KEY_ID="$RESTORE_S3_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$RESTORE_S3_SECRET_KEY" AWS_DEFAULT_REGION="$RESTORE_S3_REGION" \
  aws --endpoint-url "$RESTORE_S3_ENDPOINT" s3 sync "$MYPOCKET_BACKUP_DIR/objects" "s3://$RESTORE_S3_BUCKET/$MYPOCKET_RESTORE_S3_PREFIX/" --only-show-errors
AWS_ACCESS_KEY_ID="$RESTORE_S3_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$RESTORE_S3_SECRET_KEY" AWS_DEFAULT_REGION="$RESTORE_S3_REGION" \
  aws --endpoint-url "$RESTORE_S3_ENDPOINT" s3 sync "s3://$RESTORE_S3_BUCKET/$MYPOCKET_RESTORE_S3_PREFIX/" "$restored_objects/objects" --only-show-errors
(
  cd "$restored_objects"
  while IFS= read -r -d '' object_file; do
    shasum -a 256 "$object_file"
  done < <(find objects -type f -print0 | LC_ALL=C sort -z) > "$actual_hashes"
)
diff -u "$MYPOCKET_BACKUP_DIR/objects.sha256" "$actual_hashes"

database_name="${MYPOCKET_RESTORE_DATABASE_URL%%\?*}"
database_name="${database_name##*/}"
evidence="$MYPOCKET_BACKUP_DIR/restore-evidence.txt"
{
  echo "status=pass"
  echo "database=$database_name"
  echo "s3_bucket=$RESTORE_S3_BUCKET"
  echo "s3_prefix=$MYPOCKET_RESTORE_S3_PREFIX"
  echo "verified_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
} > "$evidence"
chmod 600 "$evidence"

echo "Restore drill passed; evidence: $evidence"
