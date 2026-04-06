BUF ?= buf
GO ?= go
DOCKER_COMPOSE ?= docker compose -f deploy/local/docker-compose.yaml

.PHONY: buf-lint buf-generate test run worker migrate seed schema-sync env-up env-down env-logs env-ps run-local worker-local migrate-local seed-local schema-sync-local check-local-ports stop-local local-up local-down local-status

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

check-local-ports:
	./scripts/check-local-ports.sh

stop-local:
	./scripts/stop-local.sh

local-up:
	./scripts/start-local-all.sh

local-down:
	./scripts/stop-local.sh
	$(DOCKER_COMPOSE) down --remove-orphans

local-status:
	./scripts/local-status.sh

run-local: check-local-ports
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/iamd

worker-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/worker

env-up:
	$(DOCKER_COMPOSE) up -d
	./scripts/wait-local-env.sh
	$(MAKE) migrate-local
	$(MAKE) seed-local
	$(MAKE) schema-sync-local

env-down:
	$(DOCKER_COMPOSE) down --remove-orphans

env-logs:
	$(DOCKER_COMPOSE) logs -f

env-ps:
	$(DOCKER_COMPOSE) ps

migrate:
	$(GO) run ./cmd/migrator

migrate-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/migrator

seed:
	$(GO) run ./cmd/seeder

seed-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/seeder

schema-sync:
	$(GO) run ./cmd/schema-sync

schema-sync-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/schema-sync
