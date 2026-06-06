# Even-APP — инструкция по разработке

Сценарий записи демо-ролика: [DEMO.md](DEMO.md). Статус API: [API_STATUS.md](API_STATUS.md).

---

## 1. Требования

| Инструмент | Версия | Зачем |
|------------|--------|-------|
| [Go](https://go.dev/dl/) | 1.23+ | сервисы, `go work` |
| [Docker](https://www.docker.com/) + Compose | актуальный | Postgres, MinIO, контейнеры сервисов |
| [just](https://github.com/casey/just) | любой | команды из `Justfile` |
| `curl` | — | health-check |
| `migrate` CLI | опционально | миграции с хоста (`brew install golang-migrate`) |
---

## 2. Первый запуск

```bash
git clone <repo>
cd Even-APP

cp .env.example .env   # скрипт just up сделает это сам
just up
```

### Пример .env
```yaml
# Copy to .env for local dev: cp .env.example .env

# --- Postgres ---
POSTGRES_USER=even
POSTGRES_PASSWORD=even
POSTGRES_HOST=localhost
POSTGRES_PORT=5432

# Per-service DSNs (docker hostnames vs localhost)
AUTH_DATABASE_URL=postgres://even:even@localhost:5432/even_auth?sslmode=disable
LEXICON_DATABASE_URL=postgres://even:even@localhost:5432/even_lexicon?sslmode=disable
CONTENT_DATABASE_URL=postgres://even:even@localhost:5432/even_content?sslmode=disable
LEARNING_DATABASE_URL=postgres://even:even@localhost:5432/even_learning?sslmode=disable

# --- MinIO / S3 ---
S3_ENDPOINT=http://localhost:9000
S3_PUBLIC_ENDPOINT=http://localhost:9000
S3_BUCKET=even-media
S3_ACCESS_KEY=minio
S3_SECRET_KEY=minio123

# --- Auth ---
JWT_SECRET=dev-change-me-in-production

# --- Service ports (local go run) ---
HTTP_PORT_GATEWAY=8080
HTTP_PORT_AUTH=8081
HTTP_PORT_LEXICON=8082
HTTP_PORT_CONTENT=8083
HTTP_PORT_LEARNING=8084

# --- Gateway upstreams (local go run; docker-compose overrides to service hostnames) ---
AUTH_URL=http://localhost:8081
LEXICON_URL=http://localhost:8082
CONTENT_URL=http://localhost:8083
LEARNING_URL=http://localhost:8084

LOG_LEVEL=info
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

### 3.1. Всё в Docker

| Команда | Действие |
|---------|----------|
| `just up` | поднять всё (миграции применяются автоматически до старта сервисов) |
| `just down` | остановить и убрать контейнеры |
| `just logs` | логи gateway + 4 сервиса (follow) |
| `just compose-logs` | логи всех контейнеров, включая postgres/minio |
| `just health-check` | smoke-тест всех `/health` и `/ready` |

Порты(можно поменять в .env):

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
2. применяет миграции (`just migrate`);
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
just migrate                                     # все БД (или просто just up — миграции там же)
# одна БД: docker compose up auth-migrate
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

## 6. Тесты

```bash
just test              # unit: libs/jwt, libs/media
just test-integration  # docker stack + smoke-api
just smoke-api         # только HTTP smoke (сервисы уже запущены)
```

---

## 7. Миграции БД

Четыре отдельные БД: `even_auth`, `even_lexicon`, `even_content`, `even_learning`.

Файлы: `services/<svc>/database/migrations/`.

### Локально (автоматически)

`just up` и `docker compose up` применяют миграции **до** старта приложений: сервисы `auth-migrate`, `lexicon-migrate`, … в `depends_on` у каждого app.

После добавления нового `.sql` — снова `just up` (или `just migrate` без перезапуска всего стека).

### Вручную

```bash
just migrate          # все четыре БД
# одна БД:
docker compose rm -sf auth-migrate && docker compose up auth-migrate
```

### С хоста

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

---
## 9. Работа над сервисом

### Структура сервиса

```
services/auth/
├── cmd/main.go              # точка входа
├── internal/
│   ├── config/              # env-конфиг
│   ├── httpapi/             # HTTP handlers + router.go
│   ├── store/               # SQL (pgx), без HTTP
│   └── gen/http/v1/         # ogen (gitignored, опционально)
├── api/http/v1/
│   ├── api.yaml             # OpenAPI
│   └── embed.go             # встроенная спека для /openapi.yaml
├── database/migrations/
├── Dockerfile
└── Justfile                 # swagger, migrate, run
```

Референсы реализованных ручек:

- **auth:** `services/auth/internal/httpapi/auth.go`, `internal/store/users.go`
- **platform media:** `services/lexicon/internal/httpapi/platform_media.go`, `internal/store/media.go`

---

### Как добавить новую ручку (пошагово)

Клиент ходит в **gateway** `:8080`. Каждый сервис — своя БД, миграции **до** старта приложения.

#### 0. Спроектировать

1. Выбрать **сервис** по URL-префиксу.
2. Решить **auth**: public / любой JWT / teacher / platform admin (`is_admin`).
3. Нужна ли новая таблица или колонка → миграция; иначе сразу store.

#### 1. Миграция (если меняется схема)

```bash
migrate create -ext sql -dir services/lexicon/database/migrations -seq add_lexemes
just migrate
# одна БД:
docker compose up lexicon-migrate
```

Файлы: `services/<svc>/database/migrations/NNNNNN_name.up.sql` и `.down.sql`.

Проверка:

```bash
docker compose exec postgres psql -U even -d even_lexicon -c '\dt'
```

#### 2. Store — SQL и данные

Новый или существующий файл в `services/<svc>/internal/store/`.

- `*pgxpool.Pool`, `context.Context` первым аргументом
- параметризованный SQL (`$1`, `$2`), без конкатенации user input
- доменные фильтры в WHERE (`scope`, `owner_id`, `language_id`, …)
- store **не** импортирует HTTP и JWT — только данные и ошибки

Несколько таблиц в одной операции — транзакция в store (`pool.Begin` → `Commit` / `Rollback`).

#### 3. Handler — бизнес-логика и HTTP

Новый файл в `services/<svc>/internal/httpapi/` (или метод в существующем handler).

Слои в handler:

1. парсинг body / query / path (`r.PathValue("id")` в Go 1.22+)
2. auth: `middleware.ClaimsFromContext(r.Context())` → `UserID`, `Role`, `IsAdmin`
3. проверка прав (например `requireAdmin` как в `platform_media.go`)
4. вызов store (+ S3 / квота / валидация из `libs/media` при необходимости)
5. ответ: `writeJSON` / `writeErr`; ошибки `{ "error", "message" }`

Регистрация маршрута в `Register(mux, jwtMW)`:

```go
// public
mux.HandleFunc("GET /api/v1/languages", h.list)

// с JWT
mux.Handle("GET /api/v1/auth/me", jwtMW(http.HandlerFunc(h.me)))
```

#### 4. Подключить в router

`services/<svc>/internal/httpapi/router.go` — создать store, handler, вызвать `Register`.

Без этого шага ручка не появится, даже если handler написан.

#### 5. Gateway

Большинство префиксов уже в `services/api-gateway/cmd/main.go`. Новый префикс — добавить в `routes` там же.

Тестировать **через gateway**:

```bash
curl http://localhost:8080/api/v1/...
```

#### 6. Конфиг (если нужны новые env)

`services/<svc>/internal/config/config.go`, при необходимости `libs/config/`, `.env.example`.

#### 7. Сборка и проверка

```bash
docker compose up --build -d lexicon   # или auth, content, learning
# либо на хосте:
just run-lexicon-local

curl -s http://localhost:8080/api/v1/... | jq .
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/...
```

Опционально: кейс в `scripts/smoke-api.sh`. После правок `libs/`: `just tidy && just build-all`.

#### 8. Документация

- [API.md](API.md) — описание ручки, коды ошибок
- [DTO.md](DTO.md) — JSON-типы
- `services/<svc>/api/http/v1/api.yaml` — OpenAPI (желательно)

#### Чеклист перед merge

- [ ] API.md + DTO.md
- [ ] миграция `.up.sql` + `.down.sql` (если нужна)
- [ ] `internal/store` + `internal/httpapi` + `router.go`
- [ ] gateway route (если новый префикс)
- [ ] `just migrate` локально
- [ ] curl / smoke через `:8080`
- [ ] нет секретов и полных JWT в логах

#### Типичные проверки прав

| Доступ | Условие в handler |
|--------|-------------------|
| Public | без `jwtMW` на route |
| Любой авторизованный | `jwtMW` |
| Platform admin | `claims.IsAdmin` |
| Teacher | `claims.Role == "teacher" \|\| claims.IsAdmin` |
| Владелец ресурса | `owner_id == claims.UserID` |

Кто загрузил файл — `uploaded_by = claims.UserID`. В какой «базе» лежит медиа — задаётся **endpoint** (`/platform/` vs `/teacher/`) и полем `scope` в store, не ролью alone.

---

### Пример: `GET /languages` (public, lexicon)

Реальный MVP-endpoint из [API_STATUS.md](API_STATUS.md): список активных языков. Таблица `languages` уже есть (миграция `000002_platform_media`), JWT не нужен.

#### Шаг 0 — контракт

**API.md** (уже описано):

- `GET /languages` → `200`, тело: `LanguageDTO[]`
- только `is_active = true`

**DTO.md:**

```typescript
LanguageDTO {
  id: string
  code: string
  name: string
  native_name: string
  direction: "ltr" | "rtl"
  is_active: boolean
}
```

Сервис: **lexicon**. Auth: **public** (без JWT на handler и в [gateway IsPublic](services/api-gateway/internal/middleware/auth.go)).

#### Шаг 1 — миграция

**Пропускаем** — таблица и seed (`evn`, `ru`) уже в `000002_platform_media.up.sql`.

Проверить данные:

```bash
docker compose exec postgres psql -U even -d even_lexicon \
  -c "SELECT code, name FROM languages WHERE is_active;"
```

#### Шаг 2 — store

Создать `services/lexicon/internal/store/languages.go`:

```go
package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Language struct {
	ID         uuid.UUID
	Code       string
	Name       string
	NativeName string
	Direction  string
	IsActive   bool
}

type LanguageStore struct {
	pool *pgxpool.Pool
}

func NewLanguageStore(pool *pgxpool.Pool) *LanguageStore {
	return &LanguageStore{pool: pool}
}

func (s *LanguageStore) ListActive(ctx context.Context) ([]Language, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, code, name, native_name, direction, is_active
		FROM languages
		WHERE is_active = true
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.ID, &l.Code, &l.Name, &l.NativeName, &l.Direction, &l.IsActive); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
```

#### Шаг 3 — handler

Создать `services/lexicon/internal/httpapi/languages.go`:

```go
package httpapi

import (
	"net/http"

	"github.com/even-app/even-app/services/lexicon/internal/store"
)

type LanguagesHandler struct {
	Store *store.LanguageStore
}

func (h *LanguagesHandler) Register(mux *http.ServeMux) {
	// Public — без jwtMW (gateway тоже пропускает GET /languages)
	mux.HandleFunc("GET /languages", h.list)
}

func (h *LanguagesHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListActive(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]map[string]any, 0, len(items))
	for _, l := range items {
		dtos = append(dtos, map[string]any{
			"id": l.ID.String(), "code": l.Code, "name": l.Name,
			"native_name": l.NativeName, "direction": l.Direction, "is_active": l.IsActive,
		})
	}
	writeJSON(w, http.StatusOK, dtos)
}
```

`writeJSON` / `writeErr` — скопировать из `platform_media.go` или вынести в `httpapi/response.go` при росте сервиса.

#### Шаг 4 — router

В `services/lexicon/internal/httpapi/router.go` после `MediaStore`:

```go
langStore := store.NewLanguageStore(pool)
(&LanguagesHandler{Store: langStore}).Register(mux)
```

#### Шаг 5 — gateway

Префикс `/languages/` → lexicon уже в `services/api-gateway/cmd/main.go`.

`GET /languages` уже в `IsPublic` — менять gateway не нужно.

Если добавляешь **новый public path** — допиши его в `services/api-gateway/internal/middleware/auth.go` и тест в `auth_test.go`.

#### Шаг 6 — конфиг

Не нужен — `DATABASE_URL` уже есть.

#### Шаг 7 — сборка и проверка

```bash
docker compose up --build -d lexicon

# напрямую lexicon
curl -s http://localhost:8082/languages | jq .

# через gateway (как у клиента)
curl -s http://localhost:8080/languages | jq .
# ожидаем: [{"code":"evn",...},{"code":"ru",...}]

# без токена — не 401 (public)
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/languages
```

Опционально — фрагмент в `scripts/smoke-api.sh`:

```bash
c=$(code http://localhost:8080/languages)
[[ "$c" == "200" ]] && pass "GET /languages → $c" || fail "GET /languages → $c"
```

#### Что если ручка с JWT или admin?

Тот же порядок, отличия:

| Требование | Что менять |
|------------|------------|
| Любой JWT | `mux.Handle("GET /path", jwtMW(http.HandlerFunc(h.fn)))` |
| Platform admin | внутри handler: `if !claims.IsAdmin { writeErr(..., 403, ...) }` |
| Новый префикс URL | `api-gateway/cmd/main.go` → `routes` |
| Новая таблица | шаг 1: `migrate create` + `just migrate` |
| Teacher / owner | фильтр `owner_id = $1` в store + проверка `claims.UserID` |

Пример с admin: `POST /platform/languages` — тот же store/handler/router, но `Register` вешает route через `jwtMW`, в handler первой строкой `requireAdmin(w, r)`.

---

### OpenAPI / ogen (опционально)

Сейчас **auth** и **platform media** — ручные handlers (см. референсы выше). Ogen можно использовать параллельно:

```bash
just -f services/auth/Justfile install-tools   # один раз: ogen + migrate
just -f services/auth/Justfile swagger         # перегенерация
```

Общие фрагменты спеки: `_shared/openapi/`. Генератор-хелпер: `_misc/openapi-handler-gen/`.

`internal/gen/` в `.gitignore` — генерируется локально.

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

JWT middleware в gateway проверяет Bearer на всех маршрутах, кроме public (auth register/login/refresh, GET `/languages/*`, health/ready/openapi). Upstream-сервисы по-прежнему валидируют JWT на своих protected routes.

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
