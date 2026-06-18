#!/usr/bin/env bash
# Ensure fixed dev accounts exist and export admin TOKEN / REFRESH.
# Usage: source scripts/lib/bootstrap-admin.sh && bootstrap_ensure_admin

bootstrap_register_if_missing() {
  local email="$1" password="$2" role="$3" display_name="$4"
  curl -sf -X POST "${GW:-http://localhost:8080}/api/v1/auth/register" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"$password\",\"role\":\"$role\",\"display_name\":\"$display_name\"}" \
    >/dev/null 2>&1 || true
}

bootstrap_ensure_dev_accounts() {
  local gw="${GW:-http://localhost:8080}"
  local root="${ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
  export GW="$gw"
  export ROOT="$root"

  export DEV_PASSWORD="${DEV_PASSWORD:-password123}"
  export DEV_TEACHER_EMAIL="${DEV_TEACHER_EMAIL:-teacher@even.local}"
  export DEV_STUDENT_EMAIL="${DEV_STUDENT_EMAIL:-student@even.local}"
  export PLATFORM_ADMIN_EMAIL="${PLATFORM_ADMIN_EMAIL:-admin@even.local}"
  export PLATFORM_ADMIN_PASSWORD="${PLATFORM_ADMIN_PASSWORD:-$DEV_PASSWORD}"

  bootstrap_register_if_missing "$DEV_TEACHER_EMAIL" "$DEV_PASSWORD" "teacher" "Dev Teacher"
  bootstrap_register_if_missing "$DEV_STUDENT_EMAIL" "$DEV_PASSWORD" "student" "Dev Student"
  bootstrap_register_if_missing "$PLATFORM_ADMIN_EMAIL" "$PLATFORM_ADMIN_PASSWORD" "teacher" "Platform Admin"

  docker compose -f "$root/docker-compose.yml" exec -T postgres psql -U "${POSTGRES_USER:-even}" -d even_auth -c \
    "UPDATE users SET is_admin=true WHERE email='$PLATFORM_ADMIN_EMAIL';" >/dev/null
}

bootstrap_login_admin() {
  local body
  body=$(curl -s -X POST "${GW:-http://localhost:8080}/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$PLATFORM_ADMIN_EMAIL\",\"password\":\"$PLATFORM_ADMIN_PASSWORD\"}")

  python3 -c "import json,sys; json.loads(sys.argv[1])['access_token']" "$body" >/dev/null \
    || { echo "bootstrap login failed: $body" >&2; return 1; }

  export TOKEN=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['access_token'])" "$body")
  export REFRESH=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['refresh_token'])" "$body")
  export USER_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['user']['id'])" "$body")
}

bootstrap_ensure_admin() {
  bootstrap_ensure_dev_accounts
  bootstrap_login_admin
}

bootstrap_print_dev_accounts() {
  echo "  Dev accounts (password: ${DEV_PASSWORD:-password123})"
  echo "    Teacher       ${DEV_TEACHER_EMAIL:-teacher@even.local}"
  echo "    Student       ${DEV_STUDENT_EMAIL:-student@even.local}"
  echo "    Teacher+Admin ${PLATFORM_ADMIN_EMAIL:-admin@even.local}"
}
