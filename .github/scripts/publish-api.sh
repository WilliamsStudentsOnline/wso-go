#!/usr/bin/env bash
# Tag HEAD as openapi-vYYYY.MM.DD[-N] and push so the release workflow attaches openapi.yaml
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

if [[ ! -f docs/swagger.yaml ]]; then
  echo "error: docs/swagger.yaml missing — run make api first (or merge the OpenAPI CI PR)" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is dirty; commit or stash before publishing" >&2
  exit 1
fi

git fetch origin --tags --quiet

day="$(date +%Y.%m.%d)"
base="openapi-v${day}"
tag="$base"
n=2
while git rev-parse -q --verify "refs/tags/${tag}" >/dev/null \
  || git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; do
  tag="${base}-${n}"
  n=$((n + 1))
done

echo "Publishing API contract as ${tag} (commit $(git rev-parse --short HEAD))"
git tag "$tag"
git push origin "refs/tags/${tag}"

echo
echo "Release workflow will attach:"
echo "  https://github.com/WilliamsStudentsOnline/wso-go/releases/download/${tag}/openapi.yaml"
