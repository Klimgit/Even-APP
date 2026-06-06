#!/usr/bin/env bash
# Deploy a git branch to isolated preview stack: /preview/<slug>/
# Usage: ./scripts/deploy-branch-preview.sh feature/my-branch
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
REPO_URL="${REPO_URL:-https://github.com/Klimgit/Even-APP.git}"
PREVIEW_HOST="${PREVIEW_HOST:-$(hostname -I 2>/dev/null | awk '{print $1}')}"

SLUG="$(branch_slug "$GIT_REF")"
if [[ -z "$SLUG" ]]; then
  echo "✗ invalid branch slug for: $GIT_REF"
  exit 1
fi

BRANCH_DIR="${PREVIEWS_ROOT}/branches/${SLUG}"
mkdir -p "$PREVIEWS_ROOT/nginx/enabled" "$PREVIEWS_ROOT/branches"

echo "→ branch ${GIT_REF} → slug ${SLUG}"

if [[ ! -d "$BRANCH_DIR/.git" ]]; then
  git clone "$REPO_URL" "$BRANCH_DIR"
fi

cd "$BRANCH_DIR"
git fetch origin --prune
if git show-ref --verify --quiet "refs/remotes/origin/${GIT_REF}"; then
  git checkout -B "preview/${SLUG}" "origin/${GIT_REF}"
else
  echo "✗ branch not found on origin: ${GIT_REF}"
  exit 1
fi

COMMIT="$(git rev-parse --short HEAD)"
read -r GW_PORT MINIO_PORT < <(preview_allocate_ports "$SLUG")
export COMPOSE_PROJECT_NAME="even-prev-${SLUG}"
export PREVIEW_GATEWAY_PORT="$GW_PORT"
export PREVIEW_MINIO_PORT="$MINIO_PORT"
export COMPOSE_FILE="docker-compose.yml:docker-compose.prod.yml:docker-compose.branch.yml"

# .env for this branch (isolated bucket per slug)
if [[ ! -f .env ]]; then
  cp deploy/env.branch-preview.example .env
  sed -i.bak \
    -e "s|PREVIEW_HOST|${PREVIEW_HOST}|g" \
    -e "s|BRANCH_SLUG|${SLUG}|g" \
    -e "s|even-media-preview|even-media-${SLUG}|g" \
    .env && rm -f .env.bak
fi

chmod +x scripts/*.sh

echo "→ migrations (project ${COMPOSE_PROJECT_NAME})..."
./scripts/migrate.sh

echo "→ building stack (gw ${GW_PORT}, minio ${MINIO_PORT})..."
docker compose up --build -d

echo "→ waiting for gateway..."
deadline=$((SECONDS + 240))
until curl -sf "http://127.0.0.1:${GW_PORT}/api/v1/ready" >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ timeout"
    docker compose ps
    exit 1
  fi
  sleep 3
done

preview_registry_upsert "$SLUG" "$GIT_REF" "$GW_PORT" "$MINIO_PORT" "$COMMIT"
"$ROOT/scripts/nginx-preview.sh" add "$SLUG" "$GW_PORT" "$MINIO_PORT"

echo "✓ preview deployed"
echo "  branch:  ${GIT_REF}"
echo "  slug:    ${SLUG}"
echo "  commit:  ${COMMIT}"
echo "  path:    http://${PREVIEW_HOST}/preview/${SLUG}/api/v1/ready"
echo "  direct:  http://127.0.0.1:${GW_PORT}/health (localhost only)"
