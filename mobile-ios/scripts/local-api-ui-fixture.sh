#!/usr/bin/env bash
set -Eeuo pipefail

# Disposable local-only API fixture for native LOCAL_API_UI tests.
# This script deliberately refuses non-loopback PostgreSQL/Redis endpoints and
# never reads or mutates production configuration.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND="$ROOT/backend"
STATE_DIR="${MP_LOCAL_API_UI_STATE_DIR:-/tmp/mypocket-local-api-ui}"
STATE_FILE="$STATE_DIR/state.env"
API_PORT="${MP_LOCAL_API_UI_PORT:-8081}"
PGHOST_VALUE="${PGHOST:-127.0.0.1}"
PGPORT_VALUE="${PGPORT:-5432}"
PGUSER_VALUE="${PGUSER:-$(id -un)}"

die() { printf 'local-api-ui-fixture: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"; }

for tool in curl createdb dropdb psql pg_isready redis-server go; do need "$tool"; done

case "$PGHOST_VALUE" in
  127.0.0.1|localhost|::1) ;;
  *) die "refusing non-loopback PostgreSQL host: $PGHOST_VALUE" ;;
esac

read_state() {
  [[ -f "$STATE_FILE" ]] || die "fixture is not running (missing $STATE_FILE)"
  # State is generated locally and contains no shell fragments; only KEY=VALUE
  # records written by this script are accepted.
  set -a
  # shellcheck disable=SC1090
  . "$STATE_FILE"
  set +a
  [[ "${DB_NAME:-}" == mypocket_native_ui_* ]] || die "invalid fixture database in state"
  [[ "${API_PORT_STATE:-}" =~ ^[0-9]+$ ]] || die "invalid API port in state"
}

free_port() {
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(('127.0.0.1', 0))
print(s.getsockname()[1])
s.close()
PY
}

wait_for() {
  local url="$1" tries="${2:-80}"
  for ((i=0; i<tries; i++)); do
    if curl --fail --silent --show-error --max-time 1 "$url" >/dev/null 2>&1; then return 0; fi
    sleep 0.25
  done
  return 1
}

start() {
  [[ ! -f "$STATE_FILE" ]] || die "fixture already appears to be running; use status or stop"
  mkdir -p "$STATE_DIR"
  if (lsof -nP -iTCP:"$API_PORT" -sTCP:LISTEN >/dev/null 2>&1); then
    die "port $API_PORT is already listening; refusing to take over another process"
  fi
  pg_isready -h "$PGHOST_VALUE" -p "$PGPORT_VALUE" -U "$PGUSER_VALUE" >/dev/null || die "PostgreSQL is not ready"

  local suffix db_name redis_port api_log redis_log api_bin redis_pid api_pid dsn
  suffix="$(date -u +%Y%m%d%H%M%S)-$$"
  db_name="mypocket_native_ui_${suffix}"
  redis_port="$(free_port)"
  createdb -h "$PGHOST_VALUE" -p "$PGPORT_VALUE" -U "$PGUSER_VALUE" "$db_name"
  dsn="postgres://${PGUSER_VALUE}@${PGHOST_VALUE}:${PGPORT_VALUE}/${db_name}?sslmode=disable"
  trap 'dropdb -h "$PGHOST_VALUE" -p "$PGPORT_VALUE" -U "$PGUSER_VALUE" --if-exists "$db_name" >/dev/null 2>&1 || true' ERR

  (cd "$BACKEND" && DATABASE_URL="$dsn" MIGRATIONS_PATH="$BACKEND/migrations" go run ./cmd/migrate up) >"$STATE_DIR/migrate.log" 2>&1 || {
    sed -n '1,120p' "$STATE_DIR/migrate.log" >&2
    die "migrations failed; generated database will be dropped"
  }
  api_bin="$STATE_DIR/api"
  (cd "$BACKEND" && go build -o "$api_bin" ./cmd/api)
  redis_log="$STATE_DIR/redis.log"
  nohup redis-server --bind 127.0.0.1 --port "$redis_port" --save '' --appendonly no --dir "$STATE_DIR" \
    </dev/null >"$redis_log" 2>&1 &
  redis_pid=$!
  api_log="$STATE_DIR/api.log"
  # Development mode intentionally permits the attachment storage adapter to
  # remain optional; this fixture never exercises external receipt storage.
  DATABASE_URL="$dsn" REDIS_URL="redis://127.0.0.1:${redis_port}/0" HTTP_ADDR=":$API_PORT" \
    OAUTH_FIXTURE_MODE=true SEED_USER_EMAIL=native-fixture@example.test \
    SEED_USER_NAME='Native UI Fixture' SEED_USER_PASSWORD='fixture-only-password' \
    JWT_SECRET='native-local-api-ui-fixture-secret-keep-local' \
    APP_ENV=development nohup "$api_bin" </dev/null >"$api_log" 2>&1 &
  api_pid=$!
  cat >"$STATE_FILE" <<EOF
DB_NAME=$db_name
DB_DSN=$dsn
REDIS_PORT=$redis_port
REDIS_PID=$redis_pid
API_PORT_STATE=$API_PORT
API_PID=$api_pid
EOF
  trap - ERR
  if ! wait_for "http://127.0.0.1:${API_PORT}/api/v1/health" 120; then
    sed -n '1,160p' "$api_log" >&2 || true
    stop || true
    die "API did not become healthy"
  fi
  local exchange_status
  exchange_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
    -X POST "http://127.0.0.1:${API_PORT}/api/v1/auth/mobile/google/exchange" \
    -H 'Content-Type: application/json' \
    --data '{"code":"fixture","code_verifier":"fixture","redirect_uri":"mypocket://oauth/callback","platform":"ios"}')"
  [[ "$exchange_status" == 200 ]] || { stop || true; die "fixture OAuth exchange returned HTTP $exchange_status"; }
  printf 'started local API fixture: http://127.0.0.1:%s/api/v1\n' "$API_PORT"
  printf 'database: %s\nredis: 127.0.0.1:%s\n' "$db_name" "$redis_port"
}

status() {
  read_state
  local api='stopped' redis='stopped'
  kill -0 "$API_PID" 2>/dev/null && api='running'
  kill -0 "$REDIS_PID" 2>/dev/null && redis='running'
  printf 'api=%s pid=%s base_url=http://127.0.0.1:%s/api/v1\n' "$api" "$API_PID" "$API_PORT_STATE"
  printf 'redis=%s pid=%s port=%s\ndatabase=%s\n' "$redis" "$REDIS_PID" "$REDIS_PORT" "$DB_NAME"
}

stop() {
  read_state
  kill "$API_PID" 2>/dev/null || true
  kill "$REDIS_PID" 2>/dev/null || true
  for _ in {1..20}; do
    kill -0 "$API_PID" 2>/dev/null || break
    sleep 0.1
  done
  dropdb -h "$PGHOST_VALUE" -p "$PGPORT_VALUE" -U "$PGUSER_VALUE" --if-exists "$DB_NAME"
  rm -f "$STATE_FILE"
  printf 'stopped local API fixture and dropped %s\n' "$DB_NAME"
}

run() {
  [[ "${2:-}" == "--" ]] || die "run requires: run -- command [args...]"
  [[ "$#" -ge 3 ]] || die "run requires a command"
  start
  local rc=0
  trap 'stop >/dev/null 2>&1 || true' EXIT INT TERM
  "${@:3}" || rc=$?
  trap - EXIT INT TERM
  stop
  return "$rc"
}

case "${1:-}" in
  start) start ;;
  status) status ;;
  stop) stop ;;
  run) run "$@" ;;
  *) printf 'usage: %s {start|status|stop|run -- command}\n' "$0" >&2; exit 2 ;;
esac
