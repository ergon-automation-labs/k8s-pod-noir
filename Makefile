# POD noir — container-oriented workflows
#
# Typical flow: `make build` then `make run` (interactive REPL in Docker).
# Host binary: `make compile` → ./bin/podnoir (Linux ELF; run via Docker or WSL).
# Native (requires local Go): `make build-native`

COMPOSE ?= docker compose
IMAGE   ?= pod-noir:local
# Extra CLI args to podnoir, e.g. RUN_EXTRA='-scenario case-002-ghost-credential'
RUN_EXTRA ?=

.PHONY: help build run compile test smoke build-native clean ensure-bin-dir

help:
	@echo "make build        - Build runtime image (pod-noir:local)"
	@echo "make run          - File cabinet menu then REPL; ~/.kube mounted; POD_NOIR_KUBE rewrites localhost API"
	@echo "                 RUN_EXTRA='-scenario case-002-...' skips menu / picks case directly"
	@echo "make smoke        - podnoir doctor (cluster connectivity check in Docker)"
	@echo "make compile      - Produce ./bin/podnoir (Linux binary via Go container)"
	@echo "make test         - go test ./... inside Docker (compose service: test)"
	@echo "make build-native - go build ./cmd/podnoir (requires local Go)"
	@echo "See docker-compose.yml for kubeconfig / host.docker.internal notes."
	@echo "Integration env: POD_NOIR_EVENTS_ADAPTER, POD_NOIR_NATS_URL, POD_NOIR_NATS_BRIDGE, ..."

build:
	$(COMPOSE) build podnoir

run:
	$(COMPOSE) run --rm podnoir $(RUN_EXTRA)

compile: ensure-bin-dir
	$(COMPOSE) run --rm compile

test:
	@echo "Running go test in Docker..."
	$(COMPOSE) run --rm test

smoke:
	$(COMPOSE) run --rm podnoir doctor

ensure-bin-dir:
	@mkdir -p bin

# Fast iteration when Go is installed locally (darwin/arm64 etc.)
build-native: ensure-bin-dir
	CGO_ENABLED=0 go build -trimpath -o bin/podnoir ./cmd/podnoir

clean:
	rm -rf bin/
	$(COMPOSE) down --rmi local 2>/dev/null || true
