#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SOURCE="$ROOT/../app/public/pwa-icon.svg"
OUTPUT="${1:-$ROOT/MyPocketIOS/Assets.xcassets/AppIcon.appiconset/AppIcon.png}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

test -f "$SOURCE"
mkdir -p "$(dirname "$OUTPUT")"

# iOS applies its own icon mask. Keep the PWA artwork but make the full-canvas
# background square instead of rasterizing the PWA's rounded outer corners.
sed 's/<rect width="512" height="512" rx="112" fill=/<rect width="512" height="512" fill=/' "$SOURCE" > "$TMP_DIR/AppIcon.svg"
if rg -q '<rect width="512" height="512" rx="112" fill=' "$TMP_DIR/AppIcon.svg"; then
  echo "PWA icon background shape changed; review the iOS mask transform" >&2
  exit 1
fi

rsvg-convert -w 1024 -h 1024 "$TMP_DIR/AppIcon.svg" -o "$OUTPUT"
