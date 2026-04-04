#!/usr/bin/env bash
# Quick cluster + binary smoke (see docs/playtest-checklist.md).
# Local: docker compose doctor + host kubectl.
# CI:    ./bin/podnoir doctor + kubectl (set CI=true). Run on host (GitHub) or inside dev image (make playtest-smoke-ci).
# Skip:  POD_NOIR_SKIP_PLAYTEST=1 or SKIP_PLAYTEST_SMOKE=1
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ "${POD_NOIR_SKIP_PLAYTEST:-}" == "1" ]] || [[ "${SKIP_PLAYTEST_SMOKE:-}" == "1" ]]; then
  echo "== playtest-smoke: skipped (POD_NOIR_SKIP_PLAYTEST or SKIP_PLAYTEST_SMOKE)"
  exit 0
fi

COMPOSE="${COMPOSE:-docker compose}"

# --- CI / kind: binary already on disk, no Compose image required ---
if [[ "${CI:-}" == "true" ]] || [[ "${PLAYTEST_MODE:-}" == "ci" ]]; then
  echo "== playtest-smoke (CI) — doctor + kubectl report =="
  if [[ ! -x ./bin/podnoir ]]; then
    echo "error: ./bin/podnoir missing — build it before playtest-smoke in CI"
    exit 1
  fi
  ./bin/podnoir doctor
  echo ""
  kubectl cluster-info --request-timeout=20s
  echo ""
  echo "== namespaces (sample) =="
  kubectl get ns -o name 2>/dev/null | head -15 || true
  echo ""
  echo "== playtest-smoke (CI): OK =="
  exit 0
fi

# --- Local developer: Compose doctor + optional kubectl ---
echo "== [1/3] podnoir doctor (API reachability from app container) =="
$COMPOSE run --rm podnoir doctor

echo ""
echo "== [2/3] kubectl on host (optional; same kubeconfig as Compose mount) =="
if command -v kubectl >/dev/null 2>&1; then
  kubectl cluster-info --request-timeout=15s
  echo ""
  if kubectl get ns pod-noir >/dev/null 2>&1; then
    echo "namespace pod-noir: present"
    kubectl get pods -n pod-noir -o wide 2>/dev/null | head -20 || true
  else
    echo "namespace pod-noir: not present yet (expected before you open a case)"
  fi
else
  echo "kubectl not on PATH — skipped. Use doctor output above, or install kubectl for host checks."
fi

echo ""
echo "== [3/3] Next steps =="
echo "    Run automated tests: make test"
echo "    Manual gameplay sample: docs/playtest-checklist.md"
echo "    Interactive: make run RUN_EXTRA='-scenario case-001-overnight-shift'"
