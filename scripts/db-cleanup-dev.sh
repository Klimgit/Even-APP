#!/usr/bin/env bash
# Remove dev/test junk: tst* languages, test lexemes/sounds, ephemeral @example.com users.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PG_USER="${POSTGRES_USER:-even}"
KEEP_EMAIL="${PLATFORM_ADMIN_EMAIL:-platform-admin@even.local}"

echo "=== DB cleanup (dev) ==="

docker compose -f "$ROOT/docker-compose.yml" exec -T postgres psql -U "$PG_USER" -d even_lexicon -v ON_ERROR_STOP=1 <<'SQL'
-- Test languages from integration scripts (cascades alphabet, lexemes, sounds, …)
DELETE FROM languages WHERE code LIKE 'tst%';

-- Leftover test data on evn/ru
DELETE FROM lexemes
WHERE language_id IN (SELECT id FROM languages WHERE code IN ('evn', 'ru'));

DELETE FROM sounds
WHERE language_id IN (SELECT id FROM languages WHERE code IN ('evn', 'ru'));

-- Only production languages in catalog
DELETE FROM languages WHERE code NOT IN ('evn', 'ru');
SQL

docker compose -f "$ROOT/docker-compose.yml" exec -T postgres psql -U "$PG_USER" -d even_media -v ON_ERROR_STOP=1 <<'SQL'
DELETE FROM media_assets
WHERE language_id IN (SELECT id FROM languages WHERE code NOT IN ('evn', 'ru'));

DELETE FROM languages WHERE code NOT IN ('evn', 'ru');
SQL

docker compose -f "$ROOT/docker-compose.yml" exec -T postgres psql -U "$PG_USER" -d even_auth -v ON_ERROR_STOP=1 <<SQL
DELETE FROM users
WHERE email <> '$KEEP_EMAIL'
  AND (
    email LIKE '%@example.com'
    OR email LIKE 'seed-%'
    OR email LIKE 'smoke-%'
    OR email LIKE 'lexicon-test-%'
    OR email LIKE 'manual-test@%'
    OR email LIKE 't%@ex.com'
  );
SQL

echo "  ✓ lexicon: only evn/ru (if present), no tst*, no test lexemes/sounds"
echo "  ✓ media: only evn/ru languages"
echo "  ✓ auth: removed ephemeral test users (kept $KEEP_EMAIL)"
echo ""
echo "=== Cleanup done ==="
