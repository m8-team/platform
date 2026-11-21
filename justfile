bin_dir := justfile_directory() / ".bin"
export PATH := bin_dir + ":" + env("PATH")

# Generate Go, gRPC, Connect, gateway, and OpenAPI code from proto/.
proto:
    buf generate

# Update buf.lock and export Protobuf dependencies locally.
proto-deps:
    buf dep update
    mkdir -p .proto-deps
    buf export buf.build/bufbuild/protovalidate -o .proto-deps
    buf export buf.build/googleapis/googleapis -o .proto-deps
    buf export buf.build/gnostic/gnostic -o .proto-deps
