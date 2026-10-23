#!/usr/bin/env bash
set -euo pipefail

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

git fetch origin master production

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
date_utc="$(date -u +%Y-%m-%d)"
branch="prod-sha/${short_approved}"

if git merge-base --is-ancestor "$prod" "$approved"; then
  if [[ "$prod" == "$approved" ]]; then
    kind="no-op (already on production)"
  else
    kind="forward"
  fi
elif git merge-base --is-ancestor "$approved" "$prod"; then
  kind="rollback"
else
  kind="diverged (Action will refuse until production history is cleaned)"
fi

title="[DEPLOY] Approve production SHA ${short_approved} (${date_utc})"

body="$(cat <<EOF
> **Warning:** Merging this PR approves a **live production deploy** of \`${approved}\`.

**Approved SHA:** \`${approved}\`
**Current production:** \`${prod}\`
**Kind:** ${kind}

## How this works
1. This PR only updates \`config/prod-sha\`
2. Get review + required checks
3. **Merge** into \`master\` (do not target \`production\` directly)
4. The prod-sha-sync Action pushes that exact SHA to \`production\`
5. Confirm the **Deploy to production** Actions run succeeds

Closes WilliamsStudentsOnline/wso-go#391
EOF
)"

# upsert branch with file change from current master
git checkout -B "$branch" origin/master

tmp="$(mktemp)"
{
  cat <<'HDR'
# Approved production deploy SHA for this repo.
# Updated by `make sync-prod` (optional SHA arg) via a PR into master.
# When that PR merges, Actions pushes this exact SHA to the production
# branch (force-with-lease). Not "master tip at merge time."
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
