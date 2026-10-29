#!/usr/bin/env bash
set -euo pipefail

# Read-only audit of a running API. This script only performs HTTP GETs and
# never applies migrations, updates a service, or sends authenticated data.
# Keep AUDIT_BASE_URL explicit so an operator cannot accidentally inspect the
# wrong environment.

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
base_url="${AUDIT_BASE_URL:-}"
openapi_path="${AUDIT_OPENAPI_PATH:-/api/v1/docs/doc.json}"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

[[ -n "${base_url}" ]] || fail 'AUDIT_BASE_URL is required (for example https://money.dungxbuif.com)'
command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v python3 >/dev/null 2>&1 || fail 'python3 is required'

base_url="${base_url%/}"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-readonly-audit.XXXXXX")"
trap 'rm -rf "${tmp_dir}"' EXIT

health_file="${tmp_dir}/health.json"
live_spec="${tmp_dir}/live-openapi.json"
local_spec="${repo_root}/backend/docs/swagger.json"
[[ -f "${local_spec}" ]] || fail "local OpenAPI document not found: ${local_spec}"

health_status="$(curl --fail --silent --show-error --location --output "${health_file}" --write-out '%{http_code}' "${base_url}/api/v1/health")" \
  || fail 'health request failed'
[[ "${health_status}" == '200' ]] || fail "health returned HTTP ${health_status}"

spec_status="$(curl --fail --silent --show-error --location --output "${live_spec}" --write-out '%{http_code}' "${base_url}${openapi_path}")" \
  || fail "OpenAPI request failed (${openapi_path})"
[[ "${spec_status}" == '200' ]] || fail "OpenAPI returned HTTP ${spec_status}"

python3 - "${health_file}" "${live_spec}" "${local_spec}" <<'PY'
import json
import os
import pathlib
import sys

health_path, live_path, local_path = map(pathlib.Path, sys.argv[1:])

def load(path):
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise SystemExit(f"FAIL: invalid JSON in {path}: {exc}")

health = load(health_path)
live = load(live_path)
local = load(local_path)

status = health.get("data", {}).get("status")
if status != "ok":
    raise SystemExit(f"FAIL: health payload status is {status!r}, expected 'ok'")

live_paths = set((live.get("paths") or {}).keys())
local_paths = set((local.get("paths") or {}).keys())
missing = sorted(local_paths - live_paths)
extra = sorted(live_paths - local_paths)

http_methods = {"get", "post", "put", "patch", "delete", "options", "head"}
def methods(document, path):
    operations = (document.get("paths") or {}).get(path) or {}
    return {method for method in operations if method in http_methods}

# A path can exist in both documents while a staged release still misses one
# operation. Report that drift separately; REQUIRE_ROUTE_PARITY turns it into
# a release gate without changing the script's default read-only behavior.
method_drift = []
security_drift = []
for path in sorted(local_paths & live_paths):
    local_methods = methods(local, path)
    live_methods = methods(live, path)
    if local_methods != live_methods:
        method_drift.append((path, sorted(local_methods - live_methods), sorted(live_methods - local_methods)))
    for method in sorted(local_methods & live_methods):
        local_operation = (local.get("paths", {}).get(path) or {}).get(method) or {}
        live_operation = (live.get("paths", {}).get(path) or {}).get(method) or {}
        # OpenAPI's `security` field is the operation-level authorization
        # contract. `None` and an empty array both mean public; a non-empty
        # array means the operation requires one of the declared schemes.
        local_requires_auth = bool(local_operation.get("security"))
        live_requires_auth = bool(live_operation.get("security"))
        if local_requires_auth != live_requires_auth:
            security_drift.append((path, method, local_requires_auth, live_requires_auth))

print("PASS: health status=ok")
print(f"LIVE_OPENAPI_VERSION: {live.get('info', {}).get('version', '<missing>')}")
print(f"LOCAL_OPENAPI_VERSION: {local.get('info', {}).get('version', '<missing>')}")
print(f"LIVE_PATHS: {len(live_paths)}")
print(f"LOCAL_PATHS: {len(local_paths)}")
if missing:
    print("CANDIDATE_ONLY_PATHS:")
    for path in missing:
        print(f"  {path}")
if extra:
    print("LIVE_ONLY_PATHS:")
    for path in extra:
        print(f"  {path}")
if method_drift:
    print("METHOD_DRIFT:")
    for path, missing_methods, extra_methods in method_drift:
        print(f"  {path}: missing_live={','.join(missing_methods) or '-'} live_only={','.join(extra_methods) or '-'}")
if security_drift:
    print("SECURITY_DRIFT:")
    for path, method, local_requires_auth, live_requires_auth in security_drift:
        print(
            f"  {method.upper()} {path}: "
            f"local_requires_auth={str(local_requires_auth).lower()} "
            f"live_requires_auth={str(live_requires_auth).lower()}"
        )

if os.environ.get("REQUIRE_ROUTE_PARITY") == "1" and (missing or method_drift or security_drift):
    raise SystemExit("FAIL: production OpenAPI route/method parity is incomplete")
PY

printf 'Read-only audit complete; no authenticated request or mutation was made.\n'
