#!/usr/bin/env bash
set -euo pipefail

existing="$(gh pr list --base production --head master --state open --json url --jq '.[0].url // empty')"
if [[ -n "$existing" ]]; then
  echo "$existing"
  exit 0
fi

behind="$(git rev-list --count origin/production..origin/master 2>/dev/null || echo '?')"
ahead="$(git rev-list --count origin/master..origin/production 2>/dev/null || echo '?')"

gh pr create --base production --head master \
  --title "Sync production to master" \
  --body "$(cat <<EOF
Fast-forward \`production\` to current \`master\` (${behind} commits behind, ${ahead} ahead on production).

After rebase-and-merge, the production deploy workflow should run on the prod runner.

Do not merge until ready for a live prod bounce.
EOF
)"
