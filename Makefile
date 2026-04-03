# POD noir — container-oriented workflows
#
# Typical flow: `make build` then `make run` (interactive REPL in Docker).
# Host binary: `make compile` → ./bin/podnoir (Linux ELF; run via Docker or WSL).
# Native (requires local Go): `make build-native`

COMPOSE ?= docker compose
IMAGE   ?= pod-noir:local
# Extra CLI args to podnoir, e.g. RUN_EXTRA='-scenario case-002-ghost-credential'
RUN_EXTRA ?=
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)

.PHONY: help build run compile test smoke build-native clean ensure-bin-dir manifests-lint lint

help:
	@echo "make build        - Build runtime image (pod-noir:local)"
	@echo "make test         - CI parity: go test ./... via Docker Compose"
	@echo "make run          - File cabinet menu then REPL; ~/.kube mounted; kube-in-Docker env rewrites API host"
	@echo "                 RUN_EXTRA='-scenario case-002-...' skips menu / picks case directly"
	@echo "make smoke        - podnoir doctor (cluster connectivity check in Docker)"
	@echo "make compile      - Produce ./bin/podnoir (Linux binary via Go container)"
	@echo "make build-native - go build ./cmd/podnoir (requires local Go)"
	@echo "make manifests-lint - validate embedded scenario YAML parses"
	@echo "make lint         - gofmt + go vet + manifests-lint (needs local Go; CI runs this first)"
	@echo "See docker-compose.yml for kubeconfig / host.docker.internal notes."
	@echo "Integration env: POD_NOIR_EVENTS_ADAPTER, POD_NOIR_NATS_URL, POD_NOIR_NATS_BRIDGE, ..."

build:
	VERSION=$(VERSION) GIT_COMMIT=$(GIT_COMMIT) $(COMPOSE) build podnoir

run:
	$(COMPOSE) run --rm podnoir $(RUN_EXTRA)

compile: ensure-bin-dir
	VERSION=$(VERSION) GIT_COMMIT=$(GIT_COMMIT) $(COMPOSE) run --rm compile

test:
	@echo "Running go test in Docker..."
	$(COMPOSE) run --rm test

smoke:
	$(COMPOSE) run --rm podnoir doctor

ensure-bin-dir:
	@mkdir -p bin

# Fast iteration when Go is installed locally (darwin/arm64 etc.)
build-native: ensure-bin-dir
	CGO_ENABLED=0 go build -trimpath \
		-ldflags="-s -w -X podnoir/internal/version.Version=$(VERSION) -X podnoir/internal/version.Commit=$(GIT_COMMIT)" \
		-o bin/podnoir ./cmd/podnoir

manifests-lint:
	go test -count=1 ./internal/scenario -run TestEmbeddedManifestsAreValidYAML

lint:
	@test -z "$$(gofmt -l cmd internal)" || (echo >&2 'gofmt needed — run: gofmt -w cmd internal' && gofmt -l cmd internal && exit 1)
	go vet ./...
	$(MAKE) manifests-lint

clean:
	rm -rf bin/
	$(COMPOSE) down --rmi local 2>/dev/null || true
