#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SKIP_BUILD=0
if [[ "${1:-}" == "--skip-build" ]]; then
  SKIP_BUILD=1
fi

cleanup() {
  echo "🧹 Cleaning up..."
  [[ -n "${API_PID:-}" ]] && kill "$API_PID" 2>/dev/null || true
  [[ -n "${PREVIEW_PID:-}" ]] && kill "$PREVIEW_PID" 2>/dev/null || true
  wait 2>/dev/null || true
  echo "✅ Done"
}
trap cleanup EXIT INT TERM

# ── 1. Build frontend with test API base URL ──
echo "🏗️  Building frontend (API → :18173)..."
cd "$ROOT_DIR/frontend"
VITE_API_BASE_URL="http://127.0.0.1:18173" npm run build

# ── 2. Start Go API (fixture mode) ──
echo "🚀 Starting Go API on :18173 (fixture mode)..."
cd "$ROOT_DIR/backend"
go run ./cmd/migrate 2>/dev/null || true
APP_ENV=development \
  DATABASE_URL="postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable" \
  PUBLIC_WEB_URL="http://127.0.0.1:4173" \
  COOKIE_SECRET="change-this-development-cookie-secret-32-bytes" \
  CSRF_SECRET="change-this-development-csrf-secret-32-bytes" \
  HTTP_ADDR="127.0.0.1:18173" \
  OAUTH_FIXTURE_MODE=true \
  go run ./cmd/api &
API_PID=$!

for i in $(seq 1 30); do
  if curl -sf http://127.0.0.1:18173/api/v1/health/live >/dev/null 2>&1; then
    echo "  ✅ API ready"
    break
  fi
  if [[ $i -eq 30 ]]; then
    echo "  ❌ API failed to start"
    exit 1
  fi
  sleep 0.5
done

# ── 3. Start Vite preview ──
echo "🚀 Starting Vite preview on :4173..."
cd "$ROOT_DIR/frontend"
npx vite preview --host 0.0.0.0 --port 4173 &
PREVIEW_PID=$!

for i in $(seq 1 15); do
  if curl -sf http://127.0.0.1:4173 >/dev/null 2>&1; then
    echo "  ✅ Preview ready"
    break
  fi
  if [[ $i -eq 15 ]]; then
    echo "  ❌ Preview failed to start"
    exit 1
  fi
  sleep 1
done

# ── 4. Run Playwright tests ──
echo "🧪 Running E2E tests..."
cd "$ROOT_DIR/frontend"
npx playwright test --config=playwright.e2e.config.ts --reporter=list "$@" 2>&1
echo "✅ All tests done"