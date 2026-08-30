#!/usr/bin/env bash
set -euo pipefail

api_url="${MYPOCKET_SMOKE_API_URL:-http://127.0.0.1:8080}"
web_url="${MYPOCKET_SMOKE_WEB_URL:-http://127.0.0.1:5173}"
s3_url="${MYPOCKET_SMOKE_S3_URL:-http://127.0.0.1:4567}"

printf '%s\n' '== MyPocket platform smoke =='
curl -fsS "$api_url/api/v1/health/live" | tee /tmp/mypocket-health-live.json
curl -fsS "$api_url/api/v1/health/ready" | tee /tmp/mypocket-health-ready.json
curl -fsS "$web_url/api/v1/health/live" | tee /tmp/mypocket-web-proxy-health-live.json
curl -fsS "$s3_url/_localstack/health" | tee /tmp/mypocket-localstack-health.json

printf '%s\n' 'health checks passed'
