#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SPEC_PATH="$ROOT_DIR/api/asyncapi/asyncapi.yaml"
OUT_DIR="$ROOT_DIR/generated/ts"

echo "Generating TypeScript models..."

mkdir -p "$OUT_DIR"

asyncapi generate models typescript "$SPEC_PATH" -o "$OUT_DIR"

echo "TypeScript models generated in $OUT_DIR"