#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ICON_ROOT="$REPO_ROOT/mobile-ios/MyPocketIOS/Assets.xcassets/Icons"
MANIFEST="${ICON_MANIFEST_PATH:-$REPO_ROOT/docs/design/references/mypocket-icon-study/manifest.json}"
BOARD="$REPO_ROOT/docs/design/references/mypocket-icon-study/comparison.html"
LICENSE="$REPO_ROOT/mobile-ios/MyPocketIOS/Resources/Phosphor-LICENSE.txt"
PROJECT="$REPO_ROOT/mobile-ios/MyPocketIOS.xcodeproj/project.pbxproj"
CATALOG="$REPO_ROOT/mobile-ios/MyPocketIOS/DesignSystem/MPIconCatalog.swift"

test -d "$ICON_ROOT"
test -f "$MANIFEST"
test -f "$BOARD"
test -f "$LICENSE"
test -f "$PROJECT"

python3 - "$MANIFEST" "$ICON_ROOT" "$BOARD" "$LICENSE" "$PROJECT" "$CATALOG" <<'PY'
import json
import pathlib
import re
import sys

manifest_path, icon_root, board_path, license_path, project_path, catalog_path = map(pathlib.Path, sys.argv[1:])
manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
catalog_text = catalog_path.read_text(encoding="utf-8")

expected_roles = {
    "tag", "home", "ledger", "assistant", "account", "feedback", "wallet", "group",
    "bankLink", "recurring", "report", "search", "bell", "add", "send", "food",
    "coffee", "snacks", "meals", "shopping", "transport", "health", "bills", "education",
    "travel", "savings", "family", "entertainment", "gifts", "insurance", "investment",
    "business", "transfer", "income", "debt", "credit", "aggregate", "adjustment",
}
expected_board = {
    "home", "ledger", "assistant", "account", "feedback", "wallet", "report", "search",
    "bell", "add", "send", "food", "coffee", "shopping", "transport", "health", "bills",
    "education", "travel", "savings",
}

source = manifest.get("source", {})
assert source.get("name") == "Phosphor Icons", "manifest source name is not Phosphor"
assert source.get("version") == "2.1.1", "manifest source version is not pinned"
assert source.get("weight") == "regular", "manifest source weight is not regular"
assert source.get("license") == "MIT", "manifest license is not MIT"
assert manifest.get("duotone", {}).get("owner_visual_approval_required") is True

roles = manifest.get("roles", [])
role_ids = {role.get("id") for role in roles}
assert len(roles) == len(expected_roles), "manifest role count is incomplete"
assert role_ids == expected_roles, "manifest roles do not match MPIconID"
enum_block = catalog_text.split("enum MPIconID", 1)[1].split("enum MPIconStyle", 1)[0]
catalog_ids = set(re.findall(r"^\s*case\s+([A-Za-z0-9_]+)", enum_block, re.MULTILINE))
assert catalog_ids == role_ids, "manifest roles do not match MPIconID source"

for role in roles:
    asset = role["asset_name"]
    assert f'"{asset}"' in catalog_text, f"catalog has no bundled asset mapping for {role['id']}"
    contents = icon_root / f"{asset}.imageset" / "Contents.json"
    vector = icon_root / f"{asset}.imageset" / f"{asset}.svg"
    assert contents.is_file(), f"missing asset contents for {role['id']}: {contents}"
    assert vector.is_file(), f"missing vector for {role['id']}: {vector}"
    contents_data = json.loads(contents.read_text(encoding="utf-8"))
    assert contents_data.get("properties", {}).get("preserves-vector-representation") is True, f"asset is not vector-preserving: {asset}"
    assert "fill=\"currentColor\"" in vector.read_text(encoding="utf-8"), f"asset is not tintable: {asset}"

board_roles = set(manifest.get("board_roles", []))
assert board_roles == expected_board, "comparison board role list is incomplete"
board_html = board_path.read_text(encoding="utf-8")
for role in sorted(expected_board):
    assert f'data-role="{role}"' in board_html, f"comparison board is missing {role}"
for source in re.findall(r'<img\s+src="([^"]+)"', board_html):
    assert (board_path.parent / source).resolve().is_file(), f"comparison preview is missing: {source}"

license_text = license_path.read_text(encoding="utf-8")
assert "MIT License" in license_text
project_text = project_path.read_text(encoding="utf-8")
assert "Phosphor-LICENSE.txt in Resources" in project_text
assert "Assets.xcassets in Resources" in project_text

print(f"icon catalog guard passed: {len(roles)} roles, {len(board_roles)} comparison roles")
PY
