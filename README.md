# POD noir

Terminal-based **Kubernetes investigation scenarios** on a **real cluster**: noir framing, real `kubectl`, and a loop that rewards observing and stating a hypothesis before you fix things.

This repo is **[k8s-pod-noir](https://github.com/ergon-automation-labs/k8s-pod-noir)** under [ergon-automation-labs](https://github.com/ergon-automation-labs).

## What you need

- A working cluster and **`kubectl`** configured (context you intend to use).
- **Docker** (recommended) **or** Go **1.23+** for a local binary.

Cases create workloads in a dedicated namespace (default `pod-noir`) and clean it up on exit unless you opt out.

**Debrief** only closes a case after a **precinct health check** passes (usually `kubectl rollout status` on the main Deployment; case 006 checks **Service endpoints**). Fix the cluster first, then debrief.

**Teaching:** the first **`observe`** and first **`examine pod`** in a scenario can surface a one-time in-world **field note**. Use **`dossier`** in the REPL to read your local folder history (opened vs **cleared** counts in `~/.pod-noir/history.db`). The file cabinet menu shows the same dossier stamps when you return.

## Quick start (Docker)

```bash
make build
make smoke    # podnoir doctor — checks kube / cluster reachability inside the container
make run      # file cabinet → pick a case → REPL
```

Jump straight to a case:

```bash
make run RUN_EXTRA='-scenario case-001-overnight-shift'
```

**Local API servers (`127.0.0.1` / `localhost`):** Compose mounts `~/.kube` and sets `POD_NOIR_KUBE_IN_DOCKER` so the app can rewrite the API URL for `host.docker.internal`. If connection fails, point that context’s server URL at `https://host.docker.internal:6443` (or your gateway). See comments in `docker-compose.yml` and `make help`.

**Linux binary on the host:** `make compile` runs `go build` inside the dev Compose service and writes `./bin/podnoir` (ELF for the container’s architecture, e.g. `linux/arm64` on Apple Silicon). Run it in a Linux environment or use `make build` / `make run` for the packaged runtime image.

## Quick start (native Go)

```bash
make build-native   # ./bin/podnoir (host OS/arch; VERSION / GIT_COMMIT from git)
./bin/podnoir doctor
./bin/podnoir
```

## CLI reference

| Invocation | Meaning |
|------------|---------|
| `podnoir` | Interactive **case menu**, then session REPL. |
| `podnoir -scenario <id>` | Skip menu; run scenario `<id>`. |
| `podnoir doctor` | Print kube/docker hints and verify **cluster-info**. |
| `podnoir version` / `podnoir --version` | Build metadata (`VERSION` / `Commit` from `-ldflags` or `dev` / `none`). |

Flags:

- `-config` — optional config YAML path.
- `-data-dir` — SQLite state directory (default `~/.pod-noir`).
- `-skip-cleanup` — leave the scenario namespace in place when you exit.

Scenarios:

| ID | Theme |
|----|--------|
| `case-001-overnight-shift` | Bad rollout / missing config path |
| `case-002-ghost-credential` | Missing Secret / `secretKeyRef` |
| `case-003-dead-letter-harbour` | Bad image tag / pull failure |
| `case-004-wrong-number` | Liveness probe wrong port (nothing listens) |
| `case-005-thin-margin` | Memory limit / tmpfs OOM |
| `case-006-ghost-wire` | Service selector ≠ pod labels (empty endpoints) |
| `case-007-waiting-on-a-witness` | Failing **initContainer** blocks the app container |

## REPL shortcuts

In the session loop, **`help`** lists full commands. Shorthand in **normal** mode:

| Input | Expands to |
|-------|------------|
| `o`, `obs` | `observe` |
| `t <name>` | `trace <name>` |
| `x <name>` | `examine pod <name>` |
| `l <name>`, `logs <name>` | `check logs <name>` |
| `l`, `logs` (no name) | repeat last logs target (after a successful `check logs`) |
| `r`, `again` | repeat last expanded command (normal) or last successful `kubectl` (solve) |
| `hist`, `history` | last ~12 commands this session |

**Solve mode** runs real `kubectl` against the case namespace. The precinct blocks cluster-wide and high-risk operations (for example `-A` / `--all-namespaces`, namespace deletes, cluster roles/bindings, node operations, `kubectl taint`, `kubectl adm`, **`kubectl kustomize` / `-k`**). **Mutating commands that use `-f` / `--filename` must pass `-n <case-namespace>`** (or `--namespace=...`); referenced manifest files are **parsed** so `metadata.namespace` must match the case namespace, **cluster-scoped kinds** are rejected, and **`-f -` / URLs** are blocked. Entering solve after a HOT accusation shows **case desk hints** (kubectl angles for that scenario). Type **`exit`** to leave solve mode.

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

## Design & direction

Product intent, tone, and mechanics are spelled out in **[docs/pod-noir-northstar.md](docs/pod-noir-northstar.md)**.

## Development

```bash
make lint-docker    # same as CI **lint** job, inside Docker (no local Go required)
make lint           # gofmt + go vet + embedded YAML test (local Go 1.23+)
make test           # go test ./... in Docker (matches CI **test** job)
make manifests-lint # embedded scenario manifests must parse as YAML
make compile        # ./bin/podnoir via Compose (Linux ELF; see Docker note above)
make help
```

Contributor orientation: **[AGENTS.md](AGENTS.md)** (Cursor rules pointer, CI parity commands). **What shipped lately:** **[PROGRESS.md](PROGRESS.md)** (index) and **[docs/progress/](docs/progress/)** (monthly notes).

## CI (GitHub Actions)

Workflow **[`.github/workflows/ci.yml`](.github/workflows/ci.yml)** (push / PR to `main` or `master`, plus **workflow_dispatch**):

1. **lint** — `make lint` (format, `go vet`, embedded manifest YAML).
2. **test** — `make test` (unit tests inside Docker, same as before).
3. **integration** — [kind](https://kind.sigs.k8s.io/) with a **digest-pinned** `kindest/node` image (linux/amd64 on the runner), build `bin/podnoir` with version ldflags, `podnoir version`, then **`podnoir doctor`** against the ephemeral cluster.

Release images get version strings from Docker **`VERSION` / `COMMIT`** build args (Compose passes `VERSION` and `GIT_COMMIT` from `make` / your shell).
