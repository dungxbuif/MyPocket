#!/usr/bin/env bash
set -euo pipefail

# Read-only check for a resolved Release .app. This intentionally does not
# sign, install, contact production, or print sensitive plist values.
APP_PATH="${APP_PATH:-${1:-}}"
EXPECTED_BUNDLE_ID="${EXPECTED_BUNDLE_ID:-com.mypocket.ios}"
EXPECTED_VERSION="${EXPECTED_VERSION:-1.2.1}"
EXPECTED_BUILD="${EXPECTED_BUILD:-4}"
EXPECTED_API_URL="${EXPECTED_API_URL:-https://money.dungxbuif.com/api/v1}"

if [[ -z "$APP_PATH" ]]; then
  echo "usage: APP_PATH=/path/MyPocketIOS.app $0" >&2
  exit 2
fi
if [[ ! -d "$APP_PATH" ]]; then
  echo "release artifact not found: $APP_PATH" >&2
  exit 2
fi
PLIST="$APP_PATH/Info.plist"
if [[ ! -f "$PLIST" ]]; then
  echo "release artifact is missing Info.plist" >&2
  exit 2
fi

plist_value() {
  /usr/libexec/PlistBuddy -c "Print :$1" "$PLIST" 2>/dev/null || true
}
fail() {
  echo "release artifact preflight failed: $1" >&2
  exit 1
}

[[ "$(plist_value CFBundleIdentifier)" == "$EXPECTED_BUNDLE_ID" ]] || \
  fail "unexpected bundle identifier"
[[ "$(plist_value CFBundleShortVersionString)" == "$EXPECTED_VERSION" ]] || \
  fail "unexpected marketing version"
[[ "$(plist_value CFBundleVersion)" == "$EXPECTED_BUILD" ]] || \
  fail "unexpected build number"
[[ "$(plist_value MP_API_BASE_URL)" == "$EXPECTED_API_URL" ]] || \
  fail "Release API base URL is not the expected production HTTPS endpoint"

ORIENTATIONS="$(plist_value UISupportedInterfaceOrientations)"
grep -qF "UIInterfaceOrientationPortrait" <<<"$ORIENTATIONS" || \
  fail "portrait orientation is not enabled"
if grep -Eq 'UIInterfaceOrientation(Landscape|PortraitUpsideDown)' <<<"$ORIENTATIONS"; then
  fail "landscape or upside-down orientation is enabled"
fi

# The packaged app must not carry obvious source-control secret files.
if find "$APP_PATH" -type f \( -name '.env' -o -iname '*secret*' -o -iname '*private*key*' \) -print -quit | grep -q .; then
  fail "secret-like file found in packaged app"
fi

echo "release artifact preflight passed"
echo "bundle=$EXPECTED_BUNDLE_ID version=$EXPECTED_VERSION build=$EXPECTED_BUILD orientation=portrait"
