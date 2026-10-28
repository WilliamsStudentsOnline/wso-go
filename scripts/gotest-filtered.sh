#!/usr/bin/env bash
# go test wrapper for gotestsum --raw-command / --watch.
# Replaces ./... (what gotestsum's `a` key sends) with the coverage-exclude package list.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ "${IGNORE_SKIPS:-}" == "1" ]]; then
  filtered=(./...)
else
  hash='#'
  skip_regex="$(awk "/^[^${hash}[:space:]]/ { print \$1 }" .github/coverage-exclude.txt | paste -sd'|' -)"
  mapfile -t filtered < <(go list ./... | grep -Ev "/(${skip_regex})(/|$)")
  if [[ ${#filtered[@]} -eq 0 ]]; then
    echo "gotest-filtered: no packages left after coverage-exclude filter" >&2
    exit 1
  fi
fi

args=()
replaced=0
for arg in "$@"; do
  if [[ "$arg" == "./..." ]]; then
    args+=("${filtered[@]}")
    replaced=1
  else
    args+=("$arg")
  fi
done

# gotestsum --watch with no package args, or `a`, should still hit the filtered set
if [[ "$replaced" -eq 0 ]]; then
  has_pkg=0
  for arg in "${args[@]}"; do
    case "$arg" in
      -* | "") ;;
      *)
        # rough: anything that looks like a package path / pattern
        if [[ "$arg" == ./* || "$arg" == github.com/* ]]; then
          has_pkg=1
          break
        fi
        ;;
    esac
  done
  if [[ "$has_pkg" -eq 0 ]]; then
    args+=("${filtered[@]}")
  fi
fi

exec go test -json "${args[@]}"
