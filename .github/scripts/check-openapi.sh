#!/usr/bin/env bash
# Regenerate OpenAPI via swag, lint with Spectral, and oasdiff against a base ref.
# Writes openapi-results/ + openapi-comment.md for the sticky PR comment.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

OPENAPI_BASE_REF="${OPENAPI_BASE_REF:-origin/master}"
SKIP_DIFF="${SKIP_OPENAPI_DIFF:-0}"
RESULTS_DIR="${OPENAPI_RESULTS_DIR:-openapi-results}"
COMMENT_OUT="${OPENAPI_COMMENT_OUT:-openapi-comment.md}"

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

rm -rf "$RESULTS_DIR"
mkdir -p "$RESULTS_DIR"
failed=0

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
  cp "$SNAP_DIR/swagger.yaml" "$SNAP_DIR/swagger.json" "$SNAP_DIR/docs.go" docs/
  echo fail >"$RESULTS_DIR/freshness.status"
  failed=1
else
  echo pass >"$RESULTS_DIR/freshness.status"
fi

echo "==> spectral lint"
set +e
spectral lint docs/swagger.yaml --ruleset .spectral.yaml --fail-severity=error \
  -f json >"$RESULTS_DIR/spectral.json"
spectral_rc=$?
set -e
if [[ "$spectral_rc" -ne 0 ]]; then
  # spectral also exits non-zero on warnings depending on version; re-check errors
  if python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); sys.exit(0 if any(x.get("severity")==0 for x in d) else 1)' \
    "$RESULTS_DIR/spectral.json" 2>/dev/null; then
    failed=1
  fi
  # Always keep human-readable spectral output in logs
  spectral lint docs/swagger.yaml --ruleset .spectral.yaml --fail-severity=error || true
fi

if [[ "$SKIP_DIFF" == "1" ]]; then
  echo "==> oasdiff skipped (SKIP_OPENAPI_DIFF=1)"
  echo pass >"$RESULTS_DIR/breaking.status"
  echo '[]' >"$RESULTS_DIR/changelog.json"
else
  echo "==> oasdiff vs ${OPENAPI_BASE_REF}"
  BASE_DIR="$(mktemp -d)"

  if ! git cat-file -e "${OPENAPI_BASE_REF}:docs/swagger.yaml" 2>/dev/null; then
    echo "error: ${OPENAPI_BASE_REF}:docs/swagger.yaml not found (fetch the base ref?)" >&2
    echo fail >"$RESULTS_DIR/breaking.status"
    echo "base ref missing: ${OPENAPI_BASE_REF}" >"$RESULTS_DIR/breaking.txt"
    echo '[]' >"$RESULTS_DIR/changelog.json"
    failed=1
  else
    git show "${OPENAPI_BASE_REF}:docs/swagger.yaml" >"$BASE_DIR/base.yaml"

    set +e
    oasdiff breaking "$BASE_DIR/base.yaml" docs/swagger.yaml --fail-on ERR --format text \
      >"$RESULTS_DIR/breaking.txt" 2>&1
    breaking_rc=$?
    set -e
    if [[ "$breaking_rc" -eq 0 ]]; then
      echo pass >"$RESULTS_DIR/breaking.status"
    else
      echo fail >"$RESULTS_DIR/breaking.status"
      failed=1
    fi

    set +e
    oasdiff changelog "$BASE_DIR/base.yaml" docs/swagger.yaml --format json \
      >"$RESULTS_DIR/changelog.json" 2>/dev/null
    set -e
    if [[ ! -s "$RESULTS_DIR/changelog.json" ]]; then
      echo '[]' >"$RESULTS_DIR/changelog.json"
    fi
  fi
fi

python3 .github/scripts/format-openapi-comment.py \
  --results-dir "$RESULTS_DIR" \
  --base-ref "$OPENAPI_BASE_REF" \
  -o "$COMMENT_OUT"

echo "==> wrote $COMMENT_OUT"
cat "$COMMENT_OUT"

if [[ "$failed" -ne 0 ]]; then
  echo "==> openapi checks failed" >&2
  exit 1
fi
echo "==> openapi checks passed"
