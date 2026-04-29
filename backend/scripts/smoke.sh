#!/usr/bin/env bash
set -euo pipefail

: "${SMOKE_BASE_URL:?SMOKE_BASE_URL is required}"

curl -fsS "${SMOKE_BASE_URL}/health" >/dev/null
curl -fsS "${SMOKE_BASE_URL}/ready" >/dev/null
curl -fsS "${SMOKE_BASE_URL}/version" >/dev/null
