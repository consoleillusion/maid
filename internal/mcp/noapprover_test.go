package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/guide"
)

// REQ-PH-007 scenario: no approver in the tool
func TestSetStatusHasNoApprover(t *testing.T) {
	// Purpose: design_set_status takes no approver.
	for _, tool := range (&Tools{}).List(context.Background()) {
		if tool.Name != "design_set_status" {
			continue
		}
		props, _ := tool.InputSchema["properties"].(map[string]any)
		if _, ok := props["approved_by"]; ok {
			t.Error("design_set_status has an approved_by property")
		}
		if strings.Contains(tool.Description, "approved_by") {
			t.Errorf("description mentions approved_by: %s", tool.Description)
		}
		return
	}
	t.Fatal("design_set_status is not listed")
}

// REQ-PH-007 scenario: no approver in the step
func TestStepDoesNotAskForApproval(t *testing.T) {
	// Purpose: the step for draft chapters tells the agent to accept them, without asking the user.
	tl := approvalProject(t)
	n, err := guide.Compute(context.Background(), tl.Svc, "", "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range n.Calls {
		if _, ok := c.Arguments["approved_by"]; ok {
			t.Errorf("call %s has an approved_by argument", c.Tool)
		}
	}
	if strings.Contains(strings.ToLower(n.Step), "approv") {
		t.Errorf("step asks for approval: %q", n.Step)
	}
	if len(n.AskUser) > 0 {
		t.Errorf("step asks the user: %v", n.AskUser)
	}
}
