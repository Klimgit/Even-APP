# Even App — общее устройство

Приложение для изучения языков (первый курс — эвенский). Клиент на **Flutter**, бэкенд на **Go**, данные в **PostgreSQL**, медиа в **S3**. MVP — **только online**.

Глоссарий доменных терминов: [CONTEXT.md](./CONTEXT.md).

## Стек

| Слой | Технология |
|------|------------|
| Клиент | Flutter (iOS, Android, Web, Desktop) |
| API | Go, REST JSON |
| БД | PostgreSQL 16 |
| Медиа | S3 (AWS) |
| Auth | JWT access + refresh, bcrypt |

Офлайн-кэш (SQLite / Drift) — **Phase 2**, не MVP.

## Пользователи и доступ

Два независимых измерения:

| | `users.role` | `users.is_admin` |
|--|--|--|
| **Назначение** | student или teacher | доступ к платформе (`/platform/*`) |
| **При регистрации** | student (default) или teacher | всегда `false` |
| **Кто меняет** | platform administrator | platform administrator или seed в БД |

| Участник | role | is_admin | Возможности |
|----------|------|----------|-------------|
| **Student** | student | false | Уроки, прогресс, повторение, словарь |
| **Teacher** | teacher | false | Курсы, редактор уроков, coverage лексики, invite-код |
| **Platform administrator** | любой | true | Lexeme repository, алфавит, звуки, языки, пользователи |

Teacher с `is_admin = true` видит **Преподавание** и **Платформа** в одной сессии, без перелогина.

\* `TeacherTab` — `role = teacher`  
\** `PlatformTab` — `is_admin = true`

UI приложения в MVP — **всегда на русском**. Поле `courses.ui_language_id` — задел на будущее.

## Запись на курс

Единственный способ попасть на Course в MVP — **Invite code**:

- Teacher создаёт Course → система генерирует один код (бессрочный, без лимита).
- Student вводит код → создаётся `course_enrollment`.
- Ручная запись по email **не поддерживается** в MVP.

Teacher с `role = teacher` может одновременно быть enrolled на Course и учиться как Student.

## Два уровня данных

### Scope: язык (Lexeme repository)

Принадлежит `language_id`. Создание и правка — **platform administrator** (`is_admin = true`).

- Lexeme, LexemeForm, переводы
- Медиа к словам (`lexeme_media`)
- Алфавит / клавиатура (`alphabet_letters`)
- Фонемы (`sounds`)

Teacher при редактировании урока выбирает слова через **read-only picker**.

### Scope: курс

Принадлежит `course_id`. Ведёт **Teacher** (owner).

- Уроки, разделы, блоки
- Enrollments (через invite code)
- `course_lexeme_usage` — какие Lexeme и **LexemeForm** introduced / exercised / referenced

Курс **не копирует** слова — только ссылается на `lexeme_id` из хранилища.

## Иерархия контента

```
Course
 └── Lesson (draft | published)
      └── LessonSection          ← sidebar «Разделы»
           └── LessonBlock       ← block_type + config (JSONB)
```

Контент и упражнения — **одна сущность** `lesson_blocks` (LessonBlock). Тип блока — enum из **42 значений** (все реализуются в MVP). См. [DTO.md](./DTO.md).

## Прохождение урока (Lesson flow)

- Student идёт по `GET /lessons/{id}/flow` — **один шаг на экран**.
- После каждых **3 gradable-блоков** — injection одного due **Review item** (если есть).
- Grable-блоки проверяются на сервере (`BlockValidatorRegistry`).

## Блоки урока

- Каждый `block_type` — свой editor, player и validator.
- Config хранит ссылки на `lexeme_id`, `media_asset_id`, не дублирует медиа.
- `section_kind = homework` и блоки essay — **Phase 2**, не MVP.

## Клавиатура

`CustomLanguageKeyboard` — буквы из `GET /languages/{code}/alphabet`. Для эвенского: `Ӈ`, `ӈ` и т.д.

## Прогресс, словарь и повторение

1. **Прогресс** — `user_block_progress` по gradable-блокам.
2. **Провал** → `user_review_items` на уровне **sub_item** (не всего блока).
3. **Успех** → Lexeme попадает в **Personal dictionary** (`user_vocabulary`).
4. **Повторение:**
   - вкладка «Повторение»;
   - injection в Lesson flow (каждые 3 gradable).

Интервалы Review: 4ч → 1д → 3д → 7д (по `failure_count`).

## Публикация и republish

1. Teacher редактирует Lesson в `draft`.
2. `POST /teacher/lessons/{id}/publish` → `published`, `version++`, `LexiconIndexer`.
3. **Republish** (повторный publish) → **сброс всего прогресса** Student по этому Lesson.
4. Student видит только `published`; клиент refetch по `version` / ETag.
5. Конфликты параллного редактирования — **409 version conflict** (`If-Match`); lesson lock — Phase 2.

## MVP scope (кратко)

| В MVP | Phase 2 |
|-------|---------|
| Все 42 BlockType | Офлайн-кэш |
| Invite code enrollment | Homework / essay |
| Forms coverage | Lesson lock |
| Online-only | Локализация UI |

См. также: [API.md](./API.md), [DTO.md](./DTO.md)
