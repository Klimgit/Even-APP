#!/usr/bin/env bash
# CI integration: docker stack + migrations + smoke-api.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
fi

cleanup() {
  docker compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

echo "→ starting infra..."
docker compose up -d postgres minio minio-init

echo "→ migrations..."
./scripts/migrate.sh

echo "→ starting app services..."
docker compose up --build -d auth lexicon content learning api-gateway

echo "→ waiting for gateway..."
deadline=$((SECONDS + 240))
until curl -sf http://localhost:8080/api/v1/ready >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ gateway timeout"
    docker compose ps
    docker compose logs --tail=50 api-gateway auth lexicon
    exit 1
  fi
  sleep 3
done

echo "→ dev seed..."
./scripts/seed-dev.sh

echo "→ smoke tests..."
./scripts/smoke-api.sh

echo "✓ CI integration passed"
