#!/usr/bin/env bash
# Regenerate OpenAPI via swag, lint with Spectral, and oasdiff against a base ref.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

OPENAPI_BASE_REF="${OPENAPI_BASE_REF:-origin/master}"
SKIP_DIFF="${SKIP_OPENAPI_DIFF:-0}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required tool not found: $1" >&2
    exit 1
  fi
}

need swag
need spectral
need oasdiff

cleanup() {
  rm -rf "${SNAP_DIR:-}" "${BASE_DIR:-}"
}
trap cleanup EXIT

echo "==> freshness check (docs must match swag output)"
SNAP_DIR="$(mktemp -d)"
cp docs/swagger.yaml docs/swagger.json docs/docs.go "$SNAP_DIR/"

swag init -g server/router.go
if command -v goimports >/dev/null 2>&1; then
  goimports -w docs/docs.go
fi

if ! diff -q "$SNAP_DIR/swagger.yaml" docs/swagger.yaml >/dev/null \
  || ! diff -q "$SNAP_DIR/swagger.json" docs/swagger.json >/dev/null \
  || ! diff -q "$SNAP_DIR/docs.go" docs/docs.go >/dev/null; then
  echo "error: docs/ is out of date relative to swag. Run: make openapi-docs && commit the result" >&2
  diff -u "$SNAP_DIR/swagger.yaml" docs/swagger.yaml | head -n 100 || true
  # restore pre-check docs so the working tree is not left half-updated on failure
  cp "$SNAP_DIR/swagger.yaml" "$SNAP_DIR/swagger.json" "$SNAP_DIR/docs.go" docs/
  exit 1
fi

echo "==> spectral lint"
spectral lint docs/swagger.yaml --ruleset .spectral.yaml --fail-severity=error

if [[ "$SKIP_DIFF" == "1" ]]; then
  echo "==> oasdiff skipped (SKIP_OPENAPI_DIFF=1)"
  exit 0
fi

echo "==> oasdiff breaking vs ${OPENAPI_BASE_REF}"
BASE_DIR="$(mktemp -d)"

if ! git cat-file -e "${OPENAPI_BASE_REF}:docs/swagger.yaml" 2>/dev/null; then
  echo "error: ${OPENAPI_BASE_REF}:docs/swagger.yaml not found (fetch the base ref?)" >&2
  exit 1
fi

git show "${OPENAPI_BASE_REF}:docs/swagger.yaml" >"$BASE_DIR/base.yaml"
oasdiff breaking "$BASE_DIR/base.yaml" docs/swagger.yaml --fail-on ERR --format text

echo "==> openapi checks passed"
