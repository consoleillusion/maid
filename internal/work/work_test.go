package work

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/tasks"
)

const reqChapter = `---
status: accepted
style: ears
owner: paul
summary: Calc.
links: []
---

# Calc

### REQ-CA-001 Add
When a caller adds two integers, the calc package shall return their sum.
`

const criteria = `Scenario: add
  Given the numbers 2 and 3
  When the caller adds them
  Then the result is 5`

func sh(t *testing.T, dir string, args ...string) string {
	// Purpose: run a command in dir and fail the test on error.
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, dir, rel, content string) {
	// Purpose: write a file, creating parent folders.
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, dir, msg string) {
	// Purpose: commit every change in a worktree.
	t.Helper()
	sh(t, dir, "git", "add", "-A")
	sh(t, dir, "git", "commit", "-q", "-m", msg)
}

// project creates a git repository with a Go module, a design and one passing test.
func project(t *testing.T) (*Service, string) {
	// Purpose: set up a small real project for merge queue tests.
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	root := t.TempDir()
	sh(t, root, "git", "init", "-q", "-b", "main")
	sh(t, root, "git", "config", "user.email", "t@example.invalid")
	sh(t, root, "git", "config", "user.name", "t")
	write(t, root, "go.mod", "module example.com/p\n\ngo 1.21\n")
	write(t, root, "design/01-a/calc.md", reqChapter)
	write(t, root, "calc/calc.go", "package calc\n")
	write(t, root, "other/other.go", "package other\n\nfunc One() int { return 1 }\n")
	write(t, root, "other/other_test.go", "package other\n\nimport \"testing\"\n\nfunc TestOne(t *testing.T) {\n\tif One() != 1 {\n\t\tt.Fatal(\"bad\")\n\t}\n}\n")
	write(t, root, ".gitignore", ".maid/\n")
	commitAll(t, root, "init")
	cfg := config.Default()
	cfg.Lint = nil
	return New(root, cfg, tasks.NewMemory()), root
}

func plan(t *testing.T, s *Service) Planned {
	// Purpose: create the standard one-item plan.
	t.Helper()
	p, err := s.CreatePlan(context.Background(), Plan{Items: []PlanItem{{
		Key: "add", Title: "Add two integers", Implements: []string{"REQ-CA-001"}, Criteria: criteria,
		AcceptanceTests: []tasks.TestRef{{File: "calc/add_test.go", Names: []string{"TestAdd"}}},
		Owns:            []string{"calc/calc.go"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

const addTest = "package calc\n\nimport \"testing\"\n\n// REQ-CA-001 scenario: add\nfunc TestAdd(t *testing.T) {\n\tif Add(2, 3) != 5 {\n\t\tt.Fatal(\"want 5\")\n\t}\n}\n"

// REQ-WI-002 REQ-WI-017 REQ-WI-018 REQ-WI-019 REQ-WI-020 REQ-ACC-007 REQ-ACC-011
func TestFullLoop(t *testing.T) {
	// Purpose: tests item, then the code item by a different agent, both through the merge queue.
	s, root := project(t)
	ctx := context.Background()
	p := plan(t, s)
	if _, err := s.Claim(ctx, "builder", p.Code["add"]); err == nil {
		t.Fatal("code item claimable before its tests exist")
	}
	// The tests author writes a real acceptance test that fails to compile on
	// main; maid must reject it (REQ-WI-020) ...
	pk, err := s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	commitAll(t, pk.Worktree, "tests")
	rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: pk.WorkItem.ID})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Accepted || rep.FailedCheck != "build" {
		t.Fatalf("non-compiling tests accepted: %+v", rep)
	}
	// ... so the tests author adds a failing stub in an owned test helper.
	pk, err = s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add is not implemented yet.\nfunc Add(a, b int) int { return 0 }\n")
	commitAll(t, pk.Worktree, "tests")
	rep, err = s.Submit(ctx, Submission{Agent: "tester", ID: pk.WorkItem.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("tests rejected: %+v", rep)
	}
	if _, err := s.Claim(ctx, "tester", p.Code["add"]); err == nil {
		t.Fatal("the test author could claim the code item")
	}
	pk, err = s.Claim(ctx, "builder", p.Code["add"])
	if err != nil {
		t.Fatal(err)
	}
	if pk.TestFiles["calc/add_test.go"] == "" || len(pk.Requirements) != 1 || pk.Criteria == "" {
		t.Fatalf("incomplete context packet: %+v", pk)
	}
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n")
	commitAll(t, pk.Worktree, "impl")
	rep, err = s.Submit(ctx, Submission{Agent: "builder", ID: pk.WorkItem.ID, Model: "haiku"})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("correct implementation rejected: %+v", rep)
	}
	if got := sh(t, root, "git", "show", rep.Commit+":calc/calc.go"); !strings.Contains(got, "return a + b") {
		t.Fatalf("main not updated: %s", got)
	}
	it, _ := s.Tasks.Get(ctx, p.Code["add"])
	if it.Status != tasks.StatusClosed {
		t.Fatalf("code item status %s", it.Status)
	}
}

// acceptedTests brings the project to the point where the code item is claimable.
func acceptedTests(t *testing.T) (*Service, string, Planned) {
	// Purpose: set up main with a stub and an accepted failing acceptance test.
	s, root := project(t)
	write(t, root, "calc/calc.go", "package calc\n\n// Add is not implemented yet.\nfunc Add(a, b int) int { return 0 }\n")
	commitAll(t, root, "stub")
	p := plan(t, s)
	ctx := context.Background()
	pk, err := s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	commitAll(t, pk.Worktree, "tests")
	if rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: pk.WorkItem.ID}); err != nil || !rep.Accepted {
		t.Fatalf("tests: %v %+v", err, rep)
	}
	return s, root, p
}

// REQ-ACC-008 REQ-WI-008 REQ-ACC-010 REQ-WI-009
func TestRejections(t *testing.T) {
	// Purpose: outside paths, changed acceptance tests and regressions are rejected; budget escalates.
	s, _, p := acceptedTests(t)
	ctx := context.Background()
	id := p.Code["add"]

	pk, err := s.Claim(ctx, "builder", id)
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n")
	write(t, pk.Worktree, "other/other.go", "package other\n\nfunc One() int { return 2 }\n")
	commitAll(t, pk.Worktree, "too much")
	rep, err := s.Submit(ctx, Submission{Agent: "builder", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Accepted || rep.FailedCheck != "owned-paths" || rep.OutsidePaths[0] != "other/other.go" {
		t.Fatalf("outside path not rejected: %+v", rep)
	}

	pk, err = s.Claim(ctx, "builder2", id)
	if err != nil {
		t.Fatal(err)
	}
	if pk.Attempt != 2 {
		t.Fatalf("attempt = %d, want 2", pk.Attempt)
	}
	s.WorkItemOwnsForTest(ctx, id, append(pk.WorkItem.Maid.Owns, "calc/add_test.go"))
	write(t, pk.Worktree, "calc/add_test.go", "package calc\n")
	commitAll(t, pk.Worktree, "delete test")
	rep, err = s.Submit(ctx, Submission{Agent: "builder2", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Accepted || rep.FailedCheck != "protected-tests" || !rep.Escalated {
		t.Fatalf("changed acceptance test not rejected and escalated: %+v", rep)
	}
	if _, err := s.Claim(ctx, "builder3", id); err == nil {
		t.Fatal("escalated item still claimable")
	}
}

// REQ-ACC-008
func TestRegressionRejected(t *testing.T) {
	// Purpose: a submission that breaks a passing test is rejected and main is unchanged.
	s, root, p := acceptedTests(t)
	ctx := context.Background()
	id := p.Code["add"]
	s.WorkItemOwnsForTest(ctx, id, []string{"calc/calc.go", "other/other.go"})
	pk, err := s.Claim(ctx, "builder", id)
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n")
	write(t, pk.Worktree, "other/other.go", "package other\n\nfunc One() int { return 2 }\n")
	commitAll(t, pk.Worktree, "breaks other")
	before := sh(t, root, "git", "rev-parse", "main")
	rep, err := s.Submit(ctx, Submission{Agent: "builder", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Accepted || rep.FailedCheck != "regressions" || len(rep.Regressions) != 1 || !strings.HasSuffix(rep.Regressions[0], "#TestOne") {
		t.Fatalf("regression not rejected: %+v", rep)
	}
	if after := sh(t, root, "git", "rev-parse", "main"); after != before {
		t.Fatal("main changed after a rejection")
	}
}

// REQ-WI-004
func TestClaimConflict(t *testing.T) {
	// Purpose: a second agent cannot claim an item that is already claimed.
	s, _ := project(t)
	ctx := context.Background()
	p := plan(t, s)
	if _, err := s.Claim(ctx, "a", p.Tests["add"]); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim(ctx, "b", ""); !errors.Is(err, tasks.ErrNoReady) {
		t.Fatalf("err = %v, want ErrNoReady", err)
	}
}

// REQ-WI-001
func TestPlanValidation(t *testing.T) {
	// Purpose: reject plans with unknown requirements or tests inside owns.
	s, _ := project(t)
	_, err := s.CreatePlan(context.Background(), Plan{Items: []PlanItem{{
		Key: "x", Title: "x", Implements: []string{"REQ-ZZ-001"}, Criteria: criteria,
		AcceptanceTests: []tasks.TestRef{{File: "calc/add_test.go", Names: []string{"TestAdd"}}},
		Owns:            []string{"calc/**"},
	}}})
	if err == nil || !strings.Contains(err.Error(), "unknown requirement") || !strings.Contains(err.Error(), "inside owns") {
		t.Fatalf("err = %v", err)
	}
}

// WorkItemOwnsForTest replaces the owned paths of a work item.
func (s *Service) WorkItemOwnsForTest(ctx context.Context, id string, owns []string) {
	// Purpose: let tests widen owns to reach a later check.
	it, err := s.Tasks.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	it.Maid.Owns = owns
	if err := s.Tasks.SetMeta(ctx, id, it.Maid); err != nil {
		panic(err)
	}
}

// REQ-WI-022
func TestTestsItemMayAddStub(t *testing.T) {
	// Purpose: tests for code that does not exist yet compile through a stub and are accepted.
	s, root := project(t)
	ctx := context.Background()
	p := plan(t, s)
	pk, err := s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n")
	commitAll(t, pk.Worktree, "tests and stub")
	rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: pk.WorkItem.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("tests with stub rejected: %+v", rep)
	}
	pk, err = s.Claim(ctx, "builder", p.Code["add"])
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n")
	commitAll(t, pk.Worktree, "impl")
	if rep, err = s.Submit(ctx, Submission{Agent: "builder", ID: pk.WorkItem.ID}); err != nil || !rep.Accepted {
		t.Fatalf("implementation rejected: %v %+v", err, rep)
	}
	_ = root
}
