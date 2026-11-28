package work

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/design"
)

// REQ-WI-007 scenario: no attempt branch
func TestClaimMakesNoBranch(t *testing.T) {
	// Purpose: an attempt works in a detached worktree; only main and maid-auto exist.
	s, root := project(t)
	pk, err := s.Claim(context.Background(), "tester", plan(t, s).Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	branches := strings.Fields(sh(t, root, "git", "branch", "--format=%(refname:short)"))
	sort.Strings(branches)
	if strings.Join(branches, " ") != "maid-auto main" {
		t.Errorf("branches = %v, want [maid-auto main]", branches)
	}
	if err := exec.Command("git", "-C", pk.Worktree, "symbolic-ref", "-q", "HEAD").Run(); err == nil {
		t.Error("the attempt worktree is on a branch, want a detached HEAD")
	}
}

// REQ-ACC-007 REQ-ACC-021 scenario: one commit per task on maid-auto
func TestAcceptedAttemptIsOneCommit(t *testing.T) {
	// Purpose: an accepted attempt becomes one commit with a task trailer on maid-auto only.
	s, root := project(t)
	before := sh(t, root, "git", "rev-parse", "main")
	id := plan(t, s).Tests["add"]
	pk, err := s.Claim(context.Background(), "tester", id)
	if err != nil {
		t.Fatal(err)
	}
	write(t, pk.Worktree, "calc/add_test.go", addTest)
	commitAll(t, pk.Worktree, "test: add the test")
	write(t, pk.Worktree, "calc/calc.go", "package calc\n\n// Add is a stub.\nfunc Add(a, b int) int { return 0 }\n")
	commitAll(t, pk.Worktree, "test: add the stub")
	rep, err := s.Submit(context.Background(), Submission{Agent: "tester", ID: id})
	if err != nil || !rep.Accepted {
		t.Fatalf("report = %+v, err = %v", rep, err)
	}
	if n := sh(t, root, "git", "rev-list", "--count", "main..maid-auto"); strings.TrimSpace(n) != "1" {
		t.Errorf("maid-auto has %s new commits, want 1", n)
	}
	if msg := sh(t, root, "git", "log", "-1", "--format=%B", "maid-auto"); !strings.Contains(msg, "Maid-Task: "+id) {
		t.Errorf("commit message %q has no Maid-Task trailer", msg)
	}
	if after := sh(t, root, "git", "rev-parse", "main"); after != before {
		t.Error("main moved")
	}
}

func createOp(path string) func(*design.Store) (design.WriteResult, error) {
	// Purpose: a design operation that creates one chapter.
	return func(ds *design.Store) (design.WriteResult, error) {
		return ds.CreateChapter(design.ChapterSpec{Path: path, ID: "extra", Title: "Extra", Style: "ste", Summary: "Extra.", Body: "# Extra\n\nText.\n"})
	}
}

// REQ-ACC-023 scenario: design change switches to the maid branch
func TestDesignChangeSwitchesToMaidBranch(t *testing.T) {
	// Purpose: a clean checkout is switched to maid-auto and the design commit lands there.
	s, root := project(t)
	if _, err := s.DesignOp(context.Background(), "a", "create", "add extra", createOp("b/extra.md")); err != nil {
		t.Fatal(err)
	}
	if cur := sh(t, root, "git", "symbolic-ref", "--short", "HEAD"); strings.TrimSpace(cur) != "maid-auto" {
		t.Errorf("checkout is on %s, want maid-auto", cur)
	}
	if subj := sh(t, root, "git", "log", "-1", "--format=%s", "maid-auto"); !strings.HasPrefix(subj, "design(") {
		t.Errorf("last commit on maid-auto is %q", subj)
	}
}

// REQ-ACC-023 scenario: uncommitted changes
func TestDesignChangeRefusedWithLocalChanges(t *testing.T) {
	// Purpose: maid does not switch a checkout that has uncommitted changes.
	s, root := project(t)
	if err := os.WriteFile(filepath.Join(root, "other/other.go"), []byte("package other\n\nfunc One() int { return 2 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.DesignOp(context.Background(), "a", "create", "add extra", createOp("b/extra.md"))
	if err == nil || !strings.Contains(err.Error(), "git switch maid-auto") {
		t.Fatalf("err = %v, want a refusal that names git switch maid-auto", err)
	}
}
