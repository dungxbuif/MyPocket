#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROJECT="$ROOT/MyPocketIOS.xcodeproj"

test -d "$PROJECT"
test -f "$ROOT/MyPocketIOS/App/MyPocketIOSApp.swift"
test -f "$ROOT/MyPocketIOS/Resources/Info.plist"
test -f "$ROOT/Config/Debug.xcconfig"
test -f "$ROOT/Config/Release.xcconfig"

LISTING="$(xcodebuild -project "$PROJECT" -list 2>/dev/null)"
grep -q "MyPocketIOS" <<<"$LISTING"
grep -q "MyPocketIOSTests" <<<"$LISTING"
grep -q "com.mypocket.ios" "$ROOT/project.yml"
grep -q "MP_API_BASE_URL" "$ROOT/Config/Debug.xcconfig"
grep -q "configFiles:" "$ROOT/project.yml"
grep -q "MP_GOOGLE_CLIENT_ID" "$ROOT/MyPocketIOS/Resources/Info.plist"
grep -q "MP_GOOGLE_REDIRECT_URI" "$ROOT/MyPocketIOS/Resources/Info.plist"

DEBUG_SETTINGS="$(xcodebuild -project "$PROJECT" -scheme MyPocketIOS -configuration Debug -showBuildSettings 2>/dev/null)"
RELEASE_SETTINGS="$(xcodebuild -project "$PROJECT" -scheme MyPocketIOS -configuration Release -showBuildSettings 2>/dev/null)"
grep -q "MP_API_BASE_URL = http://127.0.0.1:8080/api/v1" <<<"$DEBUG_SETTINGS"
grep -q "MP_API_BASE_URL = https://money.dungxbuif.com/api/v1" <<<"$RELEASE_SETTINGS"
grep -q "MP_GOOGLE_REDIRECT_URI = mypocket://oauth/callback" <<<"$RELEASE_SETTINGS"

if git -C "$ROOT/.." ls-files | rg -q '(^|/)(\.env|.*secret.*|.*google.*client.*secret.*)$'; then
  echo "tracked secret-like file found" >&2
  exit 1
fi

echo "native ios project verification passed"
