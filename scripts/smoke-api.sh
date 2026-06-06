#!/usr/bin/env bash
# Smoke test for implemented auth + platform media endpoints.
set -euo pipefail

GW="${GW:-http://localhost:8080}"
AUTH="${AUTH:-http://localhost:8081}"
LEX="${LEX:-http://localhost:8082}"
EMAIL="smoke-$(date +%s)@example.com"
PASS="password123"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }
code() { curl -s -o /tmp/smoke-body.json -w "%{http_code}" "$@"; }

echo "=== Health ==="
for url in "$GW/health" "$AUTH/health" "$LEX/health"; do
  c=$(code "$url"); [[ "$c" == "200" ]] && pass "GET $url → $c" || fail "GET $url → $c"
done

echo ""
echo "=== Auth (via gateway $GW) ==="

c=$(code -X POST "$GW/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"role\":\"teacher\",\"display_name\":\"Smoke\"}")
[[ "$c" == "201" ]] && pass "POST /auth/register → $c" || fail "POST /auth/register → $c ($(cat /tmp/smoke-body.json))"
ACCESS=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['access_token'])")
REFRESH=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['refresh_token'])")
USER_ID=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['user']['id'])")

c=$(code -X POST "$GW/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
[[ "$c" == "200" ]] && pass "POST /auth/login → $c" || fail "POST /auth/login → $c"

c=$(code -H "Authorization: Bearer $ACCESS" "$GW/api/v1/auth/me")
[[ "$c" == "200" ]] && pass "GET /auth/me → $c" || fail "GET /auth/me → $c"

c=$(code -X POST "$GW/api/v1/auth/refresh" -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
[[ "$c" == "200" ]] && pass "POST /auth/refresh → $c" || fail "POST /auth/refresh → $c"
ACCESS=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['access_token'])")

echo ""
echo "=== Promote to platform admin (DB) ==="
docker compose exec -T postgres psql -U "${POSTGRES_USER:-even}" -d even_auth -c \
  "UPDATE users SET is_admin=true WHERE id='$USER_ID';" >/dev/null
c=$(code -X POST "$AUTH/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
ACCESS=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['access_token'])")
pass "admin token issued for $EMAIL"

echo ""
echo "=== Platform media (lexicon $LEX, gateway $GW) ==="

c=$(code -H "Authorization: Bearer $ACCESS" -X POST "$LEX/api/v1/platform/media/presign" \
  -H 'Content-Type: application/json' \
  -d '{"filename":"smoke.png","mime_type":"image/png","size_bytes":68}')
[[ "$c" == "200" ]] && pass "POST /platform/media/presign → $c" || fail "POST /platform/media/presign → $c ($(cat /tmp/smoke-body.json))"
OBJ=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['object_key'])")
UPLOAD=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['upload_url'])")
MEDIA_ID=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['media_asset_id'])")

# 1x1 PNG
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xdb\x00\x00\x00\x00IEND\xaeB`\x82' > /tmp/smoke.png
curl -sf -X PUT "$UPLOAD" -H 'Content-Type: image/png' --data-binary @/tmp/smoke.png >/dev/null
pass "PUT to MinIO presigned URL"

c=$(code -H "Authorization: Bearer $ACCESS" -X POST "$LEX/api/v1/platform/media/confirm" \
  -H 'Content-Type: application/json' \
  -d "{\"object_key\":\"$OBJ\",\"mime_type\":\"image/png\",\"size_bytes\":68,\"display_name\":\"Smoke PNG\",\"ttl_seconds\":86400}")
[[ "$c" == "201" ]] && pass "POST /platform/media/confirm → $c" || fail "POST /platform/media/confirm → $c ($(cat /tmp/smoke-body.json))"

c=$(code -H "Authorization: Bearer $ACCESS" "$LEX/api/v1/platform/languages/evn/media")
[[ "$c" == "200" ]] && pass "GET /platform/languages/evn/media → $c" || fail "GET /platform/languages/evn/media → $c"
TOTAL=$(python3 -c "import json; print(json.load(open('/tmp/smoke-body.json'))['total'])")
[[ "$TOTAL" -ge 1 ]] && pass "list total=$TOTAL" || fail "list empty"

c=$(code -H "Authorization: Bearer $ACCESS" "$LEX/api/v1/platform/media/$MEDIA_ID")
[[ "$c" == "200" ]] && pass "GET /platform/media/{id} → $c" || fail "GET /platform/media/{id} → $c"

c=$(code -H "Authorization: Bearer $ACCESS" -X PATCH "$LEX/api/v1/platform/media/$MEDIA_ID" \
  -H 'Content-Type: application/json' -d '{"display_name":"Smoke PNG renamed"}')
[[ "$c" == "200" ]] && pass "PATCH /platform/media/{id} → $c" || fail "PATCH /platform/media/{id} → $c"

c=$(code -H "Authorization: Bearer $ACCESS" -X DELETE "$LEX/api/v1/platform/media/$MEDIA_ID")
[[ "$c" == "204" ]] && pass "DELETE /platform/media/{id} → $c" || fail "DELETE /platform/media/{id} → $c"

# Gateway proxy to platform
c=$(code -H "Authorization: Bearer $ACCESS" -X POST "$GW/api/v1/platform/media/presign" \
  -H 'Content-Type: application/json' \
  -d '{"filename":"gw.png","mime_type":"image/png","size_bytes":68}')
[[ "$c" == "200" ]] && pass "POST /platform/media/presign via gateway → $c" || fail "gateway presign → $c"

echo ""
echo "=== All smoke tests passed ==="
