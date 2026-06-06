#!/usr/bin/env bash
# branch_slug "feature/auth" -> feature-auth
branch_slug() {
  local raw="${1:-}"
  raw="${raw#refs/heads/}"
  echo "$raw" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g; s/-+/-/g' \
    | cut -c1-48
}
