BUF ?= buf
GO ?= go
DOCKER_COMPOSE ?= docker compose -f deploy/local/docker-compose.yaml

.PHONY: buf-lint buf-generate test run worker migrate seed schema-sync env-up env-down env-logs env-ps run-local worker-local migrate-local seed-local check-local-ports stop-local

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
	set -a; . ./deploy/local/iamd.env; set +a; \
	for addr in "$$IAM_GRPC_ADDRESS" "$$IAM_HTTP_ADDRESS"; do \
		port="$${addr##*:}"; \
		if [ -z "$$port" ]; then \
			continue; \
		fi; \
		if lsof -nP -iTCP:"$$port" -sTCP:LISTEN >/tmp/iam-port-check.$$port 2>/dev/null; then \
			echo "port $$port is already in use:"; \
			cat /tmp/iam-port-check.$$port; \
			rm -f /tmp/iam-port-check.$$port; \
			echo "run 'make stop-local' to stop a previous local iamd, or change the port in deploy/local/iamd.env"; \
			exit 1; \
		fi; \
	done

stop-local:
	set -a; . ./deploy/local/iamd.env; set +a; \
	pids="$$( { \
		lsof -tiTCP:"$${IAM_GRPC_ADDRESS##*:}" -sTCP:LISTEN 2>/dev/null || true; \
		lsof -tiTCP:"$${IAM_HTTP_ADDRESS##*:}" -sTCP:LISTEN 2>/dev/null || true; \
	} | sort -u )"; \
	if [ -z "$$pids" ]; then \
		echo "no local iamd processes found"; \
		exit 0; \
	fi; \
	echo "stopping local iamd: $$pids"; \
	kill $$pids

run-local: check-local-ports
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/iamd

worker-local:
	set -a; . ./deploy/local/iamd.env; set +a; $(GO) run ./cmd/worker

env-up:
	$(DOCKER_COMPOSE) up -d
	./scripts/wait-local-env.sh
	$(MAKE) migrate-local
	$(MAKE) seed-local

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
