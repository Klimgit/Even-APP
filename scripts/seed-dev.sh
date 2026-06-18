#!/usr/bin/env bash
# Dev bootstrap after `just up`: cleanup junk → dev accounts → languages evn/ru → Even alphabet.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export ROOT

"$ROOT/scripts/db-cleanup-dev.sh"

# shellcheck source=lib/bootstrap-admin.sh
source "$ROOT/scripts/lib/bootstrap-admin.sh"
bootstrap_ensure_dev_accounts
bootstrap_print_dev_accounts

"$ROOT/scripts/seed-languages.sh"
"$ROOT/scripts/seed-evn-alphabet.sh"

echo ""
echo "=== Dev seed complete (accounts + evn + ru + 36-letter alphabet) ==="
