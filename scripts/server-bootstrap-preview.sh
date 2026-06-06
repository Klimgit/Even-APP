#!/usr/bin/env bash
# One-time preview slot setup on the server (separate from production).
set -euo pipefail

PREVIEW_PATH="${PREVIEW_PATH:-/opt/even-app-preview}"
REPO_URL="${REPO_URL:-https://github.com/Klimgit/Even-APP.git}"

if ! command -v docker >/dev/null 2>&1; then
  echo "✗ docker not found — run scripts/server-bootstrap.sh first"
  exit 1
fi

if [[ ! -d "$PREVIEW_PATH/.git" ]]; then
  git clone "$REPO_URL" "$PREVIEW_PATH"
fi

cd "$PREVIEW_PATH"
if [[ ! -f .env ]]; then
  cp deploy/env.preview.example .env
  echo "→ edit secrets: nano ${PREVIEW_PATH}/.env"
fi

chmod +x scripts/*.sh

echo "✓ preview slot at ${PREVIEW_PATH}"
echo "  open firewall: 9080/tcp (preview API), 9010/tcp (preview MinIO, optional)"
echo "  deploy branch:   ./scripts/deploy-preview.sh my-feature-branch"
echo "  GitHub Actions:  workflow Deploy Preview (manual)"
