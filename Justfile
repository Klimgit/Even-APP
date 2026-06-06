set dotenv-load := true

default:
    @just --list

# One command: docker stack + wait + print URLs
up:
    @./scripts/up.sh

down:
    @./scripts/down.sh

# Go binaries on host, infra in docker
up-local:
    @./scripts/up-local.sh

down-local:
    @./scripts/down-local.sh

logs:
    docker compose logs -f api-gateway auth lexicon content learning

# --- Build ---

build-all:
    go build -o bin/api-gateway ./services/api-gateway/cmd
    go build -o bin/auth ./services/auth/cmd
    go build -o bin/lexicon ./services/lexicon/cmd
    go build -o bin/content ./services/content/cmd
    go build -o bin/learning ./services/learning/cmd

test:
    go test -C libs/jwt . -count=1
    go test -C libs/media . -count=1

test-integration:
    @./scripts/ci-integration.sh

tidy:
    go work sync
    cd services/auth && go mod tidy
    cd services/lexicon && go mod tidy
    cd services/content && go mod tidy
    cd services/learning && go mod tidy
    cd services/api-gateway && go mod tidy

# --- Docker ---

compose-up:
    docker compose up --build -d

compose-down:
    docker compose down

compose-logs:
    docker compose logs -f

# Postgres + MinIO only (for go run on host). Stop app containers first if ports busy.
infra-up:
    docker compose up -d postgres minio minio-init
    @echo "Postgres :5432, MinIO :9000 (console :9001)"

infra-down:
    docker compose stop postgres minio

# --- Local go run (infra-up + migrate-all; stop docker app services if ports 8080-8084 busy) ---

run-auth-local:
    HTTP_PORT=8081 DATABASE_URL="${AUTH_DATABASE_URL}" JWT_SECRET="${JWT_SECRET}" LOG_LEVEL="${LOG_LEVEL:-info}" go run ./services/auth/cmd

run-lexicon-local:
    HTTP_PORT=8082 DATABASE_URL="${LEXICON_DATABASE_URL}" JWT_SECRET="${JWT_SECRET}" \
      MEDIA_USER_QUOTA_BYTES="${MEDIA_USER_QUOTA_BYTES:-524288000}" \
      S3_ENDPOINT="${S3_ENDPOINT}" S3_PUBLIC_ENDPOINT="${S3_PUBLIC_ENDPOINT}" \
      S3_BUCKET="${S3_BUCKET}" S3_ACCESS_KEY="${S3_ACCESS_KEY}" S3_SECRET_KEY="${S3_SECRET_KEY}" \
      LOG_LEVEL="${LOG_LEVEL:-info}" go run ./services/lexicon/cmd

run-content-local:
    HTTP_PORT=8083 DATABASE_URL="${CONTENT_DATABASE_URL}" JWT_SECRET="${JWT_SECRET}" \
      S3_ENDPOINT="${S3_ENDPOINT}" S3_PUBLIC_ENDPOINT="${S3_PUBLIC_ENDPOINT}" \
      S3_BUCKET="${S3_BUCKET}" S3_ACCESS_KEY="${S3_ACCESS_KEY}" S3_SECRET_KEY="${S3_SECRET_KEY}" \
      LOG_LEVEL="${LOG_LEVEL:-info}" go run ./services/content/cmd

run-learning-local:
    HTTP_PORT=8084 DATABASE_URL="${LEARNING_DATABASE_URL}" JWT_SECRET="${JWT_SECRET}" LOG_LEVEL="${LOG_LEVEL:-info}" go run ./services/learning/cmd

run-gateway-local:
    HTTP_PORT=8080 JWT_SECRET="${JWT_SECRET}" \
      AUTH_URL="${AUTH_URL}" LEXICON_URL="${LEXICON_URL}" \
      CONTENT_URL="${CONTENT_URL}" LEARNING_URL="${LEARNING_URL}" \
      LOG_LEVEL="${LOG_LEVEL:-info}" go run ./services/api-gateway/cmd

# --- Migrations (explicit; never on app startup) ---

# Docker migrate containers (no migrate CLI on host required)
migrate:
    @./scripts/migrate.sh

migrate-all:
    DATABASE_URL="${AUTH_DATABASE_URL}" just -f services/auth/Justfile migrate-up
    DATABASE_URL="${LEXICON_DATABASE_URL}" just -f services/lexicon/Justfile migrate-up
    DATABASE_URL="${CONTENT_DATABASE_URL}" just -f services/content/Justfile migrate-up
    DATABASE_URL="${LEARNING_DATABASE_URL}" just -f services/learning/Justfile migrate-up

# --- Smoke checks ---

smoke-api:
    @./scripts/smoke-api.sh

# Branch preview on server: /preview/<slug>/
deploy-branch BRANCH:
    @./scripts/deploy-branch-preview.sh "{{BRANCH}}"

cleanup-branch BRANCH:
    @./scripts/cleanup-branch-preview.sh "{{BRANCH}}"

health-check:
    #!/usr/bin/env bash
    set -euo pipefail
    for url in \
        http://localhost:8080/health \
        http://localhost:8081/health \
        http://localhost:8082/health \
        http://localhost:8083/health \
        http://localhost:8084/health; do
      echo "GET $url"
      curl -sf "$url"
      echo
    done
    for url in \
        http://localhost:8080/api/v1/ready \
        http://localhost:8081/api/v1/ready \
        http://localhost:8082/api/v1/ready \
        http://localhost:8083/api/v1/ready \
        http://localhost:8084/api/v1/ready; do
      echo "GET $url"
      curl -sf "$url"
      echo
    done
    echo "GET http://localhost:8080/api/v1/gateway/status"
    curl -sf http://localhost:8080/api/v1/gateway/status
    echo
