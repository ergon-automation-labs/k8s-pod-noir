# Agent / contributor notes (pod-noir)

## Cursor rules

Project rules live under [`.cursor/rules/`](.cursor/rules/) (for example README sync, **progress log** sync). They apply in Cursor when this folder is the workspace.

## Progress log

- **[PROGRESS.md](PROGRESS.md)** — short index + snapshot (keep small).
- **`docs/progress/YYYY-MM.md`** — monthly bullets; add a new file when the month changes.

## Playtesting

- **`make playtest-smoke`** — scripts **[scripts/playtest-smoke.sh](scripts/playtest-smoke.sh)** (Compose `doctor` + optional host `kubectl`). Skip with **`SKIP_PLAYTEST_SMOKE=1`** or **`POD_NOIR_SKIP_PLAYTEST=1`**.
- **`make playtest-smoke-ci`** — Docker Compose service **`playtest-smoke-ci`**: build **`bin/podnoir`** in dev image, then **`scripts/playtest-smoke.sh`** with **`CI=true`** (matches integration job; no host Go).
- **Pre-commit:** **[`.pre-commit-config.yaml`](.pre-commit-config.yaml)** calls **[scripts/pre-commit-playtest.sh](scripts/pre-commit-playtest.sh)** — runs smoke when Docker + `kubectl` + cluster exist; otherwise **skip with message** (exit 0). **`SKIP_PLAYTEST_SMOKE=1`** silences.
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
