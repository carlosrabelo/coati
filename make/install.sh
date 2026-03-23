#!/bin/bash
set -euo pipefail

BINARY_NAME="coati"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
INSTALL_DIR="$HOME/.local/bin"

install -D "$ROOT_DIR/bin/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
echo "Installed to: $INSTALL_DIR/$BINARY_NAME"
