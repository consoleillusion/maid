package work

import (
	"context"
	"strings"
	"testing"
)

// REQ-WI-024 REQ-WI-025 REQ-WI-026 REQ-WI-027
func TestWorkThroughMaidOnly(t *testing.T) {
	// Purpose: an agent with no file access completes the loop with maid's file tools.
	s, root := project(t)
	ctx := context.Background()
	p := plan(t, s)

	pk, err := s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	id := pk.WorkItem.ID
	if err := s.WriteFile(ctx, "tester", id, "other/other.go", "package other\n", false); err == nil || !strings.Contains(err.Error(), "not owned") {
		t.Fatalf("write outside owns: %v", err)
	}
	if err := s.WriteFile(ctx, "tester", id, "../escape.go", "x", false); err == nil {
		t.Fatal("write outside the worktree accepted")
	}
	if err := s.WriteFile(ctx, "intruder", id, "calc/add_test.go", "x", false); err == nil {
		t.Fatal("write without the claim accepted")
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(s.WriteFile(ctx, "tester", id, "calc/add_test.go", addTest, false))
	must(s.WriteFile(ctx, "tester", id, "calc/calc.go", "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n", false))
	r, err := s.Run(ctx, "tester", id, "test")
	must(err)
	if r.Passed || r.Tests["example.com/p/calc#TestAdd"] != "fail" {
		t.Fatalf("run = %+v", r)
	}
	if _, err := s.Commit(ctx, "tester", id, "tests and stub"); err != nil {
		t.Fatal(err)
	}
	if rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: id}); err != nil || !rep.Accepted {
		t.Fatalf("tests item: %v %+v", err, rep)
	}

	pk, err = s.Claim(ctx, "builder", p.Code["add"])
	must(err)
	id = pk.WorkItem.ID
	files, err := s.Files(ctx, "builder", id)
	must(err)
	owned := ""
	for _, f := range files {
		if f.Owned {
			owned += f.Path
		}
	}
	if owned != "calc/calc.go" {
		t.Fatalf("owned files = %q", owned)
	}
	src, err := s.ReadFile(ctx, "builder", id, "calc/add_test.go")
	if err != nil || !strings.Contains(src, "TestAdd") {
		t.Fatalf("read test: %v", err)
	}
	must(s.WriteFile(ctx, "builder", id, "calc/calc.go", "package calc\n\n// Add returns a+b.\nfunc Add(a, b int) int { return a + b }\n", false))
	if r, _ := s.Run(ctx, "builder", id, "test"); !r.Passed {
		t.Fatalf("tests fail after implementation: %+v", r)
	}
	if _, err := s.Commit(ctx, "builder", id, "implement Add"); err != nil {
		t.Fatal(err)
	}
	rep, err := s.Submit(ctx, Submission{Agent: "builder", ID: id})
	if err != nil || !rep.Accepted {
		t.Fatalf("code item: %v %+v", err, rep)
	}
	if got := sh(t, root, "git", "show", "main:calc/calc.go"); !strings.Contains(got, "return a + b") {
		t.Fatalf("main = %s", got)
	}
}

// REQ-WI-028
func TestSubmitWithPatchAndFiles(t *testing.T) {
	// Purpose: a test agent submits files and a build agent submits a patch, with no other tools.
	s, root := project(t)
	ctx := context.Background()
	p := plan(t, s)
	if _, err := s.Claim(ctx, "tester", p.Tests["add"]); err != nil {
		t.Fatal(err)
	}
	rep, err := s.Submit(ctx, Submission{Agent: "tester", ID: p.Tests["add"], Message: "tests and stub", Files: map[string]string{
		"calc/add_test.go": addTest,
		"calc/calc.go":     "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n",
	}})
	if err != nil || !rep.Accepted {
		t.Fatalf("files submission: %v %+v", err, rep)
	}
	if _, err := s.Claim(ctx, "builder", p.Code["add"]); err != nil {
		t.Fatal(err)
	}
	bad := "--- a/other/other.go\n+++ b/other/other.go\n@@ -1,3 +1,3 @@\n package other\n \n-func One() int { return 1 }\n+func One() int { return 2 }\n"
	if _, err := s.Submit(ctx, Submission{Agent: "builder", ID: p.Code["add"], Patch: bad}); err == nil || !strings.Contains(err.Error(), "does not own") {
		t.Fatalf("patch outside owns: %v", err)
	}
	patch := "--- a/calc/calc.go\n+++ b/calc/calc.go\n@@ -1,4 +1,4 @@\n package calc\n \n-// Add is a stub.\n-func Add(a, b int) int { return 0 }\n+// Add returns a+b.\n+func Add(a, b int) int { return a + b }\n"
	rep, err = s.Submit(ctx, Submission{Agent: "builder", ID: p.Code["add"], Patch: patch, Message: "implement Add"})
	if err != nil || !rep.Accepted {
		t.Fatalf("patch submission: %v %+v", err, rep)
	}
	if got := sh(t, root, "git", "show", "main:calc/calc.go"); !strings.Contains(got, "return a + b") {
		t.Fatalf("main = %s", got)
	}
}
