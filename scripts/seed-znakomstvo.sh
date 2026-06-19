#!/usr/bin/env bash
# Seed MVP demo course with lesson «Знакомство» (Even basics).
# Requires: `just up` (or local stack) + `just seed-dev`.
# Exports: ZNAKOMSTVO_COURSE_ID, ZNAKOMSTVO_LESSON_ID, ZNAKOMSTVO_INVITE_CODE,
#          ZNAKOMSTVO_GRADABLE_BLOCK_ID, ZNAKOMSTVO_LEXEME_ASHATKAN
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GW="${GW:-http://localhost:8080}"
COURSE_TITLE="${ZNAKOMSTVO_COURSE_TITLE:-Эвэды: Знакомство}"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin || fail "bootstrap admin"
pass "teacher/admin token"

"$ROOT/scripts/seed-languages.sh" >/dev/null 2>&1 || true

SEED_JSON=$(python3 - "$GW" "$TOKEN" "$COURSE_TITLE" "$ROOT" <<'PY'
import json, os, subprocess, sys, urllib.error, urllib.request

gw, token, course_title, root = sys.argv[1:5]

def psql(sql):
    r = subprocess.run(
        ["docker", "compose", "-f", f"{root}/docker-compose.yml", "exec", "-T", "postgres",
         "psql", "-U", os.environ.get("POSTGRES_USER", "even"), "-d", "even_lexicon", "-tAc", sql],
        capture_output=True, text=True, check=True,
    )
    return r.stdout.strip()

ru_lex_id = psql("SELECT id FROM languages WHERE code='ru'")

def all_blocks(lesson):
    out = []
    for sec in lesson.get("sections") or []:
        out.extend(sec.get("blocks") or [])
    return out

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

def ensure_lexeme(lemma, translation):
    from urllib.parse import quote
    code, listed = req("GET", f"/api/v1/platform/languages/evn/lexicon?q={quote(lemma)}&limit=20")
    if code == 200:
        for item in listed.get("items", []):
            if item.get("lemma") == lemma:
                return item["id"]
    code, created = req("POST", "/api/v1/platform/languages/evn/lexicon", {
        "lemma": lemma,
        "translations": [{"target_language_id": ru_lex_id, "text": translation}],
    })
    if code == 201:
        return created["id"]
    raise SystemExit(f"lexeme {lemma} → {code} {created}")

code, evn = req("GET", "/languages/evn")
if code != 200:
    raise SystemExit(f"GET /languages/evn → {code}")
code, ru = req("GET", "/languages/ru")
if code != 200:
    raise SystemExit(f"GET /languages/ru → {code}")

ashatkan_id = ensure_lexeme("ашаткан", "спасибо")
distractor_id = ensure_lexeme("бодон", "человек")

code, courses = req("GET", "/api/v1/teacher/courses")
if code != 200:
    raise SystemExit(f"list courses → {code} {courses}")
course = next((c for c in courses if c.get("title") == course_title), None)
if course is None:
    code, course = req("POST", "/api/v1/teacher/courses", {
        "title": course_title,
        "target_language_id": evn["id"],
        "ui_language_id": ru["id"],
    })
    if code != 201:
        raise SystemExit(f"create course → {code} {course}")

course_id = course["id"]

code, lessons = req("GET", f"/api/v1/teacher/courses/{course_id}/lessons")
if code != 200:
    raise SystemExit(f"list lessons → {code}")
lesson = next((l for l in lessons if l.get("title") == "Знакомство"), None)
if lesson is None:
    code, lesson = req("POST", f"/api/v1/teacher/courses/{course_id}/lessons", {
        "title": "Знакомство", "sort_order": 1,
    })
    if code != 201:
        raise SystemExit(f"create lesson → {code} {lesson}")

lesson_id = lesson["id"]

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"get lesson → {code}")

sections = full.get("sections") or []
section = next((s for s in sections if s.get("title") == "Знакомство. Слова"), None)
if section is None:
    code, section = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/sections", {
        "title": "Знакомство. Слова", "sort_order": 1, "section_kind": "content",
    })
    if code != 201:
        raise SystemExit(f"create section → {code} {section}")

section_id = section["id"]
blocks = all_blocks(full)

def has_block(block_type):
    return any(b.get("block_type") == block_type for b in blocks)

if not has_block("text"):
    code, resp = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/blocks", {
        "section_id": section_id, "sort_order": 0, "block_type": "text", "title": "Приветствие",
        "config": {"items": [{"kind": "text", "text": "Урок «Знакомство». Выучим первые слова."}]},
    })
    if code != 201:
        raise SystemExit(f"block text → {code} {resp}")

if not has_block("vocabulary_set"):
    code, resp = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/blocks", {
        "section_id": section_id, "sort_order": 1, "block_type": "vocabulary_set", "title": "Слова",
        "config": {"lexeme_ids": [ashatkan_id, distractor_id], "show_images": False, "show_audio": False},
    })
    if code != 201:
        raise SystemExit(f"block vocabulary_set → {code} {resp}")

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"refresh lesson for patch → {code}")
vocab = next((b for b in all_blocks(full) if b.get("block_type") == "vocabulary_set"), None)
if vocab:
    code, resp = req("PATCH", f"/api/v1/teacher/blocks/{vocab['id']}", {
        "config": {"lexeme_ids": [ashatkan_id, distractor_id], "show_images": False, "show_audio": False},
    })
    if code != 200:
        raise SystemExit(f"patch vocabulary_set → {code} {resp}")

exercise = next((b for b in all_blocks(full) if b.get("block_type") == "prompt_choose_word"), None)
if exercise is None:
    code, exercise = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/blocks", {
        "section_id": section_id, "sort_order": 2, "block_type": "prompt_choose_word", "title": "Выбрать слово",
        "config": {
            "prompt": {"items": [{"kind": "text", "text": "Выберите перевод «спасибо»"}]},
            "choices": [{"lexeme_id": ashatkan_id}, {"lexeme_id": distractor_id}],
            "correct_index": 0,
        },
    })
    if code != 201:
        raise SystemExit(f"block exercise → {code} {exercise}")

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"refresh lesson after exercise → {code}")
exercise = next((b for b in all_blocks(full) if b.get("block_type") == "prompt_choose_word"), None)
if exercise:
    code, resp = req("PATCH", f"/api/v1/teacher/blocks/{exercise['id']}", {
        "config": {
            "prompt": {"items": [{"kind": "text", "text": "Выберите перевод «спасибо»"}]},
            "choices": [{"lexeme_id": ashatkan_id}, {"lexeme_id": distractor_id}],
            "correct_index": 0,
        },
    })
    if code != 200:
        raise SystemExit(f"patch exercise → {code} {resp}")

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"refresh lesson → {code}")
exercise = next((b for b in all_blocks(full) if b.get("block_type") == "prompt_choose_word"), None)
if exercise is None:
    raise SystemExit("missing prompt_choose_word block")

if full.get("status") != "published":
    code, resp = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/publish")
    if code != 200:
        raise SystemExit(f"publish lesson → {code} {resp}")

code, course = req("GET", f"/api/v1/teacher/courses/{course_id}")
if not course.get("is_published"):
    code, resp = req("POST", f"/api/v1/teacher/courses/{course_id}/publish")
    if code != 200:
        raise SystemExit(f"publish course → {code} {resp}")

code, course = req("GET", f"/api/v1/teacher/courses/{course_id}")
if course.get("visibility") != "public":
    code, resp = req("PATCH", f"/api/v1/teacher/courses/{course_id}", {"visibility": "public"})
    if code != 200:
        raise SystemExit(f"set course public → {code} {resp}")

code, invite = req("GET", f"/api/v1/teacher/courses/{course_id}/invite-code")
if code != 200:
    raise SystemExit(f"invite code → {code} {invite}")

print(json.dumps({
    "course_id": course_id,
    "lesson_id": lesson_id,
    "invite_code": invite["invite_code"],
    "gradable_block_id": exercise["id"],
    "lexeme_ashatkan": ashatkan_id,
}))
PY
)

export ZNAKOMSTVO_COURSE_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['course_id'])" "$SEED_JSON")
export ZNAKOMSTVO_LESSON_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['lesson_id'])" "$SEED_JSON")
export ZNAKOMSTVO_INVITE_CODE=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['invite_code'])" "$SEED_JSON")
export ZNAKOMSTVO_GRADABLE_BLOCK_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['gradable_block_id'])" "$SEED_JSON")
export ZNAKOMSTVO_LEXEME_ASHATKAN=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['lexeme_ashatkan'])" "$SEED_JSON")

pass "course $ZNAKOMSTVO_COURSE_ID"
pass "lesson «Знакомство» $ZNAKOMSTVO_LESSON_ID"
pass "invite code $ZNAKOMSTVO_INVITE_CODE"
echo ""
echo "=== Znakomstvo seed complete ==="
