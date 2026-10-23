#!/usr/bin/env bash
set -euo pipefail

ci=0
if [[ "${1:-}" == "--ci" ]]; then
  ci=1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "error: git is not installed or not on PATH" >&2
  exit 1
fi

git fetch origin master production

parse_approved() {
  grep -v '^[[:space:]]*#' | grep -v '^[[:space:]]*$' | head -1 | tr -d '[:space:]'
}

if ! git cat-file -e "origin/master:config/prod-sha" 2>/dev/null; then
  echo "error: config/prod-sha not on origin/master yet" >&2
  exit 1
fi

approved="$(git show origin/master:config/prod-sha | parse_approved)"
approved_src="REMOTE master:config/prod-sha"

if [[ ! "$approved" =~ ^[0-9a-f]{40}$ ]]; then
  echo "error: approved SHA missing or invalid in ${approved_src}" >&2
  echo "got: ${approved:-empty}" >&2
  exit 1
fi

prod="$(git rev-parse origin/production)"
master="$(git rev-parse origin/master)"
head_sha="$(git rev-parse HEAD)"

fmt_commit() {
  local sha="$1"
  local label="$2"
  echo "${label}:"
  echo "  $(git rev-parse --short "$sha")  $(git show -s --format='%ci (%cr)' "$sha")"
  echo "  $(git show -s --format='%an <%ae>' "$sha")"
  echo "  $(git show -s --format='%s' "$sha")"
}

echo "approved source: ${approved_src}"
echo
fmt_commit "$approved" "approved"
echo
fmt_commit "$prod" "production"
echo
fmt_commit "$master" "master"

if [[ "$ci" -eq 0 ]]; then
  echo
  fmt_commit "$head_sha" "local HEAD"
  if [[ -f config/prod-sha ]]; then
    local_approved="$(parse_approved <config/prod-sha)"
    if [[ "$local_approved" =~ ^[0-9a-f]{40}$ ]] && [[ "$local_approved" != "$approved" ]]; then
      echo
      fmt_commit "$local_approved" "local config/prod-sha (differs from remote)"
    fi
  fi
fi

echo
if [[ "$prod" == "$approved" ]]; then
  echo "OK: production matches approved SHA"
  exit 0
fi

echo "MISMATCH: production does not match approved SHA"
if git merge-base --is-ancestor "$prod" "$approved"; then
  echo "production < approved"
elif git merge-base --is-ancestor "$approved" "$prod"; then
  echo "note: production > approved"
else
  echo "production and approved SHAs have diverged"
fi
exit 1
