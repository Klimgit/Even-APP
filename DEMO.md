# Сценарий демо: разработка в Even-APP

Сценарий для записи экрана (~20–25 мин). Показываем полный цикл: запуск → что уже есть → добавление новой ручки → проверка.

**Цель демо:** зритель понимает, как устроен репозиторий и как за один проход добавить endpoint от контракта до `curl`.

**Документы на экране:** [DEVELOPMENT.md](DEVELOPMENT.md), [API_STATUS.md](API_STATUS.md), [API.md](API.md).

---

## Пошагово: все команды (копируй по порядку)

Шаги 1–6 — подготовка. **Шаг 7** — dummy-ручка с кодом по файлам. Шаги 8–13 — пересборка и проверка.

### 0. Один раз: клон и зависимости

```bash
git clone https://github.com/Klimgit/Even-APP.git
cd Even-APP

# macOS
brew install just jq curl
# migrate CLI — опционально
brew install golang-migrate
```

### 1. Поднять стек

```bash
cd Even-APP

# чистый старт (опционально)
just down

# .env создастся сам из .env.example
just up
```

Ждём `✓ Even-APP is up` (до ~2 мин при первой сборке).

### 2. Проверить, что всё живое

```bash
just health-check

curl -s http://localhost:8080/api/v1/gateway/status | jq .
curl -s http://localhost:8080/api/v1/ready | jq .

docker compose ps
```

### 3. Прогнать smoke-тест (auth + platform media)

```bash
just smoke-api
```

Должно закончиться на `=== All smoke tests passed ===`.

### 4. Ручки вручную (auth через gateway)

```bash
# регистрация
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "demo@example.com",
    "password": "password123",
    "display_name": "Demo",
    "role": "teacher"
  }' | jq .

# логин — сохрани токен
export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123"}' \
  | jq -r '.access_token')

echo "$TOKEN"

# кто я
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq .
```

### 5. JWT на gateway

```bash
# без токена — 401
curl -s -o /dev/null -w "auth/me без токена: %{http_code}\n" \
  http://localhost:8080/api/v1/auth/me

# login — public, не 401 от gateway (может быть 401 от auth если неверный пароль)
curl -s -o /dev/null -w "login: %{http_code}\n" \
  -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123"}'
```

### 6. Данные в БД + «ручки ещё нет»

```bash
docker compose exec postgres psql -U even -d even_lexicon \
  -c "SELECT code, name, native_name FROM languages WHERE is_active ORDER BY code;"
```

Ожидаем строки `evn` и `ru`.

**На демо (до кода):** показать, что HTTP-ручки пока нет:

```bash
curl -s -o /dev/null -w "GET /languages до реализации: %{http_code}\n" \
  http://localhost:8080/languages
# 404 — данных в БД есть, handler ещё не подключён
```

---

### 7. Куда писать, чтобы добавить новую ручку

Клиент всегда бьёт в **gateway** `:8080`. Код ручки живёт в **одном из сервисов** — по префиксу URL.

#### Какой сервис

| URL | Сервис | Порт (напрямую) |
|-----|--------|-----------------|
| `/api/v1/auth/*` | `services/auth` | 8081 |
| `/languages/*`, `/api/v1/platform/*` | `services/lexicon` | 8082 |
| `/api/v1/teacher/*` | `services/content` | 8083 |
| `/api/v1/courses/*`, `/lessons/*`, `/progress/*`, … | `services/learning` | 8084 |

#### Куда писать (по порядку)

| Шаг | Файл / папка | Что там |
|-----|--------------|---------|
| 0. Контракт | [API.md](API.md), [DTO.md](DTO.md) | метод, путь, JSON, коды ошибок |
| 1. Схема БД | `services/<svc>/database/migrations/` | `.up.sql` / `.down.sql`, если нужна новая таблица или колонка |
| 2. **Данные (SQL)** | `services/<svc>/internal/store/<ресурс>.go` | структуры, `SELECT`/`INSERT`/… — **без HTTP** |
| 3. **HTTP-логика** | `services/<svc>/internal/httpapi/<ресурс>.go` | парсинг запроса, права, вызов store, `writeJSON` / `writeErr` |
| 4. **Подключить** | `services/<svc>/internal/httpapi/router.go` | создать store, handler, вызвать `.Register(mux)` |
| 5. Gateway (если public) | `services/api-gateway/internal/middleware/auth.go` | добавить путь в `IsPublic` |
| 5. Gateway (новый префикс) | `services/api-gateway/cmd/main.go` | новая строка в `routes` (редко — большинство префиксов уже есть) |
| 6. Smoke (по желанию) | `scripts/smoke-api.sh` | `curl` на `:8080` |

**Референсы уже работающих ручек:**

- auth: `services/auth/internal/httpapi/auth.go` + `internal/store/users.go`
- platform media: `services/lexicon/internal/httpapi/platform_media.go` + `internal/store/media.go`

#### Пример: dummy demo-ручки (lexicon, все типы auth + БД)

Учебный блок в `demo.go`: каждая ручка делает `SELECT 1` и возвращает `"auth": "<тип>"`, `"db": 1`. Код — в репозитории (`store/demo.go`, `httpapi/demo.go`).

| Путь | Auth | Handler | Gateway `IsPublic` |
|------|------|---------|-------------------|
| `GET /api/v1/platform/demo/public` | public | без `jwtMW` | **да** — `auth.go` |
| `GET /api/v1/platform/demo/auth` | любой JWT | `jwtMW` | нет |
| `GET /api/v1/platform/demo/ping` | любой JWT (alias) | `jwtMW` | нет |
| `GET /api/v1/platform/demo/admin` | platform admin | `jwtMW` + `requireAdmin` | нет |
| `GET /api/v1/platform/demo/teacher` | teacher или admin | `jwtMW` + `requireTeacher` | нет |
| `GET /api/v1/platform/demo/owner?user_id=` | владелец или admin | `jwtMW` + `user_id == claims.UserID` | нет |

| Шаг | Файл | Действие |
|-----|------|----------|
| 1 | `services/lexicon/internal/store/demo.go` | SQL (`SELECT 1`) |
| 2 | `services/lexicon/internal/httpapi/demo.go` | все demo-ручки |
| 3 | `services/lexicon/internal/httpapi/router.go` | `DemoHandler` + `Register` |
| 4 | `services/api-gateway/internal/middleware/auth.go` | public-путь `/api/v1/platform/demo/public` |

---

**Файл 1 — создать** `services/lexicon/internal/store/demo.go`:

```go
package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DemoStore struct {
	pool *pgxpool.Pool
}

func NewDemoStore(pool *pgxpool.Pool) *DemoStore {
	return &DemoStore{pool: pool}
}

func (s *DemoStore) Ping(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT 1`).Scan(&n)
	return n, err
}
```

---

**Файл 2 —** `services/lexicon/internal/httpapi/demo.go` — см. готовый файл в репозитории (все 6 маршрутов + `requireTeacher`).

**Файл 4 — public на gateway** — в `IsPublic` для `GET`:

```go
"/api/v1/platform/demo/public":
```

---

**Файл 3 — дописать** `services/lexicon/internal/httpapi/router.go`

После строк с `PlatformMediaHandler` (перед `return middleware.CORS`):

```go
	(&DemoHandler{Store: store.NewDemoStore(pool)}).Register(mux, jwtMW)
```

Должно получиться примерно так:

```go
	jwtMW := middleware.JWT(jwtMgr)
	ms := store.NewMediaStore(pool)
	(&PlatformMediaHandler{Store: ms, S3: s3c, Bucket: bucket, UserQuotaBytes: userQuotaBytes}).Register(mux, jwtMW)

	(&DemoHandler{Store: store.NewDemoStore(pool)}).Register(mux, jwtMW)

	return middleware.CORS(middleware.Recovery(log, middleware.Logging(log, mux)))
```

---

**Команды после правок:**

```bash
# компиляция на хосте
go build -o /tmp/lexicon ./services/lexicon/cmd

# пересобрать контейнер
docker compose up --build -d lexicon
docker compose ps lexicon   # ждём healthy
```

**Проверка** (`$TOKEN` — teacher из шага 4; для admin — promote в smoke или SQL):

```bash
GW=http://localhost:8080
ME=$(curl -s "$GW/api/v1/auth/me" -H "Authorization: Bearer $TOKEN" | jq -r .id)

# public — без токена
curl -s "$GW/api/v1/platform/demo/public" | jq .

# любой JWT
curl -s "$GW/api/v1/platform/demo/auth" -H "Authorization: Bearer $TOKEN" | jq .

# teacher (токен teacher из register)
curl -s "$GW/api/v1/platform/demo/teacher" -H "Authorization: Bearer $TOKEN" | jq .

# owner — свой user_id
curl -s "$GW/api/v1/platform/demo/owner?user_id=$ME" -H "Authorization: Bearer $TOKEN" | jq .

# admin — после promote is_admin (как в smoke-api)
curl -s "$GW/api/v1/platform/demo/admin" -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# protected без токена → 401
curl -s -o /dev/null -w "%{http_code}\n" "$GW/api/v1/platform/demo/auth"
```

---

Полный пример с реальной таблицей — [DEVELOPMENT.md § GET /languages](DEVELOPMENT.md).

**Auth в handler (шпаргалка):**

- **public** — `mux.HandleFunc("GET /path", h.method)` без `jwtMW` + путь в `services/api-gateway/internal/middleware/auth.go` → `IsPublic`
- **с JWT** — `mux.Handle("GET /path", jwtMW(http.HandlerFunc(h.method)))`
- **platform admin** — JWT + внутри handler `requireAdmin(w, r)` (см. `platform_media.go`)

---

### 8. Пересобрать lexicon

```bash
docker compose up --build -d lexicon

# дождаться healthy
docker compose ps lexicon
```

### 9. Проверить dummy-ручку

```bash
# с токеном из шага 4
curl -s http://localhost:8080/api/v1/platform/demo/ping \
  -H "Authorization: Bearer $TOKEN" | jq .

curl -s -o /dev/null -w "без токена: %{http_code}\n" \
  http://localhost:8080/api/v1/platform/demo/ping
# 401

# smoke старых ручек всё ещё должен проходить
just smoke-api
```

Ожидаем `"auth"` в ответе (`public`, `jwt`, `teacher`, …) и `"db":1`.

**Озвучка:** «Handler + router, пересобрали lexicon, curl через gateway с Bearer.»

### 10. MinIO (опционально)

```bash
# консоль в браузере: http://localhost:9001  логин minio / minio123, bucket even-media

docker compose exec postgres psql -U even -d even_lexicon \
  -c "SELECT id, display_name, object_key, scope FROM media_assets LIMIT 5;"
```

### 11. Логи

```bash
just logs
# Ctrl+C

docker compose logs lexicon --tail 30
docker compose logs api-gateway --tail 20
```

### 12. Тесты и сборка (опционально)

```bash
just test
go test -C services/api-gateway ./internal/middleware/...

just build-all
```

### 13. Остановить

```bash
just down
```

### Шпаргалка одной строкой

```bash
just up && just health-check && just smoke-api
# шаг 7: httpapi/demo.go + router.go
docker compose up --build -d lexicon
curl -s http://localhost:8080/api/v1/platform/demo/ping -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Перед записью (чеклист)

- [ ] Docker Desktop запущен
- [ ] `just`, `jq`, `curl` установлены
- [ ] Репозиторий клонирован, ветка `main` актуальна
- [ ] `just down` — чистый старт (опционально)
- [ ] Терминал: крупный шрифт, `cd Even-APP` уже открыт
- [ ] IDE: дерево проекта видно, `.env` **не** показывать на экране (или замазать `JWT_SECRET`)
- [ ] Закрыть лишние вкладки / уведомления
- [ ] (Опционально) заранее вставить `demo.go` из шага 7 и на записи только показать diff

**Короткая версия (10 мин):** блоки 1 → 2 → 4 → 6 → 8 → 10 (без MinIO и без live coding — показать готовый PR).

---

## Блок 1. Вступление (1 мин)

**Экран:** README или корень репозитория.

**Текст:**

> Even-APP — backend для изучения эвенского языка. Пять Go-микросервисов, Postgres, MinIO, API Gateway.  
> Сейчас покажу, как локально поднять стек, дернуть API и добавить новую ручку по нашим правилам.

**Показать кратко:**

```
Even-APP/
├── services/     # auth, lexicon, content, learning, api-gateway
├── libs/         # общий код: jwt, logger, postgres, s3
├── scripts/      # up.sh, migrate.sh, smoke-api.sh
├── API.md        # полная спека
├── API_STATUS.md # что сделано / что в MVP
└── DEVELOPMENT.md
```

---

## Блок 2. Архитектура (2 мин)

**Экран:** [BACKEND.md](BACKEND.md) или схема в голове + `services/api-gateway/cmd/main.go` (routes).

**Текст:**

> Клиент ходит только в gateway на порт 8080. Gateway проксирует:
> - `/api/v1/auth/` → auth  
> - `/api/v1/platform/` → lexicon  
> - `/api/v1/teacher/` → content  
> - `/courses/`, `/lessons/`, … → learning  
> - `/languages/` → lexicon (public)  
>
> У каждого сервиса своя БД. Миграции — отдельным шагом до старта, не при boot приложения.

**На экране (30 сек):** таблица портов из DEVELOPMENT.md:

| Сервис | Порт |
|--------|------|
| gateway | 8080 |
| auth | 8081 |
| lexicon | 8082 |
| … | … |

---

## Блок 3. Поднять стек (3 мин)

**Экран:** терминал.

```bash
cd Even-APP
just up
```

**Текст во время ожидания:**

> `just up` копирует `.env`, гоняет миграции, собирает Docker-образы и ждёт ready на gateway.

**После успеха:**

```bash
just health-check
curl -s http://localhost:8080/api/v1/gateway/status | jq .
curl -s http://localhost:8080/api/v1/ready | jq .
```

**Текст:**

> Ready значит: gateway жив и все четыре backend отвечают. Postgres пингуется внутри каждого сервиса.

**Опционально (15 сек):**

```bash
docker compose ps
```

---

## Блок 4. Что уже работает (3 мин)

**Экран:** [API_STATUS.md](API_STATUS.md) — сводка + секция «Реализовано».

**Текст:**

> Из MVP сделано пока 10 бизнес-ручек: auth на четыре endpoint и platform media на шесть. Остальное — в backlog этого же файла.

**Терминал — smoke-тест:**

```bash
just smoke-api
```

**Текст по ходу вывода:**

> Smoke регистрирует пользователя, логинится, в БД ставит `is_admin`, грузит картинку в MinIO через presign → PUT → confirm, проверяет list и delete.  
> Это наш эталонный сценарий — его расширяем по мере новых фич.

**Если smoke падает** — не монтировать в демо; заранее прогнать `just up && just smoke-api`.

**Ручной curl (30 сек):**

```bash
# public login
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"teacher@example.com","password":"password123"}' | jq .
# (если пользователя нет — сначала register, как в smoke-api)
```

**Текст:**

> Gateway проверяет JWT на protected-маршрутах. Login и register — public.

---

## Блок 5. Gateway JWT (1 мин)

**Экран:** `services/api-gateway/internal/middleware/auth.go`.

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/auth/me
# 401

curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/platform/demo/ping
# 401 — protected; после шага 7 с токеном будет 200
```

**Текст:**

> Список public-путей в `IsPublic`. Всё остальное без Bearer — 401 ещё на gateway.

---

## Блок 6. Куда писать код (2 мин)

**Экран:** DEMO.md шаг 7 — таблица «куда писать» + пример dummy.

**Текст:**

> Минимум для ручки без БД: **один новый файл handler** + **одна строка в router.go**.  
> С БД добавляется **store**. Gateway трогаем только для public-путей или нового префикса.

**На демо копируем dummy `GET /api/v1/platform/demo/ping`:**

1. Создать `services/lexicon/internal/store/demo.go` — `SELECT 1`
2. Создать `services/lexicon/internal/httpapi/demo.go`
3. Дописать `(&DemoHandler{Store: store.NewDemoStore(pool)}).Register(mux, jwtMW)` в `router.go`

---

## Блок 7. Live coding (3–5 мин, опционально)

**Экран:** IDE — вставить код из шага 7 DEMO.md.

```bash
go build -o /tmp/lexicon ./services/lexicon/cmd
docker compose up --build -d lexicon
```

**Текст:** «Меняли только lexicon — пересобираем один сервис.»

---

## Блок 8. Проверка (2 мин)

```bash
curl -s http://localhost:8080/api/v1/platform/demo/ping \
  -H "Authorization: Bearer $TOKEN" | jq .

curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:8080/api/v1/platform/demo/ping
# 401 без токена
```

**Текст:**

> JWT на gateway и на lexicon. Клиент всегда бьёт в `:8080` с Bearer.

---

## Блок 9. MinIO и медиа (1.5 мин, опционально)

**Экран:** браузер `http://localhost:9001` (minio / minio123), bucket `even-media`.

**Текст:**

> Медиа не в Postgres — файлы в MinIO, метаданные в `media_assets`.  
> Загрузка: presign от API → PUT в MinIO → confirm в каталог.

```bash
docker compose exec postgres psql -U even -d even_lexicon \
  -c "SELECT id, display_name, object_key FROM media_assets LIMIT 5;"
```

---

## Блок 10. Логи и отладка (1 мин)

```bash
just logs
# Ctrl+C

docker compose logs lexicon --tail 20
```

**Текст:**

> JSON-логи в stdout: каждый запрос — method, path, status, duration_ms. Уровень — `LOG_LEVEL` в `.env`.

---

## Блок 11. Документация и merge (1 мин)

**Экран:** чеклист из DEVELOPMENT.md §9.

**Текст:**

> Dummy `demo/ping` в прод не мержим — только для демо. Реальную фичу: API.md + API_STATUS.md + store/handler.  
> Полный backlog MVP — в API_STATUS.md, порядок фич — в MVP.md.

---

## Блок 12. Финал (30 сек)

**Текст:**

> Итого цикл разработки: контракт в API.md → store/handler в нужном сервисе → migrate если нужна схема → `docker compose up --build` → curl через :8080.  
> Подробности — DEVELOPMENT.md, статус ручек — API_STATUS.md.

**Экран:** ссылки на GitHub / Issues (по желанию).

---

## Шпаргалка команд (держать под рукой)

```bash
just up
just down
just health-check
just smoke-api
just logs
just migrate
docker compose up --build -d lexicon
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/platform/demo/ping
curl http://localhost:8080/api/v1/ready
```

---

## Типичные косяки при записи

| Проблема | Что сказать / сделать |
|----------|------------------------|
| `just up` долго | Заранее прогреть образы или начать запись после `up` |
| Порт занят | `just down` и повторить |
| gateway `not_ready` | `curl localhost:8081/api/v1/ready` — какой backend упал |
| 401 на demo/ping | Нет `$TOKEN` или не передан `Authorization: Bearer` |
| 404 на demo/ping | Забыли `Register` в router.go или не пересобрали lexicon |
| smoke падает на MinIO | `docker compose up --build -d lexicon content` после pull |

---

## Варианты нарезки ролика

| Версия | Блоки | Длина |
|--------|-------|-------|
| Полная | 1–12 | ~22 мин |
| Только backend-цикл | 3, 4, 6–8, 11 | ~15 мин |
| Быстрый обзор | 1, 2, 3, 4 | ~8 мин |
| Live coding only | 6–8 | ~10 мин (для внутренней команды) |
