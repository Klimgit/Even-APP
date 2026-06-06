#!/usr/bin/env bash
# Preview branch registry: ports and metadata per slug.

preview_registry() {
  echo "${PREVIEWS_ROOT:-/opt/even-app-previews}/registry.json"
}

preview_registry_init() {
  local f
  f="$(preview_registry)"
  mkdir -p "$(dirname "$f")"
  [[ -f "$f" ]] || echo '{}' >"$f"
}

preview_port_in_use() {
  local port=$1
  ss -tln 2>/dev/null | grep -q ":${port} " && return 0
  return 1
}

preview_allocate_ports() {
  local slug=$1
  local gw_base=${PREVIEW_GW_PORT_BASE:-9100}
  local minio_base=${PREVIEW_MINIO_PORT_BASE:-9200}
  local max=${PREVIEW_PORT_SLOTS:-50}
  local reg f gw minio i existing

  reg="$(preview_registry)"
  preview_registry_init

  if command -v python3 >/dev/null 2>&1; then
    existing=$(REG_FILE="$reg" SLUG="$slug" python3 - <<'PY'
import json, os
path = os.environ["REG_FILE"]
slug = os.environ["SLUG"]
with open(path) as f:
    data = json.load(f)
e = data.get(slug)
if e:
    print(e["gateway_port"], e["minio_port"])
PY
)
    if [[ -n "$existing" ]]; then
      echo "$existing"
      return 0
    fi
  fi

  for ((i=0; i<max; i++)); do
    gw=$((gw_base + i))
    minio=$((minio_base + i))
    if preview_port_in_use "$gw" || preview_port_in_use "$minio"; then
      continue
    fi
    echo "$gw $minio"
    return 0
  done
  echo "✗ no free preview ports" >&2
  return 1
}

preview_registry_upsert() {
  local slug=$1 branch=$2 gw=$3 minio=$4 commit=$5
  local reg
  reg="$(preview_registry)"
  preview_registry_init
  REG_FILE="$reg" SLUG="$slug" BRANCH="$branch" GW="$gw" MINIO="$minio" COMMIT="$commit" \
    python3 - <<'PY'
import json, os, datetime
path = os.environ["REG_FILE"]
slug = os.environ["SLUG"]
with open(path) as f:
    data = json.load(f)
data[slug] = {
    "branch": os.environ["BRANCH"],
    "gateway_port": int(os.environ["GW"]),
    "minio_port": int(os.environ["MINIO"]),
    "commit": os.environ["COMMIT"],
    "deployed_at": datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ"),
    "path": f"/preview/{slug}/",
}
with open(path, "w") as f:
    json.dump(data, f, indent=2, sort_keys=True)
PY
}

preview_registry_remove() {
  local slug=$1
  local reg
  reg="$(preview_registry)"
  [[ -f "$reg" ]] || return 0
  REG_FILE="$reg" SLUG="$slug" python3 - <<'PY'
import json, os
path, slug = os.environ["REG_FILE"], os.environ["SLUG"]
with open(path) as f:
    data = json.load(f)
data.pop(slug, None)
with open(path, "w") as f:
    json.dump(data, f, indent=2, sort_keys=True)
PY
}
