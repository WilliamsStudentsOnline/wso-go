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
title="[PROD-SYNC] Sync production to master (${date_utc})"

body="$(cat <<EOF
> **Warning:** Merging this PR (rebase and merge only) updates \`production\` and triggers a **live production deploy** on the prod self-hosted runner. Do not merge until you intend to bounce prod.

## What this does
- Brings \`production\` up to current \`master\` (\`${behind}\` commits on master not in production; \`${ahead}\` commits on production not in master)
- After rebase-and-merge, the Deploy workflow runs on the \`prod\` runner

## How to merge
1. Wait for required checks (full CI on the prod runner)
2. Get approval (\`@stamp-buddy\` or a reviewer)
3. **Rebase and merge** only (keeps linear history / FF-style tip when production is behind master)
4. Confirm the **Deploy to production** Actions run succeeds

## If production is ahead of master
Non-zero "ahead" means tips will not match after merge; rebase-and-merge still applies master on top of production. Investigate unexpected production-only commits before merging.
EOF
)"

existing_number="$(gh pr list --base production --head master --state open --json number --jq '.[0].number // empty')"
if [[ -n "$existing_number" ]]; then
  gh pr edit "$existing_number" --title "$title" --body "$body" --add-label prod-sync >/dev/null
  gh pr view "$existing_number" --json url --jq .url
  exit 0
fi

gh pr create --base production --head master --title "$title" --body "$body" --label prod-sync
