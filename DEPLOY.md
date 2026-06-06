# Деплой Even-APP на сервер

Сервер: **91.218.245.136**  
Стек: Docker Compose (Postgres, MinIO, 5 Go-сервисов).  
Деплой вручную: SSH → `scripts/deploy.sh`.

---

## 1. Первичная настройка сервера (один раз)

Подключитесь по SSH и выполните:

```bash
git clone https://github.com/Klimgit/Even-APP.git /opt/even-app
cd /opt/even-app
sudo bash scripts/server-bootstrap.sh   # docker + git, если ещё нет
```

Отредактируйте секреты:

```bash
nano /opt/even-app/.env
```

Шаблон: [deploy/env.production.example](deploy/env.production.example). Обязательно замените:

- `POSTGRES_PASSWORD`
- `JWT_SECRET` (длинная случайная строка)
- `S3_ACCESS_KEY` / `S3_SECRET_KEY`
- `S3_PUBLIC_ENDPOINT` — публичный URL MinIO для presigned media (см. ниже)

Первый деплой:

```bash
cd /opt/even-app
./scripts/deploy.sh
```

Проверка:

```bash
curl http://localhost:8080/api/v1/ready
curl http://91.218.245.136:8080/health
```

### Firewall

| Порт | Назначение |
|------|------------|
| 8080 | API Gateway (обязательно) |
| 22   | SSH |
| 80   | nginx для branch preview (опционально) |
| 9000 | MinIO public URLs для медиа (опционально; лучше nginx) |

```bash
sudo ufw allow 22/tcp
sudo ufw allow 8080/tcp
sudo ufw enable
```

---

## 2. Production deploy

`deploy.sh`:

1. `COMPOSE_FILE=docker-compose.yml:docker-compose.prod.yml`
2. `./scripts/migrate.sh` — миграции (явный шаг)
3. `docker compose up --build -d`
4. ждёт `/api/v1/ready`

**Обновление после merge в `main`:**

```bash
cd /opt/even-app
git pull origin main
./scripts/deploy.sh
```

**Откат:**

```bash
cd /opt/even-app
git checkout <previous-commit>
./scripts/deploy.sh
```

Миграции БД откатываются отдельно (`just migrate-down` / `migrate down 1` per service).

---

## 3. Branch preview (опционально)

Изолированные стеки для feature-веток. **Production не затрагивается.**

| | Production | Branch `feature/auth` |
|---|------------|------------------------|
| Код | `/opt/even-app` | `/opt/even-app-previews/branches/feature-auth` |
| URL | `http://IP:8080/...` | `http://IP/preview/feature-auth/...` |
| Compose | `even-app` | `even-prev-feature-auth` |

Slug: `feature/auth` → `feature-auth`.

### Первичная настройка preview (один раз)

```bash
ssh DEPLOY_USER@91.218.245.136
cd /opt/even-app && git pull origin main
sudo bash /opt/even-app/scripts/server-bootstrap-previews.sh
sudo ufw allow 80/tcp
```

Sudo для nginx (если деплой не от root):

```bash
sudo tee /etc/sudoers.d/even-previews <<'EOF'
deploy ALL=(ALL) NOPASSWD: /usr/sbin/nginx, /usr/bin/systemctl reload nginx
EOF
sudo chmod 440 /etc/sudoers.d/even-previews
```

### Деплой и очистка ветки

```bash
/opt/even-app-previews/manager/scripts/deploy-branch-preview.sh feature/auth
/opt/even-app-previews/manager/scripts/cleanup-branch-preview.sh feature/auth
```

Проверка:

```bash
curl http://91.218.245.136/preview/feature-auth/api/v1/ready
cat /opt/even-app-previews/registry.json
```

Перед merge в `main` — **cleanup**, чтобы удалить контейнеры, volumes и каталог ветки.

---

## 4. Production compose

[docker-compose.prod.yml](docker-compose.prod.yml):

- `restart: unless-stopped` на сервисах
- наружу только **8080** (gateway)
- Postgres / MinIO — `127.0.0.1` (доступ с сервера или SSH-туннель)

### MinIO и presigned URL

Клиенты получают `upload_url` / `url` с хостом из `S3_PUBLIC_ENDPOINT`.

Варианты:

1. **Простой:** открыть 9000, `S3_PUBLIC_ENDPOINT=http://91.218.245.136:9000`
2. **Безопаснее:** nginx reverse proxy на `https://media.example.com` → `127.0.0.1:9000`

---

## 5. Логи и отладка

```bash
cd /opt/even-app
export COMPOSE_FILE=docker-compose.yml:docker-compose.prod.yml
docker compose ps
docker compose logs -f api-gateway auth lexicon
docker compose logs auth-migrate   # после migrate
```
