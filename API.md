# Even App — REST API

Base URL: `/api/v1`

Формат: **JSON**. Auth: **Bearer JWT** (`Authorization: Bearer <access_token>`), кроме `/auth/*` и public alphabet.

См. также [MVP.md](./MVP.md). Ниже — полная спека; секции помечены **MVP** / **Phase 2**.

---

## MVP — сводка ручек

| Сервис | Группа | Ручки |
|--------|--------|-------|
| auth | Auth | `POST /auth/register`, `/login`, `/refresh`, `GET /auth/me` |
| lexicon | Public | `GET /languages`, `/languages/{code}`, `/languages/{code}/alphabet` |
| lexicon | Platform | users, languages, alphabet, sounds, lexicon CRUD, **platform media library** |
| content | Teacher | courses, lessons, blocks, coverage, **teacher media library**, invite code, students list |
| learning | Student | `POST /courses/join`, courses, lessons, flow, progress, review, dictionary |

**Вне MVP (Phase 2):** `grammar-topics`, email-enrollment (`/teacher/students` POST), manual enrollments, block-type favorites, `POST /review/session`.

---

## Общие соглашения


| Правило               | Значение                                             |
| --------------------- | ---------------------------------------------------- |
| ID                    | UUID string                                          |
| Даты                  | ISO 8601 UTC                                         |
| Пагинация             | `?page=1&limit=20` → `{ items, total, page, limit }` |
| Поиск                 | `?q=ашат`                                            |
| Сортировка            | `?sort=lemma&order=asc`                              |
| Ошибки                | `{ error, message, details? }`                       |
| Версионирование урока | header `If-Match: <version>` при update              |


### Middleware


| Middleware                     | Где                                    |
| ------------------------------ | -------------------------------------- |
| `RequireAuth`                  | все кроме auth, public alphabet        |
| `RequireEnrollment(course_id)` | ученик читает уроки курса              |
| `RequireTeacher()`             | `/teacher/*` — `role = teacher` (или `is_admin`) |
| `RequirePlatformAdmin()`       | `/platform/*` — `is_admin = true`      |
| `RequireTeacherOfStudent(id)`  | прогресс ученика                       |
| `RequireCourseOwner()`         | CRUD курса/урока                       |


---

## Auth

### POST /auth/register

**MVP.** Регистрация. `is_admin` всегда `false`. `role` по умолчанию `student`.

**Body:**

```json
{
  "email": "student@example.com",
  "password": "secret123",
  "display_name": "Аня",
  "role": "student"
}
```

`role`: `student` | `teacher`

**Response 201:** `AuthResponse`

---

### POST /auth/login

**Body:**

```json
{ "email": "...", "password": "..." }
```

**Response 200:** `AuthResponse`

---

### POST /auth/refresh

**Body:**

```json
{ "refresh_token": "..." }
```

**Response 200:** `{ access_token, refresh_token }`

---

### GET /auth/me

**Auth:** required

**Response 200:** `UserDTO`

---

## Public — языки и алфавит

### GET /languages

Список активных языков.

**Response 200:** `LanguageDTO[]`

---

### GET /languages/{code}

**Response 200:** `LanguageDTO`

---

### GET /languages/{code}/alphabet

Буквы для клавиатуры (все роли).

**Response 200:** `AlphabetLetterDTO[]`

---

## Public — ученик: курсы и уроки

### POST /courses/join

**MVP.** Запись на курс по invite code.

**Auth:** required (`role = student`)

**Body:**

```json
{ "invite_code": "A1B2C3D4" }
```

**Response 201:** `JoinCourseResponse`

**Errors:** `404` код не найден, `409` уже записан

---

### GET /courses

Курсы текущего пользователя (из `course_enrollments`).

**Response 200:** `CourseListItemDTO[]`

---

### GET /courses/{courseId}

**RequireEnrollment**

**Response 200:** `CourseDTO`

---

### GET /courses/{courseId}/lessons

Только `status = published`.

**Response 200:**

```json
{
  "items": [
    { "id": "...", "title": "Знакомство", "sort_order": 1, "completed_percent": 0.4 }
  ]
}
```

---

### GET /lessons/{lessonId}

Полный урок с разделами и блоками. **RequireEnrollment**.

**Response 200:** `LessonDTO`

---

### GET /lessons/{lessonId}/flow

Playbook прохождения: блоки + review-injection.

**Response 200:** `LessonFlowDTO`

**Логика injection:** после каждых 3 gradable-блоков — 1 item из `user_review_items` где `due_at <= now()`.

---

## Public — прогресс

### POST /progress/blocks/{blockId}/attempt

Отправка ответа на gradable-блок.

**RequireEnrollment**

**Body:** `BlockAttemptRequest`

```json
{
  "sub_item_index": 0,
  "response": { "selected_lexeme_id": "...", "text": "ашаткан" },
  "context": "lesson"
}
```

**Response 200:** `BlockAttemptResponse`

**Side effects:**

- insert `block_attempts`
- upsert `user_block_progress`
- при провале → upsert `user_review_items`
- при успехе → update `user_vocabulary`

---

### GET /progress/lessons/{lessonId}

Прогресс по блокам урока.

**Response 200:**

```json
{
  "lesson_id": "...",
  "blocks": [ { "lesson_block_id": "...", "status": "completed", "score": 1.0 } ]
}
```

---

## Public — повторение

### GET /review

Очередь проваленных блоков (`status = pending`).

**Response 200:** `ReviewListResponse`

Query: `?status=pending|mastered`, `?due_only=true`

---

### POST /review/session

Опционально: создать сессию повторения (фильтр due/all).

**Body:**

```json
{ "due_only": true, "limit": 20 }
```

**Response 200:** `{ session_id, items: ReviewItemDTO[] }`

---

## Public — словарь

### GET /dictionary

Личный словарь ученика.

**Response 200:** `VocabularyEntryDTO[]`

Query: `?course_id=...` (опционально фильтр)

---

## Teacher — block types

### GET /teacher/block-types

Каталог 42 типов блоков по категориям.

**RequireTeacher**

**Response 200:** `BlockTypeCategoryDTO[]`

---

### POST /teacher/block-types/{blockType}/favorite

**Response 204**

---

### DELETE /teacher/block-types/{blockType}/favorite

**Response 204**

---

## Teacher — lexicon (read-only picker)

### GET /teacher/languages/{code}/lexicon

Поиск лексем для picker (read-only).

**RequireTeacher**

Query: `?q=ашат`, `?page=1&limit=20`

**Response 200:** `{ items: LexemeDTO[], total }`

---

### GET /teacher/lexemes/{lexemeId}

**Response 200:** `LexemeDTO`

---

### GET /teacher/lexemes/{lexemeId}/usage

Где слово используется в курсе.

Query: `?course_id=...` (required)

**Response 200:**

```json
{
  "lexeme_id": "...",
  "usages": [
    {
      "lesson_id": "...",
      "lesson_title": "...",
      "block_id": "...",
      "display_label": "1.3",
      "usage_kind": "exercised"
    }
  ]
}
```

---

## Teacher — курсы

### GET /teacher/courses

Курсы, где `owner_id = current_user`.

**Response 200:** `CourseDTO[]`

---

### POST /teacher/courses

**Body:**

```json
{
  "title": "Эвенский A1",
  "target_language_id": "...",
  "ui_language_id": "..."
}
```

**Response 201:** `CourseDTO`

---

### GET /teacher/courses/{courseId}

**RequireCourseOwner**

**Response 200:** `CourseDTO`

---

### PATCH /teacher/courses/{courseId}

**Body:** partial `CourseDTO`

**Response 200:** `CourseDTO`

---

### DELETE /teacher/courses/{courseId}

**Response 204**

---

### POST /teacher/courses/{courseId}/publish

Опубликовать курс (`is_published = true`).

**Response 200:** `CourseDTO`

---

## Teacher — лексика курса (coverage)

### GET /teacher/courses/{courseId}/lexicon

Сводка лексики курса.

**Response 200:**

```json
{
  "introduced_count": 42,
  "exercised_count": 38,
  "lexeme_count": 45
}
```

---

### GET /teacher/courses/{courseId}/lexicon/by-lesson

**Response 200:** `CourseLexiconByLessonDTO[]`

---

### GET /teacher/courses/{courseId}/lexicon/forms-coverage

**Response 200:** `FormsCoverageDTO[]`

---

## Teacher — уроки

### GET /teacher/courses/{courseId}/lessons

Все уроки включая draft.

**Response 200:** `LessonDTO[]` (без blocks или summary)

---

### POST /teacher/courses/{courseId}/lessons

**Body:**

```json
{ "title": "Знакомство", "sort_order": 1 }
```

**Response 201:** `LessonDTO`

---

### GET /teacher/lessons/{lessonId}

Полный урок для редактора.

**Response 200:** `LessonDTO`

---

### PATCH /teacher/lessons/{lessonId}

**Header:** `If-Match: <version>`

**Body:** `{ title?, sort_order? }`

**Response 200:** `LessonDTO`  
**Response 409:** version conflict

---

### DELETE /teacher/lessons/{lessonId}

---

### POST /teacher/lessons/{lessonId}/publish

Draft → published. Запускает `LexiconIndexer`.

**Response 200:** `LessonDTO`

---

## Teacher — разделы урока

### POST /teacher/lessons/{lessonId}/sections

**Body:**

```json
{ "title": "Знакомство. Слова", "sort_order": 1, "section_kind": "content" }
```

**Response 201:** `LessonSectionDTO`

---

### PATCH /teacher/sections/{sectionId}

**Body:** `{ title?, sort_order?, section_kind? }`

**Response 200:** `LessonSectionDTO`

---

### DELETE /teacher/sections/{sectionId}

**Response 204**

---

### POST /teacher/lessons/{lessonId}/sections/reorder

**Body:**

```json
{ "section_ids": ["uuid1", "uuid2", "..."] }
```

**Response 200:** `LessonSectionDTO[]`

---

## Teacher — блоки урока

### POST /teacher/lessons/{lessonId}/blocks

**Body:** `CreateLessonBlockRequest`

**Response 201:** `LessonBlockDTO`

---

### GET /teacher/blocks/{blockId}

**Response 200:** `LessonBlockDTO`

---

### PATCH /teacher/blocks/{blockId}

**Body:** partial block (config, title, display_label, …)

**Response 200:** `LessonBlockDTO`

---

### DELETE /teacher/blocks/{blockId}

**Response 204**

---

### POST /teacher/lessons/{lessonId}/blocks/reorder

**Body:**

```json
{ "block_ids": ["...", "..."] }
```

**Response 200:** `LessonBlockDTO[]`

---

## Teacher — invite code и ученики курса

### GET /teacher/courses/{courseId}/invite-code

**MVP.** Текущий код курса (8 символов A–Z0–9).

**RequireCourseOwner**

**Response 200:** `InviteCodeResponse`

---

### POST /teacher/courses/{courseId}/invite-code/regenerate

**MVP.** Новый код (старый перестаёт работать).

**Response 200:** `InviteCodeResponse`

---

### GET /teacher/courses/{courseId}/students

**MVP.** Ученики с активным enrollment (не email-add).

**Response 200:** `StudentDTO[]`

---

### GET /teacher/students/{studentId}/progress

**RequireTeacherOfStudent**

Query: `?course_id=...`

**Response 200:** `StudentProgressDTO`

---

## Медиа: две независимые базы

| База | `scope` | Кто видит | Кто добавляет | В урок |
|------|---------|-----------|---------------|--------|
| **Общая (platform)** | `platform` | все авторизованные | только `is_admin` | да, picker |
| **Личная учителя** | `teacher` | только `owner_id` | владелец-учитель | да, picker |

Общая база **не** собирается из загрузок учителей — только явные загрузки админов. Учитель **не** видит чужие личные медиа.

---

## Teacher — медиа (личная база)

Каталог `scope=teacher`, `owner_id=current_user`. Только свои assets; чужие — 403.

### POST /teacher/media/presign

**MVP.** **RequireTeacher** (не student).

**Body:** `PresignRequest`

**Response 200:** `PresignResponse`

---

### POST /teacher/media/confirm

**MVP.** Создаёт `media_assets` с `scope=teacher`.

**Body:** `ConfirmMediaRequest` — обязателен `display_name`; опционально `ttl_seconds` / `expires_at`. Квота: `MEDIA_USER_QUOTA_BYTES` на пользователя.

**Response 201:** `MediaAssetDTO`

---

### GET /teacher/media

**MVP.** Личный каталог медиа учителя.

Query: `?q=`, `?kind=image|audio|video`, `?language_code=evn`, `?page=`, `?limit=`

**Response 200:** `MediaListResponse`

---

### GET /teacher/media/{mediaAssetId}

**MVP.** Только свои assets.

**Response 200:** `MediaAssetDTO`

---

### PATCH /teacher/media/{mediaAssetId}

**MVP.** `PatchMediaRequest` — имя, `linked_lexeme_id`, TTL.

**Response 200:** `MediaAssetDTO`

---

### DELETE /teacher/media/{mediaAssetId}

**MVP.** Если не используется в блоках.

**Response 204**

---

### GET /teacher/languages/{code}/media/platform

**MVP.** Read-only поиск platform library при сборке урока.

Query: `?q=`, `?kind=`, `?page=`, `?limit=`

**Response 200:** `MediaListResponse`

---

## Platform — пользователи и роли

### PATCH /platform/users/{userId}

**MVP.** **RequirePlatformAdmin**

**Body:**

```json
{ "role": "teacher", "is_admin": true }
```

Поля опциональны.

**Response 200:** `UserDTO`

---

### GET /platform/users

**MVP.** Список пользователей.

Query: `?q=`, `?role=teacher|student`, `?page=`, `?limit=`

**Response 200:** `{ items: UserDTO[], total }`

---

## Platform — языки

### GET /platform/languages

**Response 200:** `LanguageDTO[]`

---

### POST /platform/languages

**Body:**

```json
{
  "code": "evn",
  "name": "Even",
  "native_name": "Эвэды",
  "direction": "ltr"
}
```

**Response 201:** `LanguageDTO`

---

### PATCH /platform/languages/{code}

**Response 200:** `LanguageDTO`

---

## Platform — алфавит / клавиатура

### GET /platform/languages/{code}/alphabet

**Response 200:** `AlphabetLetterDTO[]`

---

### POST /platform/languages/{code}/alphabet

**Body:**

```json
{
  "character": "ӈ",
  "upper_char": "Ӈ",
  "sort_order": 10,
  "label": "ng"
}
```

**Response 201:** `AlphabetLetterDTO`

---

### PATCH /platform/alphabet/{letterId}

**Response 200:** `AlphabetLetterDTO`

---

### DELETE /platform/alphabet/{letterId}

**Response 204**

---

### POST /platform/languages/{code}/alphabet/reorder

**Body:** `{ letter_ids: string[] }`

**Response 200:** `AlphabetLetterDTO[]`

---

## Platform — звуки

### GET /platform/languages/{code}/sounds

**Response 200:** `SoundDTO[]`

---

### POST /platform/languages/{code}/sounds

**Body:**

```json
{
  "ipa": "/ŋ/",
  "description": "велярный носовой",
  "audio_key": "sounds/..."
}
```

**Response 201:** `SoundDTO`

---

### PATCH /platform/sounds/{soundId}

**Response 200:** `SoundDTO`

---

### DELETE /platform/sounds/{soundId}

**Response 204**

---

### POST /platform/alphabet/{letterId}/sounds

Привязать звук к букве.

**Body:** `{ sound_id: "..." }`

**Response 204**

---

### DELETE /platform/alphabet/{letterId}/sounds/{soundId}

**Response 204**

---

## Platform — лексическое хранилище

### GET /platform/languages/{code}/lexicon

**Query:** `?q=`, `?page=`, `?limit=`

**Response 200:** `{ items: LexemeDTO[], total }`

---

### POST /platform/languages/{code}/lexicon

**Body:** `CreateLexemeRequest`

**Response 201:** `LexemeDTO`

---

### GET /platform/lexemes/{lexemeId}

**Response 200:** `LexemeDTO`

---

### PATCH /platform/lexemes/{lexemeId}

**Body:** partial lexeme

**Response 200:** `LexemeDTO`

---

### DELETE /platform/lexemes/{lexemeId}

**Response 204**

---

### POST /platform/lexemes/{lexemeId}/forms

**Body:**

```json
{ "form": "бишни", "tags": { "person": "3sg" } }
```

**Response 201:** `LexemeFormDTO`

---

### PATCH /platform/lexeme-forms/{formId}

**Response 200:** `LexemeFormDTO`

---

### DELETE /platform/lexeme-forms/{formId}

**Response 204**

---

### POST /platform/lexemes/{lexemeId}/translations

**Body:**

```json
{ "target_language_id": "...", "text": "девочка" }
```

**Response 201:** `LexemeTranslationDTO`

---

### DELETE /platform/lexeme-translations/{translationId}

**Response 204**

---

### POST /platform/lexemes/{lexemeId}/media

Привязать медиа к лексеме.

**Body:**

```json
{
  "media_asset_id": "...",
  "kind": "image",
  "label": "иллюстрация",
  "is_primary": true,
  "form_id": null
}
```

**Response 201:** `LexemeMediaDTO`

---

### DELETE /platform/lexeme-media/{lexemeMediaId}

**Response 204**

---

## Platform — медиа (общая база)

Каталог `scope=platform`, `owner_id=null`. Отдельно от личных баз учителей.

### POST /platform/media/presign

**MVP.** **RequirePlatformAdmin**

**Body:** `PresignRequest`

**Response 200:** `PresignResponse`

---

### POST /platform/media/confirm

**MVP.** `scope=platform`, `owner_id=null`.

**Body:** `ConfirmMediaRequest`

**Response 201:** `MediaAssetDTO`

---

### GET /platform/languages/{code}/media

**MVP.** Каталог общей базы (без истёкших по TTL). **Любой авторизованный** (учитель — picker при сборке урока).

Query: `?q=`, `?kind=image|audio|video`, `?page=`, `?limit=`

**Response 200:** `MediaListResponse`

---

### GET /platform/media/{mediaAssetId}

**MVP.** **Любой авторизованный** (только `scope=platform`).

**Response 200:** `MediaAssetDTO` (signed URL)

---

### PATCH /platform/media/{mediaAssetId}

**MVP.** **RequirePlatformAdmin.** `PatchMediaRequest`

**Response 200:** `MediaAssetDTO`

---

### DELETE /platform/media/{mediaAssetId}

**MVP.** **RequirePlatformAdmin.** Если не в `lexeme_media` и не в блоках.

**Response 204**

---

## Phase 2 — грамматика (topics)

> Не в MVP (нет `grammar_table` / `grammar_exercise` в 17 BlockType).

### GET /platform/languages/{code}/grammar-topics

**Phase 2**

### POST /platform/languages/{code}/grammar-topics

**Phase 2**

### PATCH /platform/grammar-topics/{topicId}

**Phase 2**

### DELETE /platform/grammar-topics/{topicId}

**Phase 2**

---

## Примеры flow

### Учитель создаёт урок

```
POST /teacher/courses/{id}/lessons
POST /teacher/lessons/{id}/sections
POST /teacher/lessons/{id}/blocks  { block_type: "vocabulary_set", config: {...} }
POST /teacher/lessons/{id}/publish
```

### Ученик проходит урок

```
GET  /courses
GET  /lessons/{id}/flow
POST /progress/blocks/{id}/attempt  (repeat per gradable block)
```

### Админ добавляет слово

```
POST /platform/media/presign
PUT  <S3 upload_url>
POST /platform/media/confirm
POST /platform/languages/evn/lexicon
POST /platform/lexemes/{id}/media
```

---
