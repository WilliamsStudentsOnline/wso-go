#!/usr/bin/env bash
# Local dev stuff
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose)
MYSQL_PASSWORD="${MYSQL_ROOT_PASSWORD:-secret-mysql-password}"
MYSQL_DATABASE="${MYSQL_DATABASE:-wso}"
AIR_VERSION="${AIR_VERSION:-v1.67.0}"
DEV_CONFIG="${DEV_CONFIG:-config/environment/development.yaml}"

CLEANING=0
AIR_PID=""

die() { echo "devenv: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing '$1'"
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

# Returns 0 when the users table exists and has at least one row.
db_has_seed_data() {
  local count
  count="$(docker exec wso-mysql mysql -uroot -p"$MYSQL_PASSWORD" -N -e \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${MYSQL_DATABASE}' AND table_name='users';" \
    2>/dev/null | tr -d '\r' || true)"
  [[ "$count" == "1" ]] || return 1
  count="$(docker exec wso-mysql mysql -uroot -p"$MYSQL_PASSWORD" -N -e \
    "SELECT COUNT(*) FROM \`${MYSQL_DATABASE}\`.users;" \
    2>/dev/null | tr -d '\r' || true)"
  [[ -n "$count" && "$count" -gt 0 ]]
}

# empty DB → migrate → synthetic seed (see db/seed/seeder).
seed_if_empty() {
  if db_has_seed_data; then
    echo "devenv: database already seeded"
    return 0
  fi
  echo "devenv: migrate + seed (empty → migrate → seed)"
  go run ./db/seed/cmd -config "$DEV_CONFIG" -model all
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
}

cmd_down() {
  echo "devenv: stopping containers"
  "${COMPOSE[@]}" down
}

cmd_reset() {
  echo "devenv: stopping containers and deleting volumes"
  "${COMPOSE[@]}" down -v
}

cmd_up() {
  prereq

  echo "devenv: starting background services"
  "${COMPOSE[@]}" up -d --wait

  seed_if_empty

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
