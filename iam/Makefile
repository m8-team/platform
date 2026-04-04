BUF ?= buf
GO ?= go
DOCKER_COMPOSE ?= docker compose -f deploy/local/docker-compose.yaml

.PHONY: buf-lint buf-generate test run worker migrate schema-sync env-up env-down env-logs env-ps run-local worker-local

buf-lint:
	$(BUF) lint

buf-generate:
	$(BUF) generate

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/iamd

worker:
	$(GO) run ./cmd/worker

run-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/iamd

worker-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/worker

env-up:
	$(DOCKER_COMPOSE) up -d
	./scripts/wait-local-env.sh

env-down:
	$(DOCKER_COMPOSE) down --remove-orphans

env-logs:
	$(DOCKER_COMPOSE) logs -f

env-ps:
	$(DOCKER_COMPOSE) ps

migrate:
	$(GO) run ./cmd/migrator

schema-sync:
	$(GO) run ./cmd/schema-sync
