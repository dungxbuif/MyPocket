#!/usr/bin/env bash
set -euo pipefail

# Repository integration tests reset the public schema. Require a dedicated DB.
: "${MYPOCKET_TEST_DATABASE_URL:?Set a dedicated mypocket_verify_* database URL}"
if [[ "$MYPOCKET_TEST_DATABASE_URL" != *"/mypocket_verify_"* ]]; then
  echo "Refusing schema-reset tests: database name must start with mypocket_verify_." >&2
  exit 1
fi
TASK_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$TASK_ROOT/backend"
go vet ./...
go test -race -p 1 -count=1 ./...
cd "$TASK_ROOT/frontend"
npm run typecheck
npm test -- --run
npm run test:e2e
cd "$TASK_ROOT"
git diff --check
