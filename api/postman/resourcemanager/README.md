# M8 Resource Manager Postman pack

This directory contains Postman-ready REST requests and copy/paste gRPC
payloads for the Resource Manager API.

## Files

- `m8-resource-manager-local.postman_environment.json` — local variables.
- `m8-resource-manager-rest.postman_collection.json` — importable Postman
  Collection v2.1 with health checks and all Resource Manager REST methods.
- `grpc/README.md` — how to load the protobuf definitions in Postman.
- `grpc/examples/*.json` — ProtoJSON payloads for every gRPC method.

The authoritative API contracts are:

- `api/proto/m8/platform/resourcemanager/v1/organization_service.proto`
- `api/proto/m8/platform/resourcemanager/v1/workspace_service.proto`
- `api/proto/m8/platform/resourcemanager/v1/service_service.proto`

## Import into Postman

1. Import `m8-resource-manager-local.postman_environment.json`.
2. Import `m8-resource-manager-rest.postman_collection.json`.
3. Select the `M8 Resource Manager - Local` environment.
4. Keep `access_token` ready for the authentication adapter; the current
   runtime does not consume bearer tokens yet.
5. Run requests in folder order. Successful create/get/update responses save
   resource IDs and versions to the environment automatically.

## Local runtime

For local gRPC testing without an authentication adapter, start the service
from the repository root with:

```bash
M8_RM_ALLOW_UNAUTHENTICATED=true go run ./cmd/resource-manager
```

Default endpoints:

- health HTTP server: `http://127.0.0.1:8080`
- plaintext gRPC server: `127.0.0.1:9090`

Current runtime support is intentionally documented here so the collection
does not create a false expectation:

| Interface | Current status |
| --- | --- |
| HTTP health (`/livez`, `/readyz`, `/startupz`, `/healthz`) | Available |
| Resource Manager REST routes | Contract exists, but no gRPC-Gateway is registered yet |
| `OrganizationService` gRPC | Registered |
| `WorkspaceService` gRPC | Contract exists, but is not registered yet |
| `ServiceService` gRPC | Contract exists, but is not registered yet |
| gRPC server reflection | Not enabled; import proto definitions manually |
| `google.longrunning.Operations` polling | Not registered; current Organization mutations return completed operations |

Until the HTTP gateway is added, the business requests in the REST collection
will return HTTP 404 from the current health-only HTTP server. They already use
the canonical paths from the current protobuf annotations:
`/resource-manager/v1/...`.

## Resource hierarchy

Requests use the following hierarchy:

```text
Organization
└── Workspace
    └── Service (environment: dev, stage, prod, ...)
```

The service environment is required at creation time and immutable. It must be
a lowercase DNS-label-compatible value. Change `service_environment` in the
Postman environment to test another environment.

## Authentication notes

The REST collection is prepared to use
`Authorization: Bearer {{access_token}}`. For gRPC, the equivalent metadata is
key `authorization` with value `Bearer {{access_token}}`. The current runtime
does not have an authentication adapter that consumes this token: it denies
business requests by default, and local access is enabled only by
`M8_RM_ALLOW_UNAUTHENTICATED=true`. That switch is for development only.
