#!/usr/bin/env bash
# Add or remove nginx snippet for a branch preview.
set -euo pipefail

ACTION="${1:-}"
SLUG="${2:-}"
GW_PORT="${3:-}"
MINIO_PORT="${4:-}"

PREVIEWS_ROOT="${PREVIEWS_ROOT:-/opt/even-app-previews}"
ENABLED_DIR="${PREVIEWS_ROOT}/nginx/enabled"
TEMPLATE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/deploy/nginx/branch-preview.conf.template"
SNIPPET="${ENABLED_DIR}/${SLUG}.conf"

usage() {
  echo "usage: $0 add <slug> <gateway_port> <minio_port>"
  echo "       $0 remove <slug>"
  exit 1
}

reload_nginx() {
  if command -v nginx >/dev/null 2>&1; then
    sudo nginx -t && sudo systemctl reload nginx
  else
    echo "⚠ nginx not installed — install for path-based previews (see DEPLOY.md)"
  fi
}

case "$ACTION" in
  add)
    [[ -n "$SLUG" && -n "$GW_PORT" && -n "$MINIO_PORT" ]] || usage
    mkdir -p "$ENABLED_DIR"
    sed -e "s/BRANCH_SLUG/${SLUG}/g" \
        -e "s/GATEWAY_PORT/${GW_PORT}/g" \
        -e "s/MINIO_PORT/${MINIO_PORT}/g" \
        "$TEMPLATE" >"$SNIPPET"
    reload_nginx
    echo "✓ nginx snippet: ${SNIPPET}"
    ;;
  remove)
    [[ -n "$SLUG" ]] || usage
    rm -f "$SNIPPET"
    reload_nginx
    echo "✓ removed nginx snippet for ${SLUG}"
    ;;
  *)
    usage
    ;;
esac
