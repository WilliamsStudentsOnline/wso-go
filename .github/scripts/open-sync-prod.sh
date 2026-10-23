#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh (GitHub CLI) is not installed or not on PATH" >&2
  echo "install: https://cli.github.com/ and run: gh auth login" >&2
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "error: gh is installed but not authenticated for this host" >&2
  echo "run: gh auth login" >&2
  gh auth status >&2 || true
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "error: git is not installed or not on PATH" >&2
  exit 1
fi

# Status dump (same as make check-prod-integrity); do not fail the sync yet
integrity_rc=0
bash "${SCRIPT_DIR}/check-prod-integrity.sh" || integrity_rc=$?
echo

requested="${1:-}"
if [[ -n "$requested" ]]; then
  if ! git cat-file -e "${requested}^{commit}" 2>/dev/null; then
    echo "error: not a commit: ${requested}" >&2
    exit 1
  fi
  approved="$(git rev-parse "${requested}^{commit}")"
else
  approved="$(git rev-parse origin/master)"
fi

if ! git merge-base --is-ancestor "$approved" origin/master; then
  echo "error: ${approved} is not reachable from origin/master" >&2
  exit 1
fi

prod="$(git rev-parse origin/production)"
short_approved="$(git rev-parse --short "$approved")"
branch="prod-sha/${short_approved}"

if [[ "$prod" == "$approved" ]]; then
  echo "OK: production already at ${approved}; nothing to approve"
  exit 0
fi

if git merge-base --is-ancestor "$prod" "$approved"; then
  kind="forward"
elif git merge-base --is-ancestor "$approved" "$prod"; then
  kind="rollback"
else
  echo "error: production and approved SHAs have diverged" >&2
  echo "clean history with a one-time force of production to a master SHA, then retry" >&2
  exit 1
fi

title="[DEPLOY] Approve production SHA ${short_approved}"

body="$(cat <<EOF
> **Warning:** Merging this PR triggers a **live production deploy** that will affect all WSO users.

**Approved SHA:** \`${approved}\`
**Operation:** ${kind}

## How to merge
1. Wait for required checks
2. Get approval
3. Merge the PR once everything lights up green
4. Confirm the **Sync production to master** Actions run succeeds
5. Confirm the **Deploy to production** Actions run succeeds

Closes WilliamsStudentsOnline/wso-go#391
EOF
)"

git checkout -B "$branch" origin/master

tmp="$(mktemp)"
{
  cat <<'HDR'
# Unless prod is in the middle of merging/deploying, it should have THIS head SHA
# Update via `make sync-prod` (can pass in any SHA, for example if you need to roll back)
# Do not edit by hand unless you know what you are doing.
HDR
  echo "$approved"
} >"$tmp"
mv "$tmp" config/prod-sha

git add config/prod-sha
if git diff --cached --quiet; then
  echo "error: config/prod-sha already equals ${approved}; nothing to approve" >&2
  exit 1
fi

git commit -m "Approve production SHA ${short_approved}"
git push -u origin "HEAD:refs/heads/${branch}" --force-with-lease

existing_number="$(gh pr list --base master --head "$branch" --state open --json number --jq '.[0].number // empty')"
if [[ -n "$existing_number" ]]; then
  gh pr edit "$existing_number" --title "$title" --body "$body" --add-label prod-sync >/dev/null
  gh pr view "$existing_number" --json url --jq .url
  exit 0
fi

gh pr create --base master --head "$branch" --title "$title" --body "$body" --label prod-sync
