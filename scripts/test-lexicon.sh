#!/usr/bin/env bash
set -euo pipefail

GW="${GW:-http://localhost:8080}"
LEX="${LEX:-http://localhost:8082}"
EMAIL="lexicon-test-$(date +%s)@example.com"
PASS="password123"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }
code() { curl -s -o /tmp/lex-body.json -w "%{http_code}" "$@"; }
json() { python3 -c "$1"; }

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin

echo "=== Public languages ==="
c=$(code "$GW/languages")
[[ "$c" == "200" ]] && pass "GET /languages → $c" || fail "GET /languages → $c ($(cat /tmp/lex-body.json))"
json "import json; d=json.load(open('/tmp/lex-body.json')); assert any(x['code']=='evn' for x in d), d"
pass "languages contains evn"

c=$(code "$GW/languages/evn")
[[ "$c" == "200" ]] && pass "GET /languages/evn → $c" || fail "GET /languages/evn → $c"

c=$(code "$GW/languages/evn/alphabet")
[[ "$c" == "200" ]] && pass "GET /languages/evn/alphabet → $c" || fail "GET /languages/evn/alphabet → $c"
json "import json; assert isinstance(json.load(open('/tmp/lex-body.json')), list)"
pass "public alphabet is a JSON array"

c=$(code "$GW/languages/zzz")
[[ "$c" == "404" ]] && pass "GET /languages/zzz → 404" || fail "GET /languages/zzz → $c"

pass "admin token ($PLATFORM_ADMIN_EMAIL)"

echo ""
echo "=== Platform languages ==="
c=$(code -H "Authorization: Bearer $TOKEN" "$LEX/api/v1/platform/languages")
[[ "$c" == "200" ]] && pass "GET /platform/languages → $c" || fail "GET /platform/languages → $c"

TST_CODE="tst$RANDOM"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages" \
  -H 'Content-Type: application/json' \
  -d "{\"code\":\"$TST_CODE\",\"name\":\"Test Lang\",\"native_name\":\"Test\",\"direction\":\"ltr\"}")
[[ "$c" == "201" ]] && pass "POST /platform/languages → $c" || fail "POST /platform/languages → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$LEX/api/v1/platform/languages/$TST_CODE" \
  -H 'Content-Type: application/json' -d '{"name":"Test Lang Updated"}')
[[ "$c" == "200" ]] && pass "PATCH /platform/languages/{code} → $c" || fail "PATCH language → $c"
json "import json; assert json.load(open('/tmp/lex-body.json'))['name']=='Test Lang Updated'"

echo ""
echo "=== Platform alphabet ($TST_CODE) ==="
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"в","upper_char":"В","sort_order":1,"label":"v"}')
[[ "$c" == "201" ]] && pass "POST letter в → $c" || fail "POST в → $c"
ID_V=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"а","upper_char":"А","sort_order":2}')
[[ "$c" == "201" ]] && pass "POST letter а → $c" || fail "POST а → $c"
ID_A=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/alphabet" \
  -H 'Content-Type: application/json' -d '{"character":"ӈ","upper_char":"Ӈ","sort_order":3,"transcription":"ŋ"}')
[[ "$c" == "201" ]] && pass "POST Even letter ӈ/Ӈ → $c" || fail "POST ӈ → $c"
ID_NG=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")
json "import json; r=json.load(open('/tmp/lex-body.json')); assert r.get('transcription')=='ŋ', r"
pass "alphabet letter includes IPA transcription ŋ"

c=$(code -H "Authorization: Bearer $TOKEN" "$LEX/api/v1/platform/languages/$TST_CODE/alphabet")
json "import json; d=json.load(open('/tmp/lex-body.json')); assert [x['character'] for x in d]==['в','а','ӈ'], d"
pass "alphabet sorted by sort_order (в, а, ӈ)"

c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$LEX/api/v1/platform/alphabet/$ID_A" \
  -H 'Content-Type: application/json' -d '{"transcription":"a"}')
[[ "$c" == "200" ]] && pass "PATCH /platform/alphabet/{id} → $c" || fail "PATCH letter → $c"
json "import json; assert json.load(open('/tmp/lex-body.json'))['transcription']=='a'"

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/alphabet/reorder" \
  -H 'Content-Type: application/json' \
  -d "{\"letter_ids\":[\"$ID_A\",\"$ID_NG\",\"$ID_V\"]}")
[[ "$c" == "200" ]] && pass "POST alphabet reorder → $c" || fail "reorder → $c"
json "import json; d=json.load(open('/tmp/lex-body.json')); assert [x['character'] for x in d]==['а','ӈ','в'], d"
pass "reorder returns а, ӈ, в"

c=$(code "$GW/languages/$TST_CODE/alphabet")
json "import json; d=json.load(open('/tmp/lex-body.json')); assert [x['character'] for x in d]==['а','ӈ','в'], d"
pass "public alphabet reflects reorder"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/alphabet/$ID_V")
[[ "$c" == "204" ]] && pass "DELETE /platform/alphabet/{id} → $c" || fail "DELETE letter → $c"
LETTER_ID="$ID_NG"

echo ""
echo "=== Platform sounds ($TST_CODE) ==="
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/sounds" \
  -H 'Content-Type: application/json' -d '{"ipa":"/a/","description":"test vowel"}')
[[ "$c" == "201" ]] && pass "POST sound → $c" || fail "POST sound → $c"
SOUND_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" "$LEX/api/v1/platform/languages/$TST_CODE/sounds")
[[ "$c" == "200" ]] && pass "GET /platform/sounds → $c" || fail "GET sounds → $c"
json "import json; assert any(x['id']=='$SOUND_ID' for x in json.load(open('/tmp/lex-body.json')))"

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/alphabet/$LETTER_ID/sounds" \
  -H 'Content-Type: application/json' -d "{\"sound_id\":\"$SOUND_ID\"}")
[[ "$c" == "204" ]] && pass "link letter-sound → $c" || fail "link → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$LEX/api/v1/platform/sounds/$SOUND_ID" \
  -H 'Content-Type: application/json' -d '{"description":"updated vowel"}')
[[ "$c" == "200" ]] && pass "PATCH /platform/sounds/{id} → $c" || fail "PATCH sound → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/alphabet/$LETTER_ID/sounds/$SOUND_ID")
[[ "$c" == "204" ]] && pass "unlink letter-sound → $c" || fail "unlink → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/sounds/$SOUND_ID")
[[ "$c" == "204" ]] && pass "DELETE /platform/sounds/{id} → $c" || fail "DELETE sound → $c"

echo ""
echo "=== Platform lexicon ($TST_CODE) ==="
RU_ID=$(docker compose exec -T postgres psql -U even -d even_lexicon -tAc "SELECT id FROM languages WHERE code='ru'")

LEMMA="lemma$RANDOM"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/languages/$TST_CODE/lexicon" \
  -H 'Content-Type: application/json' \
  -d "{\"lemma\":\"$LEMMA\",\"translations\":[{\"target_language_id\":\"$RU_ID\",\"text\":\"тест\"}]}")
[[ "$c" == "201" ]] && pass "POST lexeme → $c" || fail "POST lexeme → $c"
LEXEME_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")
TRANS_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['translations'][0]['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/lexemes/$LEXEME_ID/forms" \
  -H 'Content-Type: application/json' -d '{"form":"формы","tags":{"number":"pl"}}')
[[ "$c" == "201" ]] && pass "POST lexeme form → $c" || fail "POST form → $c"
FORM_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$LEX/api/v1/platform/lexemes/$LEXEME_ID" \
  -H 'Content-Type: application/json' -d '{"notes":"smoke note"}')
[[ "$c" == "200" ]] && pass "PATCH lexeme → $c" || fail "PATCH lexeme → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X PATCH "$LEX/api/v1/platform/lexeme-forms/$FORM_ID" \
  -H 'Content-Type: application/json' -d '{"form":"форма2"}')
[[ "$c" == "200" ]] && pass "PATCH lexeme form → $c" || fail "PATCH form → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/lexemes/$LEXEME_ID/translations" \
  -H 'Content-Type: application/json' \
  -d "{\"target_language_id\":\"$RU_ID\",\"text\":\"ещё перевод\"}")
[[ "$c" == "201" ]] && pass "POST lexeme translation → $c" || fail "POST translation → $c"
TRANS2_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

FAKE_MEDIA="00000000-0000-4000-8000-$(printf '%012x' $RANDOM$RANDOM)"
c=$(code -H "Authorization: Bearer $TOKEN" -X POST "$LEX/api/v1/platform/lexemes/$LEXEME_ID/media" \
  -H 'Content-Type: application/json' \
  -d "{\"media_asset_id\":\"$FAKE_MEDIA\",\"kind\":\"image\",\"label\":\"pic\"}")
[[ "$c" == "201" ]] && pass "POST lexeme media → $c" || fail "POST lexeme media → $c"
MEDIA_ID=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['id'])")

c=$(code -H "Authorization: Bearer $TOKEN" "$LEX/api/v1/platform/languages/$TST_CODE/lexicon?q=$LEMMA")
[[ "$c" == "200" ]] && pass "GET lexicon list → $c" || fail "GET lexicon → $c"
json "import json; assert json.load(open('/tmp/lex-body.json'))['total']>=1"

c=$(code -H "Authorization: Bearer $TOKEN" "$LEX/api/v1/platform/lexemes/$LEXEME_ID")
[[ "$c" == "200" ]] && pass "GET lexeme → $c" || fail "GET lexeme → $c"

c=$(code -H "Authorization: Bearer $TOKEN" "$GW/api/v1/platform/languages/$TST_CODE/lexicon?q=$LEMMA")
[[ "$c" == "200" ]] && pass "GET lexicon via gateway → $c" || fail "gateway lexicon → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/lexeme-media/$MEDIA_ID")
[[ "$c" == "204" ]] && pass "DELETE lexeme media → $c" || fail "DELETE lexeme media → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/lexeme-translations/$TRANS2_ID")
[[ "$c" == "204" ]] && pass "DELETE lexeme translation → $c" || fail "DELETE translation → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/lexeme-forms/$FORM_ID")
[[ "$c" == "204" ]] && pass "DELETE lexeme form → $c" || fail "DELETE form → $c"

c=$(code -H "Authorization: Bearer $TOKEN" -X DELETE "$LEX/api/v1/platform/lexemes/$LEXEME_ID")
[[ "$c" == "204" ]] && pass "DELETE lexeme → $c" || fail "DELETE lexeme → $c"

echo ""
echo "=== Auth guards ==="
c=$(code -H "Authorization: Bearer invalid" "$LEX/api/v1/platform/languages")
[[ "$c" == "401" || "$c" == "500" ]] && pass "platform without valid JWT → $c" || fail "expected 401/500 got $c"

c=$(code -X POST "$GW/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"student-$RANDOM@example.com\",\"password\":\"password123\",\"role\":\"student\"}")
STOKEN=$(json "import json; print(json.load(open('/tmp/lex-body.json'))['access_token'])" 2>/dev/null || true)
if [[ -n "${STOKEN:-}" ]]; then
  c=$(code -H "Authorization: Bearer $STOKEN" "$LEX/api/v1/platform/languages")
  [[ "$c" == "403" ]] && pass "platform non-admin → 403" || fail "expected 403 got $c"
fi

docker compose exec -T postgres psql -U even -d even_lexicon -c \
  "DELETE FROM languages WHERE code='$TST_CODE';" >/dev/null
pass "cleaned up test language $TST_CODE"

echo ""
echo "=== All lexicon tests passed ==="
