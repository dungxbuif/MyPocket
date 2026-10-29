#!/usr/bin/env bash
set -euo pipefail

# Authenticated, read-only production smoke/regression harness for a dedicated
# UAT owner. It intentionally does not create, edit, or delete financial data.
# Use a throw-away UAT account and keep credentials in the environment only.
#
# Required: UAT_BASE_URL, UAT_EMAIL, UAT_PASSWORD
# Optional: UAT_MONTH (defaults to the account-local current month)

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v jq >/dev/null 2>&1 || fail 'jq is required'

base_url="${UAT_BASE_URL:-}"
email="${UAT_EMAIL:-}"
password="${UAT_PASSWORD:-}"
month="${UAT_MONTH:-2026-10}"
conversation_id="${UAT_CONVERSATION_ID:-}"
[[ -n "$base_url" ]] || fail 'UAT_BASE_URL is required'
[[ -n "$email" ]] || fail 'UAT_EMAIL is required'
[[ -n "$password" ]] || fail 'UAT_PASSWORD is required'
base_url="${base_url%/}"

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-uat-regression.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

auth_file="$tmp_dir/auth.json"
login_code="$(curl --silent --show-error --output "$auth_file" --write-out '%{http_code}' \
  -H 'Content-Type: application/json' \
  --data "$(jq -cn --arg email "$email" --arg password "$password" '{email:$email,password:$password}')" \
  "$base_url/api/v1/login")" || fail 'login request failed'
[[ "$login_code" == 2* ]] || fail "login returned HTTP $login_code"
token="$(jq -er '.data.token // .token' "$auth_file")" || fail 'login response did not contain a token'

check_get() {
  local name endpoint output code
  name="$1"
  endpoint="$2"
  output="$tmp_dir/$(printf '%s' "$name" | tr '/?&={} ' '_').json"
  code="$(curl --silent --show-error --output "$output" --write-out '%{http_code}' \
    -H "Authorization: Bearer $token" -H 'Accept: application/json' \
    "$base_url$endpoint")" || fail "$name request failed"
  [[ "$code" == 2* ]] || { jq -c 'if .detail then {code:.code,detail:.detail} else . end' "$output" >&2 || true; fail "$name returned HTTP $code"; }
  printf 'PASS %s (%s)\n' "$name" "$code"
}

# Core authenticated surfaces.
check_get profile '/api/v1/auth/profile'
check_get home '/api/v1/home'
check_get wallets '/api/v1/wallets'
check_get categories '/api/v1/categories'
check_get transactions '/api/v1/transactions?limit=1&offset=0'
check_get reports_savings "/api/v1/reports/savings?month=$month"
check_get reports_savings_transactions "/api/v1/reports/savings/transactions?month=$month&section=surplus&kind=all&limit=1&offset=0"
check_get reports_insider "/api/v1/reports/insider?month=$month"
check_get month "/api/v1/months/$month"
check_get jars "/api/v1/jars?month=$month"
check_get recurring '/api/v1/recurring'
check_get travel_events '/api/v1/travel/events'
check_get portfolio_assets '/api/v1/portfolio/assets'
check_get portfolio_summary '/api/v1/portfolio/summary'
check_get notifications '/api/v1/notifications'
check_get api_keys '/api/v1/api-keys'
check_get feedback '/api/v1/feedback'
check_get changelog '/api/v1/changelog'
check_get sync_capabilities '/api/v1/sync/capabilities'
check_get ai_advisor_capabilities '/api/v1/ai/advisor/capabilities'
check_get ai_advisor_overview "/api/v1/ai/advisor/overview?month=$month"
if [[ -n "$conversation_id" ]]; then
  check_get ai_advisor_messages "/api/v1/ai/advisor/conversation/messages?conversation_id=$conversation_id"
else
  printf 'SKIP ai_advisor_messages (set UAT_CONVERSATION_ID after a read-only advisor session)\n'
fi
check_get ai_entry_capabilities '/api/v1/ai/entry/capabilities'
check_get budget_series '/api/v1/budget-series'
check_get budgets '/api/v1/budgets'

# Exercise parameterized read paths only when the UAT account has the relevant
# fixture. This keeps the harness safe and avoids synthetic writes.
wallets_file="$tmp_dir/wallets.json"
curl --silent --show-error -H "Authorization: Bearer $token" "$base_url/api/v1/wallets" >"$wallets_file"
while IFS= read -r wallet_id; do
  [[ -n "$wallet_id" ]] || continue
  check_get "credit_statement_$wallet_id" "/api/v1/credit/wallets/$wallet_id/statement?month=$month"
done < <(jq -r '.data[]? | select(.type == "credit") | .id' "$wallets_file")

while IFS= read -r jar_id; do
  [[ -n "$jar_id" ]] || continue
  check_get "jar_report_$jar_id" "/api/v1/jars/$jar_id/report?month=$month"
  check_get "jar_month_$jar_id" "/api/v1/jars/$jar_id/months/$month"
done < <(curl --silent --show-error -H "Authorization: Bearer $token" \
  "$base_url/api/v1/jars?month=$month" | jq -r '.data.jars[]?.id')

while IFS= read -r budget_id; do
  [[ -n "$budget_id" ]] || continue
  check_get "budget_report_$budget_id" "/api/v1/budgets/$budget_id/report?month=$month"
  check_get "budget_transactions_$budget_id" "/api/v1/budgets/$budget_id/transactions?month=$month&limit=1&offset=0"
done < <(curl --silent --show-error -H "Authorization: Bearer $token" \
  "$base_url/api/v1/budgets" | jq -r '.data.items[]?.id')

printf 'UAT regression complete: authenticated GET-only coverage passed; no production mutation was attempted.\n'
