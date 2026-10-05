#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GUARD="$ROOT/scripts/check-base-components.sh"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-base-guard.XXXXXX")"
trap 'rmdir "$TEST_ROOT" 2>/dev/null || true' EXIT

bash "$GUARD"

mkdir -p "$ROOT/MyPocketIOS/Features/GuardFixture"
trap 'rm -f "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift" 2>/dev/null || true; rmdir "$ROOT/MyPocketIOS/Features/GuardFixture" 2>/dev/null || true; rmdir "$TEST_ROOT" 2>/dev/null || true' EXIT
printf '%s\n' 'import SwiftUI' 'struct BadView: View { var body: some View { Button("bad") {} } }' > "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift"
if bash "$GUARD" >/dev/null 2>&1; then
  echo "guard failed to detect raw control" >&2
  exit 1
fi
rm -f "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift"
rmdir "$ROOT/MyPocketIOS/Features/GuardFixture"

echo "native ios base-component tests passed"
