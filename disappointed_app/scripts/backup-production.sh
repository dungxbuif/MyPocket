#!/usr/bin/env bash
set -euo pipefail

: "${MYPOCKET_BACKUP_DIR:?Set an explicit MYPOCKET_BACKUP_DIR}"

SOURCE_DATABASE_URL="${MYPOCKET_DATABASE_URL:-${DATABASE_URL:-}}"
SOURCE_S3_ENDPOINT="${MYPOCKET_S3_ENDPOINT:-${S3_ENDPOINT:-}}"
SOURCE_S3_BUCKET="${MYPOCKET_S3_BUCKET:-${S3_BUCKET:-}}"
SOURCE_S3_REGION="${MYPOCKET_S3_REGION:-${S3_REGION:-us-east-1}}"
SOURCE_S3_ACCESS_KEY="${MYPOCKET_S3_ACCESS_KEY_ID:-${S3_ACCESS_KEY:-}}"
SOURCE_S3_SECRET_KEY="${MYPOCKET_S3_SECRET_ACCESS_KEY:-${S3_SECRET_KEY:-}}"

: "${SOURCE_DATABASE_URL:?Set MYPOCKET_DATABASE_URL or DATABASE_URL}"
: "${SOURCE_S3_ENDPOINT:?Set MYPOCKET_S3_ENDPOINT or S3_ENDPOINT}"
: "${SOURCE_S3_BUCKET:?Set MYPOCKET_S3_BUCKET or S3_BUCKET}"
: "${SOURCE_S3_ACCESS_KEY:?Set MYPOCKET_S3_ACCESS_KEY_ID or S3_ACCESS_KEY}"
: "${SOURCE_S3_SECRET_KEY:?Set MYPOCKET_S3_SECRET_ACCESS_KEY or S3_SECRET_KEY}"

for command_name in pg_dump psql aws jq shasum; do
  command -v "$command_name" >/dev/null || { echo "Missing required command: $command_name" >&2; exit 1; }
done

umask 077
mkdir -p "$MYPOCKET_BACKUP_DIR"
chmod 700 "$MYPOCKET_BACKUP_DIR"
if [[ -e "$MYPOCKET_BACKUP_DIR/database.dump" || -e "$MYPOCKET_BACKUP_DIR/objects" ]]; then
  echo "Backup destination already contains MyPocket artifacts: $MYPOCKET_BACKUP_DIR" >&2
  exit 1
fi
mkdir "$MYPOCKET_BACKUP_DIR/objects"

dump_version="$(pg_dump --version | awk '{print $NF}')"
dump_major="${dump_version%%.*}"
server_version="$(psql "$SOURCE_DATABASE_URL" -At -c 'SHOW server_version')"
server_major="${server_version%%.*}"
if [[ "$dump_major" != "$server_major" ]]; then
  echo "pg_dump major ($dump_major) must match PostgreSQL server major ($server_major) for a release backup" >&2
  exit 1
fi
printf 'pg_dump=%s\nsource_server=%s\n' "$dump_version" "$server_version" > "$MYPOCKET_BACKUP_DIR/database-tool-versions.txt"

pg_dump --format=custom --no-owner --file "$MYPOCKET_BACKUP_DIR/database.dump" "$SOURCE_DATABASE_URL"

psql "$SOURCE_DATABASE_URL" -At -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name" |
while IFS= read -r table_name; do
  [[ "$table_name" =~ ^[a-zA-Z0-9_]+$ ]] || { echo "Unsafe table name in catalog" >&2; exit 1; }
  psql "$SOURCE_DATABASE_URL" -At -F $'\t' -c "SELECT '$table_name', count(*) FROM \"$table_name\""
done > "$MYPOCKET_BACKUP_DIR/database-row-counts.tsv"

AWS_ACCESS_KEY_ID="$SOURCE_S3_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$SOURCE_S3_SECRET_KEY" AWS_DEFAULT_REGION="$SOURCE_S3_REGION" \
  aws --endpoint-url "$SOURCE_S3_ENDPOINT" s3api list-objects-v2 --bucket "$SOURCE_S3_BUCKET" --output json |
  jq -S '{objects: [(.Contents // [])[] | {Key, Size, ETag, ChecksumAlgorithm, LastModified}]}' > "$MYPOCKET_BACKUP_DIR/s3-inventory.json"
AWS_ACCESS_KEY_ID="$SOURCE_S3_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$SOURCE_S3_SECRET_KEY" AWS_DEFAULT_REGION="$SOURCE_S3_REGION" \
  aws --endpoint-url "$SOURCE_S3_ENDPOINT" s3 sync "s3://$SOURCE_S3_BUCKET" "$MYPOCKET_BACKUP_DIR/objects" --only-show-errors

(
  cd "$MYPOCKET_BACKUP_DIR"
  while IFS= read -r -d '' object_file; do
    shasum -a 256 "$object_file"
  done < <(find objects -type f -print0 | LC_ALL=C sort -z) > objects.sha256
  shasum -a 256 database.dump database-row-counts.tsv database-tool-versions.txt s3-inventory.json objects.sha256 > backup-files.sha256
)
chmod 600 "$MYPOCKET_BACKUP_DIR"/database.dump "$MYPOCKET_BACKUP_DIR"/*.tsv "$MYPOCKET_BACKUP_DIR"/*.txt "$MYPOCKET_BACKUP_DIR"/*.json "$MYPOCKET_BACKUP_DIR"/*.sha256

echo "Backup complete: $MYPOCKET_BACKUP_DIR"
