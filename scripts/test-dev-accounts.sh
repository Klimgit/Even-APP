#!/usr/bin/env bash
# Verify fixed dev accounts exist after seed-dev and have expected roles.
set -euo pipefail

GW="${GW:-http://localhost:8080}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BODY=/tmp/test-dev-accounts-body.json
PASS=0
FAIL=0

pass() { echo "  ✓ $1" >&2; PASS=$((PASS + 1)); }
fail() { echo "  ✗ $1" >&2; FAIL=$((FAIL + 1)); }
expect() { local want=$1 got=$2 msg=$3; [[ "$got" == "$want" ]] && pass "$msg → $got" || fail "$msg expected $want got $got"; }
code() { curl -s -o "$BODY" -w "%{http_code}" "$@"; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_dev_accounts

DEV_PASSWORD="${DEV_PASSWORD:-password123}"
TEACHER="${DEV_TEACHER_EMAIL:-teacher@even.local}"
STUDENT="${DEV_STUDENT_EMAIL:-student@even.local}"
ADMIN="${PLATFORM_ADMIN_EMAIL:-admin@even.local}"

echo "=== Dev accounts ==="

login_as() {
  local email=$1
  c=$(code -X POST "$GW/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"$DEV_PASSWORD\"}")
  expect 200 "$c" "login $email"
  python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['access_token'])" "$BODY"
}

TTOKEN=$(login_as "$TEACHER")
c=$(code -H "Authorization: Bearer $TTOKEN" "$GW/api/v1/auth/me"); expect 200 "$c" "GET /auth/me teacher"
ROLE=$(python3 -c "import json; print(json.load(open('$BODY'))['role'])")
IS_ADMIN=$(python3 -c "import json; print(json.load(open('$BODY'))['is_admin'])")
[[ "$ROLE" == "teacher" ]] && pass "teacher role" || fail "teacher role ($ROLE)"
[[ "$IS_ADMIN" == "False" ]] && pass "teacher not admin" || fail "teacher is_admin=$IS_ADMIN"

STOKEN=$(login_as "$STUDENT")
c=$(code -H "Authorization: Bearer $STOKEN" "$GW/api/v1/auth/me"); expect 200 "$c" "GET /auth/me student"
ROLE=$(python3 -c "import json; print(json.load(open('$BODY'))['role'])")
[[ "$ROLE" == "student" ]] && pass "student role" || fail "student role ($ROLE)"

ATOKEN=$(login_as "$ADMIN")
c=$(code -H "Authorization: Bearer $ATOKEN" "$GW/api/v1/auth/me"); expect 200 "$c" "GET /auth/me admin"
ROLE=$(python3 -c "import json; print(json.load(open('$BODY'))['role'])")
IS_ADMIN=$(python3 -c "import json; print(json.load(open('$BODY'))['is_admin'])")
[[ "$ROLE" == "teacher" ]] && pass "admin role teacher" || fail "admin role ($ROLE)"
[[ "$IS_ADMIN" == "True" ]] && pass "admin is_admin" || fail "admin is_admin=$IS_ADMIN"

c=$(code -X POST "$GW/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$STUDENT\",\"password\":\"$DEV_PASSWORD\",\"role\":\"student\"}")
expect 409 "$c" "duplicate register student rejected"

c=$(code -H "Authorization: Bearer $STOKEN" "$GW/api/v1/auth/demo/teacher"); expect 403 "$c" "student demo/teacher forbidden"

c=$(code -H "Authorization: Bearer $TTOKEN" "$GW/api/v1/auth/demo/teacher"); expect 200 "$c" "teacher demo/teacher ok"
c=$(code -H "Authorization: Bearer $ATOKEN" "$GW/api/v1/auth/demo/admin/stats"); expect 200 "$c" "admin demo/admin/stats ok"
c=$(code -H "Authorization: Bearer $TTOKEN" "$GW/api/v1/auth/demo/admin/stats"); expect 403 "$c" "teacher demo/admin forbidden"

echo ""
if [[ "$FAIL" -eq 0 ]]; then
  echo "=== DEV ACCOUNTS OK ($PASS checks) ==="
else
  echo "=== DEV ACCOUNTS FAILED ($PASS passed, $FAIL failed) ==="
  exit 1
fi
