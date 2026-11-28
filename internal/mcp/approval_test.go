package mcp

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/guide"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

// approvalProject creates a git repository with a draft vision chapter.
func approvalProject(t *testing.T) *Tools {
	// Purpose: set up a project whose vision waits for approval.
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.email", "t@example.invalid"}, {"config", "user.name", "t"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	tl := &Tools{Svc: work.New(root, config.Default(), tasks.NewMemory())}
	vision := `{"path":"a/vision.md","id":"vision","title":"Vision","style":"ste","summary":"Why.","body":"# Vision\n\n### Goal\nIt adds numbers.\n"}`
	if r := tl.Call(context.Background(), "design_create", json.RawMessage(vision), "a"); r["isError"] == true {
		t.Fatalf("design_create: %v", r)
	}
	return tl
}

// scenario: tool without approved_by
func TestSetStatusWithoutApprover(t *testing.T) {
	// Purpose: design_set_status accepts without approved_by and names the agent in the commit.
	tl := approvalProject(t)
	r := tl.Call(context.Background(), "design_set_status", json.RawMessage(`{"ids":["vision"],"status":"accepted"}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if r["isError"] == true || !strings.Contains(txt, `"ok": true`) {
		t.Fatalf("design_set_status without approved_by = %s", txt)
	}
	ds, err := tl.Svc.Design()
	if err != nil {
		t.Fatal(err)
	}
	if got := ds.Chapter("vision").FM.String("status"); got != "accepted" {
		t.Errorf("status = %q, want accepted", got)
	}
	out, _ := exec.Command("git", "-C", tl.Svc.Root, "log", "-1", "--format=%s").Output()
	if got := strings.TrimSpace(string(out)); got != "design: accepted vision (by a)" {
		t.Errorf("last commit subject = %q, want %q", got, "design: accepted vision (by a)")
	}
}

// scenario: the agent accepts chapters itself
func TestStepOffersDelegation(t *testing.T) {
	// Purpose: maid_next tells the agent to accept the chapters itself.
	tl := approvalProject(t)
	n, err := guide.Compute(context.Background(), tl.Svc, "", "a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(n.Step, "accept them all with ONE design_set_status call") {
		t.Errorf("step = %q", n.Step)
	}
}
