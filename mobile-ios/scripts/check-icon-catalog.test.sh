#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GUARD="$ROOT/mobile-ios/scripts/check-icon-catalog.sh"
MANIFEST="$ROOT/docs/design/references/mypocket-icon-study/manifest.json"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-icon-catalog.XXXXXX")"
trap 'rm -rf "$TEST_ROOT"' EXIT

bash "$GUARD"

python3 - "$MANIFEST" "$TEST_ROOT/missing-role.json" <<'PY'
import json
import sys

source, target = sys.argv[1:]
with open(source, encoding="utf-8") as handle:
    manifest = json.load(handle)
manifest["roles"] = manifest["roles"][:-1]
with open(target, "w", encoding="utf-8") as handle:
    json.dump(manifest, handle)
PY

if ICON_MANIFEST_PATH="$TEST_ROOT/missing-role.json" bash "$GUARD" >/dev/null 2>&1; then
    echo "icon catalog guard failed to reject an incomplete manifest" >&2
    exit 1
fi

echo "icon catalog guard tests passed"
