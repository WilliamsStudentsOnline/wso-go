#!/usr/bin/env bash
# Print go packages to include in CI tests/coverage, excluding prefixes listed in
# .github/coverage-exclude.txt (or $COVERAGE_EXCLUDE_FILE).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXCLUDE_FILE="${COVERAGE_EXCLUDE_FILE:-$ROOT/.github/coverage-exclude.txt}"

if [[ ! -f "$EXCLUDE_FILE" ]]; then
  echo "coverage exclude file not found: $EXCLUDE_FILE" >&2
  exit 1
fi

cd "$ROOT"
module="$(go list -m)"

prefixes=()
while IFS= read -r line || [[ -n "$line" ]]; do
  # Strip comments and surrounding whitespace.
  line="${line%%#*}"
  line="${line#"${line%%[![:space:]]*}"}"
  line="${line%"${line##*[![:space:]]}"}"
  [[ -z "$line" ]] && continue

  if [[ "$line" == /* || "$line" == *..* ]]; then
    echo "invalid coverage exclude path: $line" >&2
    exit 1
  fi
  if [[ ! -d "$ROOT/$line" ]]; then
    echo "coverage exclude path does not exist: $line" >&2
    exit 1
  fi
  prefixes+=("$line")
done < "$EXCLUDE_FILE"

should_exclude() {
  local rel="$1"
  local prefix
  for prefix in "${prefixes[@]}"; do
    if [[ "$rel" == "$prefix" || "$rel" == "$prefix"/* ]]; then
      return 0
    fi
  done
  return 1
}

while IFS= read -r pkg; do
  rel="${pkg#"$module"/}"
  if [[ "$rel" == "$pkg" ]]; then
    # Module root package itself.
    rel="."
  fi
  if should_exclude "$rel"; then
    continue
  fi
  printf '%s\n' "$pkg"
done < <(go list ./...)
