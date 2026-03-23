#!/bin/bash
set -euo pipefail

BINARY_NAME="coati"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

rm -f "$ROOT_DIR/bin/$BINARY_NAME"
find "$ROOT_DIR/out" -mindepth 1 ! -name .gitkeep -delete
