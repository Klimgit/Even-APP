# Even-APP — backend

Краткий обзор backend-скелета. **Полная инструкция по разработке: [DEVELOPMENT.md](DEVELOPMENT.md).**

## Сервисы

| Сервис | Порт | БД | Назначение (позже) |
|--------|------|-----|-------------------|
| api-gateway | 8080 | — | reverse proxy, merged OpenAPI |
| auth | 8081 | even_auth | JWT, пользователи |
| lexicon | 8082 | even_lexicon | языки, лексика, медиа |
| content | 8083 | even_content | курсы, уроки, блоки |
| learning | 8084 | even_learning | прогресс, словарь |

Системные эндпоинты (сейчас единственные):

- `GET /health`, `/api/v1/health` — процесс жив
- `GET /api/v1/ready` — Postgres (сервисы) / ping upstream (gateway)
- `GET /api/v1/openapi.yaml` — OpenAPI-заглушка

## Быстрый старт

```bash
just up       # поднять всё в Docker
just down     # остановить
just logs     # логи
```

Локальная разработка на хосте: `just up-local` — см. [DEVELOPMENT.md](DEVELOPMENT.md).

## Структура

```
libs/           — config, logger, http, postgres
services/       — 5 микросервисов
scripts/        — up.sh, down.sh, up-local.sh
deploy/         — init Postgres / MinIO
_shared/        — OpenAPI-фрагменты
_misc/          — codegen helpers
```

Конфигурация — только env-переменные.
