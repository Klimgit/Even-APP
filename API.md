# Even App — REST API

Base URL: `/api/v1`

Формат: **JSON**. Auth: **Bearer JWT** (`Authorization: Bearer <access_token>`), кроме `/auth/*`.

DTO описаны в [DTO.md](./DTO.md).

---

## Общие соглашения

| Правило | Значение |
|---------|----------|
| ID | UUID string |
| Даты | ISO 8601 UTC |
| Пагинация | `?page=1&limit=20` → `{ items, total, page, limit }` |
| Поиск | `?q=ашат` |
| Сортировка | `?sort=lemma&order=asc` |
| Ошибки | `{ error, message, details? }` |
| Версионирование урока | header `If-Match: <version>` при update |

### Middleware

| Middleware | Где |
|------------|-----|
| `RequireAuth` | все кроме auth, public alphabet |
| `RequireEnrollment(course_id)` | ученик читает уроки курса |
| `RequireTeacher()` | `/teacher/*` — `role = teacher` |
| `RequirePlatformAdmin()` | `/platform/*` — `is_admin = true` |
| `RequireTeacherOfStudent(id)` | прогресс ученика на курсе owner |
| `RequireCourseOwner()` | CRUD курса/урока |

---

## Auth

### POST /auth/register

Регистрация ученика или учителя. `is_admin` всегда `false`.

**Body:**
```json
{
  "email": "student@example.com",
  "password": "secret123",
  "display_name": "Аня",
  "role": "student"
}
```

Правила:
- `role` опционально, default `"student"`.
- Допустимые значения: `"student"` | `"teacher"`.
- Поле `is_admin` в body → **400 Bad Request**.

**Response 201:** `AuthResponse`

JWT claims: `{ sub, role, is_admin }`.

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

## Public — ученик: запись на курс

### POST /courses/join

**RequireAuth** (любой role)

Запись на Course по invite code. Единственный способ Enrollment в MVP.

**Body:** `JoinCourseRequest`
```json
{ "invite_code": "EVNA1X7K" }
```

**Response 201:** `JoinCourseResponse`

**Response 404:** код не найден  
**Response 409:** уже записан на этот курс

---

## Public — ученик: курсы и уроки

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

Playbook прохождения: блоки + review-injection. UI: **один item flow на экран**.

**Response 200:** `LessonFlowDTO`

**Логика injection:** после каждых 3 gradable-блоков — 1 due **Review item** (`sub_item_index`) из `user_review_items` где `due_at <= now()`.

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
- insert `block_attempts` (с `sub_item_index`)
- upsert `user_block_progress`
- при провале → upsert `user_review_items` для `(lesson_block_id, sub_item_index)`
- при успехе → upsert `user_vocabulary` (Personal dictionary)

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

Очередь Review items (`status = pending`), привязка к `lesson_block_id` + `sub_item_index`.

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

Personal dictionary — Lexeme после **успешных** gradable-ответов.

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

**Response 201:** `CourseDTO` (включая `invite_code`)

При создании курса генерируется `invite_code` (8 символов, уникальный).

---

### GET /teacher/courses/{courseId}

**RequireCourseOwner**

**Response 200:** `CourseDTO` (owner видит `invite_code`)

---

### PATCH /teacher/courses/{courseId}

**RequireCourseOwner**

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

Отчёт по **всем LexemeForm** курса: introduced / exercised.

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
**Response 409:** version conflict — урок изменён другим редактором; клиент refetch и merge/retry. Lesson lock — Phase 2.

---

### DELETE /teacher/lessons/{lessonId}

**Response 204**

---

### POST /teacher/lessons/{lessonId}/publish

Draft → published. Запускает `LexiconIndexer`.

**Republish** (урок уже `published`): `version++`, пересборка lexicon refs, **сброс всего прогресса** Student по этому уроку (`user_block_progress`, связанные `user_review_items`).

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

## Teacher — ученики курса

### GET /teacher/courses/{courseId}/students

**RequireCourseOwner**

Список Student, записанных на курс через invite code.

**Response 200:** `EnrolledStudentDTO[]`

---

### GET /teacher/courses/{courseId}/students/{studentId}/progress

**RequireCourseOwner**

Query: `?course_id=` (должен совпадать с `{courseId}`)

**Response 200:** `StudentProgressDTO`

---

## Teacher — медиа (для блоков урока)

### POST /teacher/media/presign

**Body:** `PresignRequest`

**Response 200:** `PresignResponse`

---

### POST /teacher/media/confirm

**Body:** `ConfirmMediaRequest`

**Response 201:** `MediaAssetDTO`

---

## Platform — пользователи

### PATCH /platform/users/{userId}

**RequirePlatformAdmin**

Назначение platform admin, смена роли student ↔ teacher.

**Body:** `UpdateUserRequest`
```json
{ "is_admin": true }
```

```json
{ "role": "teacher" }
```

Правила:
- `{ "is_admin": true }` — выдать доступ к `/platform/*`.
- `{ "is_admin": false }` — снять.
- `{ "role": "student" | "teacher" }` — смена роли (поддержка).
- Нельзя назначить себе `is_admin`, если текущий пользователь не platform admin.

**Response 200:** `UserDTO`

---

### GET /platform/users

**RequirePlatformAdmin**

Query: `?role=student|teacher`, `?is_admin=true|false`, `?page=`, `?limit=`

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

## Platform — медиа (для лексем)

### POST /platform/media/presign

**Body:** `PresignRequest`

**Response 200:** `PresignResponse`

---

### POST /platform/media/confirm

**Body:** `ConfirmMediaRequest`

**Response 201:** `MediaAssetDTO`

---

### GET /platform/media/{mediaAssetId}

**Response 200:** `MediaAssetDTO` (с signed URL)

---

## Platform — грамматика (topics)

### GET /platform/languages/{code}/grammar-topics

**Response 200:** `GrammarTopicDTO[]`

---

### POST /platform/languages/{code}/grammar-topics

**Body:**
```json
{
  "title": "Спряжение биш-",
  "body_richtext": "...",
  "table_data": { "rows": [] }
}
```

**Response 201:** `GrammarTopicDTO`

---

### PATCH /platform/grammar-topics/{topicId}

**Response 200:** `GrammarTopicDTO`

---

### DELETE /platform/grammar-topics/{topicId}

**Response 204**

---

## Webhooks / будущее (v2)

| Endpoint | Назначение |
|----------|------------|
| `POST /teacher/lessons/{id}/lock` | блокировка при редактировании |
| `DELETE /teacher/lessons/{id}/lock` | снять блокировку |
| `GET /health` | healthcheck |
| `GET /ready` | readiness (DB, S3) |

---

## Матрица доступа (кратко)

| Endpoint group | student | teacher | is_admin |
|----------------|---------|---------|----------|
| /auth/* | ✓ | ✓ | ✓ |
| /courses, /lessons, /progress, /review, /dictionary | ✓ | ✓* | ✓* |
| /teacher/* | | ✓ | ✓** |
| /platform/* | ✓*** | ✓*** | ✓ |

\* если enrolled на курс (любой role)  
\** `/teacher/*` только при `role = teacher`; `is_admin` без `teacher` не даёт доступ  
\*** `/platform/*` при `is_admin = true` (любой role)

---

## Примеры flow

### Ученик записывается на курс

```
POST /courses/join  { "invite_code": "EVNA1X7K" }
GET  /courses
GET  /lessons/{id}/flow
POST /progress/blocks/{id}/attempt  (repeat per gradable sub_item)
```

### Учитель создаёт курс и урок

```
POST /teacher/courses
POST /teacher/courses/{id}/lessons
POST /teacher/lessons/{id}/sections
POST /teacher/lessons/{id}/blocks  { block_type: "vocabulary_set", config: {...} }
POST /teacher/lessons/{id}/publish
```

### Platform administrator добавляет слово

```
POST /platform/media/presign
PUT  <S3 upload_url>
POST /platform/media/confirm
POST /platform/languages/evn/lexicon
POST /platform/lexemes/{id}/media
```

---

См. также: [APP.md](./APP.md), [DTO.md](./DTO.md)
