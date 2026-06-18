#!/usr/bin/env bash
# Learning outline + progress smoke using fixed dev student account.
set -euo pipefail

GW="${GW:-http://localhost:8080}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BODY=/tmp/test-learning-outline.json

pass() { echo "  ✓ $1" >&2; }
fail() { echo "  ✗ $1" >&2; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin >/dev/null
# shellcheck source=seed-znakomstvo.sh
source "$ROOT/scripts/seed-znakomstvo.sh" >/dev/null

DEV_PASSWORD="${DEV_PASSWORD:-password123}"
STUDENT="${DEV_STUDENT_EMAIL:-student@even.local}"

STOKEN=$(curl -sf -X POST "$GW/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$STUDENT\",\"password\":\"$DEV_PASSWORD\"}" | python3 -c "import json,sys; print(json.load(sys.stdin)['access_token'])")

curl -sf -X POST "$GW/api/v1/courses/join" -H "Authorization: Bearer $STOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"invite_code\":\"$ZNAKOMSTVO_INVITE_CODE\"}" >/dev/null 2>&1 || true

code=$(curl -s -o "$BODY" -w "%{http_code}" -H "Authorization: Bearer $STOKEN" \
  "$GW/api/v1/courses/$ZNAKOMSTVO_COURSE_ID/outline")
[[ "$code" == "200" ]] || fail "GET outline → $code ($(cat "$BODY"))"

python3 - "$BODY" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
assert d.get("course_id"), "missing course_id"
assert d.get("lessons"), "missing lessons"
lesson = d["lessons"][0]
assert lesson.get("sections"), "missing sections"
blocks = lesson["sections"][0].get("blocks") or []
assert blocks, "missing blocks"
assert any(b.get("is_gradable") for b in blocks), "expected gradable block"
print("outline ok:", d["title"], "blocks", len(blocks))
PY

pass "course outline tree for enrolled student"

# Wrong block attempt then outline progress update
code=$(curl -s -o "$BODY" -w "%{http_code}" -H "Authorization: Bearer $STOKEN" \
  -X POST "$GW/api/v1/progress/blocks/$ZNAKOMSTVO_GRADABLE_BLOCK_ID/attempt" \
  -H 'Content-Type: application/json' -d '{"response":{"selected_index":1}}')
[[ "$code" == "200" ]] || fail "wrong attempt → $code"
python3 -c "import json; d=json.load(open('$BODY')); assert d['is_correct'] is False" || fail "wrong attempt should score incorrect"

code=$(curl -s -o "$BODY" -w "%{http_code}" -H "Authorization: Bearer $STOKEN" \
  -X POST "$GW/api/v1/progress/blocks/$ZNAKOMSTVO_GRADABLE_BLOCK_ID/attempt" \
  -H 'Content-Type: application/json' -d '{"response":{"selected_index":0}}')
[[ "$code" == "200" ]] || fail "correct attempt → $code"
python3 -c "import json; d=json.load(open('$BODY')); assert d['is_correct'] is True" || fail "correct attempt should score correct"

pass "gradable block attempt grading"

code=$(curl -s -o "$BODY" -w "%{http_code}" -H "Authorization: Bearer $STOKEN" \
  "$GW/api/v1/progress/lessons/$ZNAKOMSTVO_LESSON_ID")
[[ "$code" == "200" ]] || fail "lesson progress → $code"

pass "lesson progress after attempts"
echo "=== LEARNING OUTLINE OK ==="
