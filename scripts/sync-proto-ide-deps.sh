#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/.idea/proto-deps"

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

buf export buf.build/bufbuild/protovalidate \
  --path buf/validate \
  --output "$OUT_DIR"

buf export buf.build/googleapis/googleapis \
  --path google/api \
  --path google/protobuf \
  --output "$OUT_DIR"

buf export buf.build/gnostic/gnostic \
  --path gnostic/openapi/v3 \
  --output "$OUT_DIR"

echo "Proto IDE deps exported to $OUT_DIR"
