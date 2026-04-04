# M8 Platform IAM

M8 Platform IAM is a modular monolith for identity and access management of the M8 SaaS control plane.

Implemented repository outputs:

- protobuf-first gRPC contracts under `api/proto`
- `buf` configuration, generated Go stubs, grpc-gateway handlers, and OpenAPI spec under `gen/proto` and `gen/openapi`
- modular Go services for `identity`, `authz`, `graph`, `support`, `audit`, `ops`
- adapters for `YDB`, `Redis`, `Keycloak`, `SpiceDB`, `Temporal`, and `YDB Topics`
- HTTP REST gateway layered on top of the gRPC API
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
make local-up
make local-down
make local-status
make run-local
make worker-local
make stop-local
make migrate-local
make seed-local
make run
make worker
make migrate
make seed
make schema-sync
```

## Local Run

1. Start the full local environment in one command:

```bash
make local-up
```

This starts:

- dependency stack: `YDB`, `Redis`, `Keycloak`, `SpiceDB`, `Temporal`
- bootstrap steps: migrations, seed data, SpiceDB schema sync
- local processes in background: `iamd`, `worker`, `UI`

The background process logs are written to `${TMPDIR:-/tmp}/m8-platform-iam-local/logs`.

2. Local endpoints:

- `YDB` on `grpc://127.0.0.1:2136/local` with UI on `http://127.0.0.1:8765`
- `Redis` on `127.0.0.1:6379`
- `Keycloak` on `http://127.0.0.1:8081` with imported realm `m8`
- `SpiceDB` on `127.0.0.1:50051`
- `Temporal` on `127.0.0.1:7233` with UI on `http://127.0.0.1:8233`
- gRPC API on `127.0.0.1:8080`
- REST gateway on `http://127.0.0.1:8082`
- OpenAPI JSON on `http://127.0.0.1:8082/openapi/iam.swagger.json`
- Admin UI on `http://127.0.0.1:5173`

3. Check status or stop everything:

```bash
make local-status
make local-down
```

4. If you need the old step-by-step flow, you can still start only the dependency stack:

```bash
make env-up
```

This starts only:

- `YDB` on `grpc://127.0.0.1:2136/local` with UI on `http://127.0.0.1:8765`
- `Redis` on `127.0.0.1:6379`
- `Keycloak` on `http://127.0.0.1:8081` with imported realm `m8`
- `SpiceDB` on `127.0.0.1:50051`
- `Temporal` on `127.0.0.1:7233` with UI on `http://127.0.0.1:8233`

The `env-up` target also:

- applies `migrations/*.sql`
- loads local demo data from `testdata/seed/*.json`

5. Generate stubs and verify contracts:

```bash
buf dep update
buf lint
buf generate
go test ./...
```

6. Run the API server against the local stack:

```bash
make run-local
```

This starts:

- gRPC API on `127.0.0.1:8080`
- REST gateway on `http://127.0.0.1:8082`
- generated OpenAPI JSON on `http://127.0.0.1:8082/openapi/iam.swagger.json`

7. Run the Temporal worker against the local stack:

```bash
make worker-local
```

8. Print migrations and SpiceDB schema for bootstrap:

```bash
make migrate-local
make seed-local
go run ./cmd/schema-sync
```

9. Inspect or stop the stack:

```bash
make stop-local
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
- `local-up` uses the same env file and additionally starts the Vite UI with `IAM_UI_HOST`, `IAM_UI_PORT`, and `VITE_IAM_API_BASE_URL`.
- `run-local` checks `IAM_GRPC_ADDRESS` and `IAM_HTTP_ADDRESS` before startup and prints the conflicting PID if a previous `iamd` is still running.
- `stop-local` stops managed local `iamd`, `worker`, and `UI` processes.
- `IAM_HTTP_ADDRESS` controls the REST gateway listener and defaults to `:8082`.
- `IAM_UI_PORT` controls the Vite UI listener and defaults to `5173`.
- `IAM_OPENAPI_DIR` points to generated OpenAPI artifacts and defaults to `gen/openapi`.
- `migrate-local` applies `migrations/*.sql`, creates `schema_migrations` if needed, backfills bootstrap migrations when tables already exist, and applies new unapplied schema updates.
- `seed-local` idempotently upserts demo tenants, users, memberships, groups, group members, service accounts, OAuth clients, and access bindings from `testdata/seed/*.json`.
- `buf generate` now produces grpc-gateway handlers and a merged OpenAPI document at `gen/openapi/iam.swagger.json`.
- `YDB` runs with `platform: linux/amd64` and in-memory PDisks. On Apple Silicon, make sure Docker Desktop has Rosetta support enabled as recommended by YDB's Docker documentation.

## Local Seed Data

The default local dataset includes:

- tenant `tenant-demo` with admin, analyst, support, operations group, bot service account, and admin UI client
- tenant `tenant-sandbox` with owner, developer, developers group, CI service account, and sandbox UI client
- access bindings for tenant management, project viewer/editor roles, and support case access

## Postman

Postman assets for the local stack live under `postman/`:

- `postman/m8-platform-iam-local.postman_environment.json`
- `postman/m8-platform-iam-rest.postman_collection.json`
- `postman/grpc/README.md`
- `postman/grpc/examples/*.json`

The REST collection is a curated smoke and workflow set for local IAM scenarios.
For the full REST surface, import `gen/openapi/iam.swagger.json` into Postman.

For gRPC requests in Postman, import protobuf definitions from `api/proto` and use the ready JSON payloads from `postman/grpc/examples/`.

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
- `testdata/seed/sandbox.json`
