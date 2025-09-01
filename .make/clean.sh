#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-coati}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR"
go clean
rm -f "$ROOT_DIR/bin/$BINARY_NAME"
