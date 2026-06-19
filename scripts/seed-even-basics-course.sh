#!/usr/bin/env bash
# Seed Even lexicon + demo course «Эвэды: Первые шаги» with full exercises,
# admin personal lexicon, demo students, and varied progress.
# Requires: `just up` (or local stack) + `just seed-dev`.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GW="${GW:-http://localhost:8080}"
COURSE_TITLE="${EVEN_BASICS_COURSE_TITLE:-Эвэды: Первые шаги}"
LESSON_TITLE="${EVEN_BASICS_LESSON_TITLE:-Буквы и слова}"
DEV_PASSWORD="${DEV_PASSWORD:-password123}"

pass() { echo "  ✓ $1"; }
fail() { echo "  ✗ $1"; exit 1; }

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_admin || fail "bootstrap admin"
pass "teacher/admin token"

"$ROOT/scripts/seed-languages.sh" >/dev/null 2>&1 || true
"$ROOT/scripts/seed-evn-alphabet.sh" >/dev/null 2>&1 || true

SEED_JSON=$(python3 - "$GW" "$TOKEN" "$COURSE_TITLE" "$LESSON_TITLE" "$ROOT" "$DEV_PASSWORD" <<'PY'
import json, os, subprocess, sys, urllib.error, urllib.request

gw, token, course_title, lesson_title, root, dev_password = sys.argv[1:7]
lexicon_path = os.path.join(root, "scripts", "data", "evn-lexicon-basic.json")
personal_path = os.path.join(root, "scripts", "data", "evn-lexicon-personal.json")

DEMO_STUDENTS = [
    ("student@even.local", "Dev Student"),
    ("student2@even.local", "Айа Эвэ"),
    ("student3@even.local", "Пётр Нагай"),
    ("student4@even.local", "Мария Олон"),
]

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

def req(method, path, body=None, auth_token=None):
    url = f"{gw}{path}"
    hdrs = {"Content-Type": "application/json"}
    tok = auth_token or token
    if tok:
        hdrs["Authorization"] = f"Bearer {tok}"
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

def ensure_platform_lexeme(lemma, translation):
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
    raise SystemExit(f"platform lexeme {lemma} → {code} {created}")

def ensure_teacher_lexeme(lemma, translation):
    from urllib.parse import quote
    code, listed = req("GET", f"/api/v1/teacher/languages/evn/lexicon?q={quote(lemma)}&limit=20")
    if code == 200:
        for item in listed.get("items", []):
            if item.get("lemma") == lemma:
                return item["id"]
    code, created = req("POST", "/api/v1/teacher/languages/evn/lexicon", {
        "lemma": lemma,
        "translations": [{"target_language_id": ru_lex_id, "text": translation}],
    })
    if code == 201:
        return created["id"]
    if code == 409:
        code, listed = req("GET", f"/api/v1/teacher/languages/evn/lexicon?q={quote(lemma)}&limit=20")
        for item in listed.get("items", []):
            if item.get("lemma") == lemma:
                return item["id"]
    raise SystemExit(f"teacher lexeme {lemma} → {code} {created}")

def ensure_student(email, display_name):
    req("POST", "/api/v1/auth/register", {
        "email": email,
        "password": dev_password,
        "role": "student",
        "display_name": display_name,
    })

def login(email):
    code, body = req("POST", "/api/v1/auth/login", {
        "email": email,
        "password": dev_password,
    })
    if code != 200:
        raise SystemExit(f"login {email} → {code} {body}")
    return body["access_token"]

def ensure_block(lesson_id, section_id, block_type, sort_order, title, config, blocks):
    existing = next(
        (b for b in blocks if b.get("section_id") == section_id and b.get("sort_order") == sort_order),
        None,
    )
    if existing:
        patch_body = {"config": config, "title": title}
        if existing.get("block_type") != block_type:
            patch_body["block_type"] = block_type
        code, resp = req("PATCH", f"/api/v1/teacher/blocks/{existing['id']}", patch_body)
        if code != 200:
            raise SystemExit(f"patch {block_type}#{sort_order} → {code} {resp}")
        return existing["id"], block_type, config
    code, created = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/blocks", {
        "section_id": section_id,
        "sort_order": sort_order,
        "block_type": block_type,
        "title": title,
        "config": config,
    })
    if code != 201:
        raise SystemExit(f"create {block_type}#{sort_order} → {code} {created}")
    return created["id"], block_type, config

def enroll_student(email, course_id):
    code, resp = req("POST", "/api/v1/teacher/students", {"email": email, "course_id": course_id})
    if code in (201, 409):
        return
    raise SystemExit(f"enroll {email} → {code} {resp}")

def attempt(token, block_id, response, sub_item_index=0):
    code, resp = req(
        "POST",
        f"/api/v1/progress/blocks/{block_id}/attempt",
        {"response": response, "sub_item_index": sub_item_index},
        auth_token=token,
    )
    if code != 200:
        raise SystemExit(f"attempt {block_id} → {code} {resp}")
    return resp

def correct_response(block_type, config):
    if block_type in ("prompt_choose_word", "word_choose_translation", "listen_choose_word", "gap_sentence_choose_word"):
        idx = config.get("correct_index", 0)
        return {"selected_index": idx}
    if block_type in ("prompt_type_word", "listen_type_word"):
        text = config.get("correct_text") or ""
        return {"text": text}
    if block_type in ("prompt_sentence_type", "listen_sentence_type"):
        return {"text": config.get("correct_text", "")}
    if block_type in ("prompt_sentence_word_order", "listen_sentence_word_order"):
        sub = (config.get("sub_items") or [{}])[0]
        order = sub.get("correct_order") or config.get("correct_order") or []
        return {"order": order}
    return {}

# --- Platform lexicon ---
with open(lexicon_path, encoding="utf-8") as f:
    lexicon_items = json.load(f)
lex_ids = {}
for item in lexicon_items:
    lex_ids[item["lemma"]] = ensure_platform_lexeme(item["lemma"], item["translation"])

# --- Teacher personal lexicon (admin@even.local) ---
with open(personal_path, encoding="utf-8") as f:
    personal_items = json.load(f)
personal_ids = {}
for item in personal_items:
    personal_ids[item["lemma"]] = ensure_teacher_lexeme(item["lemma"], item["translation"])

code, evn = req("GET", "/languages/evn")
if code != 200:
    raise SystemExit(f"GET /languages/evn → {code}")
code, ru = req("GET", "/languages/ru")
if code != 200:
    raise SystemExit(f"GET /languages/ru → {code}")

# --- Course & lesson ---
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
lesson = next((l for l in lessons if l.get("title") == lesson_title), None)
if lesson is None:
    code, lesson = req("POST", f"/api/v1/teacher/courses/{course_id}/lessons", {
        "title": lesson_title, "sort_order": 1,
    })
    if code != 201:
        raise SystemExit(f"create lesson → {code} {lesson}")
lesson_id = lesson["id"]

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"get lesson → {code}")

sections = full.get("sections") or []
section_intro = next((s for s in sections if s.get("title") in ("Введение", "Основной раздел")), None)
if section_intro is None:
    code, section_intro = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/sections", {
        "title": "Введение", "sort_order": 1, "section_kind": "content",
    })
    if code != 201:
        raise SystemExit(f"create intro section → {code} {section_intro}")
elif section_intro.get("title") == "Основной раздел":
    req("PATCH", f"/api/v1/teacher/sections/{section_intro['id']}", {"title": "Введение"})

section_practice = next((s for s in sections if s.get("title") == "Упражнения"), None)
if section_practice is None:
    code, section_practice = req("POST", f"/api/v1/teacher/lessons/{lesson_id}/sections", {
        "title": "Упражнения", "sort_order": 2, "section_kind": "content",
    })
    if code != 201:
        raise SystemExit(f"create practice section → {code} {section_practice}")

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"refresh lesson sections → {code}")
blocks = all_blocks(full)
sections = full.get("sections") or []
section_intro = next((s for s in sections if s.get("title") in ("Введение", "Основной раздел")), None)
section_practice = next((s for s in sections if s.get("title") == "Упражнения"), None)
if section_intro is None or section_practice is None:
    raise SystemExit("missing lesson sections after refresh")

intro_id = section_intro["id"]
practice_id = section_practice["id"]

for b in list(blocks):
    if b.get("section_id") == intro_id and b.get("sort_order", 0) >= 4:
        code, _ = req("DELETE", f"/api/v1/teacher/blocks/{b['id']}")
        if code not in (204, 404):
            raise SystemExit(f"delete orphan block {b['id']} → {code}")

code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
blocks = all_blocks(full)

gradable = []

def L(lemma):
    return lex_ids[lemma]

def lbl(lemma):
    return {L(lemma): lemma}

# --- Intro blocks ---
ensure_block(lesson_id, intro_id, "text", 0, "Добро пожаловать", {
    "items": [{
        "kind": "text",
        "text": (
            "Урок «Буквы и слова» — первый шаг в изучении эвенского языка. "
            "Вы познакомитесь с базовой лексикой природы и быта: дом, вода, солнце, лес. "
            "Для ввода ответов используйте встроенную эвенскую клавиатуру — в ней есть буква ӈ (ŋ), "
            "которой нет в обычной русской раскладке."
        ),
    }],
}, blocks)

ensure_block(lesson_id, intro_id, "note", 1, "Как проходить урок", {
    "body": (
        "1. Прочитайте текст и выучите слова из набора.\n"
        "2. Отвечайте на вопросы — можно ошибаться, система запомнит прогресс.\n"
        "3. Нажмите ⇧ на клавиатуре для заглавных букв."
    ),
}, blocks)

vocab_all = [
    "ашаткан", "бодон", "илен", "нюрген", "гур", "эн", "наад",
    "ураан", "чурок", "урум", "арка", "диил", "мурэ", "тирэ", "хэй",
]
vocab_ids = [L(w) for w in vocab_all]
labels = {L(w): w for w in vocab_all}

ensure_block(lesson_id, intro_id, "vocabulary_set", 2, "Слова урока (15)", {
    "lexeme_ids": vocab_ids,
    "lexeme_labels": labels,
    "show_images": True,
    "show_audio": True,
}, blocks)

ensure_block(lesson_id, intro_id, "text", 3, "Примеры употребления", {
    "items": [{
        "kind": "text",
        "text": (
            "• ашаткан — «спасибо» (после помощи)\n"
            "• эн наад — «дом у воды»\n"
            "• нюрген гур — «солнце в лесу»\n"
            "• бодон бэлэ — «хороший человек» (из личной базы преподавателя)"
        ),
    }],
}, blocks)

# --- Practice: choose word ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_choose_word", 0, "Выбор: спасибо", {
    "prompt": {"items": [{"kind": "text", "text": "Какое слово означает «спасибо»?"}]},
    "choices": [
        {"lexeme_id": L("ашаткан"), "lemma": "ашаткан"},
        {"lexeme_id": L("бодон"), "lemma": "бодон"},
        {"lexeme_id": L("хэй"), "lemma": "хэй"},
    ],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_choose_word", 1, "Выбор: солнце", {
    "prompt": {"items": [{"kind": "text", "text": "Выберите слово «солнце»"}]},
    "choices": [
        {"lexeme_id": L("нюрген"), "lemma": "нюрген"},
        {"lexeme_id": L("гур"), "lemma": "гур"},
        {"lexeme_id": L("илен"), "lemma": "илен"},
    ],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_choose_word", 2, "Выбор: собака", {
    "prompt": {"items": [{"kind": "text", "text": "Как будет «собака» по-эвенски?"}]},
    "choices": [
        {"lexeme_id": L("чурок"), "lemma": "чурок"},
        {"lexeme_id": L("диил"), "lemma": "диил"},
        {"lexeme_id": L("тирэ"), "lemma": "тирэ"},
    ],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

# --- Translation ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "word_choose_translation", 3, "Перевод: ашаткан", {
    "prompt": {"items": [{"kind": "lexeme", "lexeme_id": L("ашаткан"), "lemma": "ашаткан"}]},
    "choices": [{"text": "спасибо"}, {"text": "человек"}, {"text": "дом"}],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "word_choose_translation", 4, "Перевод: наад", {
    "prompt": {"items": [{"kind": "lexeme", "lexeme_id": L("наад"), "lemma": "наад"}]},
    "choices": [{"text": "вода"}, {"text": "лёд"}, {"text": "река"}],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "word_choose_translation", 5, "Перевод: ураан", {
    "prompt": {"items": [{"kind": "text", "text": "Выберите перевод слова «ураан»"}]},
    "choices": [{"text": "язык"}, {"text": "день"}, {"text": "год"}],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

# --- Type word ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_type_word", 6, "Напишите: спасибо", {
    "prompt": {"items": [{"kind": "text", "text": "Напишите по-эвенски: спасибо"}]},
    "correct_lexeme_id": L("ашаткан"),
    "correct_text": "ашаткан",
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_type_word", 7, "Напишите: дом", {
    "prompt": {"items": [{"kind": "text", "text": "Напишите по-эвенски: дом"}]},
    "correct_lexeme_id": L("эн"),
    "correct_text": "эн",
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_type_word", 8, "Напишите: вода", {
    "prompt": {"items": [{"kind": "lexeme", "lexeme_id": L("наад"), "lemma": "наад"}]},
    "correct_lexeme_id": L("наад"),
    "correct_text": "наад",
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_type_word", 9, "Напишите: зима", {
    "prompt": {"items": [{"kind": "text", "text": "Как пишется слово «зима»?"}]},
    "correct_lexeme_id": L("урум"),
    "correct_text": "урум",
}, blocks)
gradable.append((bid, btype, cfg))

# --- Sentence type ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_sentence_type", 10, "Напишите: лес", {
    "prompt": {"items": [{"kind": "text", "text": "Введите слово «лес» на эвенском"}]},
    "correct_text": "гур",
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_sentence_type", 11, "Напишите: друг", {
    "prompt": {"items": [{"kind": "text", "text": "Напишите слово «друг»"}]},
    "correct_text": "тирэ",
}, blocks)
gradable.append((bid, btype, cfg))

# --- Gap ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "gap_sentence_choose_word", 12, "Пропуск: дом", {
    "body": "Эн {{gap:0}} — наад илен.",
    "gaps": [{"id": 0, "correct_lexeme_id": L("эн")}],
    "choices": [
        {"lexeme_id": L("эн"), "lemma": "эн"},
        {"lexeme_id": L("гур"), "lemma": "гур"},
        {"lexeme_id": L("илен"), "lemma": "илен"},
    ],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

# --- Word order ---
bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_sentence_word_order", 13, "Порядок слов", {
    "sub_items": [{
        "prompt": {"items": [{"kind": "text", "text": "Составьте: «спасибо, человек» (ашаткан бодон)"}]},
        "tokens": ["ашаткан", "бодон"],
        "correct_order": [0, 1],
    }],
}, blocks)
gradable.append((bid, btype, cfg))

bid, btype, cfg = ensure_block(lesson_id, practice_id, "prompt_choose_word", 14, "Выбор: да", {
    "prompt": {"items": [{"kind": "text", "text": "Какое слово означает «да»?"}]},
    "choices": [
        {"lexeme_id": L("хэй"), "lemma": "хэй"},
        {"lexeme_id": L("урум"), "lemma": "урум"},
        {"lexeme_id": L("арка"), "lemma": "арка"},
    ],
    "correct_index": 0,
}, blocks)
gradable.append((bid, btype, cfg))

# --- Publish ---
code, full = req("GET", f"/api/v1/teacher/lessons/{lesson_id}")
if code != 200:
    raise SystemExit(f"refresh lesson → {code}")

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

# --- Students & progress ---
for email, name in DEMO_STUDENTS:
    ensure_student(email, name)
    enroll_student(email, course_id)

# Progress profiles: (email, fraction of gradable blocks to complete)
progress_plan = [
    ("student2@even.local", 1.0),
    ("student@even.local", 0.75),
    ("student3@even.local", 0.45),
    ("student4@even.local", 0.15),
]

for email, fraction in progress_plan:
    stoken = login(email)
    count = max(1, int(len(gradable) * fraction))
    for i, (block_id, block_type, config) in enumerate(gradable[:count]):
        if email == "student4@even.local" and i == 0 and block_type == "prompt_choose_word":
            attempt(stoken, block_id, {"selected_index": 1})
        resp = correct_response(block_type, config)
        attempt(stoken, block_id, resp)

print(json.dumps({
    "course_id": course_id,
    "lesson_id": lesson_id,
    "invite_code": invite["invite_code"],
    "platform_lexeme_count": len(lex_ids),
    "personal_lexeme_count": len(personal_ids),
    "block_count": len(all_blocks(full)),
    "gradable_block_count": len(gradable),
    "students": [e for e, _ in DEMO_STUDENTS],
}))
PY
)

export EVEN_BASICS_COURSE_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['course_id'])" "$SEED_JSON")
export EVEN_BASICS_LESSON_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['lesson_id'])" "$SEED_JSON")
export EVEN_BASICS_INVITE_CODE=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['invite_code'])" "$SEED_JSON")

pass "course $EVEN_BASICS_COURSE_ID"
pass "lesson «${LESSON_TITLE}» $EVEN_BASICS_LESSON_ID"
pass "invite code $EVEN_BASICS_INVITE_CODE"
pass "platform lexicon: $(python3 -c "import json,sys; print(json.loads(sys.argv[1])['platform_lexeme_count'])" "$SEED_JSON") words"
pass "personal lexicon (admin): $(python3 -c "import json,sys; print(json.loads(sys.argv[1])['personal_lexeme_count'])" "$SEED_JSON") words"
pass "blocks: $(python3 -c "import json,sys; print(json.loads(sys.argv[1])['block_count'])" "$SEED_JSON") ($(python3 -c "import json,sys; print(json.loads(sys.argv[1])['gradable_block_count'])" "$SEED_JSON") exercises)"
pass "students: $(python3 -c "import json,sys; print(', '.join(json.loads(sys.argv[1])['students']))" "$SEED_JSON")"

"$ROOT/scripts/seed-lexicon-media.sh" >/dev/null 2>&1 || true
pass "lexicon media (images + audio)"

echo ""
echo "=== Even basics seed complete ==="
