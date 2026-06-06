#!/usr/bin/env bash
# One-time server setup (run on 91.218.245.136 as root or sudo user).
# Usage: curl -fsSL ... | bash   OR   scp + ssh bash scripts/server-bootstrap.sh
set -euo pipefail

DEPLOY_PATH="${DEPLOY_PATH:-/opt/even-app}"
REPO_URL="${REPO_URL:-https://github.com/Klimgit/Even-APP.git}"

echo "→ installing docker (if missing)..."
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi

echo "→ installing git (if missing)..."
command -v git >/dev/null 2>&1 || apt-get update && apt-get install -y git curl

echo "→ cloning repo to ${DEPLOY_PATH}..."
if [[ ! -d "$DEPLOY_PATH/.git" ]]; then
  mkdir -p "$DEPLOY_PATH"
  git clone "$REPO_URL" "$DEPLOY_PATH"
else
  echo "  repo already exists at ${DEPLOY_PATH}"
fi

cd "$DEPLOY_PATH"
if [[ ! -f .env ]]; then
  cp deploy/env.production.example .env
  echo "→ created .env from template — EDIT SECRETS before deploy:"
  echo "  nano ${DEPLOY_PATH}/.env"
fi

chmod +x scripts/*.sh

echo ""
echo "✓ bootstrap done"
echo "  1. Edit ${DEPLOY_PATH}/.env (passwords, JWT_SECRET)"
echo "  2. Open firewall: 8080/tcp (API gateway)"
echo "  3. Add GitHub secrets for CI/CD (see DEPLOY.md)"
echo "  4. Run: cd ${DEPLOY_PATH} && ./scripts/deploy.sh"
