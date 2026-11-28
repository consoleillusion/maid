#!/usr/bin/env bash
# Purpose: build design/index.typ as a light PDF and a dark PDF.
set -euo pipefail

root="${MAID_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
cd "$root"

dark_only=0
case "${1:-}" in
  "") ;;
  --dark-only) dark_only=1 ;;
  -h | --help)
    echo "usage: scripts/compile-pdf.sh [--dark-only]" >&2
    exit 0
    ;;
  *)
    echo "unknown option: $1" >&2
    exit 2
    ;;
esac

typst="${TYPST:-typst}"
if ! command -v "$typst" >/dev/null 2>&1; then
  echo "typst 0.15 or later is required" >&2
  exit 1
fi
if [[ ! -f design/index.md || ! -f design/index.typ ]]; then
  echo "design/index.md and design/index.typ are required" >&2
  exit 1
fi

title="$(awk '/^# /{print substr($0, 3); exit}' design/index.md)"
slug="$(printf '%s' "${title:-design}" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g')"
[[ -n "$slug" ]] || slug=design

inputs=()
if [[ -f design/cover_image.png ]]; then
  inputs+=(--input cover=yes)
fi

if ((dark_only == 0)); then
  "$typst" compile "${inputs[@]}" design/index.typ "${slug}-light.pdf" >&2
fi
"$typst" compile "${inputs[@]}" --input mode=dark design/index.typ "${slug}-dark.pdf" >&2

# Print the output paths on stdout so callers can use them.
if ((dark_only == 0)); then
  printf '%s\n' "${slug}-light.pdf"
fi
printf '%s\n' "${slug}-dark.pdf"
