# Как добавить новую ручку

---

## 0. Спланировать

### Выбрать уровень авторизации

| Уровень | OpenAPI | Gateway | Handler |
|---------|---------|---------|---------|
| **Public** | без `security:` | путь в `IsPublic()` | JWT не нужен |
| **Любой JWT** | `security: [bearerAuth: []]` | JWT middleware | `ClaimsFromContext(ctx)` |
| **Teacher** | `bearerAuth` | JWT | `claims.Role == "teacher" \|\| claims.IsAdmin` |
| **Platform admin** | `bearerAuth` | JWT | `claims.IsAdmin` |
| **Владелец ресурса** | `bearerAuth` | JWT | `owner_id == claims.UserID` в SQL + handler |

JWT claims: `uid`, `role` (`student` \| `teacher`), `is_admin`.

---

## 1. Контракт — `api/http/v1/api.yaml`

Файл: `services/<svc>/api/http/v1/api.yaml`.

Добавить path, `operationId`, теги, request/response schemas, коды ошибок.

**Публичная ручка** — тег `[Public]`, без `security`:

```yaml
/api/v1/auth/demo/public:
  get:
    operationId: demoPublic
    tags: [Auth, Public, Demo]
    summary: Public demo
    responses:
      "200":
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/DemoPublicResponse"
```

**Защищённая ручка** — `security` + при необходимости `403`:

```yaml
/api/v1/auth/demo/admin/stats:
  get:
    operationId: demoAdminStats
    tags: [Auth, Demo]
    security:
      - bearerAuth: []
    responses:
      "200":
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/DemoAdminStatsResponse"
      "403":
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ErrorResponse"
```

---

## 2. Схема БД (если нужна новая таблица)

```bash
migrate create -ext sql -dir services/<svc>/database/migrations -seq add_items
```

Создать пару файлов:

- `services/<svc>/database/migrations/NNNNNN_add_items.up.sql`
- `services/<svc>/database/migrations/NNNNNN_add_items.down.sql`

Применить:

```bash
just migrate                    # все сервисы (docker)
# или одна БД:
DATABASE_URL="postgres://even:even@localhost:5432/even_lexicon?sslmode=disable" \
  just -f services/lexicon/Justfile migrate-up
```

При `just up` миграции запускаются автоматически (`*-migrate` в docker-compose).

Проверка:

```bash
docker compose exec postgres psql -U even -d even_lexicon -c '\dt'
```

---

## 3. SQL-запросы — `database/queries/*.sql`

Один файл на ресурс (`users.sql`, `languages.sql`, …). Имена запросов — для sqlc.

```sql
-- name: CountUsers :one
SELECT count(*)::int AS count FROM users;

-- name: UserStats :one
SELECT
  count(*)::int AS total_users,
  count(*) FILTER (WHERE role = 'student')::int AS students,
  count(*) FILTER (WHERE role = 'teacher')::int AS teachers,
  count(*) FILTER (WHERE is_admin = true)::int AS admins
FROM users;

-- name: GetItemByID :one
SELECT id, title, owner_id, created_at
FROM items
WHERE id = $1;
```

Типы результата sqlc:

| Аннотация | Go-метод возвращает |
|-----------|---------------------|
| `:one` | одну строку (или `pgx.ErrNoRows`) |
| `:many` | слайс |
| `:exec` | только `error` |

Параметры: `$1`, `$2`, … — именованные поля в `query.SomeParams{...}`.

---

## 4. Генерация sqlc

```bash
just -f services/<svc>/Justfile sqlc
# или из корня:
just sqlc-all
```

Результат: `services/<svc>/internal/gen/query/*.sql.go`.

В сервисе вызывать так:

```go
row, err := s.q.UserStats(ctx)
if errors.Is(err, pgx.ErrNoRows) {
    return nil, domain.ErrNotFound
}
```

---

## 5. Domain — `internal/domain/`

- сущности (`User`, `Language`, …)
- доменные ошибки (`ErrNotFound`, `ErrForbidden`, …)
- входы service-слоя при необходимости

Handler и service **не** объявляют свои struct-модели данных — только `domain` + ogen-типы на границе HTTP.

---

## 6. Service — `internal/service/`

Бизнес-логика, вызовы `*query.Queries`, маппинг `query.*` → `domain.*`:

```go
func (s *AuthService) DemoAdminStats(ctx context.Context, isAdmin bool) (*DemoStats, error) {
    if !isAdmin {
        return nil, domain.ErrForbidden
    }
    row, err := s.q.UserStats(ctx)
    if err != nil {
        return nil, err
    }
    return &DemoStats{
        TotalUsers: int(row.TotalUsers),
        Students:   int(row.Students),
        Teachers:   int(row.Teachers),
        Admins:     int(row.Admins),
    }, nil
}
```
---

## 7. Генерация HTTP (ogen)

Один раз установить инструменты:

```bash
just -f services/<svc>/Justfile install-tools
```

После правок `api.yaml`:

```bash
just -f services/<svc>/Justfile generate   # swagger + sqlc
# или по отдельности:
just -f services/<svc>/Justfile swagger
just -f services/<svc>/Justfile sqlc
```

`swagger` делает:

1. `ogen` → `internal/gen/http/v1/` (роутер, типы, `Handler` interface)
2. `openapi-handler-gen` → `openapi_handler.go` (отдача `GET /api/v1/openapi.yaml` на сервисе)

Если `openapi-handler-gen` падает с ошибкой go.work — запустить ogen вручную, затем из **корня репо**:

```bash
cd services/<svc>
ogen --target internal/gen/http/v1 --package http_v1 \
  --config api/http/v1/.ogen.yml --clean api/http/v1/api.yaml

cd ../..
go run ./_misc/openapi-handler-gen/ \
  services/<svc>/internal/gen/http/v1/openapi_handler.go http_v1 openapi.yaml
```

---

## 8. Handler — `internal/handler/http_api.go`

Реализовать методы интерфейса `http_v1.Handler` (имена = `operationId` в PascalCase).

Паттерн:

```go
func (h *HTTPHandler) DemoMe(ctx context.Context) (http_v1.DemoMeRes, error) {
    claims, ok := middleware.ClaimsFromContext(ctx)
    if !ok {
        return nil, domain.ErrUnauthorized
    }
    u, err := h.svc.DemoMe(ctx, claims.UserID)
    if err != nil {
        return nil, err
    }
    return &http_v1.DemoMeResponse{
        User:         mapUser(*u),
        TokenRole:    http_v1.DemoMeResponseTokenRole(claims.Role),
        TokenIsAdmin: claims.IsAdmin,
    }, nil
}
```

**Ошибки:**

- доменные (`ErrNotFound`, …) → `return nil, err` → `NewError` в `errors.go` мапит в HTTP-код
- ожидаемые 4xx с телом → typed response, например `(*http_v1.DemoTeacherForbidden, error)`

**JWT на сервисе:** `internal/handler/security.go` — `HandleBearerAuth` парсит токен и кладёт claims в context. Отдельный `router.go` не нужен — ogen регистрирует пути из `api.yaml`.

**Wiring** уже в `cmd/main.go`:

```go
oasServer, _ := http_v1.NewServer(
    handler.NewHTTPHandler(svc),
    handler.NewSecurityHandler(jwtMgr),
)
mux.Handle("/", oasServer)
```

---

## 9. Gateway

### Прокси

Большинство префиксов уже настроены в `router.go`. Новый префикс — добавить в `routes` там же.

### Public-пути

Если ручка **без JWT**, gateway должен пропускать запрос до прокси. Добавить путь в:

- `services/api-gateway/internal/middleware/auth.go` → `IsPublic()`
- `services/api-gateway/internal/middleware/auth_test.go` → тест-кейс

Пример: `GET /api/v1/auth/demo/public`.

### Swagger merge — автоматически

Отдельно в gateway **ничего прописывать не нужно**.

При старте gateway:

1. Скачивает `GET /api/v1/openapi.yaml` с каждого backend (порядок: auth → media → lexicon → content → learning)
2. Сливает paths и components через **libopenapi** (`services/api-gateway/internal/swagger/`)
3. Отдаёт единый spec на `http://localhost:8080/api/v1/openapi.yaml`

Достаточно добавить ручку в `api.yaml` сервиса, пересобрать сервис и gateway — новая операция появится в merged OpenAPI.

Проверка:

```bash
curl -s http://localhost:8080/api/v1/openapi.yaml | grep demoAdminStats
```

---

## 10. Сборка и запуск

### Docker (рекомендуется)

```bash
just up
```