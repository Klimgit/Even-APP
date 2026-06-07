#!/usr/bin/env bash
# Load the standard Even Cyrillic alphabet via platform API.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GW="${GW:-http://localhost:8080}"
LEX="${LEX:-http://localhost:8082}"
DATA="$ROOT/scripts/data/evn-alphabet.json"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin

python3 - "$LEX" "$TOKEN" "$DATA" <<'PY'
import json, sys, urllib.request, urllib.error

lex, token, data_path = sys.argv[1:4]
letters = json.load(open(data_path))
canonical = {row["character"] for row in letters}

def req(method, url, body=None):
    data = None if body is None else json.dumps(body).encode()
    r = urllib.request.Request(
        url, data=data, method=method,
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(r) as resp:
            return resp.status, json.loads(resp.read() or b"null")
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            payload = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            payload = {"raw": raw.decode(errors="replace")}
        return e.code, payload

_, current = req("GET", f"{lex}/api/v1/platform/languages/evn/alphabet")
removed = 0
for row in current:
    if row["character"] not in canonical:
        code, resp = req("DELETE", f"{lex}/api/v1/platform/alphabet/{row['id']}")
        if code != 204:
            raise SystemExit(f"DELETE {row['character']} → {code} {resp}")
        removed += 1

created = skipped = updated = 0
for i, row in enumerate(letters):
    body = {
        "character": row["character"],
        "upper_char": row.get("upper_char"),
        "sort_order": i,
    }
    if row.get("label"):
        body["label"] = row["label"]
    if "transcription" in row:
        body["transcription"] = row["transcription"]
    code, resp = req("POST", f"{lex}/api/v1/platform/languages/evn/alphabet", body)
    if code == 201:
        created += 1
    elif code == 409:
        skipped += 1
    else:
        raise SystemExit(f"POST {row['character']} → {code} {resp}")

_, current = req("GET", f"{lex}/api/v1/platform/languages/evn/alphabet")
by_char = {x["character"]: x for x in current}
for row in letters:
    cur = by_char.get(row["character"])
    if not cur:
        continue
    patch = {}
    want_upper = row.get("upper_char")
    if want_upper and cur.get("upper_char") != want_upper:
        patch["upper_char"] = want_upper
    want_tr = row.get("transcription")
    if want_tr != cur.get("transcription"):
        if want_tr is not None:
            patch["transcription"] = want_tr
        elif cur.get("transcription"):
            patch["transcription"] = ""
    if patch:
        code, resp = req("PATCH", f"{lex}/api/v1/platform/alphabet/{cur['id']}", patch)
        if code != 200:
            raise SystemExit(f"PATCH {row['character']} → {code} {resp}")
        updated += 1

_, current = req("GET", f"{lex}/api/v1/platform/languages/evn/alphabet")
by_id = {x["character"]: x["id"] for x in current}
ids = [by_id[row["character"]] for row in letters]
code, reordered = req("POST", f"{lex}/api/v1/platform/languages/evn/alphabet/reorder", {"letter_ids": ids})
if code != 200:
    raise SystemExit(f"reorder → {code} {reordered}")

chars = [x["character"] for x in reordered]
expected = [row["character"] for row in letters]
if chars != expected:
    raise SystemExit(f"order mismatch: {chars} != {expected}")

print(f"removed={removed} created={created} skipped={skipped} updated={updated} total={len(letters)}")
PY
pass "36 Even letters loaded and ordered"

c=$(curl -s -o /tmp/seed-body.json -w "%{http_code}" "$GW/languages/evn/alphabet")
[[ "$c" == "200" ]] || fail "public alphabet → $c"
python3 -c "
import json
d = json.load(open('/tmp/seed-body.json'))
exp = json.load(open('$DATA'))
assert len(d) == len(exp), (len(d), len(exp))
assert [x['character'] for x in d] == [x['character'] for x in exp]
ng = d[15]
assert ng['character'] == 'ӈ' and ng.get('upper_char') == 'Ӈ' and ng.get('transcription') == 'ŋ', ng
"
pass "GET /languages/evn/alphabet OK"
