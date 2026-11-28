package work

import (
	"context"
	"testing"

	"github.com/consoleillusion/maid/internal/tasks"
)

// scenario: a tests item redefines an existing test
func TestTestsItemMayRedefineTest(t *testing.T) {
	// Purpose: a tests item that changes what an existing test expects is not a regression.
	s, _ := project(t)
	ctx := context.Background()
	p, err := s.CreatePlan(ctx, Plan{Items: []PlanItem{{
		Key: "two", Title: "One returns two", Implements: []string{"calc#add"}, Criteria: criteria,
		AcceptanceTests: []tasks.TestRef{{File: "other/other_test.go", Names: []string{"TestOne"}}},
		Owns:            []string{"other/other.go"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	id := p.Tests["two"]
	pk, err := s.Claim(ctx, "tester", id)
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "other/other_test.go", "package other\n\nimport \"testing\"\n\n// design section: calc#add\nfunc TestOne(t *testing.T) {\n\tif One() != 2 {\n\t\tt.Fatal(\"want 2\")\n\t}\n}\n")
	commitAll(t, pk.Worktree, "redefine TestOne")
	rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("report = %+v, want accepted", rep)
	}
}
