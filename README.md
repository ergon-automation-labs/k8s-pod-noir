# POD noir

The cluster doesn’t lie—it just doesn’t volunteer. **POD noir** is a terminal **detective game** played on a **real Kubernetes cluster**: you get a case file, a wire-room REPL, and actual `kubectl`. No toy sandbox—the weirdness in `describe` and `events` is the same weirdness you’d see when something fails for real.

This repo is **[k8s-pod-noir](https://github.com/ergon-automation-labs/k8s-pod-noir)** under [ergon-automation-labs](https://github.com/ergon-automation-labs).

## The world you’re stepping into

You’re not “doing labs.” You’re working the **Cluster Agency** floor: manila folders on the cabinet, rain in the margins of the briefing, a **precinct** that won’t stamp **CLOSED** until the workload stops bleeding. Each **case** is a scenario manifest applied into a dedicated namespace—usually `pod-noir`—with a story that sounds like noir because **fiction gives the debugging stakes a spine**. Underneath it is the same discipline good on-calls use: **look before you patch**, say what you think out loud (**accuse**), then earn **solve mode** and fix the cluster like an adult.

The game rewards **observe → hypothesis → fix**. Skip the theory and you’re just typing; nail a **HOT** accusation and the desk unlocks **raw kubectl**—within precinct rules, because the job is to learn, not to melt the cluster.

## What you need

- A working cluster and **`kubectl`** pointed at the context you intend to use.
- **Docker** (recommended) **or** Go **1.23+** if you build a binary yourself.

Opening a case applies workloads into a namespace and tears them down on exit unless you pass **`-skip-cleanup`**.

**Debrief** is the stamp at the end: it only runs after a **precinct health check**—typically `kubectl rollout status` on the main Deployment; one case checks **Service endpoints** instead. Heal the cluster first; then close the file.

**Field notes & dossier:** the first **`observe`** and first **`examine pod`** in a scenario can surface a one-time teaching beat in-world. **`dossier`** reads your local history (`~/.pod-noir/history.db`—opened vs **cleared** folders). The file cabinet shows the same stamps when you come back.

## Quick start — get on the wire

```bash
make build
make smoke    # podnoir doctor — proves the container can talk API
make run      # file cabinet → pick a case → REPL
```

Jump a case directly:

```bash
make run RUN_EXTRA='-scenario case-001-overnight-shift'
```

**Local API (`127.0.0.1` / `localhost`):** Compose mounts `~/.kube` and can rewrite the server URL for `host.docker.internal`. If the phone doesn’t ring, point that context at `https://host.docker.internal:6443` (or your gateway). Details in `docker-compose.yml` and `make help`.

**Linux ELF on disk:** `make compile` writes `./bin/podnoir` via the dev image (architecture matches the container—e.g. `linux/arm64` on Apple Silicon). Run it where that ELF runs, or stay in the runtime image with **`make run`**.

## Quick start (native Go)

```bash
make build-native   # ./bin/podnoir (host OS/arch; VERSION / GIT_COMMIT from git)
./bin/podnoir doctor
./bin/podnoir
```

## Commands

| Invocation | Meaning |
|------------|---------|
| `podnoir` | **File cabinet**, then the session REPL. |
| `podnoir -scenario <id>` | Skip the drawer; open that case id. |
| `podnoir doctor` | Kube/docker hints and **`cluster-info`**. |
| `podnoir version` / `podnoir --version` | Build metadata (`VERSION` / `Commit` from `-ldflags`, or `dev` / `none`). |

Flags:

- `-config` — optional config YAML path.
- `-data-dir` — SQLite state (default `~/.pod-noir`).
- `-skip-cleanup` — leave the scenario namespace when you exit.

## Cases on the desk

| ID | What’s wrong (the short version) |
|----|-----------------------------------|
| `case-001-overnight-shift` | Bad rollout / missing config path |
| `case-002-ghost-credential` | Missing Secret / `secretKeyRef` |
| `case-003-dead-letter-harbour` | Bad image tag / pull failure |
| `case-004-wrong-number` | Liveness probe wrong port (nothing listens) |
| `case-005-thin-margin` | Memory limit / tmpfs OOM |
| `case-006-ghost-wire` | Service selector ≠ pod labels (empty endpoints) |
| `case-007-waiting-on-a-witness` | Failing **initContainer** blocks the app |
| `case-008-the-red-tape-room` | **ResourceQuota** vs Pod **requests** |
| `case-009-evidence-locker-blues` | **PVC** Pending / bad **StorageClass** |
| `case-010-the-silent-corridor` | **NetworkPolicy** egress blocks what the pod needs |

## In the session (REPL)

Type **`help`** for the full command list. Shortcuts in **normal** mode:

| Input | Expands to |
|-------|------------|
| `o`, `obs` | `observe` |
| `t <name>` | `trace <name>` |
| `x <name>` | `examine pod <name>` |
| `l <name>`, `logs <name>` | `check logs <name>` |
| `l`, `logs` (no name) | repeat last logs target (after a successful `check logs`) |
| `r`, `again` | repeat last expanded command (normal) or last successful `kubectl` (solve) |
| `hist`, `history` | last ~12 commands this session |

**Solve mode** is the back alley: real `kubectl` against the case namespace, after a **HOT** accusation. The **precinct** blocks cluster-wide damage (`-A`, namespace deletes, cluster-admin plays, node ops, `taint`, `adm`, **`kustomize` / `-k`**). Mutations with **`-f` / `--filename`** must carry **`-n`** (or `--namespace=...`); manifests are **checked** so namespaces and cluster-scoped kinds can’t smuggle past the story. **`-f -`** and URLs are blocked. **Case desk hints** print when you enter solve. Type **`exit`** to leave solve mode.

## Environment (`POD_NOIR_*`)

| Variable | Role |
|----------|------|
| `POD_NOIR_EVENTS_ADAPTER` | `stdout` (default), `nats`, `both`, `none`. |
| `POD_NOIR_NATS_URL` | NATS server URL when using NATS. |
| `POD_NOIR_NATS_SUBJECT_PREFIX` | Subject prefix (default `pod-noir`). |
| `POD_NOIR_NATS_BRIDGE` | Publish bridged envelope (`true` / `false`). |
| `POD_NOIR_LLM_PROVIDER` | `mock` (default), `anthropic`, `openai`, `ollama`. |
| `POD_NOIR_LLM_API_KEY`, `POD_NOIR_LLM_MODEL`, `POD_NOIR_LLM_BASE_URL` | Provider credentials / model / base URL. |
| `POD_NOIR_LLM_FALLBACK_MOCK` | Use mock when HTTP LLM errors (default on unless `false`). |
| `POD_NOIR_LLM_REPAIR` | After a bad JSON accusation response, one automatic retry (default on unless `false`). |
| `POD_NOIR_KUBE_IN_DOCKER`, `POD_NOIR_KUBE_API_HOST`, `POD_NOIR_KUBE_TLS_INSECURE` | Docker ↔ host API rewriting and TLS (see `doctor` output). |

## Why it exists (design)

Tone, learning loop, and non-goals live in **[docs/pod-noir-northstar.md](docs/pod-noir-northstar.md)**—the constitution for the world this README only introduces.

## Playtesting

- **`make playtest-smoke`** — `podnoir doctor` plus optional host **`kubectl`** checks (Compose locally; see script for skip env vars).
- **`make playtest-smoke-ci`** — same **`doctor` + kubectl report** as the **integration** job, **inside Docker** (builds `./bin/podnoir` in the dev image — no host **Go** required).
- **[docs/playtest-checklist.md](docs/playtest-checklist.md)** — a **small scenario matrix** and a tight loop so you don’t have to play every folder before a release.

**Pre-commit (optional):** `pip install pre-commit && pre-commit install` runs **[`.pre-commit-config.yaml`](.pre-commit-config.yaml)** — including **`scripts/pre-commit-playtest.sh`**, which runs **`playtest-smoke`** when **Docker**, **`kubectl`**, and a **reachable cluster** are present; otherwise it **prints a skip reason** and exits **0** (so commits are not blocked offline). Set **`SKIP_PLAYTEST_SMOKE=1`** or **`POD_NOIR_SKIP_PLAYTEST=1`** to silence the hook.

## Development

```bash
make lint-docker    # same as CI **lint** job, inside Docker (no local Go required)
make lint           # gofmt + go vet + embedded YAML test (local Go 1.23+)
make test           # go test ./... in Docker (matches CI **test** job)
make manifests-lint # embedded scenario manifests must parse as YAML
make compile        # ./bin/podnoir via Compose (Linux ELF; see above)
make playtest-smoke    # doctor + optional kubectl (see Playtesting)
make playtest-smoke-ci # integration parity: build + smoke in Docker (no host Go)
make help
```

Contributors: **[AGENTS.md](AGENTS.md)**. **What shipped:** **[PROGRESS.md](PROGRESS.md)** and **[docs/progress/](docs/progress/)**.

## CI (GitHub Actions)

Workflow **[`.github/workflows/ci.yml`](.github/workflows/ci.yml)** (push / PR to `main` or `master`, plus **workflow_dispatch**):

1. **lint** — `make lint` (format, `go vet`, embedded manifest YAML).
2. **test** — `make test` (unit tests inside Docker).
3. **integration** — [kind](https://kind.sigs.k8s.io/) with a **digest-pinned** `kindest/node` image (linux/amd64 on the runner), build `bin/podnoir`, `podnoir version`, then **`scripts/playtest-smoke.sh`** with **`CI=true`** (**`podnoir doctor`** + **`kubectl` report** — same path as **`make playtest-smoke-ci`**).

Release images take **`VERSION` / `COMMIT`** from Docker build args (Compose passes `VERSION` and `GIT_COMMIT` from `make` / your shell).
