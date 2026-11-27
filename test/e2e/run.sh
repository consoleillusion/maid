#!/usr/bin/env bash
# End-to-end check of maid with the real binary and real Beads:
# init, plan, tests item by one agent, code item by another, merge to main.
# Requires go, git, bd. Usage: test/e2e/run.sh [maid-binary]
set -euo pipefail

fail() {
  # Purpose: stop with a message.
  echo "E2E FAIL: $*" >&2
  exit 1
}

field() {
  # Purpose: read a top-level field from JSON on stdin.
  python3 -c 'import json,sys; d=json.load(sys.stdin); print(d.get(sys.argv[1]) if isinstance(d,dict) else "")' "$1"
}

setup() {
  # Purpose: create a project with a design, a stub and one passing test.
  dir=$(mktemp -d)
  cd "$dir"
  git init -q -b main
  git config user.email e2e@example.invalid
  git config user.name e2e
  mkdir -p design/01-a calc
  cat > design/01-a/calc.md <<'EOF'
---
id: calc
title: Calc
part: 01-a
status: accepted
style: ears
owner: e2e
summary: Calc.
links: []
---

# Calc

### REQ-CA-001 Add
When a caller adds two integers, the calc package shall return their sum.
EOF
  cat > design/01-a/vision.md <<'EOF'
---
id: vision
title: Vision
part: 01-a
status: accepted
style: ste
owner: e2e
summary: A calculator.
links: []
---

# Vision

### G-001 Add numbers
The calc package adds numbers.

### NG-001 No floats
The calc package does not handle floating point numbers.
EOF
  printf 'module example.com/p\n\ngo 1.21\n' > go.mod
  printf 'package calc\n\n// Add is not implemented yet.\nfunc Add(a, b int) int { return 0 }\n' > calc/calc.go
  "$MAID" init >/dev/null
  git add -A && git commit -q -m init
}

write_plan() {
  # Purpose: write a one-item plan.
  cat > /tmp/maid-e2e-plan.json <<'EOF'
{"items":[{"key":"add","title":"Add two integers","implements":["REQ-CA-001"],
  "criteria":"Scenario: add\n  Given 2 and 3\n  When the caller adds them\n  Then the result is 5",
  "acceptance_tests":[{"file":"calc/add_test.go","names":["TestAdd"]}],
  "owns":["calc/calc.go"]}]}
EOF
}

main() {
  # Purpose: run the whole loop and check the result on main.
  MAID=$(readlink -f "${1:-maid}")
  setup
  write_plan
  plan=$("$MAID" plan /tmp/maid-e2e-plan.json)
  tests_id=$(echo "$plan" | python3 -c 'import json,sys; print(json.load(sys.stdin)["tests"]["add"])')
  code_id=$(echo "$plan" | python3 -c 'import json,sys; print(json.load(sys.stdin)["code"]["add"])')

  wt=$("$MAID" claim -agent tester | field worktree)
  [ -d "$wt" ] || fail "no worktree for tests item"
  printf 'package calc\n\nimport "testing"\n\n// REQ-CA-001 scenario: add\nfunc TestAdd(t *testing.T) {\n\tif Add(2, 3) != 5 {\n\t\tt.Fatal("want 5")\n\t}\n}\n' > "$wt/calc/add_test.go"
  git -C "$wt" add -A && git -C "$wt" commit -q -m tests
  "$MAID" submit -agent tester "$tests_id" | field accepted | grep -qx True || fail "tests item rejected"

  "$MAID" claim -agent tester "$code_id" >/dev/null 2>&1 && fail "test author claimed the code item"
  wt=$("$MAID" claim -agent builder | field worktree)
  printf 'package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n' > "$wt/calc/calc.go"
  git -C "$wt" add -A && git -C "$wt" commit -q -m impl
  "$MAID" submit -agent builder -model haiku "$code_id" | field accepted | grep -qx True || fail "code item rejected"

  git show main:calc/calc.go | grep -q 'return a + b' || fail "main was not updated"
  go test ./... >/dev/null || fail "tests fail on main"
  "$MAID" status | grep -q '7-verify' || fail "phase is not 7-verify"
  echo "E2E PASS ($dir)"
}

main "$@"
