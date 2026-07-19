# Resource Manager gRPC requests in Postman

Postman needs service definitions because the current Resource Manager server
does not expose gRPC reflection. Export the Resource Manager proto files and
all of their imports from the repository root:

```bash
postman_proto_dir="$(mktemp -d)"
buf export . \
  --path api/proto/m8/platform/resourcemanager/v1 \
  --output "$postman_proto_dir"
echo "$postman_proto_dir"
```

In Postman:

1. Create a new gRPC request.
2. Set the server URL to `{{grpc_target}}` and leave TLS disabled for local use.
3. Open **Service definition** > **Import a .proto file** and select the
   relevant file under
   `$postman_proto_dir/m8/platform/resourcemanager/v1/`, for example
   `organization_service.proto`.
4. In the request's **Import paths** tab, add the printed
   `$postman_proto_dir` directory. Imports such as
   `google/longrunning/operations.proto` then resolve from that proto root.
5. Select one of these services:
   - `m8.platform.resourcemanager.v1.OrganizationService`
   - `m8.platform.resourcemanager.v1.WorkspaceService`
   - `m8.platform.resourcemanager.v1.ServiceService`
6. Select a method and paste the matching object from `examples/*.json` into
   the Message tab. Each payload is stored under `methods.<MethodName>`; do not
   paste the example file's outer `service` and `methods` wrapper.
7. Once an authentication adapter is connected, add metadata
   `authorization: Bearer {{access_token}}`. The current runtime does not
   consume bearer metadata; use the local unauthenticated switch described in
   the parent README for development calls.

Postman's JSON message editor uses protobuf JSON field names. Accordingly,
these examples use `organizationId`, `workspaceId`, `pageSize`, `updateMask`,
and other lowerCamelCase names. Protobuf `int64` versions are represented as
JSON strings.

After a create/get/update/undelete call, copy the returned resource `id` and
`version` into the matching Postman environment variables. Mutation responses
place the resource under `response.organization`, `response.workspace`, or
`response.service`; direct get responses return the resource at the top level.

Postman stores gRPC requests in a multi-protocol collection. That format is
separate from the exportable HTTP Collection v2.1 file, so this repository
keeps the REST collection and the gRPC payload pack separate instead of
inventing a non-standard exported gRPC collection.

## Runtime status

- `OrganizationService` can be called on the current local server.
- `WorkspaceService` and `ServiceService` payloads match the contracts but
  calls currently return `Unimplemented` because those adapters are not
  registered in the composition root.
- Registered Organization mutation methods currently return a completed
  `google.longrunning.Operation`; there is no registered Operations service to
  poll afterward. Workspace and Service mutations remain `Unimplemented` with
  the current composition root.
