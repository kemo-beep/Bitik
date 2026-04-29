#!/usr/bin/env bash
set -euo pipefail

: "${DEPLOY_HOST:?DEPLOY_HOST is required}"
: "${DEPLOY_USER:?DEPLOY_USER is required}"
: "${DEPLOY_PATH:?DEPLOY_PATH is required}"
: "${IMAGE_TAG:?IMAGE_TAG is required}"

ssh "${DEPLOY_USER}@${DEPLOY_HOST}" <<EOF
set -euo pipefail
mkdir -p "${DEPLOY_PATH}"
cd "${DEPLOY_PATH}"
export IMAGE_TAG="${IMAGE_TAG}"
docker compose pull
docker compose up -d
EOF
