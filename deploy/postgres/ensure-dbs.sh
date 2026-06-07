#!/bin/sh
# Create application databases if missing (idempotent; safe on existing volumes).
set -eu

PGHOST="${PGHOST:-postgres}"
PGUSER="${POSTGRES_USER:-even}"
export PGPASSWORD="${POSTGRES_PASSWORD:-even}"

for db in even_auth even_media even_lexicon even_content even_learning; do
  exists="$(psql -h "$PGHOST" -U "$PGUSER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '${db}'")"
  if [ "$exists" != "1" ]; then
    echo "→ creating database ${db}"
    psql -h "$PGHOST" -U "$PGUSER" -d postgres -c "CREATE DATABASE ${db};"
  fi
done

echo "✓ databases ready"
