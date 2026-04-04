# Playtest checklist

Use this for **spot-checking gameplay** without walking all scenarios. Pair with **`make playtest-smoke`** (cluster reachability) and **`make test`** (automated). CI and **`make playtest-smoke-ci`** use the same **`doctor` + kubectl report** (**`CI=true`** in the script; local parity runs **in Docker** via **`docker-compose.yml`**).

## Before you start

- [ ] Cluster reachable: `make playtest-smoke` passes **`podnoir doctor`**, or **`make playtest-smoke-ci`** (build + smoke in Docker, no host Go).
- [ ] Optional: **`make git-hooks`** — **`core.hooksPath`** → **`githooks/`** (**pre-commit** + **pre-push**); hooks print **skip** reasons when Docker/kube/cluster are missing; **`SKIP_PLAYTEST_SMOKE=1`** skips **both**; **`git push --no-verify`** bypasses **pre-push** only.
- [ ] Optional: `make test` and `make lint-docker` green after your changes.
- [ ] Mock LLM is fine for flow testing: default `POD_NOIR_LLM_PROVIDER=mock` (or unset).

## Representative scenarios (pick 3–4 per release)

| Case ID | Stress | Victory |
|---------|--------|---------|
| `case-001-overnight-shift` | Rollout / undo | rollout |
| `case-002-ghost-credential` | Secret / env | rollout |
| `case-006-ghost-wire` | Service / endpoints | **endpoints** (different code path) |
| `case-008-the-red-tape-room` | ResourceQuota | rollout |
| `case-009-evidence-locker-blues` | PVC / StorageClass | rollout |
| `case-010-the-silent-corridor` | NetworkPolicy egress | rollout |

You don’t need every row every time—rotate **one “wiring” case (006)** with **one policy case (008–010)** and **one classic workload case (001–005)**.

## Minimal loop (per scenario, ~10–15 min)

Use: `make run RUN_EXTRA='-scenario <case-id>'` (skips file cabinet).

1. [ ] **Briefing** — read the box; note namespace `pod-noir`.
2. [ ] **observe** — pods/events load; field note if shown (first time).
3. [ ] **examine pod** / **check logs** / **trace** — at least one deep dive.
4. [ ] **accuse** — get a **HOT** once (adjust wording if needed); confirm solve unlocks.
5. [ ] **solve** — one mutating `kubectl` that fits precinct (`-n pod-noir`, no `-A` / `-k` tricks).
6. [ ] Fix the cluster until **observe** looks healthy.
7. [ ] **debrief** — precinct victory check passes; mock debrief prints.
8. [ ] **quit** — namespace cleanup (unless `-skip-cleanup`).

## Quick kubectl-only triage (no REPL)

After opening a case (namespace exists), you can validate **broken state** without dialogue:

```bash
kubectl get pods,svc,quota,pvc,networkpolicy -n pod-noir
kubectl describe pod -n pod-noir -l app=<workload>
```

## Optional: real LLM

After mock flow feels solid, set `POD_NOIR_LLM_PROVIDER` and API vars (see README), repeat steps 4–7 on **one** scenario and check tone + errors.

## What full coverage means

All **10** scenarios should be walked **before a major release** or when changing **solve policy**, **victory checks**, or **mock HOT rules**. Routine PRs can rely on **automated tests + this checklist + smoke**.
