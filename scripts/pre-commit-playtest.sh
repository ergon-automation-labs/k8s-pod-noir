#!/usr/bin/env bash
# Invoked from githooks/pre-commit and githooks/pre-push (POD_NOIR_GIT_HOOK).
# Runs playtest-smoke when Docker + kubectl + a reachable cluster exist.
# Otherwise skip with a clear message (does not block offline work).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
HOOK="${POD_NOIR_GIT_HOOK:-git-hook}"

if [[ "${SKIP_PLAYTEST_SMOKE:-}" == "1" ]] || [[ "${POD_NOIR_SKIP_PLAYTEST:-}" == "1" ]]; then
  echo "${HOOK} playtest-smoke: skipped (SKIP_PLAYTEST_SMOKE / POD_NOIR_SKIP_PLAYTEST)"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "${HOOK} playtest-smoke: skipped — docker not on PATH"
  exit 0
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "${HOOK} playtest-smoke: skipped — kubectl not on PATH"
  exit 0
fi

if ! kubectl cluster-info --request-timeout=8s >/dev/null 2>&1; then
  echo "${HOOK} playtest-smoke: skipped — no reachable cluster (kubectl cluster-info failed)."
  echo "  Tip: run against kind/minikube, or export SKIP_PLAYTEST_SMOKE=1 to silence."
  exit 0
fi

echo "${HOOK} playtest-smoke: cluster reachable — running smoke (see output below)"
exec "$ROOT/scripts/playtest-smoke.sh"
