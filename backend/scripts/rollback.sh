#!/usr/bin/env bash
set -euo pipefail

: "${DEPLOY_HOST:?DEPLOY_HOST is required}"
: "${DEPLOY_USER:?DEPLOY_USER is required}"
: "${DEPLOY_PATH:?DEPLOY_PATH is required}"
: "${ROLLBACK_IMAGE_TAG:?ROLLBACK_IMAGE_TAG is required}"

ssh "${DEPLOY_USER}@${DEPLOY_HOST}" <<EOF
set -euo pipefail
cd "${DEPLOY_PATH}"
export IMAGE_TAG="${ROLLBACK_IMAGE_TAG}"
docker compose up -d
EOF
