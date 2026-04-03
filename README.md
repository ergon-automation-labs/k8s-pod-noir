# POD noir

Terminal-based **Kubernetes investigation scenarios** on a **real cluster**: noir framing, real `kubectl`, and a loop that rewards observing and stating a hypothesis before you fix things.

This repo is **[k8s-pod-noir](https://github.com/ergon-automation-labs/k8s-pod-noir)** under [ergon-automation-labs](https://github.com/ergon-automation-labs).

## What you need

- A working cluster and **`kubectl`** configured (context you intend to use).
- **Docker** (recommended) **or** Go **1.23+** for a local binary.

Cases create workloads in a dedicated namespace (default `pod-noir`) and clean it up on exit unless you opt out.

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

## Quick start (native Go)

```bash
make build-native   # ./bin/podnoir
./bin/podnoir doctor
./bin/podnoir
```

## CLI reference

| Invocation | Meaning |
|------------|---------|
| `podnoir` | Interactive **case menu**, then session REPL. |
| `podnoir -scenario <id>` | Skip menu; run scenario `<id>`. |
| `podnoir doctor` | Print kube/docker hints and verify **cluster-info**. |

Flags:

- `-config` — optional config YAML path.
- `-data-dir` — SQLite state directory (default `~/.pod-noir`).
- `-skip-cleanup` — leave the scenario namespace in place when you exit.

Scenarios (examples):

- `case-001-overnight-shift`
- `case-002-ghost-credential`
- `case-003-dead-letter-harbour`

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
| `POD_NOIR_KUBE_IN_DOCKER`, `POD_NOIR_KUBE_API_HOST`, `POD_NOIR_KUBE_TLS_INSECURE` | Docker ↔ host API rewriting and TLS (see `doctor` output). |

## Design & direction

Product intent, tone, and mechanics are spelled out in **[docs/pod-noir-northstar.md](docs/pod-noir-northstar.md)**.

## Development

```bash
make test   # go test ./... in Docker
make help
```
