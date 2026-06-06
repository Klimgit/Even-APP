#!/usr/bin/env bash
# Remove branch preview: containers, volumes, nginx route, worktree.
# Usage: ./scripts/cleanup-branch-preview.sh feature/my-branch
set -euo pipefail

GIT_REF="${1:-}"
if [[ -z "$GIT_REF" ]]; then
  echo "usage: $0 <branch-name>"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=scripts/lib/branch-slug.sh
source "$ROOT/scripts/lib/branch-slug.sh"
# shellcheck source=scripts/lib/preview-registry.sh
source "$ROOT/scripts/lib/preview-registry.sh"

PREVIEWS_ROOT="${PREVIEWS_ROOT:-/opt/even-app-previews}"
SLUG="$(branch_slug "$GIT_REF")"
BRANCH_DIR="${PREVIEWS_ROOT}/branches/${SLUG}"
PROJECT="even-prev-${SLUG}"

echo "→ cleanup preview: ${GIT_REF} (${SLUG})"

if [[ -d "$BRANCH_DIR" ]]; then
  cd "$BRANCH_DIR"
  export COMPOSE_PROJECT_NAME="$PROJECT"
  export COMPOSE_FILE="docker-compose.yml:docker-compose.prod.yml:docker-compose.branch.yml"
  docker compose down -v --remove-orphans 2>/dev/null || true
fi

"$ROOT/scripts/nginx-preview.sh" remove "$SLUG" || true
preview_registry_remove "$SLUG"
rm -rf "$BRANCH_DIR"

echo "✓ preview removed: ${SLUG}"
echo "  safe to merge/delete branch ${GIT_REF} on GitHub"
