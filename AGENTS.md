# Agent / contributor notes (pod-noir)

## Setup from scratch

**[docs/setup.md](docs/setup.md)** — install Docker, `kubectl`, local cluster (kind / alternatives), kubeconfig for Docker, then `make build` / `make run`.

## Cursor rules

Project rules live under [`.cursor/rules/`](.cursor/rules/): **[readme-sync](.cursor/rules/readme-sync.mdc)** (player-facing README), **[progress-sync](.cursor/rules/progress-sync.mdc)** (PROGRESS + monthly log), **[northstar-sync](.cursor/rules/northstar-sync.mdc)** (constitution **[docs/pod-noir-northstar.md](docs/pod-noir-northstar.md)** + **[docs/architecture-decisions.md](docs/architecture-decisions.md)**). They apply in Cursor when this folder is the workspace.

## Progress log

- **[PROGRESS.md](PROGRESS.md)** — short index + snapshot (keep small).
- **`docs/progress/YYYY-MM.md`** — monthly bullets; add a new file when the month changes.

## Contacts (NPCs)

Four **wire-room** personas are implemented: **Senior Detective** (`hint senior`; roster via bare **`hint`**), **Sysadmin**, **Network Engineer**, **Archivist**. Unlock rules, **per-scenario wire copy** (`wire_messages.go`), and **`WireRoster`** live under **`internal/contacts/`**; session routing in **`internal/session/session.go`**. **LLM wire contacts:** HTTP providers implement **`llm.ContactWirer`** — **`POD_NOIR_LLM_CONTACT_WIRE`** (default on); static fallback on mock/disabled/error (see **`internal/llm/contact_wire.go`**).

## Playtesting

- **`make playtest-smoke`** — scripts **[scripts/playtest-smoke.sh](scripts/playtest-smoke.sh)** (Compose `doctor` + optional host `kubectl`). Skip with **`SKIP_PLAYTEST_SMOKE=1`** or **`POD_NOIR_SKIP_PLAYTEST=1`**.
- **`make playtest-smoke-ci`** — Docker Compose service **`playtest-smoke-ci`**: build **`bin/podnoir`** in dev image, then **`scripts/playtest-smoke.sh`** with **`CI=true`** (matches integration job; no host Go).
- **Git hooks:** Tracked under **[`githooks/`](githooks/)** (committed). **`make git-hooks`** or **[scripts/setup-git-hooks.sh](scripts/setup-git-hooks.sh)** sets **`core.hooksPath=githooks`**. **[`githooks/pre-commit`](githooks/pre-commit)** and **[`githooks/pre-push`](githooks/pre-push)** → **[scripts/pre-commit-playtest.sh](scripts/pre-commit-playtest.sh)** — same smoke on commit and on **push** (push catches **`--no-verify`** / skip-on-commit). **`SKIP_PLAYTEST_SMOKE=1`** silences both; **`git push --no-verify`** bypasses only push. Optional **[`.pre-commit-config.yaml`](.pre-commit-config.yaml)** is for **`pre-commit run`** only (whitespace hooks), not for installing these hooks (`pre-commit install` writes `.git/hooks`, ignored when **`hooksPath`** is set).
- **[docs/playtest-checklist.md](docs/playtest-checklist.md)** — recommended scenario subset and minimal REPL loop.

## Build & test (no local Go)

- **`make test`** — `go test ./...` inside Docker Compose (matches CI **test**).
- **`make lint-docker`** — gofmt + `go vet` + embedded YAML test in Docker (matches CI **lint**).
- **`make compile`** — Linux `bin/podnoir` via Compose.
- **`make run`** — interactive REPL; see [`README.md`](README.md) for kube-in-Docker.

## With local Go 1.23+

- **`make lint`** — same as **lint-docker** but uses host `go` / `gofmt`.
- **`make build-native`** — host-arch binary in `./bin/podnoir`.

## CI (GitHub Actions)

[`.github/workflows/ci.yml`](.github/workflows/ci.yml): **lint** → **test** → **integration** (kind + **`playtest-smoke.sh`** with **`CI=true`** — includes **`podnoir doctor`** + `kubectl`). When changing CLI, solve policy, scenarios, Docker, or CI, update **`README.md`** (and **`PROGRESS.md` / current month under `docs/progress/`** when shipping meaningful work—see `.cursor/rules/progress-sync.mdc`).
