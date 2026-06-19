#!/usr/bin/env bash
# Run Flutter web against the local API gateway (USE_API_AUTH=true).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/frontend/flutter"

DEVICE="${FLUTTER_WEB_DEVICE:-web-server}"
PORT="${FLUTTER_WEB_PORT:-5173}"
API_BASE="${API_BASE_URL:-http://localhost:8080/api/v1}"

flutter pub get
exec flutter run -d "$DEVICE" --web-port="$PORT" \
  --dart-define=USE_API_AUTH=true \
  --dart-define="API_BASE_URL=$API_BASE"
