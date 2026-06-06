#!/usr/bin/env bash
# Apply DB migrations explicitly (not on app startup). Requires postgres.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
fi

# shellcheck disable=SC1091
set -a && source .env && set +a

echo "→ ensuring postgres is up..."
docker compose up -d postgres
docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-even}" >/dev/null

echo "→ applying migrations (auth, lexicon, content, learning)..."
# COMPOSE_FILE may include docker-compose.prod.yml (set by deploy.sh)
docker compose --profile migrate up auth-migrate lexicon-migrate content-migrate learning-migrate

echo "✓ migrations applied"
