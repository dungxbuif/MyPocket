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

if rg -n --glob '*.swift' '\.font\(\.' "$FEATURES"; then
  echo "system-font override found in a feature; use MPText/MPFont bases" >&2
  exit 1
fi

DESIGN_SYSTEM="$ROOT/MyPocketIOS/DesignSystem"
if rg -n --glob '*.swift' '\.font\(\.(body|caption|footnote|headline|subheadline|title)' "$DESIGN_SYSTEM"; then
  echo "system-font override found in DesignSystem; use MPFont roles" >&2
  exit 1
fi

echo "native ios base-component guard passed"
