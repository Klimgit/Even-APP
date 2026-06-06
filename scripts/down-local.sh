#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "→ stopping local processes..."
for f in .dev/pids/*.pid; do
  [[ -f "$f" ]] || continue
  pid=$(cat "$f")
  if kill "$pid" 2>/dev/null; then
    echo "  stopped pid $pid"
  fi
  rm -f "$f"
done

echo "→ stopping docker infra..."
docker compose stop postgres minio 2>/dev/null || true

echo "✓ stopped (docker app images unchanged; use 'just up' for full docker stack)"
