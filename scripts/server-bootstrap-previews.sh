#!/usr/bin/env bash
# One-time setup for multi-branch previews + nginx path routing.
set -euo pipefail

PREVIEWS_ROOT="${PREVIEWS_ROOT:-/opt/even-app-previews}"
REPO_URL="${REPO_URL:-https://github.com/Klimgit/Even-APP.git}"
MANAGER="${PREVIEWS_ROOT}/manager"

echo "→ docker..."
command -v docker >/dev/null 2>&1 || { curl -fsSL https://get.docker.com | sh; systemctl enable --now docker; }

echo "→ nginx..."
if ! command -v nginx >/dev/null 2>&1; then
  apt-get update && apt-get install -y nginx
fi

mkdir -p "${PREVIEWS_ROOT}/branches" "${PREVIEWS_ROOT}/nginx/enabled"

if [[ ! -d "${MANAGER}/.git" ]]; then
  git clone "$REPO_URL" "$MANAGER"
fi

# Main nginx include (production :8080 stays separate)
if [[ ! -f /etc/nginx/sites-enabled/even-previews.conf ]]; then
  cp "${MANAGER}/deploy/nginx/even-previews.conf" /etc/nginx/sites-available/even-previews.conf
  ln -sf /etc/nginx/sites-available/even-previews.conf /etc/nginx/sites-enabled/even-previews.conf
  nginx -t && systemctl reload nginx
fi

chmod +x "${MANAGER}/scripts/"*.sh "${MANAGER}/scripts/lib/"*.sh 2>/dev/null || true

echo "✓ previews root: ${PREVIEWS_ROOT}"
echo "  production:    /opt/even-app :8080 (unchanged)"
echo "  branch URL:    http://HOST/preview/<slug>/api/v1/ready"
echo "  deploy branch: ${MANAGER}/scripts/deploy-branch-preview.sh <branch>"
echo "  cleanup:       ${MANAGER}/scripts/cleanup-branch-preview.sh <branch>"
