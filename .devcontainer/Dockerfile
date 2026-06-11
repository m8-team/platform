FROM mcr.microsoft.com/devcontainers/go:1-bookworm

RUN apt-get update && apt-get install -y \
    make \
    git \
    curl \
    jq \
    unzip \
    ca-certificates \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/bufbuild/buf/cmd/buf@latest \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest \
    && go install github.com/air-verse/air@latest