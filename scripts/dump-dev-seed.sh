#!/usr/bin/env bash
# Refresh db/seed/development.sql.gz from the running devenv MySQL.
#
# Typical flow after schema changes:
#   1. make dev-reset && make dev   # import current gz, then migrate
#   2. (optional) tweak data in the DB
#   3. ./scripts/dump-dev-seed.sh   # rewrite the gz from this DB
#   4. commit db/seed/development.sql.gz
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
MYSQL_PASSWORD="${MYSQL_ROOT_PASSWORD:-secret-mysql-password}"
MYSQL_DATABASE="${MYSQL_DATABASE:-wso}"
OUT="${SEED_SQL:-$ROOT/db/seed/development.sql.gz}"

COMPOSE=(docker compose -p "$COMPOSE_PROJECT_NAME")

die() { echo "dump-dev-seed: $*" >&2; exit 1; }

"${COMPOSE[@]}" ps --status running --services 2>/dev/null | grep -qx mysql \
  || die "mysql is not running (start with make dev / ./scripts/devenv.sh up)"

echo "dump-dev-seed: dumping ${MYSQL_DATABASE} → ${OUT}"
"${COMPOSE[@]}" exec -T mysql mysqldump \
  -uroot -p"$MYSQL_PASSWORD" \
  --single-transaction \
  --routines \
  --triggers \
  --default-character-set=utf8mb4 \
  "$MYSQL_DATABASE" \
  | gzip -c >"$OUT"

echo "dump-dev-seed: wrote $(wc -c <"$OUT" | tr -d ' ') bytes"
