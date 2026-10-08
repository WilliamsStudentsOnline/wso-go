#!/usr/bin/env bash
# Local API lifecycle: Compose MySQL+Redis + host-built wso-backend
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DEVENV_DIR="${DEVENV_DIR:-$ROOT/.devenv}"
PID_FILE="$DEVENV_DIR/api.pid"
LOG_FILE="$DEVENV_DIR/api.log"
COMPOSE=(docker compose)
MYSQL_PASSWORD="${MYSQL_ROOT_PASSWORD:-secret-mysql-password}"
MYSQL_DATABASE="${MYSQL_DATABASE:-wso}"
SEED_SQL="${SEED_SQL:-$ROOT/db/seed/development.sql.gz}"
API_PORT="${API_PORT:-8080}"
BINARY="${BINARY_NAME:-wso-backend}"

die() { echo "devenv: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing '$1'"
}

prereq() {
  need docker
  need go
  docker info >/dev/null 2>&1 || die "docker daemon not running"
  docker compose version >/dev/null 2>&1 || die "docker compose plugin required"
}

ensure_secrets() {
  if [[ ! -f config/secrets.yaml ]]; then
    cp config/secrets_example.yaml config/secrets.yaml
    echo "devenv: created config/secrets.yaml from secrets_example.yaml"
  fi
}

compose_running() {
  local id
  id="$("${COMPOSE[@]}" ps -q mysql 2>/dev/null || true)"
  [[ -n "$id" ]]
}

api_pid() {
  if [[ -f "$PID_FILE" ]]; then
    cat "$PID_FILE"
  fi
}

api_alive() {
  local pid
  pid="$(api_pid || true)"
  [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null
}

port_in_use() {
  if command -v ss >/dev/null 2>&1; then
    ss -ltn "( sport = :$API_PORT )" 2>/dev/null | grep -q ":$API_PORT"
  else
    # fallback
    (echo >/dev/tcp/127.0.0.1/"$API_PORT") >/dev/null 2>&1
  fi
}

wait_healthy() {
  local svc="$1"
  local i h
  for i in $(seq 1 60); do
    h="$(docker inspect --format='{{.State.Health.Status}}' "wso-$svc" 2>/dev/null || echo starting)"
    if [[ "$h" == "healthy" ]]; then
      return 0
    fi
    sleep 1
  done
  die "$svc not healthy"
}

db_has_users_table() {
  docker exec wso-mysql mysql -uroot -p"$MYSQL_PASSWORD" -N -e \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${MYSQL_DATABASE}' AND table_name='users';" \
    2>/dev/null | tr -d '\r' | grep -qx '1'
}

import_seed_if_empty() {
  if db_has_users_table; then
    echo "devenv: mysql already has users table; skipping seed"
    return 0
  fi
  if [[ ! -f "$SEED_SQL" ]]; then
    die "seed missing at $SEED_SQL (generate with make seed-dump)"
  fi
  echo "devenv: importing seed into empty mysql ($MYSQL_DATABASE)"
  case "$SEED_SQL" in
    *.gz) gunzip -c "$SEED_SQL" | docker exec -i wso-mysql mysql -uroot -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" ;;
    *) docker exec -i wso-mysql mysql -uroot -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" <"$SEED_SQL" ;;
  esac
}

cmd_up() {
  prereq
  ensure_secrets
  mkdir -p "$DEVENV_DIR"

  if api_alive; then
    die "API already running (pid $(api_pid)); try make dev-down"
  fi
  if port_in_use; then
    die "port $API_PORT already in use"
  fi
  if compose_running; then
    die "compose already up; try make dev-down or make dev-logs"
  fi

  echo "devenv: starting mysql + redis"
  "${COMPOSE[@]}" up -d
  wait_healthy mysql
  wait_healthy redis
  import_seed_if_empty

  echo "devenv: building $BINARY"
  if ! make -s "$BINARY"; then
    die "build failed"
  fi

  echo "devenv: starting API (logs: $LOG_FILE)"
  : >"$LOG_FILE"
  nohup "./$BINARY" --development >>"$LOG_FILE" 2>&1 &
  echo $! >"$PID_FILE"
  sleep 1
  if ! api_alive; then
    die "API failed to start; see $LOG_FILE"
  fi
  echo "devenv: up — http://127.0.0.1:${API_PORT}/docs (pid $(api_pid))"
}

cmd_down() {
  if api_alive; then
    echo "devenv: stopping API pid $(api_pid)"
    kill "$(api_pid)" 2>/dev/null || true
    # wait briefly
    for _ in $(seq 1 20); do
      api_alive || break
      sleep 0.2
    done
    if api_alive; then
      kill -9 "$(api_pid)" 2>/dev/null || true
    fi
  fi
  rm -f "$PID_FILE"
  if compose_running || "${COMPOSE[@]}" ps -q 2>/dev/null | grep -q .; then
    echo "devenv: compose down (volumes kept)"
    "${COMPOSE[@]}" down
  fi
  echo "devenv: down"
}

cmd_reset() {
  cmd_down
  echo "devenv: wiping volumes"
  "${COMPOSE[@]}" down -v
  # fresh up + seed
  cmd_up
}

cmd_logs() {
  if [[ -f "$LOG_FILE" ]]; then
    echo "===== API ($LOG_FILE) ====="
    tail -n 80 "$LOG_FILE" || true
  else
    echo "devenv: no API log yet"
  fi
  echo "===== compose ====="
  "${COMPOSE[@]}" logs --tail=80 || true
}

case "${1:-}" in
  up) cmd_up ;;
  down) cmd_down ;;
  reset) cmd_reset ;;
  logs) cmd_logs ;;
  *)
    echo "usage: $0 {up|down|reset|logs}" >&2
    exit 2
    ;;
esac
