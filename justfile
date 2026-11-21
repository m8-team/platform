buf := "go run github.com/bufbuild/buf/cmd/buf@v1.54.0"

# Generate Go, gRPC, Connect, gateway, and OpenAPI code from proto/.
proto: proto-deps
    {{buf}} generate

# Update buf.lock and export Protobuf dependencies locally.
proto-deps:
    {{buf}} dep update
    mkdir -p .proto-deps
    {{buf}} export buf.build/bufbuild/protovalidate -o .proto-deps
    {{buf}} export buf.build/googleapis/googleapis -o .proto-deps
    {{buf}} export buf.build/gnostic/gnostic -o .proto-deps
