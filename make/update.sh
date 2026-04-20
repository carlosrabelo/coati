#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

sudo cp -f "$ROOT_DIR/data/gen/etc/hosts" /etc/hosts
cp -f "$ROOT_DIR/data/gen/ssh/config" ~/.ssh/config
