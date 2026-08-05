#!/usr/bin/env sh
set -eu
REPO_URL="${1:-https://github.com/appoloncel283-debug/pulsenet.git}"
git init
git branch -M main
git add .
git commit -m "Release PulseNet 2.0.0"
if git remote get-url origin >/dev/null 2>&1; then
  git remote set-url origin "$REPO_URL"
else
  git remote add origin "$REPO_URL"
fi
git push -u origin main
printf 'Published to %s\n' "$REPO_URL"
