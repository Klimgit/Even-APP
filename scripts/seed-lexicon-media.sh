#!/usr/bin/env bash
# Upload synthetic placeholder images + audio for platform lexemes (evn basics).
# Requires: `just up` + `just seed-dev` + `just seed-languages`.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GW="${GW:-http://localhost:8080}"
DATA="$ROOT/scripts/data/evn-lexicon-basic.json"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin || fail "bootstrap admin"
pass "admin token"

"$ROOT/scripts/seed-languages.sh" >/dev/null 2>&1 || true

python3 - "$GW" "$TOKEN" "$DATA" <<'PY'
import json, struct, sys, urllib.error, urllib.request, zlib

gw, token, data_path = sys.argv[1:4]
items = json.load(open(data_path, encoding="utf-8"))

PALETTE = [
    (231, 76, 60), (46, 204, 113), (52, 152, 219), (155, 89, 182),
    (241, 196, 15), (230, 126, 34), (26, 188, 156), (149, 165, 166),
    (192, 57, 43), (41, 128, 185), (142, 68, 173), (39, 174, 96),
    (243, 156, 18), (211, 84, 0), (22, 160, 133),
]

def req(method, path, body=None):
    url = f"{gw}{path}"
    hdrs = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = None if body is None else json.dumps(body).encode()
    r = urllib.request.Request(url, data=data, method=method, headers=hdrs)
    try:
        with urllib.request.urlopen(r) as resp:
            raw = resp.read()
            return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            payload = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            payload = {"raw": raw.decode(errors="replace")}
        return e.code, payload

def png_bytes(r, g, b, size=64):
    row = b"\x00" + bytes([r, g, b]) * size
    raw = row * size
    compressed = zlib.compress(raw, 9)

    def chunk(tag, data):
        crc = zlib.crc32(tag + data) & 0xFFFFFFFF
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", crc)

    ihdr = struct.pack(">IIBBBBB", size, size, 8, 2, 0, 0, 0)
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", ihdr) + chunk(b"IDAT", compressed) + chunk(b"IEND", b"")

def wav_bytes(freq=440, duration=0.35, rate=22050):
    n = int(rate * duration)
    amp = 8000
    frames = bytearray()
    for i in range(n):
        sample = int(amp * __import__("math").sin(2 * __import__("math").pi * freq * i / rate))
        frames += struct.pack("<h", sample)
    data_size = len(frames)
    header = struct.pack(
        "<4sI4s4sIHHIIHH4sI",
        b"RIFF", 36 + data_size, b"WAVE", b"fmt ", 16, 1, 1, rate, rate * 2, 2, 16, b"data", data_size,
    )
    return header + frames

def upload_platform(filename, mime, body, lang_id, display_name):
    code, presign = req("POST", "/api/v1/platform/media/presign", {
        "filename": filename,
        "mime_type": mime,
        "size_bytes": len(body),
        "language_id": lang_id,
    })
    if code not in (200, 201):
        raise SystemExit(f"presign {filename} → {code} {presign}")
    put = urllib.request.Request(
        presign["upload_url"], data=body, method="PUT",
        headers={"Content-Type": mime},
    )
    with urllib.request.urlopen(put) as resp:
        if resp.status not in (200, 204):
            raise SystemExit(f"put {filename} → {resp.status}")
    code, asset = req("POST", "/api/v1/platform/media/confirm", {
        "object_key": presign["object_key"],
        "mime_type": mime,
        "size_bytes": len(body),
        "language_id": lang_id,
        "display_name": display_name,
    })
    if code not in (200, 201):
        raise SystemExit(f"confirm {filename} → {code} {asset}")
    return asset["id"]

def ensure_lexeme_id(lemma):
    from urllib.parse import quote
    code, listed = req("GET", f"/api/v1/platform/languages/evn/lexicon?q={quote(lemma)}&limit=20")
    if code == 200:
        for item in listed.get("items", []):
            if item.get("lemma") == lemma:
                return item["id"]
    raise SystemExit(f"lexeme not found: {lemma}")

def attach_media(lexeme_id, asset_id, kind):
    code, resp = req("POST", f"/api/v1/platform/lexemes/{lexeme_id}/media", {
        "media_asset_id": asset_id,
        "kind": kind,
        "is_primary": True,
    })
    if code == 201:
        return
    if code == 409:
        return
    raise SystemExit(f"attach {kind} to {lexeme_id} → {code} {resp}")

code, evn = req("GET", "/languages/evn")
if code != 200:
    raise SystemExit(f"GET /languages/evn → {code}")
lang_id = evn["id"]

attached = 0
for i, item in enumerate(items):
    lemma = item["lemma"]
    lex_id = ensure_lexeme_id(lemma)
    r, g, b = PALETTE[i % len(PALETTE)]
    img = png_bytes(r, g, b)
    wav = wav_bytes(freq=220 + i * 40)
    img_id = upload_platform(f"{lemma}.png", "image/png", img, lang_id, f"{lemma} (картинка)")
    aud_id = upload_platform(f"{lemma}.wav", "audio/wav", wav, lang_id, f"{lemma} (аудио)")
    attach_media(lex_id, img_id, "image")
    attach_media(lex_id, aud_id, "audio_word")
    attached += 1

print(json.dumps({"lexemes_with_media": attached}))
PY

pass "synthetic image+audio attached to platform lexemes"
echo ""
echo "=== Lexicon media seed complete ==="
