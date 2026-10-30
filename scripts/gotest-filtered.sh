#!/usr/bin/env bash
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

if [[ "$replaced" -eq 0 ]]; then
  has_pkg=0
  for arg in "${args[@]}"; do
    case "$arg" in
      -* | "") ;;
      *)
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
