package work

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/tasks"
)

// scenario: missing formatter
func TestMissingFormatterIsEnvironmentError(t *testing.T) {
	// Purpose: a missing tool is an environment error that does not use an attempt.
	s, _ := project(t)
	s.Cfg.FormatCheck = []string{"no-such-tool-xyz"}
	ctx := context.Background()
	p := plan(t, s)
	id := p.Tests["add"]
	pk, err := s.Claim(ctx, "tester", id)
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n")
	commitAll(t, pk.Worktree, "tests")
	rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Accepted || rep.FailedCheck != "environment" || !rep.Environment {
		t.Fatalf("report = %+v, want failed check environment", rep)
	}
	if !strings.Contains(rep.Reason, "no-such-tool-xyz") {
		t.Errorf("reason %q does not name the command", rep.Reason)
	}
	it, err := s.Tasks.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if it.Maid.Attempts != 0 || it.Assignee != "tester" || it.Status != tasks.StatusInProgress {
		t.Errorf("after an environment error: attempts %d, assignee %q, status %s; want 0, tester, in_progress",
			it.Maid.Attempts, it.Assignee, it.Status)
	}
}

// scenario: missing tools listed
func TestMissingToolsListed(t *testing.T) {
	// Purpose: MissingTools names each configured command that is not installed.
	s, _ := project(t)
	s.Cfg.Lint = []string{"no-such-tool-xyz", "./..."}
	if got := s.MissingTools(); !reflect.DeepEqual(got, []string{"no-such-tool-xyz"}) {
		t.Errorf("MissingTools() = %v, want [no-such-tool-xyz]", got)
	}
}
