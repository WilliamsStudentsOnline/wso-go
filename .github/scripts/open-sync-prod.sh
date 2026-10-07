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

behind="$(git rev-list --count origin/production..origin/master)"
ahead="$(git rev-list --count origin/master..origin/production)"
date_utc="$(date -u +%Y-%m-%d)"
title="[DEPLOY] Sync production to master (${date_utc})"

body="$(cat <<EOF
> **Warning:** Merging this PR triggers a **live production deploy** that will affect all wso-go users.

## What this does
STATUS: \`production\` is ${behind} commit(s) behind, ${ahead} commit(s) ahead of \`master\`

## How to merge
1. Wait for required checks
2. Get approval
3. **Rebase and merge** the PR once everything lights up green
4. Confirm the **Deploy to production** Actions run succeeds
EOF
)"

existing_number="$(gh pr list --base production --head master --state open --json number --jq '.[0].number // empty')"
if [[ -n "$existing_number" ]]; then
  gh pr edit "$existing_number" --title "$title" --body "$body" --add-label prod-sync >/dev/null
  gh pr view "$existing_number" --json url --jq .url
  exit 0
fi

gh pr create --base production --head master --title "$title" --body "$body" --label prod-sync
