#!/bin/bash
set -euo pipefail

BINARY_NAME="coati"
INSTALL_DIR="$HOME/.local/bin"

rm -f "$INSTALL_DIR/$BINARY_NAME"
echo "Uninstalled: $INSTALL_DIR/$BINARY_NAME"
