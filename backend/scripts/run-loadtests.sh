#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
export BASE_URL

k6 run ./loadtests/catalog.js
k6 run ./loadtests/detail.js
k6 run ./loadtests/search.js
k6 run ./loadtests/cart-checkout.js
k6 run ./loadtests/notification-fanout.js
