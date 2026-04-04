#!/usr/bin/env bash
# Point this repo at tracked hooks under githooks/ (core.hooksPath).
# Run once per clone: ./scripts/setup-git-hooks.sh   or   make git-hooks
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
git config core.hooksPath githooks
echo "git config core.hooksPath=githooks  (this repository only)"
echo "Hooks run from: $ROOT/githooks/"
