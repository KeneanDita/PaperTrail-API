#!/usr/bin/env bash
set -euo pipefail

# PaperTrail API route smoke-test (curl-based)
# Usage:
#   PAPERTRAIL_BASE_URL=http://localhost:8080 ./scripts/test_routes.sh
#   PAPERTRAIL_JWT="<paste token>" ./scripts/test_routes.sh
#   SKIP_PRIVATE=1 ./scripts/test_routes.sh

BASE_URL="${PAPERTRAIL_BASE_URL:-http://localhost:8080}"
API_BASE="$BASE_URL/api"

curl_json() {
  local method="$1"; shift
  local url="$1"; shift
  local body="${1:-}"; shift || true

  echo
  echo "==> $method $url"

  local args=(-sS -X "$method" "$url" -H 'Accept: application/json')

  if [[ -n "${PAPERTRAIL_JWT:-}" ]]; then
    args+=(-H "Authorization: Bearer ${PAPERTRAIL_JWT}")
  fi

  if [[ -n "$body" ]]; then
    args+=(-H 'Content-Type: application/json' --data "$body")
  fi

  # Print body + status sentinel
  curl "${args[@]}" -w "\nHTTP_STATUS:%{http_code}\n"
}

# 1) Public health check
curl_json GET "$BASE_URL/health" | sed '/^HTTP_STATUS:/d'

# 2) Public user bootstrapping routes
email="curltest+$(date +%Y%m%d%H%M%S)@example.com"
user_resp=$(curl_json POST "$API_BASE/users" "{\"email\":\"$email\"}")
echo "$user_resp" | sed '/^HTTP_STATUS:/d'

if [[ "${SKIP_PRIVATE:-}" == "1" ]]; then
  echo
  echo "Skipping authenticated /api/* routes (SKIP_PRIVATE=1)."
  exit 0
fi

if [[ -z "${PAPERTRAIL_JWT:-}" ]]; then
  echo
  echo "No PAPERTRAIL_JWT set; skipping authenticated routes."
  exit 0
fi

# Papers
curl_json GET "$API_BASE/papers" | sed '/^HTTP_STATUS:/d'

# NOTE: The current codebase appears to use numeric IDs for papers/reviews/comments internally.
# These next calls assume paper id = 1 exists; they may 404/500 on a fresh DB.
curl_json GET "$API_BASE/papers/1" | sed '/^HTTP_STATUS:/d'

# Reviews
curl_json GET "$API_BASE/papers/1/reviews" | sed '/^HTTP_STATUS:/d'
curl_json POST "$API_BASE/papers/1/reviews" '{"reviewer_id":"1","rating":5,"comments":"Looks good"}' | sed '/^HTTP_STATUS:/d'

# Comments
curl_json GET "$API_BASE/papers/1/comments" | sed '/^HTTP_STATUS:/d'
curl_json POST "$API_BASE/papers/1/comments" '{"user_id":"1","body":"Nice paper"}' | sed '/^HTTP_STATUS:/d'

echo
echo "Done."