#!/usr/bin/env bash
# Local dev stuff
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-$(basename "$ROOT")}"
MYSQL_HOST_PORT="${MYSQL_HOST_PORT:-3306}"
REDIS_HOST_PORT="${REDIS_HOST_PORT:-6379}"
TEST_MYSQL_HOST_PORT="${TEST_MYSQL_HOST_PORT:-3307}"
MYSQL_PASSWORD="${MYSQL_ROOT_PASSWORD:-secret-mysql-password}"
MYSQL_DATABASE="${MYSQL_DATABASE:-wso}"
SEED_SQL="${SEED_SQL:-$ROOT/db/seed/development.sql.gz}"
AIR_VERSION="${AIR_VERSION:-v1.67.0}"
DEV_CONFIG="${DEV_CONFIG:-config/environment/development.yaml}"

export COMPOSE_PROJECT_NAME
export MYSQL_HOST_PORT REDIS_HOST_PORT TEST_MYSQL_HOST_PORT
export WSO_MYSQL_PORT="$MYSQL_HOST_PORT"
export WSO_REDIS_ADDR="127.0.0.1:${REDIS_HOST_PORT}"

COMPOSE=(docker compose -p "$COMPOSE_PROJECT_NAME")
TEST_COMPOSE=(docker compose -p "${COMPOSE_PROJECT_NAME}-test" -f docker-compose.test.yml)

CLEANING=0
AIR_PID=""

die() { echo "devenv: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing '$1'"
}

stop_test_mysql() {
  local wipe="${1:-}"
  if [[ "$wipe" == "wipe" ]]; then
    "${TEST_COMPOSE[@]}" down -v >/dev/null 2>&1 || true
  else
    "${TEST_COMPOSE[@]}" down >/dev/null 2>&1 || true
  fi
}

prereq() {
  need docker
  need go
  docker info >/dev/null 2>&1 || die "docker daemon not running"
  docker compose version >/dev/null 2>&1 || die "docker compose plugin required"
  if [[ ! -f config/secrets.yaml ]]; then
    die "missing config/secrets.yaml; run: cp config/secrets_example.yaml config/secrets.yaml"
  fi
}

db_has_users_table() {
  "${COMPOSE[@]}" exec -T mysql mysql -uroot -p"$MYSQL_PASSWORD" -N -e \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${MYSQL_DATABASE}' AND table_name='users';" \
    2>/dev/null | tr -d '\r' | grep -qx '1'
}

# MySQL 8 dumps use utf8mb4_0900_*; MariaDB 10.11 rejects those.
# Rewrite to the server collation from docker-compose (utf8mb4_unicode_ci).
rewrite_mysql8_collations() {
  sed -E 's/utf8mb4_0900_[A-Za-z0-9_]+/utf8mb4_unicode_ci/g'
}

import_seed_sql() {
  local src="$1"
  echo "devenv: importing seed from $src"
  case "$src" in
    *.gz)
      gunzip -c "$src" | rewrite_mysql8_collations \
        | "${COMPOSE[@]}" exec -T mysql mysql -uroot -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE"
      ;;
    *)
      rewrite_mysql8_collations <"$src" \
        | "${COMPOSE[@]}" exec -T mysql mysql -uroot -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE"
      ;;
  esac
}

import_seed_if_empty() {
  if db_has_users_table; then
    return 0
  fi
  [[ -f "$SEED_SQL" ]] || die "seed missing at $SEED_SQL"
  import_seed_sql "$SEED_SQL"
}

# Apply any Atlas migrations newer than the dump (or catch up an existing volume).
migrate_db() {
  echo "devenv: running migrations"
  go run ./db/seed/cmd -config "$DEV_CONFIG" -model none
}

cleanup() {
  # ignore further Ctrl+C while tearing down
  if [[ "$CLEANING" -eq 1 ]]; then
    return 0
  fi
  CLEANING=1
  trap '' INT TERM

  if [[ -n "$AIR_PID" ]] && kill -0 "$AIR_PID" 2>/dev/null; then
    kill -TERM "$AIR_PID" 2>/dev/null || true
    # wso-backend may outlive go-run briefly
    pkill -P "$AIR_PID" 2>/dev/null || true
    wait "$AIR_PID" 2>/dev/null || true
  fi
  pkill -f './tmp/wso-backend' 2>/dev/null || true

  echo "devenv: stopped"
  "${COMPOSE[@]}" down >/dev/null 2>&1 || true
  stop_test_mysql
}

cmd_down() {
  echo "devenv: stopping containers"
  "${COMPOSE[@]}" down
  stop_test_mysql
}

cmd_reset() {
  echo "devenv: stopping containers and deleting volumes"
  "${COMPOSE[@]}" down -v
  stop_test_mysql wipe
}

cmd_up() {
  prereq

  echo "devenv: starting background services"
  "${COMPOSE[@]}" up -d --wait

  import_seed_if_empty
  migrate_db

  trap cleanup EXIT INT TERM
  echo "devenv: starting API! Ctrl+C to stop"
  go run "github.com/air-verse/air@${AIR_VERSION}" -c .air.toml &
  AIR_PID=$!
  wait "$AIR_PID" || true
}

case "${1:-}" in
  up|down|reset) "cmd_$1" ;;
  *) die "usage: $0 {up|down|reset}" ;;
esac
