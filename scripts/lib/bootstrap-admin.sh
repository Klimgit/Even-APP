#!/usr/bin/env bash
# Ensure a fixed platform admin exists and export TOKEN / REFRESH.
# Usage: source scripts/lib/bootstrap-admin.sh && bootstrap_ensure_admin

bootstrap_ensure_admin() {
  local gw="${GW:-http://localhost:8080}"
  local root="${ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
  export PLATFORM_ADMIN_EMAIL="${PLATFORM_ADMIN_EMAIL:-platform-admin@even.local}"
  export PLATFORM_ADMIN_PASSWORD="${PLATFORM_ADMIN_PASSWORD:-password123}"
  export GW="$gw"
  export ROOT="$root"

  curl -sf -X POST "$gw/api/v1/auth/register" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$PLATFORM_ADMIN_EMAIL\",\"password\":\"$PLATFORM_ADMIN_PASSWORD\",\"role\":\"teacher\",\"display_name\":\"Platform Admin\"}" \
    >/dev/null 2>&1 || true

  docker compose -f "$root/docker-compose.yml" exec -T postgres psql -U "${POSTGRES_USER:-even}" -d even_auth -c \
    "UPDATE users SET is_admin=true WHERE email='$PLATFORM_ADMIN_EMAIL';" >/dev/null

  local body
  body=$(curl -s -X POST "$gw/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$PLATFORM_ADMIN_EMAIL\",\"password\":\"$PLATFORM_ADMIN_PASSWORD\"}")

  python3 -c "import json,sys; json.loads(sys.argv[1])['access_token']" "$body" >/dev/null \
    || { echo "bootstrap login failed: $body" >&2; return 1; }

  export TOKEN=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['access_token'])" "$body")
  export REFRESH=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['refresh_token'])" "$body")
  export USER_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['user']['id'])" "$body")
}
