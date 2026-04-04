# Playtest checklist

Use this for **spot-checking gameplay** without walking all scenarios. Pair with **`make playtest-smoke`** (cluster reachability) and **`make test`** (automated). CI and **`make playtest-smoke-ci`** use the same **`doctor` + kubectl report** (**`CI=true`** in the script; local parity runs **in Docker** via **`docker-compose.yml`**).

## Suggested first run (~10–15 min)

- **Case:** **`case-001-overnight-shift`** — clearest mock **HOT** path, rollout undo / patch story, field notes on **`observe`** / **`examine`**. Good default when someone asks “which folder should I open first?”
- **Command:** `make run RUN_EXTRA='-scenario case-001-overnight-shift'` (after **`make smoke`** looks good).
- **Goal:** complete the minimal loop below once before rotating harder cases (policy / NetworkPolicy / PVC rows).

### Mock `accuse` on case-001 (why “warm” happens)

With the default **mock** LLM (`POD_NOIR_LLM_PROVIDER=mock` or unset), **`accuse` is scored against keyword lists**, not open-ended “grading.” For **case-001** specifically see **`internal/llm/mock.go`** (`HotHints` for `Case001` and `accusationHot`):

- **HOT** if your theory **includes `settings.json`**, *or* your lowercased text matches **two or more** of the case **HotHints** substrings (`settings.json`, `config`, `entrypoint`, `/app/config`, `missing file`, `start.sh`). Each distinct hint phrase counts once.
- **Warm** if you match **exactly one** HotHint (e.g. only “missing file”) — the reply is nudging you to name the mechanism more tightly.
- **Cold** if you’re not matching those cues yet (or you’re only hitting **WarmHints** like “crash” without hot cues — different branch).

So a line like **“missing file”** alone is *supposed* to read as warm: you’re in the right neighborhood; add **`settings.json`** and/or a second cue (e.g. **`entrypoint`** + **`config`**) to earn **HOT** and unlock **`solve`**. With a **real** HTTP LLM, tone differs but you still want a concrete theory aligned with evidence.

## If you’re stuck (too hard, thin clues, or rusty kubectl)

The game is meant to be **harder than a tutorial** but not a dead end.

1. **Change folders** — **`quit`** (cleanup) and open **`case-001-overnight-shift`** or another **001–005** row before **008–010** (quota / PVC / NetworkPolicy need more pattern recognition). See the scenario table above.
2. **`help`** — full REPL command list; **`status`** — what you’ve logged this session (notes, accused/hot/solve flags, contact flags).
3. **`hint`** — wire **roster**: who’s locked, what behavior unlocks them (**Senior** after logs+trace or a non-HOT **accuse**; **Sysadmin** after **examine pod**; **Network** after **trace**; **Archivist** after **dossier**). Then **`hint senior`**, **`hint sysadmin`**, etc., when open — **one** message per contact per case.
4. **Treat `accuse` as a dial, not a verdict** — **cold** / **warm** tell you to gather different evidence or tighten wording; iterate. **HOT** unlocks **solve**. On **mock**, each case uses **HotHints** substring scoring (see **“Mock accuse on case-001”** above); a real LLM is looser but still expects evidence-shaped theories.
5. **After HOT, enter `solve`** — the desk prints **SolveHints** (precinct-safe angles for that case).
6. **First-time field notes** — the first **`observe`** / **`examine pod`** in a scenario can surface a short in-world teaching beat.
7. **If kubectl itself feels opaque** — POD noir assumes you can run basic **`kubectl`** (see **[docs/setup.md](setup.md)** and **[docs/pod-noir-northstar.md](pod-noir-northstar.md)** § audience). A short refresher on pods, events, and **`describe`** elsewhere is fair prep; the REPL is not a substitute for that baseline.
8. **Debrief after the cluster is healthy** — the mock debrief’s **“what to study”** block points at concepts to revisit even when you struggled mid-case.

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
4. [ ] **accuse** — get a **HOT** once (adjust wording if needed); confirm solve unlocks. Optional: **`hint`** (roster), **`hint senior`** / other contacts after their unlocks.
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
