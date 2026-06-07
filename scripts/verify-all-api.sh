#!/usr/bin/env bash
set -euo pipefail

GW="${GW:-http://localhost:8080}"
LEX="${LEX:-http://localhost:8082}"
MEDIA="${MEDIA:-http://localhost:8085}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; FAIL=1; }
expect() { local want=$1 got=$2 msg=$3; [[ "$got" == "$want" ]] && pass "$msg → $got" || fail "$msg expected $want got $got"; }

FAIL=0
BODY=/tmp/verify-api-body.json

code() { curl -s -o "$BODY" -w "%{http_code}" "$@"; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin || { fail "bootstrap admin"; exit 1; }
pass "login platform-admin@even.local"

export EVN_MEDIA_ID=$(docker compose exec -T postgres psql -U even -d even_media -tAc "SELECT id FROM languages WHERE code='evn'")
export RU_LEX_ID=$(docker compose exec -T postgres psql -U even -d even_lexicon -tAc "SELECT id FROM languages WHERE code='ru'")

echo "=== 1. System / health ==="
for url in "$GW/health" "$GW/api/v1/ready" "$GW/api/v1/gateway/status" \
  "http://localhost:8081/health" "http://localhost:8082/health" "http://localhost:8085/health" \
  "http://localhost:8083/health" "http://localhost:8084/health"; do
  c=$(code "$url"); expect 200 "$c" "GET $url"
done
c=$(code "$GW/api/v1/openapi.yaml"); [[ "$c" == "200" ]] && pass "GET openapi.yaml → 200" || fail "openapi → $c"

echo ""
echo "=== 2. Auth ==="
STU_EMAIL="verify-stu-$RANDOM@example.com"
c=$(code -X POST "$GW/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$STU_EMAIL\",\"password\":\"password123\",\"role\":\"student\"}")
expect 201 "$c" "POST /auth/register"

c=$(code -X POST "$GW/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$STU_EMAIL\",\"password\":\"password123\"}")
expect 200 "$c" "POST /auth/login student"
STOKEN=$(python3 -c "import json; print(json.load(open('$BODY'))['access_token'])")

c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/auth/me")
expect 200 "$c" "GET /auth/me admin"

c=$(code -X POST "$GW/api/v1/auth/refresh" -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
expect 200 "$c" "POST /auth/refresh"
TOKEN=$(python3 -c "import json; print(json.load(open('$BODY'))['access_token'])")

echo ""
echo "=== 3. Auth demo ==="
c=$(code "$GW/api/v1/auth/demo/public"); expect 200 "$c" "GET demo/public"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/auth/demo/me"); expect 200 "$c" "GET demo/me"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/auth/demo/teacher"); expect 200 "$c" "GET demo/teacher"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/auth/demo/admin/stats"); expect 200 "$c" "GET demo/admin/stats"
c=$(code -H "Authorization: Bearer $STOKEN" "$GW/api/v1/auth/demo/admin/stats"); expect 403 "$c" "GET demo/admin/stats student"

echo ""
echo "=== 4. Public languages ==="
c=$(code "$GW/languages"); expect 200 "$c" "GET /languages"
python3 -c "import json; d=json.load(open('$BODY')); assert any(x['code']=='evn' for x in d)" || fail "languages missing evn"
pass "languages contains evn"
c=$(code "$GW/languages/evn"); expect 200 "$c" "GET /languages/evn"
c=$(code "$GW/languages/evn/alphabet"); expect 200 "$c" "GET /languages/evn/alphabet"
python3 -c "import json; d=json.load(open('$BODY')); assert len(d)==36" || fail "alphabet count"
pass "alphabet has 36 letters"
c=$(code "$GW/languages/zzz"); expect 404 "$c" "GET /languages/zzz"

echo ""
echo "=== 5. Platform languages ==="
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages"); expect 200 "$c" "GET platform/languages"
V_CODE="vfy$RANDOM"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages" \
  -H 'Content-Type: application/json' \
  -d "{\"code\":\"$V_CODE\",\"name\":\"Verify\",\"native_name\":\"V\",\"direction\":\"ltr\"}")
expect 201 "$c" "POST platform/languages"
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/languages/$V_CODE" \
  -H 'Content-Type: application/json' -d '{"name":"Verify Updated"}')
expect 200 "$c" "PATCH platform/languages/{code}"

echo ""
echo "=== 6. Platform alphabet ($V_CODE) ==="
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"в","upper_char":"В","sort_order":1}')
expect 201 "$c" "POST letter в"
ID_V=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"а","upper_char":"А","sort_order":2}')
expect 201 "$c" "POST letter а"
ID_A=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"ӈ","upper_char":"Ӈ","sort_order":3,"transcription":"ŋ"}')
expect 201 "$c" "POST letter ӈ"
ID_NG=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages/$V_CODE/alphabet"); expect 200 "$c" "GET platform alphabet"
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/alphabet/$ID_A" \
  -H 'Content-Type: application/json' -d '{"transcription":"a"}'); expect 200 "$c" "PATCH alphabet letter"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/alphabet/reorder" \
  -H 'Content-Type: application/json' \
  -d "{\"letter_ids\":[\"$ID_A\",\"$ID_NG\",\"$ID_V\"]}"); expect 200 "$c" "POST alphabet reorder"
c=$(code "$GW/languages/$V_CODE/alphabet"); expect 200 "$c" "GET public alphabet $V_CODE"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/alphabet/$ID_V" \
  -H "Authorization: Bearer $TOKEN"); expect 204 "$c" "DELETE alphabet letter"

echo ""
echo "=== 7. Platform sounds ($V_CODE) ==="
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/sounds" \
  -H 'Content-Type: application/json' -d '{"ipa":"/a/","description":"vowel"}'); expect 201 "$c" "POST sound"
SOUND_ID=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages/$V_CODE/sounds"); expect 200 "$c" "GET sounds"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/alphabet/$ID_NG/sounds" \
  -H 'Content-Type: application/json' -d "{\"sound_id\":\"$SOUND_ID\"}"); expect 204 "$c" "link letter-sound"
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/sounds/$SOUND_ID" \
  -H 'Content-Type: application/json' -d '{"description":"upd"}'); expect 200 "$c" "PATCH sound"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/alphabet/$ID_NG/sounds/$SOUND_ID"); expect 204 "$c" "unlink letter-sound"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/sounds/$SOUND_ID"); expect 204 "$c" "DELETE sound"

echo ""
echo "=== 8. Platform lexicon ($V_CODE) ==="
LEMMA="vfy$RANDOM"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/languages/$V_CODE/lexicon" \
  -H 'Content-Type: application/json' \
  -d "{\"lemma\":\"$LEMMA\",\"translations\":[{\"target_language_id\":\"$RU_LEX_ID\",\"text\":\"тест\"}]}")
expect 201 "$c" "POST lexeme"
LEXEME_ID=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
TRANS_ID=$(python3 -c "import json; print(json.load(open('$BODY'))['translations'][0]['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/lexemes/$LEXEME_ID/forms" \
  -H 'Content-Type: application/json' -d '{"form":"формы","tags":{"n":"pl"}}'); expect 201 "$c" "POST form"
FORM_ID=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/lexemes/$LEXEME_ID" \
  -H 'Content-Type: application/json' -d '{"notes":"n"}'); expect 200 "$c" "PATCH lexeme"
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/lexeme-forms/$FORM_ID" \
  -H 'Content-Type: application/json' -d '{"form":"форма"}'); expect 200 "$c" "PATCH form"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/lexemes/$LEXEME_ID/translations" \
  -H 'Content-Type: application/json' \
  -d "{\"target_language_id\":\"$RU_LEX_ID\",\"text\":\"ещё\"}"); expect 201 "$c" "POST translation"
TRANS2=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
FAKE="00000000-0000-4000-8000-$(printf '%012x' $RANDOM$RANDOM)"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/lexemes/$LEXEME_ID/media" \
  -H 'Content-Type: application/json' \
  -d "{\"media_asset_id\":\"$FAKE\",\"kind\":\"image\"}"); expect 201 "$c" "POST lexeme media"
LMEDIA=$(python3 -c "import json; print(json.load(open('$BODY'))['id'])")
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages/$V_CODE/lexicon?q=$LEMMA"); expect 200 "$c" "GET lexicon list"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/lexemes/$LEXEME_ID"); expect 200 "$c" "GET lexeme"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/lexeme-media/$LMEDIA"); expect 204 "$c" "DELETE lexeme media"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/lexeme-translations/$TRANS2"); expect 204 "$c" "DELETE translation"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/lexeme-forms/$FORM_ID"); expect 204 "$c" "DELETE form"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/lexemes/$LEXEME_ID"); expect 204 "$c" "DELETE lexeme"

echo ""
echo "=== 9. Platform media ==="
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/media/presign" \
  -H 'Content-Type: application/json' \
  -d "{\"filename\":\"v.png\",\"mime_type\":\"image/png\",\"size_bytes\":68,\"language_id\":\"$EVN_MEDIA_ID\"}")
expect 200 "$c" "POST media presign"
OBJ=$(python3 -c "import json; print(json.load(open('$BODY'))['object_key'])")
UP=$(python3 -c "import json; print(json.load(open('$BODY'))['upload_url'])")
MID=$(python3 -c "import json; print(json.load(open('$BODY'))['media_asset_id'])")
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xdb\x00\x00\x00\x00IEND\xaeB`\x82' > /tmp/verify.png
c=$(code -X PUT "$UP" -H 'Content-Type: image/png' --data-binary @/tmp/verify.png); expect 200 "$c" "PUT MinIO"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$GW/api/v1/platform/media/confirm" \
  -H 'Content-Type: application/json' \
  -d "{\"object_key\":\"$OBJ\",\"mime_type\":\"image/png\",\"size_bytes\":68,\"display_name\":\"Verify\",\"language_id\":\"$EVN_MEDIA_ID\",\"ttl_seconds\":86400}")
expect 201 "$c" "POST media confirm"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages/evn/media"); expect 200 "$c" "GET languages/evn/media"
c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/media/$MID"); expect 200 "$c" "GET media/{id}"
c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$GW/api/v1/platform/media/$MID" \
  -H 'Content-Type: application/json' -d '{"display_name":"V2"}'); expect 200 "$c" "PATCH media"
c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$GW/api/v1/platform/media/$MID"); expect 204 "$c" "DELETE media"

echo ""
echo "=== 10. Auth guards ==="
c=$(code -H "Authorization: Bearer invalid" "$GW/api/v1/platform/languages")
[[ "$c" == "401" || "$c" == "500" ]] && pass "invalid JWT → $c" || fail "invalid JWT → $c"
c=$(code -H "Authorization: Bearer $STOKEN" "$GW/api/v1/platform/languages"); expect 403 "$c" "student platform"

echo ""
echo "=== 11. Cleanup verify lang ==="
docker compose exec -T postgres psql -U even -d even_lexicon -c \
  "DELETE FROM languages WHERE code='$V_CODE';" >/dev/null
pass "removed $V_CODE"

echo ""
if [[ "${FAIL:-0}" -eq 0 ]]; then
  echo "=== ALL API CHECKS PASSED ==="
else
  echo "=== SOME CHECKS FAILED ==="
  exit 1
fi
