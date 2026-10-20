#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROJECT="$ROOT/MyPocketIOS.xcodeproj"

test -d "$PROJECT"
test -f "$ROOT/MyPocketIOS/App/MyPocketIOSApp.swift"
test -f "$ROOT/MyPocketIOS/Resources/Info.plist"
test -f "$ROOT/MyPocketIOS/Resources/Manrope-Variable.ttf"
test -f "$ROOT/MyPocketIOS/Resources/OFL-Manrope.txt"
ICONSET="$ROOT/MyPocketIOS/Assets.xcassets/AppIcon.appiconset"
test -f "$ICONSET/Contents.json"
test -f "$ICONSET/AppIcon.png"
test -x "$ROOT/scripts/generate-app-icon.sh"
test -x "$ROOT/scripts/release-artifact-preflight.sh"
test -x "$ROOT/scripts/release-artifact-preflight.test.sh"
test -f "$ROOT/scripts/fixtures/release-Info.plist"
grep -q "MyPocketIOS/Assets.xcassets" "$ROOT/project.yml"
ICON_SIZE="$(sips -g pixelWidth -g pixelHeight "$ICONSET/AppIcon.png" 2>/dev/null)"
grep -q "pixelWidth: 1024" <<<"$ICON_SIZE"
grep -q "pixelHeight: 1024" <<<"$ICON_SIZE"
ICON_CHECK_DIR="$(mktemp -d)"
trap 'rm -rf "$ICON_CHECK_DIR"' EXIT
"$ROOT/scripts/generate-app-icon.sh" "$ICON_CHECK_DIR/AppIcon.png"
cmp -s "$ICONSET/AppIcon.png" "$ICON_CHECK_DIR/AppIcon.png"
test -f "$ROOT/Config/Debug.xcconfig"
test -f "$ROOT/Config/Release.xcconfig"

LISTING="$(xcodebuild -project "$PROJECT" -list 2>/dev/null)"
grep -q "MyPocketIOS" <<<"$LISTING"
grep -q "MyPocketIOSTests" <<<"$LISTING"
grep -q "com.mypocket.ios" "$ROOT/project.yml"
grep -q "MP_API_BASE_URL" "$ROOT/Config/Debug.xcconfig"
grep -q "configFiles:" "$ROOT/project.yml"
grep -q "Manrope-Variable.ttf" "$ROOT/MyPocketIOS/Resources/Info.plist"
grep -q "UISupportedInterfaceOrientations" "$ROOT/MyPocketIOS/Resources/Info.plist"
grep -q "UIInterfaceOrientationPortrait" "$ROOT/MyPocketIOS/Resources/Info.plist"
if grep -q "UIInterfaceOrientationLandscape" "$ROOT/MyPocketIOS/Resources/Info.plist"; then
  echo "native iOS app must remain portrait-only" >&2
  exit 1
fi
if grep -q "MP_GOOGLE_\|auth-google\|GooglePKCEClient" "$ROOT/MyPocketIOS/Resources/Info.plist" "$ROOT/Config/Debug.xcconfig" "$ROOT/Config/Release.xcconfig" "$ROOT/MyPocketIOS/Features/Auth/AuthView.swift" "$ROOT/MyPocketIOS/Features/Auth/AuthViewModel.swift"; then
  echo "native password-only auth must not ship Google login configuration or UI" >&2
  exit 1
fi

DEBUG_SETTINGS="$(xcodebuild -project "$PROJECT" -scheme MyPocketIOS -configuration Debug -showBuildSettings 2>/dev/null)"
RELEASE_SETTINGS="$(xcodebuild -project "$PROJECT" -scheme MyPocketIOS -configuration Release -showBuildSettings 2>/dev/null)"
grep -q "MP_API_BASE_URL = http://127.0.0.1:8080/api/v1" <<<"$DEBUG_SETTINGS"
grep -q "ENABLE_PREVIEWS = YES" <<<"$DEBUG_SETTINGS"
grep -q "MP_API_BASE_URL = https://money.dungxbuif.com/api/v1" <<<"$RELEASE_SETTINGS"

if git -C "$ROOT/.." ls-files | rg -q '(^|/)(\.env|.*secret.*|.*google.*client.*secret.*)$'; then
  echo "tracked secret-like file found" >&2
  exit 1
fi

echo "native ios project verification passed"
