#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FEATURES="$ROOT/MyPocketIOS/Features"

if [ ! -d "$FEATURES" ]; then
  echo "native ios base-component guard passed (no feature sources yet)"
  exit 0
fi

if rg -n --glob '*.swift' '\b(Button|TextField|SecureField|DatePicker|ProgressView|Text|Image)\s*\(' "$FEATURES"; then
  echo "raw SwiftUI visual/control found outside DesignSystem" >&2
  exit 1
fi

if rg -n --glob '*.swift' '\.(sheet|fullScreenCover|alert)\s*\(' "$FEATURES"; then
  echo "modal/alert modifier found outside DesignSystem" >&2
  exit 1
fi

echo "native ios base-component guard passed"
