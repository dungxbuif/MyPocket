#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT/scripts/release-artifact-preflight.sh"
FIXTURE="$ROOT/scripts/fixtures/release-Info.plist"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

make_app() {
  local name="$1"
  mkdir -p "$TMP/$name.app"
  cp "$FIXTURE" "$TMP/$name.app/Info.plist"
  printf '%s\n' "$TMP/$name.app"
}

APP="$(make_app valid)"
"$SCRIPT" "$APP" >/dev/null

BAD_VERSION="$(make_app bad-version)"
/usr/libexec/PlistBuddy -c 'Set :CFBundleVersion 99' "$BAD_VERSION/Info.plist"
if "$SCRIPT" "$BAD_VERSION" >/dev/null 2>&1; then
  echo "expected version mismatch to fail" >&2
  exit 1
fi

BAD_URL="$(make_app bad-url)"
/usr/libexec/PlistBuddy -c 'Set :MP_API_BASE_URL http://127.0.0.1:8080/api/v1' "$BAD_URL/Info.plist"
if "$SCRIPT" "$BAD_URL" >/dev/null 2>&1; then
  echo "expected API URL mismatch to fail" >&2
  exit 1
fi

BAD_ORIENTATION="$(make_app bad-orientation)"
/usr/libexec/PlistBuddy -c 'Add :UISupportedInterfaceOrientations:1 string UIInterfaceOrientationLandscapeLeft' "$BAD_ORIENTATION/Info.plist"
if "$SCRIPT" "$BAD_ORIENTATION" >/dev/null 2>&1; then
  echo "expected landscape orientation to fail" >&2
  exit 1
fi

echo "release artifact preflight tests passed"
