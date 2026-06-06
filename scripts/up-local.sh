#!/usr/bin/env bash
# Infra in Docker, Go binaries on host. For day-to-day backend dev.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "→ created .env from .env.example"
fi

# shellcheck disable=SC1091
set -a && source .env && set +a

mkdir -p .dev/logs .dev/pids

echo "→ starting postgres + minio..."
docker compose up -d postgres minio minio-init

echo "→ applying migrations..."
./scripts/migrate.sh

echo "→ stopping docker app containers (free ports 8080-8084)..."
docker compose stop api-gateway auth lexicon content learning 2>/dev/null || true

# Stop previous local processes if any
for f in .dev/pids/*.pid; do
  [[ -f "$f" ]] || continue
  pid=$(cat "$f")
  kill "$pid" 2>/dev/null || true
  rm -f "$f"
done

echo "→ building binaries..."
just build-all

start() {
  local svc="$1"
  local cmd="$2"
  local log=".dev/logs/${svc}.log"
  local pidfile=".dev/pids/${svc}.pid"
  setsid bash -c "$cmd" >"$log" 2>&1 &
  echo $! >"$pidfile"
  echo "  started $svc (pid $(cat "$pidfile"), log $log)"
}

echo "→ starting services on host..."
start auth \
  "HTTP_PORT=8081 DATABASE_URL='${AUTH_DATABASE_URL}' JWT_SECRET='${JWT_SECRET}' LOG_LEVEL='${LOG_LEVEL:-info}' ./bin/auth"
start lexicon \
  "HTTP_PORT=8082 DATABASE_URL='${LEXICON_DATABASE_URL}' JWT_SECRET='${JWT_SECRET}' \
   S3_ENDPOINT='${S3_ENDPOINT}' S3_PUBLIC_ENDPOINT='${S3_PUBLIC_ENDPOINT}' \
   S3_BUCKET='${S3_BUCKET}' S3_ACCESS_KEY='${S3_ACCESS_KEY}' S3_SECRET_KEY='${S3_SECRET_KEY}' \
   LOG_LEVEL='${LOG_LEVEL:-info}' ./bin/lexicon"
start content \
  "HTTP_PORT=8083 DATABASE_URL='${CONTENT_DATABASE_URL}' JWT_SECRET='${JWT_SECRET}' \
   S3_ENDPOINT='${S3_ENDPOINT}' S3_PUBLIC_ENDPOINT='${S3_PUBLIC_ENDPOINT}' \
   S3_BUCKET='${S3_BUCKET}' S3_ACCESS_KEY='${S3_ACCESS_KEY}' S3_SECRET_KEY='${S3_SECRET_KEY}' \
   LOG_LEVEL='${LOG_LEVEL:-info}' ./bin/content"
start learning \
  "HTTP_PORT=8084 DATABASE_URL='${LEARNING_DATABASE_URL}' JWT_SECRET='${JWT_SECRET}' LOG_LEVEL='${LOG_LEVEL:-info}' ./bin/learning"

# Backends must be up before gateway ready check passes
sleep 2
start api-gateway \
  "HTTP_PORT=8080 JWT_SECRET='${JWT_SECRET}' AUTH_URL='${AUTH_URL}' LEXICON_URL='${LEXICON_URL}' \
   CONTENT_URL='${CONTENT_URL}' LEARNING_URL='${LEARNING_URL}' LOG_LEVEL='${LOG_LEVEL:-info}' ./bin/api-gateway"

echo "→ waiting for gateway..."
deadline=$((SECONDS + 60))
until curl -sf http://localhost:8080/api/v1/ready >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ timeout — check .dev/logs/"
    exit 1
  fi
  sleep 1
done

echo ""
echo "✓ Even-APP running locally (binaries on host)"
echo "  Gateway http://localhost:8080"
echo "  Logs     tail -f .dev/logs/*.log"
echo "  Stop     ./scripts/down-local.sh"
echo ""
