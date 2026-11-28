#!/usr/bin/env bash
# Purpose: run the same PDF build script that maid installs in projects.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export MAID_ROOT="$root"
exec "$root/internal/setup/templates/compile-pdf.sh" "$@"
