# Agent / contributor notes (pod-noir)

## Cursor rules

Project rules live under [`.cursor/rules/`](.cursor/rules/) (for example README sync, **progress log** sync). They apply in Cursor when this folder is the workspace.

## Progress log

- **[PROGRESS.md](PROGRESS.md)** — short index + snapshot (keep small).
- **`docs/progress/YYYY-MM.md`** — monthly bullets; add a new file when the month changes.

## Build & test (no local Go)

- **`make test`** — `go test ./...` inside Docker Compose (matches CI **test**).
- **`make lint-docker`** — gofmt + `go vet` + embedded YAML test in Docker (matches CI **lint**).
- **`make compile`** — Linux `bin/podnoir` via Compose.
- **`make run`** — interactive REPL; see [`README.md`](README.md) for kube-in-Docker.

## With local Go 1.23+

- **`make lint`** — same as **lint-docker** but uses host `go` / `gofmt`.
- **`make build-native`** — host-arch binary in `./bin/podnoir`.

## CI (GitHub Actions)

[`.github/workflows/ci.yml`](.github/workflows/ci.yml): **lint** → **test** → **integration** (kind + `podnoir doctor`). When changing CLI, solve policy, scenarios, Docker, or CI, update **`README.md`** (and **`PROGRESS.md` / current month under `docs/progress/`** when shipping meaningful work—see `.cursor/rules/progress-sync.mdc`).
