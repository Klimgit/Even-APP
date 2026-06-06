#!/usr/bin/env bash
# Deploy a git branch/ref to the preview slot (ports 9080, isolated DB/MinIO).
# Usage: ./scripts/deploy-preview.sh feature/my-branch
set -euo pipefail

GIT_REF="${1:-}"
if [[ -z "$GIT_REF" ]]; then
  echo "usage: $0 <branch-or-ref>"
  echo "example: $0 feature/auth-fix"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  echo "✗ missing .env — copy deploy/env.preview.example to .env"
  exit 1
fi

export COMPOSE_FILE="docker-compose.yml:docker-compose.prod.yml:docker-compose.preview.yml"
export COMPOSE_PROJECT_NAME="even-preview"

echo "→ fetching ${GIT_REF}..."
git fetch origin --prune
if git show-ref --verify --quiet "refs/remotes/origin/${GIT_REF}"; then
  git checkout -B preview-deploy "origin/${GIT_REF}"
elif git rev-parse --verify "${GIT_REF}" >/dev/null 2>&1; then
  git checkout -B preview-deploy "${GIT_REF}"
else
  echo "✗ ref not found: ${GIT_REF}"
  exit 1
fi

echo "→ preview deploy @ $(git rev-parse --short HEAD) — ${GIT_REF}"
chmod +x scripts/*.sh

echo "→ applying migrations..."
./scripts/migrate.sh

echo "→ building and starting preview stack..."
docker compose up --build -d

echo "→ waiting for preview gateway (port 9080)..."
deadline=$((SECONDS + 180))
until curl -sf http://localhost:9080/api/v1/ready >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "✗ timeout — preview gateway not ready"
    docker compose ps
    exit 1
  fi
  sleep 3
done

HOST_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
echo "✓ preview ready"
echo "  branch: ${GIT_REF}"
echo "  commit: $(git rev-parse --short HEAD)"
echo "  url:    http://${HOST_IP:-localhost}:9080/api/v1/ready"
