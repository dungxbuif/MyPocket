#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GUARD="$ROOT/scripts/check-base-components.sh"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-base-guard.XXXXXX")"
trap 'rmdir "$TEST_ROOT" 2>/dev/null || true' EXIT

bash "$GUARD"

mkdir -p "$ROOT/MyPocketIOS/Features/GuardFixture"
trap 'rm -f "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift" 2>/dev/null || true; rmdir "$ROOT/MyPocketIOS/Features/GuardFixture" 2>/dev/null || true; rmdir "$TEST_ROOT" 2>/dev/null || true' EXIT
for expression in 'Button("bad") {}' 'Text("bad")' 'Image(systemName: "xmark")'; do
  printf '%s\n' 'import SwiftUI' "struct BadView: View { var body: some View { $expression } }" > "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift"
  if bash "$GUARD" >/dev/null 2>&1; then
    echo "guard failed to detect raw feature visual/control: $expression" >&2
    exit 1
  fi
done
rm -f "$ROOT/MyPocketIOS/Features/GuardFixture/BadView.swift"
rmdir "$ROOT/MyPocketIOS/Features/GuardFixture"

mkdir -p "$ROOT/MyPocketIOS/Features/GuardFixture"
printf '%s\n' 'import SwiftUI' 'struct BadFontView: View { var body: some View { MPText("bad").font(.body) } }' > "$ROOT/MyPocketIOS/Features/GuardFixture/BadFontView.swift"
if bash "$GUARD" >/dev/null 2>&1; then
  echo "guard failed to detect a feature system-font override" >&2
  exit 1
fi
rm -f "$ROOT/MyPocketIOS/Features/GuardFixture/BadFontView.swift"
rmdir "$ROOT/MyPocketIOS/Features/GuardFixture"

echo "native ios base-component tests passed"
