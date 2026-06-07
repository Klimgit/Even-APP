# API — статус реализации (MVP)

Сводка: **что уже в коде** и **что осталось для MVP**, с контрактами.

Полная спека: [API.md](API.md). Типы JSON: [DTO.md](DTO.md). Домен: [CONTEXT.md](CONTEXT.md). Как добавить ручку: [DEVELOPMENT.md](DEVELOPMENT.md) §9.

**Base URL (клиент):** `http://localhost:8080/api/v1` через gateway.

**Легенда:** ✅ реализовано · ⬜ MVP, не реализовано · 🚫 Phase 2 (вне MVP)

---

## Сводка

| Область | ✅ | ⬜ MVP |
|---------|---|--------|
| Auth | 4 | 0 |
| System (health/ready) | все сервисы | — |
| Platform media | 6 | 0 |
| Public languages | 0 | 3 |
| Platform admin | 0 | ~30 |
| Teacher | 0 | ~35 |
| Student (learning) | 0 | 9 |

**Итого бизнес-ручек:** 10 / ~87 MVP (без Phase 2).

---

## Общие соглашения

| Правило | Значение |
|---------|----------|
| Auth | `Authorization: Bearer <access_token>` |
| Ошибка | `{ "error": string, "message": string }` |
| Пагинация | `?page=1&limit=20` → `{ items, total, page, limit }` |
| ID | UUID string |
| Даты | ISO 8601 UTC |

### JWT (access token)

Поля в claims: `uid` (user id), `role` (`student` \| `teacher`), `is_admin` (bool).

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

## ⬜ MVP — нужно реализовать

Рекомендуемый порядок: **public languages → platform users/lexicon → teacher media/editor → student flow → seed урока «Знакомство»**.

---

### Public — языки (`lexicon`)

Без JWT (кроме если позже решите иначе). Gateway: `/languages/` → lexicon.

| Метод | Путь | Response |
|-------|------|----------|
| GET | `/languages` | `LanguageDTO[]` |
| GET | `/languages/{code}` | `LanguageDTO` |
| GET | `/languages/{code}/alphabet` | `AlphabetLetterDTO[]` |

```typescript
LanguageDTO {
  id, code, name, native_name, direction: "ltr"|"rtl", is_active
}
AlphabetLetterDTO {
  id, language_id, character, upper_char?, sort_order, label?
}
```

---

### Platform — пользователи (`lexicon`)

| Метод | Путь | Auth | Body / Query | Response |
|-------|------|------|--------------|----------|
| GET | `/platform/users` | admin | `?q=`, `?role=`, `?page=`, `?limit=` | `{ items: UserDTO[], total }` |
| PATCH | `/platform/users/{userId}` | admin | `{ role?, is_admin? }` | `UserDTO` |

---

### Platform — языки, алфавит, звуки (`lexicon`)

| Метод | Путь | Auth | Примечание |
|-------|------|------|------------|
| GET | `/platform/languages` | admin | `LanguageDTO[]` |
| POST | `/platform/languages` | admin | body: `code, name, native_name, direction` → 201 |
| PATCH | `/platform/languages/{code}` | admin | partial → `LanguageDTO` |
| GET | `/platform/languages/{code}/alphabet` | admin | `AlphabetLetterDTO[]` |
| POST | `/platform/languages/{code}/alphabet` | admin | body: `character, upper_char?, sort_order, label?` → 201 |
| PATCH | `/platform/alphabet/{letterId}` | admin | partial → `AlphabetLetterDTO` |
| DELETE | `/platform/alphabet/{letterId}` | admin | 204 |
| POST | `/platform/languages/{code}/alphabet/reorder` | admin | body: `{ letter_ids: string[] }` → 204 |
| GET | `/platform/languages/{code}/sounds` | admin | `SoundDTO[]` |
| POST | `/platform/languages/{code}/sounds` | admin | body: `ipa?, description?, audio_media_id?` → 201 |
| PATCH | `/platform/sounds/{soundId}` | admin | partial → `SoundDTO` |
| DELETE | `/platform/sounds/{soundId}` | admin | 204 |
| POST | `/platform/alphabet/{letterId}/sounds` | admin | body: `{ sound_id }` → 204 |
| DELETE | `/platform/alphabet/{letterId}/sounds/{soundId}` | admin | 204 |

```typescript
SoundDTO { id, language_id, ipa?, description?, audio_url? }
```

---

### Platform — лексикон (`lexicon`)

| Метод | Путь | Auth | Body / Response |
|-------|------|------|-----------------|
| GET | `/platform/languages/{code}/lexicon` | admin | query `?q=`, `?page=`, `?limit=` → `{ items: LexemeDTO[], total }` |
| POST | `/platform/languages/{code}/lexicon` | admin | `CreateLexemeRequest` → 201 `LexemeDTO` |
| GET | `/platform/lexemes/{lexemeId}` | admin | `LexemeDTO` |
| PATCH | `/platform/lexemes/{lexemeId}` | admin | partial → `LexemeDTO` |
| DELETE | `/platform/lexemes/{lexemeId}` | admin | 204 |
| POST | `/platform/lexemes/{lexemeId}/forms` | admin | `{ form, tags? }` → `LexemeFormDTO` |
| PATCH | `/platform/lexeme-forms/{formId}` | admin | partial → `LexemeFormDTO` |
| DELETE | `/platform/lexeme-forms/{formId}` | admin | 204 |
| POST | `/platform/lexemes/{lexemeId}/translations` | admin | `{ target_language_id, text, target_lexeme_id? }` → `LexemeTranslationDTO` |
| DELETE | `/platform/lexeme-translations/{translationId}` | admin | 204 |
| POST | `/platform/lexemes/{lexemeId}/media` | admin | `{ media_asset_id, kind, label?, is_primary?, form_id? }` → `LexemeMediaDTO` |
| DELETE | `/platform/lexeme-media/{lexemeMediaId}` | admin | 204 |

```typescript
LexemeDTO {
  id, language_id, lemma, part_of_speech?, notes?,
  translations: LexemeTranslationDTO[],
  forms: LexemeFormDTO[],
  media: LexemeMediaDTO[],
  primary_image_url?, primary_audio_url?
}
CreateLexemeRequest {
  lemma, part_of_speech?, notes?,
  translations?: { target_language_id, text }[]
}
```

---

### Teacher — медиа (`content` или `lexicon` по решению)

`scope=teacher`, `owner_id=current_user`. Те же DTO, что platform media (`PresignRequest`, `ConfirmMediaRequest`, `MediaAssetDTO`).

| Метод | Путь | Auth | Статус |
|-------|------|------|--------|
| POST | `/teacher/media/presign` | teacher | 200 |
| POST | `/teacher/media/confirm` | teacher | 201 |
| GET | `/teacher/media` | teacher | 200 `MediaListResponse` |
| GET | `/teacher/media/{mediaAssetId}` | teacher (owner) | 200 |
| PATCH | `/teacher/media/{mediaAssetId}` | teacher (owner) | 200 |
| DELETE | `/teacher/media/{mediaAssetId}` | teacher (owner) | 204 |
| GET | `/teacher/languages/{code}/media/platform` | teacher | 200 read-only platform picker |

Query для GET `/teacher/media`: `?q=`, `?kind=`, `?language_code=`, `?page=`, `?limit=`.

---

### Teacher — lexicon picker (`content` → lexicon read-only)

| Метод | Путь | Auth | Response |
|-------|------|------|----------|
| GET | `/teacher/languages/{code}/lexicon` | teacher | `{ items: LexemeDTO[], total }` |
| GET | `/teacher/lexemes/{lexemeId}` | teacher | `LexemeDTO` |
| GET | `/teacher/lexemes/{lexemeId}/usage` | teacher | query `?course_id=` required |

```json
{
  "lexeme_id": "uuid",
  "usages": [{
    "lesson_id", "lesson_title", "block_id",
    "display_label", "usage_kind": "introduced|exercised|referenced"
  }]
}
```

---

### Teacher — block types (`content`)

| Метод | Путь | Auth | Response |
|-------|------|------|----------|
| GET | `/teacher/block-types` | teacher | `BlockTypeCategoryDTO[]` |

Каталог типов блоков для редактора (MVP: 6 content + 11 gradable типов, см. [MVP.md](MVP.md)).

---

### Teacher — курсы и уроки (`content`)

| Метод | Путь | Auth | Body / Response |
|-------|------|------|-----------------|
| GET | `/teacher/courses` | teacher | `CourseDTO[]` |
| POST | `/teacher/courses` | teacher | `{ title, target_language_id, ui_language_id }` → 201 |
| GET | `/teacher/courses/{courseId}` | owner | `CourseDTO` |
| PATCH | `/teacher/courses/{courseId}` | owner | partial → `CourseDTO` |
| DELETE | `/teacher/courses/{courseId}` | owner | 204 |
| POST | `/teacher/courses/{courseId}/publish` | owner | `CourseDTO` |
| GET | `/teacher/courses/{courseId}/lessons` | owner | `LessonDTO[]` (summary) |
| POST | `/teacher/courses/{courseId}/lessons` | owner | `{ title, sort_order? }` → 201 |
| GET | `/teacher/lessons/{lessonId}` | owner | `LessonDTO` (full + sections + blocks) |
| PATCH | `/teacher/lessons/{lessonId}` | owner | partial; header `If-Match: <version>` |
| DELETE | `/teacher/lessons/{lessonId}` | owner | 204 |
| POST | `/teacher/lessons/{lessonId}/publish` | owner | `LessonDTO` |
| POST | `/teacher/lessons/{lessonId}/sections` | owner | `{ title, section_kind?, sort_order? }` → 201 |
| PATCH | `/teacher/sections/{sectionId}` | owner | partial → `LessonSectionDTO` |
| DELETE | `/teacher/sections/{sectionId}` | owner | 204 |
| POST | `/teacher/lessons/{lessonId}/sections/reorder` | owner | `{ section_ids: string[] }` → 204 |
| POST | `/teacher/lessons/{lessonId}/blocks` | owner | `CreateLessonBlockRequest` → 201 |
| GET | `/teacher/blocks/{blockId}` | owner | `LessonBlockDTO` |
| PATCH | `/teacher/blocks/{blockId}` | owner | partial config → `LessonBlockDTO` |
| DELETE | `/teacher/blocks/{blockId}` | owner | 204 |
| POST | `/teacher/lessons/{lessonId}/blocks/reorder` | owner | `{ block_ids: string[] }` → 204 |

```typescript
CourseDTO {
  id, title, target_language_id, ui_language_id,
  owner_id, is_published, invite_code?
}
LessonDTO {
  id, course_id, title, sort_order, version,
  status: "draft"|"published", sections: LessonSectionDTO[]
}
LessonBlockDTO {
  id, section_id?, sort_order, display_label?, title?,
  block_type, config: Record<string, unknown>,
  is_homework, is_gradable
}
```

---

### Teacher — coverage, invite, students (`content`)

| Метод | Путь | Auth | Response |
|-------|------|------|----------|
| GET | `/teacher/courses/{courseId}/lexicon` | owner | `{ introduced_count, exercised_count, lexeme_count }` |
| GET | `/teacher/courses/{courseId}/lexicon/by-lesson` | owner | `CourseLexiconByLessonDTO[]` |
| GET | `/teacher/courses/{courseId}/lexicon/forms-coverage` | owner | `FormsCoverageDTO[]` |
| GET | `/teacher/courses/{courseId}/invite-code` | owner | `{ invite_code }` |
| POST | `/teacher/courses/{courseId}/invite-code/regenerate` | owner | `{ invite_code }` |
| GET | `/teacher/courses/{courseId}/students` | owner | `StudentDTO[]` |
| GET | `/teacher/students/{studentId}/progress` | owner | `StudentProgressDTO` |

---

### Student — enrollment и уроки (`learning`)

| Метод | Путь | Auth | Body / Response |
|-------|------|------|-----------------|
| POST | `/courses/join` | student | `{ invite_code }` → 201 `JoinCourseResponse` |
| GET | `/courses` | JWT | `CourseListItemDTO[]` |
| GET | `/courses/{courseId}` | enrollment | `CourseDTO` |
| GET | `/courses/{courseId}/lessons` | enrollment | `{ items: [{ id, title, sort_order, completed_percent }] }` |
| GET | `/lessons/{lessonId}` | enrollment | `LessonDTO` (published) |
| GET | `/lessons/{lessonId}/flow` | enrollment | `LessonFlowDTO` |

```typescript
JoinCourseRequest { invite_code: string }
JoinCourseResponse { course_id, enrollment_id }
LessonFlowDTO {
  lesson_id, version,
  items: ( { kind: "lesson_block", block } | { kind: "review_injection", review } )[]
}
```

**Flow:** после каждых 3 gradable-блоков — 1 review injection (`due_at <= now()`).

---

### Student — progress, review, dictionary (`learning`)

| Метод | Путь | Auth | Body / Response |
|-------|------|------|-----------------|
| POST | `/progress/blocks/{blockId}/attempt` | enrollment | `BlockAttemptRequest` → `BlockAttemptResponse` |
| GET | `/progress/lessons/{lessonId}` | enrollment | `{ lesson_id, blocks: UserBlockProgressDTO[] }` |
| GET | `/review` | JWT | query `?status=`, `?due_only=` → `ReviewListResponse` |
| GET | `/dictionary` | JWT | query `?course_id=` → `VocabularyEntryDTO[]` |

```typescript
BlockAttemptRequest {
  sub_item_index?: number,
  response: Record<string, unknown>,
  context?: "lesson"|"review_tab"|"injected"
}
BlockAttemptResponse {
  is_correct, score, correct_answer?, block_progress
}
ReviewListResponse { pending_count, due_count, items: ReviewItemDTO[] }
VocabularyEntryDTO { lexeme: LexemeDTO, first_seen_at, mastery }
```

Side effects attempt: `block_attempts`, `user_block_progress`, `user_review_items`, `user_vocabulary`.

---

## 🚫 Phase 2 (не MVP)

Не реализовывать до закрытия MVP end-to-end:

| Ручки | Причина |
|-------|---------|
| `POST/DELETE /teacher/block-types/{type}/favorite` | удобство редактора |
| `POST /review/session` | достаточно `GET /review` |
| `GET/POST/PATCH/DELETE /platform/.../grammar-topics` | грамматика после MVP |
| `POST /teacher/students` (assign by email) | только invite code в MVP |

---

## Gateway — маршрутизация

| Префикс | Сервис |
|---------|--------|
| `/api/v1/auth/` | auth |
| `/api/v1/platform/` | lexicon |
| `/api/v1/teacher/` | content |
| `/api/v1/courses/`, `/lessons/`, `/progress/`, `/review/`, `/dictionary/` | learning |
| `/languages/` | lexicon |

Gateway проверяет JWT на protected маршрутах; public — см. `services/api-gateway/internal/middleware/auth.go`. Upstream-сервисы дублируют проверку на своих handlers.

---

## Обновление этого документа

При добавлении ручки:

1. Перенести строку из ⬜ в ✅ с актуальным контрактом.
2. Обновить счётчик в сводке.
3. Добавить кейс в `scripts/smoke-api.sh` для критичных сценариев.
