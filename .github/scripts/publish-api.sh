#!/usr/bin/env bash
# Tag origin/master as openapi-vYYYY.MM.DD[-N] and push so the release workflow attaches openapi.yaml
# Must be run on a clean local master that matches origin/master
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" != "master" ]]; then
  echo "error: make publish-api only runs on master (currently on ${branch})" >&2
  echo "checkout master, pull, then retry" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is dirty; commit or stash before publishing" >&2
  exit 1
fi

if [[ ! -f docs/swagger.yaml ]]; then
  echo "error: docs/swagger.yaml missing — run make api and commit on master first" >&2
  exit 1
fi

git fetch origin master --tags --quiet

local_sha="$(git rev-parse HEAD)"
remote_sha="$(git rev-parse origin/master)"
if [[ "$local_sha" != "$remote_sha" ]]; then
  echo "error: local master ($(git rev-parse --short "$local_sha")) != origin/master ($(git rev-parse --short "$remote_sha"))" >&2
  echo "git pull --ff-only (or reset to origin/master) and retry" >&2
  exit 1
fi

day="$(date +%Y.%m.%d)"
base="openapi-v${day}"
tag="$base"
n=2
while git rev-parse -q --verify "refs/tags/${tag}" >/dev/null \
  || git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; do
  tag="${base}-${n}"
  n=$((n + 1))
done

echo "Publishing API contract as ${tag} (origin/master $(git rev-parse --short HEAD))"
git tag "$tag"
git push origin "refs/tags/${tag}"

echo
echo "Release workflow will attach:"
echo "  https://github.com/WilliamsStudentsOnline/wso-go/releases/download/${tag}/openapi.yaml"
