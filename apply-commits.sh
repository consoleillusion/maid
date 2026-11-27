#!/usr/bin/env bash
# Adds the design write tools as one commit. Apply the Nix archive first.
# Unzip into the repository root, run ./apply-commits.sh, push yourself.
set -euo pipefail

TRAILER=''

die() {
  # Purpose: print an error and stop.
  echo "apply-commits: $*" >&2
  exit 1
}

main() {
  # Purpose: commit the design write tools, the accept command and their tests.
  git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git repository"
  [ "$(git rev-parse --show-toplevel)" = "$(pwd -P)" ] || die "run this from the repository root"
  git diff --cached --quiet || die "the index has staged changes; commit or unstage them first"
  git ls-files --error-unmatch cmd/maid/main.go >/dev/null 2>&1 || die "apply the milestone 1 archive first"
  git add -- README.md cmd/maid/main.go design/07-delivery/01-roadmap.md \
    internal/design/store.go internal/design/write.go internal/design/write_test.go \
    internal/guide/guide.go internal/mcp/server_test.go internal/mcp/tools.go internal/work/designops.go
  if git diff --cached --quiet; then
    echo "skip (no changes): already applied"
    return
  fi
  printf '%s\n\n%s\n' "feat(design): let agents write and accept the design through maid

Adds design_write (section replace with version check), design_append,
design_create, design_set_status, design_propose and design_question.
Accepted chapters reject direct writes and take proposals. Accepting
needs approved_by and a chapter without check errors. Each change is
committed under a design lock. maid accept CHAPTER records a human
acceptance from the CLI. A project with no design/ folder now reaches
the bootstrap step instead of failing." "$TRAILER" | git commit -q -F -
  echo "commit: feat(design): let agents write and accept the design through maid"
}

main "$@"
