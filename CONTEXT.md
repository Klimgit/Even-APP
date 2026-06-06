# Even App

Приложение для изучения языков: курсы с уроками, лексическое хранилище на уровне языка, прогресс и повторение заданий.

## Language

**User**:
Человек с аккаунтом в системе (email + пароль).
_Avoid_: account, member, profile

**Role**:
Тип участия пользователя в продукте: `student` (учится) или `teacher` (создаёт курсы и уроки). Одна роль на пользователя.
_Avoid_: тип аккаунта, user type

**Platform administrator**:
Пользователь с флагом `is_admin = true`. Доступ к языковому хранилищу, алфавиту, звукам и управлению пользователями. Не отдельная роль — дополнение к любому User. Первые назначаются вручную через БД (seed); далее — через PATCH /platform/users/{id} существующим platform administrator. При регистрации всегда `false`.
_Avoid_: admin role, супер-админ, модератор

**Student**:
User с `role = student`. Проходит уроки, накапливает прогресс, повторяет проваленные задания, ведёт личный словарь.
_Avoid_: ученик (в UI допустимо), learner

**Teacher**:
User с `role = teacher`. Ведёт курсы: создаёт уроки, смотрит coverage лексики. Может одновременно быть записан на Course и учиться как Student. Становится Teacher при регистрации (выбор role) или сменой role platform administrator.
_Avoid_: редактор, автор, преподаватель (в UI допустимо)

**Course**:
Набор уроков одного преподавателя по одному целевому языку. Принадлежит одному Teacher (owner). Студенты попадают на Course через Enrollment.
_Avoid_: программа, класс, группа

**Enrollment**:
Связь Student ↔ Course. Без Enrollment студент не видит уроки курса и не отправляет прогресс. Создаётся только через Invite code (ручная запись по email в MVP нет).
_Avoid_: подписка, регистрация на курс

**Invite code**:
Код (или ссылка), который Teacher создаёт для Course. Student вводит код в приложении и получает Enrollment. Основной способ попадания на курс в MVP. Один код на курс, бессрочный, без лимита использований.
_Avoid_: промокод, referral link

### Language scope

**Language**:
Целевой или UI-язык в системе (например, эвенский). Владеет алфавитом, звуками и Lexeme repository.
_Avoid_: locale, lang

**Lexeme**:
Единица словаря языка: lemma, формы, переводы, медиа. Живёт в Language scope, не копируется в Course.
_Avoid_: word, vocabulary item, термин

**Lexeme repository**:
Языковое хранилище всех Lexeme для данного Language. Создание и правка — только Platform administrator. Teacher при редактировании урока выбирает Lexeme через read-only picker.
_Avoid_: словарь платформы, word bank, база слов

**Platform media library** (общая база):
Отдельный каталог медиа языка (`scope=platform`, `owner_id=null`). Наполняется **только** platform administrator — это не агрегат и не «сборка со всех учителей». Просмотр (поиск, picker) доступен всем авторизованным; загрузка и правка — только admin. При сборке урока учитель может вставить медиа из общей базы.
_Avoid_: S3 browser, assets folder, общая папка учителей

**Teacher media library** (личная база):
Изолированный каталог учителя (`scope=teacher`, `owner_id=teacher`). Учитель видит и управляет **только своими** файлами; медиа других учителей недоступны. Каждая загрузка — в хранилище с `display_name`, опционально TTL. Квота: `MEDIA_USER_QUOTA_BYTES`. При сборке урока учитель может вставить медиа из **своей** базы или из **общей** (read-only).
_Avoid_: teacher uploads, личные картинки

**Lesson media picker**:
При редактировании блока урока источник медиа — ровно один из двух: `platform` library или `teacher` library (своя). Ссылка в блоке — `media_asset_id`; сервис проверяет, что asset либо `scope=platform`, либо `scope=teacher` с `owner_id=current_teacher`.

**LexemeForm**:
Грамматическая форма Lexeme (например, «бишни» 3sg). Отслеживается в Course coverage: introduced / exercised по каждой форме.
_Avoid_: inflection, variant

### Course scope

**Lesson**:
Часть Course: разделы и блоки. Бывает в статусе draft (видит Teacher) или published (видит Student с Enrollment).
_Avoid_: занятие, unit, module

**LessonBlock**:
Единая единица контента и задания в Lesson. Имеет фиксированный тип (BlockType) и config. Контент и упражнение — одна сущность, не два слоя.
_Avoid_: exercise, step, slide, widget

**Course lexeme usage**:
Учёт использования Lexeme и LexemeForm в Course: где introduced, где exercised, в каком LessonBlock. Teacher видит coverage-отчёты; MVP отслеживает все формы.
_Avoid_: word tracking, vocabulary coverage

**Publish**:
Действие Teacher, переводящее Lesson из draft в published и увеличивающее version. **Republish** — повторный Publish; сбрасывает весь прогресс Student по этому Lesson.
_Avoid_: release, deploy, save

**Review item**:
Проваленное задание в очереди повторения Student. Привязано к LessonBlock + sub_item_index (не ко всему блоку целиком). Появляется во вкладке Review и inject'ится в Lesson flow после каждых 3 gradable-блоков (если есть due items).
_Avoid_: ошибка, mistake, failed exercise

**Lesson flow**:
Упорядоченная последовательность шагов прохождения Lesson для Student: блоки урока + Review item injection. UI показывает один шаг на экран.
_Avoid_: playbook, timeline, path

**Personal dictionary**:
Личный словарь Student (user_vocabulary): Lexeme, которые Student успешно отработал в gradable-заданиях (или intro с подтверждённым успехом). Наполняется автоматически, не вручную в MVP.
_Avoid_: словарик, word list, glossary (для student view)

## Flagged ambiguities

_(Все пункты ниже синхронизированы с APP.md, DTO.md, API.md.)_

## Example dialogue

> **Dev:** Учитель может зайти в хранилище слов?  
> **Expert:** Да, если у него стоит platform administrator. Иначе только picker при редактировании урока — read-only.  
> **Dev:** Нужен второй логин?  
> **Expert:** Нет. Один User, одна сессия. Role teacher + is_admin — видит и курсы, и платформу.
