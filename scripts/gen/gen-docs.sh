#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SPEC_PATH="$ROOT_DIR/api/asyncapi/asyncapi.yaml"
OUT_DIR="$ROOT_DIR/generated/docs"

echo "Generating AsyncAPI HTML docs..."

mkdir -p "$OUT_DIR"

asyncapi generate fromTemplate "$SPEC_PATH" @asyncapi/html-template \
  -o "$OUT_DIR" \
  --force-write

echo "Docs generated in $OUT_DIR"