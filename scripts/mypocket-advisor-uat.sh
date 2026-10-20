#!/usr/bin/env bash
set -euo pipefail

# Explicit live-provider UAT for the Finance Assistant. This is intentionally
# separate from the read-only regression harness because POSTing a question
# may consume provider quota. Credentials remain in the environment and are
# never printed.
#
# Required: UAT_BASE_URL, UAT_EMAIL, UAT_PASSWORD
# Optional: UAT_ADVISOR_PROMPT, UAT_TIMEOUT_SECONDS (default 90),
#           UAT_POLL_SECONDS (default 2), UAT_CLIENT_REQUEST_ID. For a
#           completed submission whose temporary evidence was lost, set both
#           UAT_RUN_ID and UAT_CONVERSATION_ID to verify it without another
#           provider call.

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v jq >/dev/null 2>&1 || fail 'jq is required'

base_url="${UAT_BASE_URL:-}"
email="${UAT_EMAIL:-}"
password="${UAT_PASSWORD:-}"
prompt="${UAT_ADVISOR_PROMPT:-Tổng chi tiêu tháng này của tôi là bao nhiêu? Chỉ trả lời dựa trên dữ liệu tài chính của tài khoản này.}"
timeout_seconds="${UAT_TIMEOUT_SECONDS:-90}"
poll_seconds="${UAT_POLL_SECONDS:-2}"
request_id="${UAT_CLIENT_REQUEST_ID:-mypocket-uat-advisor-$(date +%s)-$RANDOM}"
run_id="${UAT_RUN_ID:-}"
conversation_id="${UAT_CONVERSATION_ID:-}"
[[ -n "$base_url" ]] || fail 'UAT_BASE_URL is required'
[[ -n "$email" ]] || fail 'UAT_EMAIL is required'
[[ -n "$password" ]] || fail 'UAT_PASSWORD is required'
[[ "$timeout_seconds" =~ ^[0-9]+$ && "$timeout_seconds" -gt 0 ]] || fail 'UAT_TIMEOUT_SECONDS must be a positive integer'
[[ "$poll_seconds" =~ ^[0-9]+$ && "$poll_seconds" -gt 0 ]] || fail 'UAT_POLL_SECONDS must be a positive integer'
base_url="${base_url%/}"

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/mypocket-advisor-uat.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

auth_file="$tmp_dir/auth.json"
login_code="$(curl --silent --show-error --output "$auth_file" --write-out '%{http_code}' \
  -H 'Content-Type: application/json' \
  --data "$(jq -cn --arg email "$email" --arg password "$password" '{email:$email,password:$password}')" \
  "$base_url/api/v1/login")" || fail 'login request failed'
[[ "$login_code" == 2* ]] || fail "login returned HTTP $login_code"
token="$(jq -er '.data.token // .token' "$auth_file")" || fail 'login response did not contain a token'

if [[ -z "$run_id" && -z "$conversation_id" ]]; then
  submit_file="$tmp_dir/submit.json"
  submit_code="$(curl --silent --show-error --output "$submit_file" --write-out '%{http_code}' \
    -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
    --data "$(jq -cn --arg id "$request_id" --arg text "$prompt" '{client_request_id:$id,text:$text}')" \
    "$base_url/api/v1/ai/advisor/messages/async")" || fail 'advisor submit request failed'
  [[ "$submit_code" == 2* ]] || { printf 'advisor submit returned HTTP %s\n' "$submit_code" >&2; exit 1; }
  run_id="$(jq -er '.data.run_id // .run_id' "$submit_file")" || fail 'submit response did not contain run_id'
  conversation_id="$(jq -er '.data.conversation_id // .conversation_id' "$submit_file")" || fail 'submit response did not contain conversation_id'
  printf 'SUBMITTED run_id=%s conversation_id=%s\n' "$run_id" "$conversation_id"
elif [[ -z "$run_id" || -z "$conversation_id" ]]; then
  fail 'UAT_RUN_ID and UAT_CONVERSATION_ID must be provided together'
else
  printf 'VERIFY_ONLY run_id=%s conversation_id=%s\n' "$run_id" "$conversation_id"
fi

run_file="$tmp_dir/run.json"
deadline=$(( $(date +%s) + timeout_seconds ))
terminal=''
while (( $(date +%s) < deadline )); do
  run_code="$(curl --silent --show-error --output "$run_file" --write-out '%{http_code}' \
    -H "Authorization: Bearer $token" -H 'Accept: application/json' \
    "$base_url/api/v1/ai/advisor/runs/$run_id")" || fail 'advisor run poll failed'
  [[ "$run_code" == 2* ]] || fail "advisor run returned HTTP $run_code"
  terminal="$(jq -r '.data.status // .status // empty' "$run_file")"
  case "$terminal" in
    completed|failed|cancelled|interrupted|purged) break;;
  esac
  sleep "$poll_seconds"
done
[[ "$terminal" == completed ]] || fail "advisor run did not complete (status=${terminal:-timeout}); no automatic resubmission was attempted"

history_file="$tmp_dir/history.json"
history_code="$(curl --silent --show-error --output "$history_file" --write-out '%{http_code}' \
  -H "Authorization: Bearer $token" -H 'Accept: application/json' \
  "$base_url/api/v1/ai/advisor/conversation/messages?conversation_id=$conversation_id")" || fail 'advisor history request failed'
[[ "$history_code" == 2* ]] || fail "advisor history returned HTTP $history_code"
message_count="$(jq -r '(.data // .messages // []) | length' "$history_file")"
(( message_count >= 2 )) || fail "advisor history has fewer than two messages (count=$message_count)"
printf 'PASS advisor run completed; conversation_id=%s messages=%s\n' "$conversation_id" "$message_count"
printf 'Evidence is stored only in a temporary directory and is removed on exit.\n'
