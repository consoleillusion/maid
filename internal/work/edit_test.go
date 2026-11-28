package work

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// claimTests plans the standard item and claims its tests item as tester.
func claimTests(t *testing.T) (*Service, string, string) {
	// Purpose: give each edit test a claimed item and its worktree.
	t.Helper()
	s, _ := project(t)
	id := plan(t, s).Tests["add"]
	pk, err := s.Claim(context.Background(), "tester", id)
	if err != nil {
		t.Fatal(err)
	}
	return s, id, pk.Worktree
}

func readFile(t *testing.T, dir, rel string) string {
	// Purpose: read a worktree file and fail the test on error.
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// scenario: replace text
func TestEditReplacesText(t *testing.T) {
	// Purpose: an exact replacement changes only the matched text.
	s, id, wt := claimTests(t)
	write(t, wt, "calc/calc.go", "package calc\n\nfunc Add(a, b int) int { return 0 }\n")
	err := s.EditFile(context.Background(), "tester", id, []Edit{{Path: "calc/calc.go", Old: "return 0", New: "return a + b"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, wt, "calc/calc.go"); got != "package calc\n\nfunc Add(a, b int) int { return a + b }\n" {
		t.Errorf("calc/calc.go = %q", got)
	}
}

// scenario: ambiguous match
func TestEditRefusesAmbiguousMatch(t *testing.T) {
	// Purpose: an old text that matches twice refuses every edit in the call.
	s, id, wt := claimTests(t)
	before := "package calc\n\nvar x = 1 // x\n"
	write(t, wt, "calc/calc.go", before)
	err := s.EditFile(context.Background(), "tester", id, []Edit{
		{Path: "calc/calc.go", Old: "package calc", New: "package calc // edited"},
		{Path: "calc/calc.go", Old: "x", New: "y"},
	})
	if err == nil || !strings.Contains(err.Error(), "2 matches") {
		t.Fatalf("err = %v, want an error that contains \"2 matches\"", err)
	}
	if got := readFile(t, wt, "calc/calc.go"); got != before {
		t.Errorf("file changed after a refused edit: %q", got)
	}
}

// scenario: patch with wrong line counts
func TestPatchWithWrongCounts(t *testing.T) {
	// Purpose: a model-written diff with wrong counts and spaces for tabs still applies.
	s, id, wt := claimTests(t)
	write(t, wt, "calc/add_test.go", addTest)
	write(t, wt, "calc/calc.go", "package calc\n\nfunc helper() int {\n\treturn 1\n}\n")
	commitAll(t, wt, "tests")
	patch := "--- a/calc/calc.go\n+++ b/calc/calc.go\n@@ -1,2 +1,3 @@\n package calc\n \n func helper() int {\n     return 1\n }\n+\n+// Add is a stub.\n+func Add(a, b int) int { return 0 }\n"
	rep, err := s.Submit(context.Background(), Submission{Agent: "tester", ID: id, Patch: patch, Message: "stub"})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("report = %+v, want accepted", rep)
	}
}

// scenario: edits in a submission
func TestSubmitEdits(t *testing.T) {
	// Purpose: work_submit can carry edits instead of whole files.
	s, id, wt := claimTests(t)
	write(t, wt, "calc/add_test.go", addTest)
	commitAll(t, wt, "tests")
	rep, err := s.Submit(context.Background(), Submission{Agent: "tester", ID: id, Message: "stub",
		Edits: []Edit{{Path: "calc/calc.go", Old: "package calc\n", New: "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Accepted {
		t.Fatalf("report = %+v, want accepted", rep)
	}
}
