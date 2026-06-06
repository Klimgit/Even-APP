# Сценарий демо: разработка в Even-APP

Сценарий для записи экрана (~20–25 мин). Показываем полный цикл: запуск → что уже есть → добавление новой ручки → проверка.

**Цель демо:** зритель понимает, как устроен репозиторий и как за один проход добавить endpoint от контракта до `curl`.

**Документы на экране:** [DEVELOPMENT.md](DEVELOPMENT.md), [API_STATUS.md](API_STATUS.md), [API.md](API.md).

---

## Пошагово: все команды (копируй по порядку)

Шаги 1–6 — подготовка. **Шаг 7** — куда писать код новой ручки (на примере `GET /languages`). Шаги 8–13 — пересборка и проверка.

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

#### Пример: `GET /languages` (lexicon, public)

Миграция **не нужна** — таблица `languages` уже в БД (шаг 6).

| Что | Куда |
|-----|------|
| SQL: выборка активных языков | **новый** `services/lexicon/internal/store/languages.go` |
| HTTP: `GET /languages` → JSON-массив | **новый** `services/lexicon/internal/httpapi/languages.go` |
| Подключить handler | **правка** `services/lexicon/internal/httpapi/router.go` — 2 строки: `NewLanguageStore` + `LanguagesHandler.Register` |
| Public без JWT на gateway | уже в `auth.go` → `IsPublic` для `/languages` |
| Прокси на lexicon | уже в `api-gateway/cmd/main.go` |

Полный код примера — в [DEVELOPMENT.md § GET /languages](DEVELOPMENT.md).

**Auth в handler:**

- **public** — `mux.HandleFunc("GET /path", h.method)` без `jwtMW`
- **с JWT** — `mux.Handle("GET /path", jwtMW(http.HandlerFunc(h.method)))`
- **platform admin** — внутри handler: `claims.IsAdmin` (см. `platform_media.go` → `requireAdmin`)

```bash
# проверить, что компилируется
go build -o /tmp/lexicon ./services/lexicon/cmd
```

---

### 8. Пересобрать lexicon

```bash
docker compose up --build -d lexicon

# дождаться healthy
docker compose ps lexicon
```

### 9. Проверить новую ручку

```bash
# напрямую на сервис
curl -s http://localhost:8082/languages | jq .

# через gateway (как клиент)
curl -s http://localhost:8080/languages | jq .

# public — 200 без токена
curl -s -o /dev/null -w "GET /languages: %{http_code}\n" \
  http://localhost:8080/languages

# полный smoke с новой ручкой
just smoke-api
```

Ожидаем JSON с `evn` и `ru`; smoke — `GET /languages → 200`.

**Озвучка:** «От контракта в API.md до curl — store, handler, router, один сервис пересобрали.»

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
# шаг 7: store/*.go → httpapi/*.go → router.go
docker compose up --build -d lexicon
curl -s http://localhost:8080/languages | jq .
just smoke-api
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
- [ ] (Опционально) заранее реализовать `GET /languages` в отдельной ветке и на демо только показать diff — если боишься опечаток в live coding

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

curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/languages
# 404 сейчас (ручки ещё нет) — после реализации будет 200 без токена
```

**Текст:**

> Список public-путей в `IsPublic`. Всё остальное без Bearer — 401 ещё на gateway.

---

## Блок 6. Куда писать код (2 мин)

**Экран:** DEMO.md шаг 7 или дерево `services/lexicon/internal/`.

**Текст:**

> Новая ручка — это три места в сервисе: **store** (SQL), **httpapi** (логика и JSON), **router.go** (подключение).  
> Контракт — в API.md. Если нужна новая таблица — миграция в `database/migrations/`.  
> Клиент ходит на gateway `:8080`; public-пути — в `api-gateway/.../auth.go`.

**Показать на примере `GET /languages`:**

1. `internal/store/languages.go` — запрос к таблице `languages`
2. `internal/httpapi/languages.go` — handler + `Register`
3. `internal/httpapi/router.go` — две строки wiring

Код — в [DEVELOPMENT.md](DEVELOPMENT.md), не набираем на демо целиком, если мало времени.

---

## Блок 7. Live coding (5–7 мин, опционально)

**Экран:** IDE — по очереди открыть три файла из блока 6.

```bash
docker compose exec postgres psql -U even -d even_lexicon \
  -c "SELECT code, name FROM languages WHERE is_active;"
```

```bash
docker compose up --build -d lexicon
```

**Текст:** «Меняли только lexicon — пересобираем один сервис.»

---

## Блок 8. Проверка (2 мин)

```bash
# напрямую
curl -s http://localhost:8082/languages | jq .

# через gateway — как у клиента
curl -s http://localhost:8080/languages | jq .

# public: без токена
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/languages
# 200
```

**Текст:**

> Два языка из seed-миграции. Gateway проксирует `/languages/` на lexicon без JWT.

**Опционально — добавить в smoke (20 сек):**

Показать 3 строки в `scripts/smoke-api.sh` из DEVELOPMENT.md (не обязательно запускать, если времени мало).

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

> Перед merge: обновить API_STATUS.md — перенести `GET /languages` в ✅.  
> API.md и DTO.md уже описывают контракт.  
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
curl http://localhost:8080/languages
curl http://localhost:8080/api/v1/ready
```

---

## Типичные косяки при записи

| Проблема | Что сказать / сделать |
|----------|------------------------|
| `just up` долго | Заранее прогреть образы или начать запись после `up` |
| Порт занят | `just down` и повторить |
| gateway `not_ready` | `curl localhost:8081/api/v1/ready` — какой backend упал |
| 401 на /languages | Забыли public в handler или gateway `IsPublic` |
| Пустой список языков | `just migrate` не применён |
| smoke падает на MinIO | `docker compose up --build -d lexicon content` после pull |

---

## Варианты нарезки ролика

| Версия | Блоки | Длина |
|--------|-------|-------|
| Полная | 1–12 | ~22 мин |
| Только backend-цикл | 3, 4, 6–8, 11 | ~15 мин |
| Быстрый обзор | 1, 2, 3, 4 | ~8 мин |
| Live coding only | 6–8 | ~10 мин (для внутренней команды) |
