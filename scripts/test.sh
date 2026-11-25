#!/usr/bin/env bash
# Run tests by package and/or name, e.g. scripts/test.sh services/feedback TestCreateFeedback
set -euo pipefail

usage() {
  echo "VERBOSE=1 usage: scripts/test.sh [pkg ...] [TestName ...] [-- go test args]" >&2
  exit 2
}

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  . ./.env
  set +a
fi

pkgs=()
names=()
extra=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help) usage ;;
    --)
      shift
      extra=("$@")
      break
      ;;
    Test*) names+=("$1") ;;
    .* | */*)
      p="${1%/}"
      [[ "$p" == .* ]] || p="./$p"
      pkgs+=("$p")
      ;;
    *)
      echo "test.sh: not a package path or test name: $1" >&2
      usage
      ;;
  esac
  shift
done

if [[ ${#pkgs[@]} -eq 0 ]]; then
  module="$(go list -m)"
  if [[ "${IGNORE_SKIPS:-}" == "1" ]]; then
    list="$(go list ./...)"
  else
    hash='#'
    skip_regex="$(awk "/^[^${hash}[:space:]]/ { print \$1 }" .github/coverage-exclude.txt | paste -sd'|' -)"
    list="$(go list ./... | grep -Ev "/(${skip_regex})(/|$)")"
  fi
  while IFS= read -r p; do
    pkgs+=(".${p#"$module"}")
  done <<<"$list"
fi

run_args=()
if [[ ${#names[@]} -gt 0 ]]; then
  joined="$(IFS='|'; echo "${names[*]}")"
  run_args=(-run "^(${joined})$")
fi

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-$(basename "$ROOT")}"
TEST_COMPOSE_PROJECT="${TEST_COMPOSE_PROJECT:-${COMPOSE_PROJECT_NAME}-test}"
docker compose -p "$TEST_COMPOSE_PROJECT" -f docker-compose.test.yml --profile redis up -d --wait

export CI=1
export WSO_TEST_MYSQL_HOST="${WSO_TEST_MYSQL_HOST:-${TEST_MYSQL_HOST:-127.0.0.1}}"
export WSO_TEST_MYSQL_PORT="${WSO_TEST_MYSQL_PORT:-${TEST_MYSQL_PORT:-${TEST_MYSQL_HOST_PORT:-3307}}}"
export WSO_TEST_MYSQL_USER="${WSO_TEST_MYSQL_USER:-root}"
export WSO_TEST_MYSQL_PASSWORD="${WSO_TEST_MYSQL_PASSWORD:-${TEST_MYSQL_ROOT_PASSWORD:-secret-mysql-password}}"
export WSO_TEST_REDIS_ADDR="${WSO_TEST_REDIS_ADDR:-127.0.0.1:${TEST_REDIS_HOST_PORT:-6380}}"

# bash 3.2 treats "${empty[@]}" as unbound under set -u
go_args=(-race ${run_args[@]+"${run_args[@]}"} ${extra[@]+"${extra[@]}"} "${pkgs[@]}")
fmt=pkgname-and-test-fails
if [[ "${VERBOSE:-}" == "1" ]]; then
  fmt=standard-verbose
  go_args=(-v "${go_args[@]}")
fi
if command -v gotestsum >/dev/null 2>&1; then
  exec gotestsum --format="$fmt" --format-hide-empty-pkg --hide-summary=output -- "${go_args[@]}"
fi
exec go test "${go_args[@]}"
