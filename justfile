bin := ".bin"
buf_version := "1.75.0"
buf := bin / "buf"

# Download the pinned Buf CLI for the current platform.
[private]
[script("sh", "-e")]
install-buf:
    [ "$("{{buf}}" --version 2>/dev/null)" = "{{buf_version}}" ] && exit 0
    mkdir -p "{{bin}}"
    curl -fsSL --remove-on-error "https://github.com/bufbuild/buf/releases/download/v{{buf_version}}/buf-$(uname -s)-$(uname -m)" -o "{{buf}}"
    chmod +x "{{buf}}"

# Generate Go, gRPC, Connect, gateway, and OpenAPI code from proto/.
proto: proto-deps
    "{{buf}}" generate

# Update buf.lock and export Protobuf dependencies locally.
proto-deps: install-buf
    "{{buf}}" dep update
    mkdir -p .proto-deps
    "{{buf}}" export buf.build/bufbuild/protovalidate -o .proto-deps
    "{{buf}}" export buf.build/googleapis/googleapis -o .proto-deps
    "{{buf}}" export buf.build/gnostic/gnostic -o .proto-deps
