# Even-APP — инструкция по разработке

Руководство для backend-разработки: локальный запуск, сборка, миграции, логи, отладка.

Связанные документы: [BACKEND.md](BACKEND.md) (архитектура скелета), [API.md](API.md), [DTO.md](DTO.md), [CONTEXT.md](CONTEXT.md).

---

## 1. Требования

| Инструмент | Версия | Зачем |
|------------|--------|-------|
| [Go](https://go.dev/dl/) | 1.23+ | сервисы, `go work` |
| [Docker](https://www.docker.com/) + Compose | актуальный | Postgres, MinIO, контейнеры сервисов |
| [just](https://github.com/casey/just) | любой | команды из `Justfile` |
| `curl` | — | health-check |
| `migrate` CLI | опционально | миграции с хоста (`brew install golang-migrate`) |

Без `migrate` на хосте миграции всё равно применяются через Docker при `just up`.

---

## 2. Первый запуск

```bash
git clone <repo>
cd Even-APP

cp .env.example .env   # скрипт just up сделает это сам
just up
```

`just up` (или `./scripts/up.sh`):

1. создаёт `.env`, если файла нет;
2. собирает и поднимает весь `docker-compose`;
3. ждёт `http://localhost:8080/api/v1/ready` (до 120 с);
4. печатает URL сервисов.

Проверка:

```bash
just health-check
curl http://localhost:8080/api/v1/gateway/status
```

Остановка:

```bash
just down
```

---

## 3. Режимы локальной разработки

### 3.1. Всё в Docker (рекомендуется для старта)

Удобно, когда нужно просто поднять стек и не трогать Go на хосте.

| Команда | Действие |
|---------|----------|
| `just up` | поднять всё |
| `just down` | остановить и убрать контейнеры |
| `just logs` | логи gateway + 4 сервиса (follow) |
| `just compose-logs` | логи всех контейнеров, включая postgres/minio |
| `just health-check` | smoke-тест всех `/health` и `/ready` |

Порты:

| Сервис | URL |
|--------|-----|
| API Gateway | http://localhost:8080 |
| auth | http://localhost:8081 |
| lexicon | http://localhost:8082 |
| content | http://localhost:8083 |
| learning | http://localhost:8084 |
| Postgres | `localhost:5432` (user/pass: `even` / `even`) |
| MinIO API | http://localhost:9000 |
| MinIO Console | http://localhost:9001 (`minio` / `minio123`) |

После изменения Go-кода:

```bash
docker compose up --build -d <service>   # например auth
# или пересобрать всё:
just up
```

### 3.2. Go на хосте, инфра в Docker (`up-local`)

Удобно при активной разработке одного сервиса: быстрый перезапуск бинарника без пересборки образа.

```bash
just up-local
```

Скрипт `scripts/up-local.sh`:

1. поднимает Postgres + MinIO;
2. гоняет миграции через Docker (`*-migrate` контейнеры);
3. останавливает docker-контейнеры приложений (освобождает 8080–8084);
4. `just build-all` → бинарники в `bin/`;
5. запускает 5 процессов в фоне, логи в `.dev/logs/`, PID в `.dev/pids/`.

Остановка:

```bash
just down-local
```

Логи:

```bash
tail -f .dev/logs/auth.log
tail -f .dev/logs/*.log
```

Пересборка одного сервиса после правок:

```bash
go build -o bin/auth ./services/auth/cmd
kill $(cat .dev/pids/auth.pid)
HTTP_PORT=8081 DATABASE_URL="$AUTH_DATABASE_URL" JWT_SECRET="$JWT_SECRET" ./bin/auth \
  >> .dev/logs/auth.log 2>&1 &
echo $! > .dev/pids/auth.pid
```

(или проще — снова `just up-local`).

### 3.3. Ручной запуск одного сервиса

Когда нужен один сервис в foreground с hot-reload через `go run`:

```bash
just infra-up                                    # только postgres + minio
docker compose up auth-migrate                   # миграции одного сервиса
docker compose stop auth                         # если порт занят

just run-auth-local                              # foreground, Ctrl+C для остановки
```

Аналоги: `run-lexicon-local`, `run-content-local`, `run-learning-local`, `run-gateway-local`.

**Важно:** gateway запускать последним — его `/api/v1/ready` пингует остальные сервисы.

---

## 4. Переменные окружения

Файл `.env` в корне (не коммитится). Шаблон — `.env.example`.

### Общие

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `JWT_SECRET` | — | секрет для JWT (обязателен) |

### Postgres

| Переменная | Описание |
|------------|----------|
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | учётка Postgres |
| `AUTH_DATABASE_URL` | DSN для `even_auth` |
| `LEXICON_DATABASE_URL` | DSN для `even_lexicon` |
| `CONTENT_DATABASE_URL` | DSN для `even_content` |
| `LEARNING_DATABASE_URL` | DSN для `even_learning` |

На хосте в DSN всегда `localhost:5432`. Внутри Docker compose подставляет `postgres:5432` сам.

### MinIO / S3

| Переменная | Описание |
|------------|----------|
| `S3_ENDPOINT` | внутренний URL (для сервисов: `http://minio:9000` в docker, `http://localhost:9000` на хосте) |
| `S3_PUBLIC_ENDPOINT` | URL для presigned-ссылок клиенту |
| `S3_BUCKET` | имя бакета (`even-media`) |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | ключи MinIO |

Нужны сервисам **lexicon** и **content**.

### Gateway upstream

| Переменная | На хосте | В Docker (compose) |
|------------|----------|-------------------|
| `AUTH_URL` | `http://localhost:8081` | `http://auth:8081` |
| `LEXICON_URL` | `http://localhost:8082` | `http://lexicon:8082` |
| `CONTENT_URL` | `http://localhost:8083` | `http://content:8083` |
| `LEARNING_URL` | `http://localhost:8084` | `http://learning:8084` |

### Порт сервиса

Каждый сервис читает **`HTTP_PORT`** (не `HTTP_PORT_AUTH` из `.env`). Рецепты `run-*-local` и compose задают его явно.

---

## 5. Сборка

### Весь backend

```bash
just build-all     # bin/api-gateway, bin/auth, ...
just tidy          # go work sync + go mod tidy во всех модулях
```

### Один сервис

```bash
go build -o bin/auth ./services/auth/cmd
go run ./services/auth/cmd          # из корня monorepo
```

Monorepo на `go.work` — зависимости `libs/*` подключаются через `replace` в `go.mod` каждого сервиса.

### Docker-образ

```bash
docker compose build auth
docker compose build               # все сервисы
```

Dockerfile каждого сервиса: multi-stage, `GOWORK=off`, копирует `libs/` + свой `services/<name>/`.

---

## 6. Миграции БД

Четыре отдельные БД: `even_auth`, `even_lexicon`, `even_content`, `even_learning`.

Файлы: `services/<svc>/database/migrations/`.

### Через Docker (без migrate на хосте)

```bash
docker compose up auth-migrate
docker compose up lexicon-migrate content-migrate learning-migrate
```

При `just up` миграции запускаются автоматически перед стартом сервисов.

### С хоста (нужен CLI migrate)

```bash
just migrate-all
```

Или для одного сервиса:

```bash
cd services/auth
DATABASE_URL="postgres://even:even@localhost:5432/even_auth?sslmode=disable" just migrate-up
```

Откат последней миграции:

```bash
DATABASE_URL="..." just -f services/auth/Justfile migrate-down
```

### Новая миграция

```bash
migrate create -ext sql -dir services/auth/database/migrations -seq add_users_table
```

Появятся `NNNNNN_add_users_table.up.sql` и `.down.sql`.

---

## 7. Логирование

### Как устроено

- Пакет `libs/core/logger` — JSON-логи в **stdout** через `slog`.
- Уровень из `LOG_LEVEL` (`debug` / `info` / `warn` / `error`).
- Middleware `libs/http/middleware`:
  - **Logging** — каждый HTTP-запрос: method, path, status, duration_ms;
  - **Recovery** — panic → 500 + stack в лог;
  - **CORS** — `*` для локальной разработки.

Пример строки лога запроса:

```json
{"time":"...","level":"INFO","msg":"request","method":"GET","path":"/health","status":200,"duration_ms":0}
```

Пример panic:

```json
{"time":"...","level":"ERROR","msg":"panic","err":"...","stack":"..."}
```

### Где смотреть

| Режим | Команда |
|-------|---------|
| Docker | `just logs` или `docker compose logs -f auth` |
| up-local | `tail -f .dev/logs/auth.log` |
| go run | прямо в терминале |

### Уровень debug

В `.env`:

```
LOG_LEVEL=debug
```

Перезапусти сервис. В коде используй переданный `*slog.Logger`:

```go
logr.Debug("loaded config", "port", cfg.Base.HTTPPort)
logr.Info("user created", "id", userID)
logr.Error("db query failed", "err", err)
```

### Логи в проде (позже)

Сейчас stdout → Docker logs / systemd journal. На VPS: собрать через Loki, Datadog или `docker compose logs` + ротация. Структурированный JSON упрощает парсинг.

---

## 8. Health и отладка

### Эндпоинты

| Путь | Смысл |
|------|-------|
| `GET /health` | процесс жив |
| `GET /api/v1/health` | то же |
| `GET /api/v1/ready` | готов принимать трафик |
| `GET /api/v1/openapi.yaml` | OpenAPI-заглушка сервиса |
| `GET /api/v1/gateway/status` | только gateway — список upstream |

**ready у сервисов** — ping Postgres (`pool.Ping`).

**ready у gateway** — GET `/api/v1/ready` на каждый upstream; при сбое `503`:

```json
{"status":"not_ready","reason":"auth: ready returned 503"}
```

### Полезные команды

```bash
just health-check

docker compose ps
docker compose logs auth --tail 50

# Postgres
docker compose exec postgres psql -U even -d even_auth -c '\dt'

# MinIO — бакет even-media создаётся init-скриптом
open http://localhost:9001
```

### Частые проблемы

| Симптом | Решение |
|---------|---------|
| `port already in use` | `docker compose stop api-gateway auth lexicon content learning` или `just down-local` |
| gateway `not_ready` | проверь, что все 4 backend подняты: `curl localhost:8081/api/v1/ready` |
| `required env X is not set` | проверь `.env`, для `go run` используй `just run-*-local` |
| миграции не применились | `docker compose up auth-migrate` и смотри вывод |
| после смены libs не собирается | `just tidy && just build-all` |

---

## 9. Работа над сервисом

### Структура сервиса

```
services/auth/
├── cmd/main.go              # точка входа
├── internal/
│   ├── config/              # env-конфиг
│   └── gen/http/v1/         # ogen (gitignored, генерируется)
├── api/http/v1/
│   ├── api.yaml             # OpenAPI
│   └── embed.go             # встроенная спека для /openapi.yaml
├── database/migrations/
├── Dockerfile
└── Justfile                 # swagger, migrate, run
```

### Типичный цикл

1. Правишь `api/http/v1/api.yaml`.
2. `just -f services/auth/Justfile swagger` — codegen (нужен ogen: `just -f services/auth/Justfile install-tools`).
3. Добавляешь миграцию + handler.
4. Локально: `just run-auth-local` или `just up-local`.
5. Проверяешь через gateway: `curl http://localhost:8080/api/v1/auth/...`.

### OpenAPI / ogen

```bash
just -f services/auth/Justfile install-tools   # один раз: ogen + migrate
just -f services/auth/Justfile swagger         # перегенерация handlers
```

Общие фрагменты спеки: `_shared/openapi/`. Генератор-хелпер: `_misc/openapi-handler-gen/`.

`internal/gen/` в `.gitignore` — генерируется локально и в CI.

---

## 10. API Gateway

Прокси-префиксы (запросы с клиента → upstream):

| Префикс | Сервис |
|---------|--------|
| `/api/v1/auth/` | auth |
| `/api/v1/platform/` | lexicon |
| `/api/v1/teacher/` | content |
| `/api/v1/courses/`, `/lessons/`, `/progress/`, `/review/`, `/dictionary/` | learning |
| `/languages/` | lexicon |

Merged OpenAPI: `GET http://localhost:8080/api/v1/openapi.yaml` (скелет, полный merge позже).

JWT middleware в gateway — в планах, пока не реализован.

---

## 11. Общие пакеты (`libs/`)

| Пакет | Назначение |
|-------|------------|
| `libs/config` | `LoadBase`, `LoadS3`, `MustGetenv` |
| `libs/core/logger` | JSON slog |
| `libs/http/middleware` | logging, recovery, CORS |
| `libs/http/server` | `Run`, `RegisterHealth`, `RegisterReady` |
| `libs/postgres` | pgx pool из `DATABASE_URL` |

Изменения в `libs/` затрагивают все сервисы — после правок `just build-all` или пересборка нужных контейнеров.

---

## 12. Справочник команд

```bash
just --list              # все рецепты

# Жизненный цикл
just up / just down
just up-local / just down-local

# Сборка и зависимости
just build-all
just tidy

# Docker
just compose-up
just compose-down
just logs
just compose-logs

# Инфра отдельно
just infra-up
just migrate-all
just health-check

# Один сервис (foreground)
just run-auth-local
just run-gateway-local
```

---

## 13. Деплой на VPS (черновик)

Полного CI/CD пока нет. Общая схема:

1. На сервере: Docker + Compose, клон репозитория.
2. `.env.prod` с боевыми секретами (`JWT_SECRET`, пароли Postgres, MinIO на том же VPS).
3. `S3_ENDPOINT` — внутренний (`http://minio:9000`), `S3_PUBLIC_ENDPOINT` — публичный URL для клиентов.
4. `docker compose up --build -d` (или `just up`).
5. Reverse proxy (nginx/caddy) на `:443` → gateway `:8080`.
6. TLS, firewall, смена дефолтных паролей MinIO/Postgres.

Volumes `postgres_data` и `minio_data` в compose — данные переживают перезапуск контейнеров.

---

## 14. Что дальше

Текущий скелет без бизнес-логики. Порядок разработки по [MVP.md](MVP.md):

1. auth — register/login, JWT, `is_admin`
2. lexicon — языки, алфавит, presign
3. content — курсы, уроки, invite code
4. learning — enrollment, flow, progress
5. gateway — JWT middleware
6. Flutter `apps/mobile/`

При добавлении фичи: миграция → handler → тест через `just health-check` и curl через gateway `:8080`.
