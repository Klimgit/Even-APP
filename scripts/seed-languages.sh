#!/usr/bin/env bash
# Bootstrap evn + ru via lexicon API. Uses fixed platform-admin@even.local.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GW="${GW:-http://localhost:8080}"
LEX="${LEX:-http://localhost:8082}"
DATA="$ROOT/scripts/data/languages.json"
PG_USER="${POSTGRES_USER:-even}"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin
pass "admin $PLATFORM_ADMIN_EMAIL"

python3 - "$LEX" "$TOKEN" "$DATA" <<'PY'
import json, sys, urllib.request, urllib.error

lex, token, data_path = sys.argv[1:4]
langs = json.load(open(data_path))

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

created = skipped = 0
for row in langs:
    code, resp = req("POST", f"{lex}/api/v1/platform/languages", row)
    if code == 201:
        created += 1
    elif code == 409:
        skipped += 1
    else:
        raise SystemExit(f"POST {row['code']} → {code} {resp}")

_, public = req("GET", f"{lex}/languages")
codes = {x["code"] for x in public}
missing = [x["code"] for x in langs if x["code"] not in codes]
if missing:
    raise SystemExit(f"missing public languages: {missing}")

print(f"lexicon created={created} skipped={skipped}")
PY
pass "lexicon: evn, ru"

python3 - "$DATA" "$ROOT" "$PG_USER" <<'PY'
import json, subprocess, sys

data_path, root, pg_user = sys.argv[1:4]
langs = json.load(open(data_path))
values = ", ".join(
    f"('{l['code']}', '{l['name']}', '{l['native_name']}')" for l in langs
)
sql = f"""
INSERT INTO languages (code, name, native_name)
VALUES {values}
ON CONFLICT (code) DO NOTHING;
"""
subprocess.run(
    ["docker", "compose", "-f", f"{root}/docker-compose.yml", "exec", "-T", "postgres",
     "psql", "-U", pg_user, "-d", "even_media", "-v", "ON_ERROR_STOP=1", "-c", sql],
    check=True,
)
PY
pass "media: evn, ru synced"
