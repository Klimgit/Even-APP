# API — статус реализации (MVP)

---


## ✅ Реализовано

### System (все сервисы + gateway)

| Метод | Путь | Auth | Ответ |
|-------|------|------|-------|
| GET | `/health`, `/api/v1/health` | нет | `{ "status": "ok", "service": "<name>" }` |
| GET | `/api/v1/ready` | нет | `{ "status": "ready" }` или 503 |
| GET | `/api/v1/openapi.yaml` | нет | YAML |
| GET | `/api/v1/gateway/status` | нет | `{ "status", "service", "backends" }` (только gateway) |

---

### Auth — `auth` (:8081)

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| POST | `/auth/register` | нет | 201 |
| POST | `/auth/login` | нет | 200 |
| POST | `/auth/refresh` | нет | 200 |
| GET | `/auth/me` | JWT | 200 |

**POST /auth/register** — body:

```json
{
  "email": "user@example.com",
  "password": "secret123",
  "display_name": "Имя",
  "role": "student"
}
```

`password` ≥ 8 символов. `role`: `student` \| `teacher` (default `student`). `is_admin` всегда `false` при регистрации.

**Response 201 / login 200** — `AuthResponse`:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "user": {
    "id": "uuid",
    "email": "...",
    "display_name": "...",
    "role": "student",
    "is_admin": false,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

**POST /auth/refresh** — body: `{ "refresh_token": "..." }`  
**Response 200:** `{ "access_token", "refresh_token" }`

**GET /auth/me** — **Response 200:** `UserDTO` (без токенов).

**Ошибки:** 400 validation, 401 credentials, 409 email exists.

---

### Platform media — `media` (:8085)

Каталог `scope=platform`, `owner_id=null`. Запись в S3: presign → PUT → confirm.

| Метод | Путь | Кто | Статус |
|-------|------|-----|--------|
| POST | `/platform/media/presign` | platform admin | 200 |
| POST | `/platform/media/confirm` | platform admin | 201 |
| GET | `/platform/languages/{code}/media` | любой JWT | 200 |
| GET | `/platform/media/{id}` | любой JWT | 200 / 410 expired |
| PATCH | `/platform/media/{id}` | platform admin | 200 |
| DELETE | `/platform/media/{id}` | platform admin | 204 |

**POST /platform/media/presign** — body (`PresignRequest` + опционально `language_id`):

```json
{
  "filename": "photo.png",
  "mime_type": "image/png",
  "size_bytes": 12345,
  "language_id": "uuid"
}
```

Если `language_id` не передан — default язык `evn`.

**Response 200** (`PresignResponse`):

```json
{
  "upload_url": "presigned PUT URL",
  "object_key": "media/<uuid>.png",
  "media_asset_id": "uuid"
}
```

Далее клиент: `PUT upload_url` с телом файла → **POST confirm**.

**POST /platform/media/confirm** — body (`ConfirmMediaRequest`):

```json
{
  "object_key": "media/....png",
  "mime_type": "image/png",
  "size_bytes": 12345,
  "display_name": "Обязательное имя",
  "language_id": "uuid",
  "linked_lexeme_id": "uuid",
  "ttl_seconds": 86400,
  "expires_at": "2026-12-31T00:00:00Z",
  "width": 100,
  "height": 100,
  "duration_ms": 5000
}
```

`display_name` обязателен. `ttl_seconds` или `expires_at` — опционально (бессрочно если не задано).

**Response 201:** `MediaAssetDTO`:

```json
{
  "id": "uuid",
  "scope": "platform",
  "language_id": "uuid",
  "display_name": "...",
  "mime_type": "image/png",
  "media_kind": "image",
  "url": "signed GET URL",
  "size_bytes": 12345,
  "linked_lexeme_id": "uuid",
  "expires_at": "2026-12-31T00:00:00Z",
  "width": 100,
  "height": 100,
  "duration_ms": null,
  "created_at": "..."
}
```

**GET /platform/languages/{code}/media** — query: `?q=`, `?kind=image|audio|video`, `?page=`, `?limit=`  
**Response 200:** `MediaListResponse` `{ items: MediaAssetDTO[], total, page, limit }`

**PATCH /platform/media/{id}** — body (`PatchMediaRequest`):

```json
{
  "display_name": "Новое имя",
  "linked_lexeme_id": "uuid",
  "ttl_seconds": 0,
  "expires_at": null
}
```

**Квота:** `MEDIA_USER_QUOTA_BYTES` на `uploaded_by`; admin не ограничен.

**Проверка:** `just smoke-api` (auth + platform media end-to-end).

---

### Public languages — `lexicon` (:8082)

Gateway: `/languages/` → lexicon. Без JWT.

| Метод | Путь | Статус |
|-------|------|--------|
| GET | `/languages` | 200 |
| GET | `/languages/{code}` | 200 |
| GET | `/languages/{code}/alphabet` | 200 |

---

### Platform users — `auth` (:8081)

Gateway: `/api/v1/platform/users` → auth (не lexicon).

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| GET | `/platform/users` | admin | 200 `{ items, total }` |
| PATCH | `/platform/users/{userId}` | admin | 200 `UserDTO` |

Query GET: `?q=`, `?role=`, `?page=`, `?limit=`.

---

### Platform lexicon — `lexicon` (:8082)

Языки, алфавит, звуки, лексикон (29 ручек). Gateway: `/api/v1/platform/*` → lexicon.

| Группа | Пути | Auth |
|--------|------|------|
| Languages | `GET/POST/PATCH /platform/languages`, alphabet CRUD, reorder | admin |
| Grammar | `GET/POST /platform/languages/{code}/grammar-topics`, `PATCH/DELETE /platform/grammar-topics/{id}` | admin |
| Sounds | `GET/POST/PATCH/DELETE /platform/.../sounds`, letter links | admin |
| Lexicon | `GET/POST /platform/languages/{code}/lexicon`, lexeme/forms/translations/media CRUD | admin |

**Проверка:** `just verify-api` (секции 4–8).

---

### Teacher lexicon picker — `lexicon` (:8082)

Read-only для редактора. Gateway: `/api/v1/teacher/languages/{code}/lexicon`, `/api/v1/teacher/lexemes/` → lexicon.

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| GET | `/teacher/languages/{code}/lexicon` | teacher | 200 |
| GET | `/teacher/lexemes/{lexemeId}` | teacher | 200 |
| GET | `/teacher/lexemes/{lexemeId}/usage` | teacher | 200 |

---

### Teacher media — `media` (:8085)

`scope=teacher`, `owner_id=current_user`. Gateway: `/api/v1/teacher/media/` → media.

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| POST | `/teacher/media/presign` | teacher | 200 |
| POST | `/teacher/media/confirm` | teacher | 201 |
| GET | `/teacher/media` | teacher | 200 |
| GET | `/teacher/media/{id}` | teacher (owner) | 200 |
| PATCH | `/teacher/media/{id}` | teacher (owner) | 200 |
| DELETE | `/teacher/media/{id}` | teacher (owner) | 204 |
| GET | `/teacher/languages/{code}/media/platform` | teacher | 200 read-only picker |

---

### Content editor — `content` (:8083)

Gateway: `/api/v1/teacher/*` (кроме media/lexicon picker) → content. ~29 ручек.

| Группа | Пути | Auth |
|--------|------|------|
| Block types | `GET /teacher/block-types` | teacher |
| Courses | CRUD + publish, lessons CRUD + publish | owner |
| Sections/blocks | CRUD, reorder | owner |
| Coverage | `/teacher/courses/{id}/lexicon`, by-lesson, forms-coverage | owner |
| Invite | get/regenerate invite code | owner |
| Students | list students, progress, `POST /teacher/students` (email enroll) | owner |

Контракт JSON `config` для 17 MVP block types: [`services/content/docs/BLOCK_TYPES.md`](services/content/docs/BLOCK_TYPES.md).

---

### Learning (student flow) — `learning` (:8084)

Gateway: `/api/v1/courses/`, `/lessons/`, `/progress/`, `/review/`, `/dictionary/` → learning.

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| GET | `/courses/public` | public (optional JWT) | 200 published catalog |
| POST | `/courses/join` | student | 201 |
| GET | `/courses`, `/courses/{id}`, `/courses/{id}/lessons`, `/courses/{id}/outline` | enrollment | 200 (`target_language` filled) |
| GET | `/lessons/{id}`, `/lessons/{id}/flow` | enrollment | 200 (`resolved_lexemes` on lesson) |
| POST | `/progress/blocks/{id}/attempt` | enrollment | 200 (11 gradable types) |
| GET | `/progress/lessons/{id}` | enrollment | 200 |
| GET | `/progress/summary` | JWT | 200 aggregated stats |
| GET | `/review` | JWT | 200 |
| POST | `/review/session` | JWT | 200 start review session |
| GET | `/dictionary` | JWT | 200 |

Learning читает опубликованные уроки из `even_content` через `CONTENT_DATABASE_URL` (snapshots при join) и языки/лексемы из `even_lexicon` через `LEXICON_DATABASE_URL`.

**Seed + e2e:** `just seed-znakomstvo`, `just verify-api` (секция 15).

---

## ⬜ Phase 2 — вне MVP

| Фича | Статус |
|------|--------|
| Grammar topics CRUD | ✅ platform admin (`/platform/languages/{code}/grammar-topics`) |
| `POST /review/session` | ✅ |
| Email enroll `POST /teacher/students` | ✅ |
| Block config validation | ✅ on create/patch block |
| Block-type favorites | не в OpenAPI |
| Manual enrollments (admin) | не в OpenAPI |
| Strict JSON schema validation для block `config` | частично (required keys per type) |
| Bulk lexicon import, audit log, `GET /platform/stats` | не в OpenAPI |

