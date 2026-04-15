#!/bin/bash
set -euo pipefail

BINARY_NAME="coati"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

mkdir -p "$ROOT_DIR/bin"
cd "$ROOT_DIR"
go build -o "$ROOT_DIR/bin/$BINARY_NAME" "./coati/cmd/$BINARY_NAME"
echo "Binary ready at: $ROOT_DIR/bin/$BINARY_NAME"
