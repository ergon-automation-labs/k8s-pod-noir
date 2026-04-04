#!/usr/bin/env bash
# Quick cluster + binary smoke before a manual playtest (see docs/playtest-checklist.md).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE="${COMPOSE:-docker compose}"

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
