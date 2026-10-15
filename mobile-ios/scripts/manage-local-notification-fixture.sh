#!/bin/sh
set -eu

mode="${1:-prepare}"
case "$mode" in
  prepare|cleanup) ;;
  *) printf 'usage: %s [prepare|cleanup]\n' "$0" >&2; exit 2 ;;
esac

: "${DATABASE_URL:?set DATABASE_URL to the local mypocket_stage_v1 database}"

case "$DATABASE_URL" in
  postgres://*@127.0.0.1:5432/mypocket_stage_v1*|postgresql://*@127.0.0.1:5432/mypocket_stage_v1*) ;;
  *) printf 'refusing notification fixture operation: DATABASE_URL must target 127.0.0.1:5432/mypocket_stage_v1\n' >&2; exit 1 ;;
esac

database_name="$(psql "$DATABASE_URL" -X -qAt -c 'SELECT current_database();')"
if [ "$database_name" != 'mypocket_stage_v1' ]; then
  printf 'refusing notification fixture operation outside local mypocket_stage_v1 (got %s)\n' "$database_name" >&2
  exit 1
fi

fixture_id='native-ui-notification-inbox-fixture-20261008'
owner_id='69427fb4-bf5a-43cc-b41f-f8035ec7c01f'

if [ "$mode" = cleanup ]; then
  psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -qAt \
    -v fixture_id="$fixture_id" -v owner_id="$owner_id" <<'SQL'
DELETE FROM notifications
WHERE id = :'fixture_id'
  AND owner_id = :'owner_id'
  AND title = 'Native UI notification fixture';
SQL
  exit 0
fi

result="$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -qAt \
  -v fixture_id="$fixture_id" -v owner_id="$owner_id" <<'SQL'
INSERT INTO notifications (id, owner_id, kind, title, body, dedupe_key, read_at)
VALUES (
  :'fixture_id',
  :'owner_id',
  'budget',
  'Native UI notification fixture',
  'Synthetic unread notification for the native UI test.',
  :'fixture_id',
  NULL
)
ON CONFLICT (id) DO UPDATE SET
  kind = EXCLUDED.kind,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  dedupe_key = EXCLUDED.dedupe_key,
  read_at = NULL,
  created_at = now()
WHERE notifications.owner_id = EXCLUDED.owner_id
  AND notifications.title = EXCLUDED.title
RETURNING id;
SQL
)"

if [ "$result" != "$fixture_id" ]; then
  printf 'notification fixture was not inserted or reset safely\n' >&2
  exit 1
fi
printf 'prepared unread notification fixture %s\n' "$result"
