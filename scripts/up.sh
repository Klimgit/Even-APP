#!/usr/bin/env bash
# Start full Even-APP stack in Docker and wait until gateway is ready.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "→ created .env from .env.example"
fi

echo "→ building and starting containers..."
docker compose up --build -d

echo "→ waiting for gateway (http://localhost:8080/api/v1/ready)..."
deadline=$((SECONDS + 120))
until curl -sf http://localhost:8080/api/v1/ready >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ timeout — stack did not become ready in 120s"
    docker compose ps
    exit 1
  fi
  sleep 2
done

echo ""
echo "✓ Even-APP is up"
echo ""
echo "  Gateway   http://localhost:8080"
echo "  Auth      http://localhost:8081"
echo "  Lexicon   http://localhost:8082"
echo "  Content   http://localhost:8083"
echo "  Learning  http://localhost:8084"
echo "  Postgres  localhost:5432  (even / even)"
echo "  MinIO     http://localhost:9000  (console http://localhost:9001, minio / minio123)"
echo ""
echo "  Health    curl http://localhost:8080/api/v1/ready"
echo "  Logs      just logs   or   docker compose logs -f"
echo "  Stop      just down   or   ./scripts/down.sh"
echo ""
