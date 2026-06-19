#!/usr/bin/env bash
# Run colleague Flutter app (frontend/flutter) on web.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/frontend/flutter"

DEVICE="${FLUTTER_WEB_DEVICE:-web-server}"
PORT="${FLUTTER_WEB_PORT:-5173}"

flutter pub get
exec flutter run -d "$DEVICE" --web-port="$PORT"
