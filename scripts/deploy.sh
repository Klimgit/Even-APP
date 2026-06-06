#!/usr/bin/env bash
# Production deploy on server: migrate → rebuild → health check.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  echo "✗ missing .env — copy deploy/env.production.example to .env and set secrets"
  exit 1
fi

export COMPOSE_FILE="docker-compose.yml:docker-compose.prod.yml"

echo "→ applying migrations..."
./scripts/migrate.sh

echo "→ building and starting stack..."
docker compose up --build -d

echo "→ waiting for gateway..."
deadline=$((SECONDS + 180))
until curl -sf http://localhost:8080/api/v1/ready >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ timeout — gateway not ready"
    docker compose ps
    exit 1
  fi
  sleep 3
done

echo "✓ deploy complete — http://$(hostname -I 2>/dev/null | awk '{print $1}'):8080/api/v1/ready"
