#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOK="$ROOT/.git/hooks/prepare-commit-msg"
SRC="$ROOT/scripts/git-hooks/prepare-commit-msg"

cp "$SRC" "$HOOK"
chmod +x "$HOOK"
echo "Installed prepare-commit-msg hook → $HOOK"
