# Even App — общее устройство

Приложение для изучения языков (первый курс — эвенский). Клиент на **Flutter**, бэкенд на **Go**, данные в **PostgreSQL**, медиа в **S3**.

## Стек


| Слой          | Технология                                           |
| ------------- | ---------------------------------------------------- |
| Клиент        | Flutter (iOS, Android, Web, Desktop)                 |
| Локальный кэш | Drift (SQLite) — офлайн-чтение опубликованных уроков |
| API           | Go (chi/gin), REST JSON                              |
| БД            | PostgreSQL 16                                        |
| Медиа         | S3-совместимое хранилище (MinIO / AWS / Yandex)      |
| Auth          | JWT access + refresh, bcrypt                         |


## Роли пользователей


| Роль        | `users.role` | Возможности                                        |
| ----------- | ------------ | -------------------------------------------------- |
| **Ученик**  | `student`    | Уроки, прогресс, повторение, словарь               |
| **Учитель** | `teacher`    | Курсы, редактор уроков, ученики, coverage лексики курса |
| **Админ**   | `admin`      | Языковое хранилище, алфавит, звуки, языки|



## Два уровня данных

### Scope: язык (общее хранилище)

Принадлежит `language_id`. Наполняет **админ**.

- Лексемы, формы, переводы
- Картинки и аудио к словам (`lexeme_media`)
- Алфавит / клавиатура (`alphabet_letters`)
- Фонемы (`sounds`)

Используется **всеми курсами** этого языка. Загрузил один раз — переиспользуй везде.

### Scope: курс

Принадлежит `course_id`. Ведёт **учитель**.

- Уроки, разделы, блоки
- Запись учеников (`course_enrollments`)
- Учёт лексики (`course_lexeme_usage`) — что выдано, где использовано, какие формы тренировались

## Иерархия контента

```
Course
 └── Lesson (draft | published)
      └── LessonSection          ← sidebar «Разделы»
           └── LessonBlock       ← block_type + config (JSONB)
```

## Блоки урока

- Каждый `block_type` — свой виджет редактора, player и валидатор на бэкенде.
- Grable-блоки проверяются сервером (`BlockValidatorRegistry`).
- Config хранит ссылки на `lexeme_id`, `media_asset_id`, не дублирует медиа.

Категории типов: Изображения, Аудио/видео, Слова и пропуски, Базовые задания, Аудирование, Чтение, Грамматика, Тесты, Расставить, Текст, Прочее.

## Клавиатура

Для блоков с вводом текста на целевом языке — `CustomLanguageKeyboard`, буквы из `alphabet_letters` API. Для эвенского: `Ӈ`, `ӈ` и т.д.

## Прогресс и повторение

1. **Прогресс** — `user_block_progress` по каждому gradable-блоку.
2. **Провал** → `user_review_items` (очередь повторения).
3. **Два канала повторения:**
  - вкладка «Повторение» (явно);
  - автоподстановка в урок через `GET /lessons/{id}/flow` (playbook с `review_injection`).

Интервалы автоподстановки: 4ч → 1д → 3д → 7д (по `failure_count`).

## Flutter — вкладки

```dart
LessonTab()       // все роли (если enrolled)
ReviewTab()       // все
DictionaryTab()   // все
if (user.isTeacher)  TeacherTab()   // курсы, редактор, ученики
if (user.isPlatformAdmin) PlatformTab()  // хранилище, клавиатура, звуки
ProfileTab()
```

## Go Backend — структура

```
backend/
├── cmd/api/main.go
├── internal/
│   ├── auth/              # JWT, RequireTeacher, RequirePlatformAdmin, RequireEnrollment
│   ├── domain/
│   ├── repository/
│   ├── service/
│   │   ├── review/        # ReviewScheduler, flow injection
│   │   └── lexicon/       # LexiconIndexer
│   ├── handler/
│   │   ├── public/
│   │   ├── teacher/
│   │   └── platform/
│   ├── blocks/            # BlockValidatorRegistry
│   └── storage/           # S3 presign
└── migrations/
```

## Публикация контента

1. Учитель редактирует урок в статусе `draft`.
2. `POST /teacher/lessons/{id}/publish` → `status = published`, `version++`.
3. `LexiconIndexer` пересобирает `course_lexeme_usage` и `block_lexeme_refs`.
4. Ученики видят только `published`; клиент refetch по `version` / ETag.

## MVP — фазы

Детальный scope, 17 BlockType и критерии приёмки: **[MVP.md](./MVP.md)**.

**Фаза 1 (MVP):** auth, invite code, lexicon (platform), lesson editor + player для 17 block types, progress, review, урок «Знакомство».

**Фаза 2:** оставшиеся 25 block types, офлайн-кэш, homework/essay, reading, grammar, тесты, ProgressMe-шаблоны «Слова и пропуски».

## Связанные документы

- [MVP.md](./MVP.md) — scope MVP и минимальный набор block types
- [DTO.md](./DTO.md) — таблицы БД и DTO
- [Even app.txt](./Even%20app.txt) — исходное ТЗ
- [lesson_example.pdf](./lesson_example.pdf) — эталон урока

