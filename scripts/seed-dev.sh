#!/usr/bin/env bash
# Dev bootstrap after `just up`: cleanup junk → languages evn/ru → Even alphabet.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export ROOT

"$ROOT/scripts/db-cleanup-dev.sh"
"$ROOT/scripts/seed-languages.sh"
"$ROOT/scripts/seed-evn-alphabet.sh"

echo ""
echo "=== Dev seed complete (evn + ru + 36-letter alphabet) ==="
