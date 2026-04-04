# Postman Assets

This folder contains Postman assets for the local IAM stack.

Files:

- `m8-platform-iam-local.postman_environment.json`: local variables for `http://127.0.0.1:8082` and `127.0.0.1:8080`
- `m8-platform-iam-rest.postman_collection.json`: curated REST collection for the main IAM flows
- `grpc/README.md`: how to use the gRPC API in Postman
- `grpc/examples/*.json`: ready request payloads for gRPC methods

Recommended import order:

1. Import `m8-platform-iam-local.postman_environment.json`.
2. Import `m8-platform-iam-rest.postman_collection.json`.
3. Select the `M8 Platform IAM Local` environment.
4. Start the local stack with `make env-up`, `make run-local`, and `make worker-local`.

Notes:

- The REST collection is curated around local smoke and workflow scenarios.
- For the full REST surface, Postman can also import the generated OpenAPI document from `gen/openapi/iam.swagger.json`.
- gRPC requests are provided as a Postman-oriented pack under `grpc/` because Postman uses protobuf definitions directly for gRPC methods.
