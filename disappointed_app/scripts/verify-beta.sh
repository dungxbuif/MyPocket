#!/usr/bin/env bash
set -euo pipefail

# Repository integration tests reset the public schema. Require a dedicated DB.
: "${MYPOCKET_TEST_DATABASE_URL:?Set a dedicated mypocket_verify_* database URL}"
if [[ "$MYPOCKET_TEST_DATABASE_URL" != *"/mypocket_verify_"* ]]; then
  echo "Refusing schema-reset tests: database name must start with mypocket_verify_." >&2
  exit 1
fi
TASK_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

required_files=(
  docs/work/completion/BASELINE-2026-09-11.md
  docs/work/completion/DATA-SAFETY.md
  docs/work/completion/BUSINESS-CORRECTNESS.md
  docs/work/completion/PHYSICAL-DEVICE-UAT.md
  docs/work/completion/API-CONTRACT.md
  docs/work/completion/AGENT-REVIEW-FIRST.md
  docs/work/completion/OCR-AGENT-TOOL.md
)
for file in "${required_files[@]}"; do
  [[ -s "$TASK_ROOT/$file" ]] || { echo "Missing release evidence: $file" >&2; exit 1; }
done
grep -q '^Status: \*\*PASS' "$TASK_ROOT/docs/work/completion/PHYSICAL-DEVICE-UAT.md" || {
  echo "Physical iPhone/Safari UAT is not PASS" >&2
  exit 1
}
for key in MYPOCKET_AI_BASE_URL MYPOCKET_AI_API_KEY MYPOCKET_AI_MODEL MYPOCKET_OCR_BASE_URL MYPOCKET_OCR_API_KEY; do
  [[ -n "${!key:-}" ]] || { echo "Missing required provider setting: $key" >&2; exit 1; }
done
[[ -f "$TASK_ROOT/backend/migrations/0016_agent_image_tools.sql" ]] || { echo "Migration 0016 is missing" >&2; exit 1; }
: "${MYPOCKET_BACKUP_DIR:?Set MYPOCKET_BACKUP_DIR containing verified release backup evidence}"
[[ -s "$MYPOCKET_BACKUP_DIR/backup-files.sha256" ]] || { echo "Verified backup checksum manifest is missing" >&2; exit 1; }
[[ -s "$MYPOCKET_BACKUP_DIR/restore-evidence.txt" ]] || { echo "Restore drill evidence is missing" >&2; exit 1; }
grep -q '^status=pass$' "$MYPOCKET_BACKUP_DIR/restore-evidence.txt" || { echo "Restore drill did not pass" >&2; exit 1; }

cd "$TASK_ROOT/backend"
go vet ./...
go test -race -p 1 -count=1 ./...
cd "$TASK_ROOT/frontend"
npm run typecheck
npm test -- --run
npm run build
npm run test:e2e
cd "$TASK_ROOT/frontend/docs"
npm run build
cd "$TASK_ROOT"
git diff --check
git diff --quiet -- backend frontend scripts docs .env.example || { echo "Release inputs differ from the tested commit" >&2; exit 1; }
