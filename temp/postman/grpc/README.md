# Postman gRPC Pack

Use this folder together with the `M8 Platform IAM Local` Postman environment.

## Import

1. Start the local API with `make run-local`.
2. In Postman create a new `gRPC` request.
3. Use server `{{grpc_target}}`.
4. Import protobuf definitions from the repo root `api/proto`.

Import the whole `api/proto` tree, not a single file. The IAM protos reference each other with relative imports.

## Method Names

- `saas.iam.identity.v1.IdentityService/*`
- `saas.iam.identity.v1.OAuthFacadeService/*`
- `saas.iam.authz.v1.AuthorizationFacadeService/*`
- `saas.iam.graph.v1.GraphService/*`
- `saas.iam.support.v1.SupportAccessService/*`
- `saas.iam.audit.v1.AuditService/*`
- `saas.iam.ops.v1.OperationsService/*`

## Payload Examples

Files under `examples/` map Postman method names to JSON request payloads:

- `identity-service.json`
- `oauth-facade-service.json`
- `authz-service.json`
- `graph-service.json`
- `support-service.json`
- `audit-service.json`
- `ops-service.json`

The payload values use the same variables as `m8-platform-iam-local.postman_environment.json`.
