#!/usr/bin/env bash
# Beads spike: checks the task-store behaviour that maid needs.
# Requires: bd on PATH (npm install -g @beads/bd, or brew install beads), git, python3.
# Usage: spikes/beads/run.sh [workdir]
set -euo pipefail

fail() {
  # Purpose: print a failure message and stop the spike.
  echo "FAIL: $*" >&2
  exit 1
}

field() {
  # Purpose: read one field of the issue JSON that bd show returns on stdin.
  python3 -c 'import json,sys; d=json.load(sys.stdin); d=d[0] if isinstance(d,list) else d; print(d.get(sys.argv[1]))' "$1"
}

setup_repo() {
  # Purpose: create an empty git repository with beads initialised for maid.
  rm -rf "$1" && mkdir -p "$1" && cd "$1"
  git init -q
  git config user.email spike@example.invalid
  git config user.name spike
  bd init --non-interactive --skip-agents --skip-hooks --quiet
  [ ! -e AGENTS.md ] || fail "bd init wrote AGENTS.md despite --skip-agents"
}

test_graph() {
  # Purpose: create a dependency graph with maid metadata and check ready work.
  A=$(bd create "store schema" --spec-id REQ-WI-001 --acceptance "TestStore passes" \
        --metadata '{"maid":{"owns":["internal/store/**"]}}' --silent)
  B=$(bd create "claim tool" --deps "$A" --silent)
  C=$(bd create "ready tool" --deps "$A" --silent)
  ready=$(bd ready --json | python3 -c 'import json,sys; print(" ".join(i["id"] for i in json.load(sys.stdin)))')
  [[ " $ready " == *" $A "* ]] || fail "root item not ready"
  [[ " $ready " != *" $B "* ]] || fail "blocked item listed as ready"
  [ "$(bd show "$A" --json | field spec_id)" = "REQ-WI-001" ] || fail "spec_id did not round-trip"
  echo "PASS graph, ready, metadata"
}

test_same_item_race() {
  # Purpose: six processes claim one item; exactly one must win.
  for i in 1 2 3 4 5 6; do
    ( BEADS_ACTOR="agent$i" bd update "$A" --claim --json >/dev/null 2>&1 && echo win ) &
  done | grep -c win > wins.txt || true
  wait
  [ "$(cat wins.txt)" = "1" ] || fail "expected 1 winner, got $(cat wins.txt)"
  echo "PASS same-item race: 1 of 6 claims won"
}

test_close_releases() {
  # Purpose: closing an item makes its dependents ready.
  owner=$(bd show "$A" --json | field assignee)
  BEADS_ACTOR="$owner" bd close "$A" --reason done >/dev/null || fail "owner $owner could not close $A"
  ready=$(bd ready --json | python3 -c 'import json,sys; print(" ".join(i["id"] for i in json.load(sys.stdin)))')
  [[ " $ready " == *" $B "* && " $ready " == *" $C "* ]] || fail "dependents not ready after close"
  echo "PASS close releases dependents"
}

test_parallel_claims() {
  # Purpose: twenty processes claim twenty different items; all must succeed.
  ids=()
  for i in $(seq 1 20); do ids+=("$(bd create "load$i" --silent)"); done
  start=$(date +%s)
  for i in $(seq 0 19); do
    ( BEADS_ACTOR="w$i" bd update "${ids[$i]}" --claim >/dev/null 2>&1 || echo "x" ) &
  done | grep -c x > fails.txt || true
  wait
  [ "$(cat fails.txt)" = "0" ] || fail "$(cat fails.txt) parallel claims failed"
  echo "PASS 20 parallel claims in $(( $(date +%s) - start ))s"
}

test_claim_next() {
  # Purpose: eight processes take the next ready item; each must get a different one.
  for i in $(seq 1 8); do bd create "pool$i" -l pool --silent >/dev/null; done
  for i in $(seq 1 8); do
    ( BEADS_ACTOR="p$i" bd ready --claim -l pool --json 2>/dev/null | field title ) &
  done | sort | uniq -d > dups.txt
  wait
  [ ! -s dups.txt ] || fail "two processes claimed the same item: $(cat dups.txt)"
  echo "PASS claim-next: no duplicate claims"
}

main() {
  # Purpose: run all spike checks in a scratch repository.
  command -v bd >/dev/null || fail "bd not found on PATH"
  echo "bd: $(bd version)"
  setup_repo "${1:-${TMPDIR:-/tmp}/maid-beads-spike}"
  test_graph
  test_same_item_race
  test_close_releases
  test_parallel_claims
  test_claim_next
  echo "ALL PASS"
}

main "$@"
