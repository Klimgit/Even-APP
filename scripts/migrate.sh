#!/usr/bin/env bash
# Apply DB migrations (local dev). Normally runs automatically via docker compose depends_on on `just up`.
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
# Recreate migrate containers so new .sql files are picked up on each run.
docker compose rm -sf auth-migrate lexicon-migrate content-migrate learning-migrate >/dev/null 2>&1 || true
docker compose up auth-migrate lexicon-migrate content-migrate learning-migrate

echo "✓ migrations applied"
