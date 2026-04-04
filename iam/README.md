# M8 Platform IAM

M8 Platform IAM is a modular monolith for identity and access management of the M8 SaaS control plane.

Implemented repository outputs:

- protobuf-first gRPC contracts under `api/proto`
- `buf` configuration and generated Go stubs under `gen/proto`
- modular Go services for `identity`, `authz`, `graph`, `support`, `audit`, `ops`
- adapters for `YDB`, `Redis`, `Keycloak`, `SpiceDB`, `Temporal`, and `YDB Topics`
- Temporal workflows for service-account lifecycle, support access, read-model rebuild and relationship sync
- seed data, YDB migrations, SpiceDB schema, deployment manifests, and basic unit tests

## Layout

```text
iam/
  api/proto/
  cmd/
  deploy/
  docs/
  gen/proto/
  internal/
  migrations/
  testdata/seed/
```

## Commands

```bash
make buf-lint
make buf-generate
make test
make env-up
make env-down
make env-logs
make env-ps
make run-local
make worker-local
make migrate-local
make run
make worker
make migrate
make schema-sync
```

## Local Run

1. Start the local dependency stack:

```bash
make env-up
```

This starts:

- `YDB` on `grpc://127.0.0.1:2136/local` with UI on `http://127.0.0.1:8765`
- `Redis` on `127.0.0.1:6379`
- `Keycloak` on `http://127.0.0.1:8081` with imported realm `m8`
- `SpiceDB` on `127.0.0.1:50051`
- `Temporal` on `127.0.0.1:7233` with UI on `http://127.0.0.1:8233`

2. Generate stubs and verify contracts:

```bash
buf dep update
buf lint
buf generate
go test ./...
```

3. Run the API server against the local stack:

```bash
make run-local
```

4. Run the Temporal worker against the local stack:

```bash
make worker-local
```

5. Print migrations and SpiceDB schema for bootstrap:

```bash
make migrate-local
go run ./cmd/schema-sync
```

6. Inspect or stop the stack:

```bash
make env-ps
make env-logs
make env-down
```

## Local Environment Notes

- The local `Keycloak` stack imports realm `m8` from `deploy/local/keycloak/m8-realm.json`.
- `Keycloak` credentials:
  - admin console: `admin / admin`
  - test realm user: `test-admin / admin`
- `run-local` and `worker-local` load environment from `deploy/local/iamd.env`.
- `migrate-local` applies `migrations/*.sql`, creates `schema_migrations` if needed, backfills bootstrap migrations when tables already exist, and applies new unapplied schema updates.
- `YDB` runs with `platform: linux/amd64` and in-memory PDisks. On Apple Silicon, make sure Docker Desktop has Rosetta support enabled as recommended by YDB's Docker documentation.

## Current Notes

- `buf lint`, `buf generate`, and `go test ./...` pass in the repository.
- The repository contains compile-ready adapter layers and workflow wiring. The YDB and SpiceDB adapter methods are intentionally conservative and require environment-specific query/write implementations before production rollout.
- Fallback authorization logic exists for local validation and tests, but production `CheckAccess` is expected to run through SpiceDB.

## Key Files

- `docs/spicedb/schema.zed`
- `migrations/001_init_identity.sql`
- `migrations/002_init_authorization.sql`
- `migrations/003_init_async.sql`
- `testdata/seed/demo.json`
