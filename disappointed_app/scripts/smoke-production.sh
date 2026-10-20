#!/usr/bin/env bash
set -euo pipefail

base_url="${1:-https://money.dungxbuif.com}"
: "${MYPOCKET_SMOKE_API_KEY:?Set the dedicated smoke-account API key}"
: "${MYPOCKET_SMOKE_RECEIPT_PATH:?Set a real JPEG, PNG, or WebP receipt image path}"
: "${MYPOCKET_SMOKE_RECEIPT_CONTENT_TYPE:?Set image/jpeg, image/png, or image/webp}"
[[ -f "$MYPOCKET_SMOKE_RECEIPT_PATH" ]] || { echo "Smoke receipt image does not exist" >&2; exit 1; }
case "$MYPOCKET_SMOKE_RECEIPT_CONTENT_TYPE" in image/jpeg|image/png|image/webp) ;; *) echo "Unsupported smoke receipt content type" >&2; exit 1 ;; esac

for command_name in basename curl jq shasum mktemp uuidgen wc; do
  command -v "$command_name" >/dev/null || { echo "Missing required command: $command_name" >&2; exit 1; }
done

smoke_dir="$(mktemp -d)"
chmod 700 "$smoke_dir"
auth_config="$smoke_dir/auth.conf"
printf 'header = "Authorization: Bearer %s"\n' "$MYPOCKET_SMOKE_API_KEY" >"$auth_config"
chmod 600 "$auth_config"
run_id="$(uuidgen | tr '[:upper:]' '[:lower:]')"
wallet_a=""
wallet_b=""
income_category=""
expense_category=""
income_tx=""
transfer_tx=""
child_key_id=""
agent_draft_id=""

api_call() {
  local method="$1" path="$2" body="${3:-}" output="$4" idempotency="${5:-}"
  local args=(--fail-with-body --silent --show-error --config "$auth_config" -X "$method" -H 'Accept: application/json')
  [[ -z "$idempotency" ]] || args+=(-H "Idempotency-Key: $idempotency")
  [[ -z "$body" ]] || args+=(-H 'Content-Type: application/json' --data "$body")
  curl "${args[@]}" "$base_url/api/v1$path" >"$output"
}

archive_if_present() {
	local resource="$1" id="$2" version="$3"
	[[ -n "$id" ]] || return 0
	api_call POST "/$resource/$id/archive" "{\"base_version\":$version}" "$smoke_dir/cleanup.json" "cleanup-$run_id-$resource-$id" >/dev/null 2>&1 || true
}

cleanup() {
  set +e
  if [[ -n "$agent_draft_id" ]]; then
    api_call GET /transaction-drafts '' "$smoke_dir/drafts-cleanup.json" >/dev/null 2>&1
    draft_version="$(jq -r --arg id "$agent_draft_id" '.drafts[] | select(.id == $id) | .version' "$smoke_dir/drafts-cleanup.json" 2>/dev/null | head -1)"
    [[ -z "$draft_version" ]] || api_call POST "/transaction-drafts/$agent_draft_id/reject" "{\"version\":$draft_version}" "$smoke_dir/reject.json" "reject-$run_id" >/dev/null 2>&1
  fi
  archive_if_present transactions "$transfer_tx" 1
  archive_if_present transactions "$income_tx" 1
  archive_if_present categories "$expense_category" 1
  archive_if_present categories "$income_category" 1
  api_call GET /wallets '' "$smoke_dir/wallets-cleanup.json" >/dev/null 2>&1
  wallet_b_version="$(jq -r --arg id "$wallet_b" '.wallets[] | select(.id == $id) | .version' "$smoke_dir/wallets-cleanup.json" 2>/dev/null | head -1)"
  wallet_a_version="$(jq -r --arg id "$wallet_a" '.wallets[] | select(.id == $id) | .version' "$smoke_dir/wallets-cleanup.json" 2>/dev/null | head -1)"
  [[ -z "$wallet_b_version" ]] || archive_if_present wallets "$wallet_b" "$wallet_b_version"
  [[ -z "$wallet_a_version" ]] || archive_if_present wallets "$wallet_a" "$wallet_a_version"
  if [[ -n "$child_key_id" ]]; then
    api_call POST "/api-keys/$child_key_id/revoke" '{}' "$smoke_dir/revoke-cleanup.json" "revoke-$run_id" >/dev/null 2>&1
  fi
  rm -rf "$smoke_dir"
}
trap cleanup EXIT

curl --fail --silent --show-error "$base_url/api/v1/health/live" >"$smoke_dir/live.json"
curl --fail --silent --show-error "$base_url/api/v1/health/ready" >"$smoke_dir/ready.json"
curl --fail --silent --show-error "$base_url/manifest.webmanifest" >"$smoke_dir/manifest.json"
curl --fail --silent --show-error "$base_url/sw.js" >"$smoke_dir/sw.js"
curl --fail --silent --show-error "$base_url/api/v1/openapi.json" >"$smoke_dir/openapi.json"
curl --fail --silent --show-error --location "$base_url/docs/api/openapi" >"$smoke_dir/docs.html"
curl --fail --silent --show-error --location "$base_url/docs/skills/mypocket-api" >"$smoke_dir/skill.html"
jq -e '.openapi == "3.1.0" and (.paths["/agent/messages"].post != null)' "$smoke_dir/openapi.json" >/dev/null

api_call GET /me '' "$smoke_dir/me.json"
jq -e '.user.id and .user.email' "$smoke_dir/me.json" >/dev/null

api_call POST /wallets "{\"name\":\"Smoke A $run_id\",\"type\":\"cash\"}" "$smoke_dir/wallet-a.json" "wallet-a-$run_id"
wallet_a="$(jq -er '.wallet.id' "$smoke_dir/wallet-a.json")"
api_call POST /wallets "{\"name\":\"Smoke B $run_id\",\"type\":\"cash\"}" "$smoke_dir/wallet-b.json" "wallet-b-$run_id"
wallet_b="$(jq -er '.wallet.id' "$smoke_dir/wallet-b.json")"
api_call POST /categories "{\"kind\":\"income\",\"name\":\"Smoke income $run_id\"}" "$smoke_dir/income-category.json" "income-category-$run_id"
income_category="$(jq -er '.category.id' "$smoke_dir/income-category.json")"
api_call POST /categories "{\"kind\":\"expense\",\"name\":\"Smoke expense $run_id\"}" "$smoke_dir/expense-category.json" "expense-category-$run_id"
expense_category="$(jq -er '.category.id' "$smoke_dir/expense-category.json")"

occurred_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
report_from="$(TZ=Asia/Ho_Chi_Minh date +%Y-%m-01)"
report_to="$(TZ=Asia/Ho_Chi_Minh date +%Y-%m-%d)"
api_call POST /transactions "{\"type\":\"income\",\"source_wallet_id\":\"$wallet_a\",\"category_id\":\"$income_category\",\"amount_vnd\":1000000,\"occurred_at\":\"$occurred_at\",\"note\":\"release smoke $run_id\"}" "$smoke_dir/income.json" "income-$run_id"
income_tx="$(jq -er '.transaction.id' "$smoke_dir/income.json")"
api_call POST /transactions "{\"type\":\"transfer\",\"source_wallet_id\":\"$wallet_a\",\"destination_wallet_id\":\"$wallet_b\",\"amount_vnd\":250000,\"occurred_at\":\"$occurred_at\",\"note\":\"release transfer $run_id\"}" "$smoke_dir/transfer.json" "transfer-$run_id"
transfer_tx="$(jq -er '.transaction.id' "$smoke_dir/transfer.json")"
api_call GET /wallets '' "$smoke_dir/wallets.json"
api_call GET "/reports/cash-flow?from=$report_from&to=$report_to" '' "$smoke_dir/report.json"
jq -e --arg a "$wallet_a" --arg b "$wallet_b" 'any(.wallets[]; .id == $a) and any(.wallets[]; .id == $b)' "$smoke_dir/wallets.json" >/dev/null
jq -e '.report.summary | .income_vnd == 1000000 and .expense_vnd == 0 and .net_income_vnd == 1000000' "$smoke_dir/report.json" >/dev/null

api_call POST /api-keys "{\"name\":\"release-smoke-$run_id\"}" "$smoke_dir/child-key.json" "child-key-$run_id"
child_key_id="$(jq -er '.key.id' "$smoke_dir/child-key.json")"
api_call POST "/api-keys/$child_key_id/revoke" '{}' "$smoke_dir/revoked.json" "revoke-$run_id"
revoked_plaintext="$(jq -er '.key.plaintext' "$smoke_dir/child-key.json")"
revoked_status="$(curl --silent --output /dev/null --write-out '%{http_code}' -H "Authorization: Bearer $revoked_plaintext" "$base_url/api/v1/me")"
if [[ "$revoked_status" != 401 ]]; then
  echo "Revoked API key returned unexpected HTTP status: $revoked_status" >&2
  exit 1
fi
unset revoked_plaintext
child_key_id=""

api_call POST /exports '{"datasets":["transactions"]}' "$smoke_dir/export.json" "export-$run_id"
export_id="$(jq -er '.job.id' "$smoke_dir/export.json")"
for _ in {1..60}; do
  api_call GET "/exports/$export_id" '' "$smoke_dir/export-status.json"
  export_status="$(jq -r '.job.status' "$smoke_dir/export-status.json")"
  [[ "$export_status" == completed ]] && break
  [[ "$export_status" == failed ]] && { echo "Export smoke failed" >&2; exit 1; }
  sleep 2
done
[[ "${export_status:-}" == completed ]] || { echo "Export smoke timed out" >&2; exit 1; }
api_call GET "/exports/$export_id/download" '' "$smoke_dir/export-download.json"
curl --fail --silent --show-error "$(jq -er '.download_url' "$smoke_dir/export-download.json")" >"$smoke_dir/export.csv"
test -s "$smoke_dir/export.csv"

receipt_checksum="$(shasum -a 256 "$MYPOCKET_SMOKE_RECEIPT_PATH" | awk '{print $1}')"
receipt_size="$(wc -c <"$MYPOCKET_SMOKE_RECEIPT_PATH" | tr -d ' ')"
receipt_name="$(basename "$MYPOCKET_SMOKE_RECEIPT_PATH")"
(( receipt_size > 0 && receipt_size <= 15728640 )) || { echo "Smoke receipt size is outside 1 byte–15 MiB" >&2; exit 1; }
api_call POST /files/presign "{\"filename\":$(jq -Rn --arg value "$receipt_name" '$value'),\"content_type\":\"$MYPOCKET_SMOKE_RECEIPT_CONTENT_TYPE\",\"size_bytes\":$receipt_size,\"checksum_sha256\":\"$receipt_checksum\"}" "$smoke_dir/presign.json" "receipt-$run_id"
receipt_id="$(jq -er '.file.id' "$smoke_dir/presign.json")"
curl --fail --silent --show-error -X PUT -H "Content-Type: $MYPOCKET_SMOKE_RECEIPT_CONTENT_TYPE" --data-binary "@$MYPOCKET_SMOKE_RECEIPT_PATH" "$(jq -er '.upload_url' "$smoke_dir/presign.json")" >/dev/null
api_call POST /agent/messages "{\"kind\":\"transaction_draft\",\"message\":\"Analyze this release smoke receipt and propose a review-only expense\",\"receipt_id\":\"$receipt_id\"}" "$smoke_dir/agent.json" "agent-$run_id"
agent_run_id="$(jq -er '.run.id' "$smoke_dir/agent.json")"
for _ in {1..90}; do
  api_call GET "/agent/runs/$agent_run_id" '' "$smoke_dir/agent-status.json"
  agent_status="$(jq -r '.run.status' "$smoke_dir/agent-status.json")"
  [[ "$agent_status" == completed ]] && break
  [[ "$agent_status" == failed ]] && { echo "Agent/OCR smoke failed" >&2; exit 1; }
  sleep 2
done
[[ "${agent_status:-}" == completed ]] || { echo "Agent/OCR smoke timed out" >&2; exit 1; }
agent_draft_id="$(jq -r '.run.draft_ids[0] // empty' "$smoke_dir/agent-status.json")"
[[ -n "$agent_draft_id" ]] || { echo "Agent completed without a reviewable draft" >&2; exit 1; }

echo "Production smoke passed for $base_url (run $run_id); cleanup will now archive/reject created finance state."
