package mcp

import (
	"context"
	"testing"
)

// scenario: work_edit is listed
func TestWorkEditListed(t *testing.T) {
	// Purpose: agents can find work_edit and the edits property of work_submit.
	found := map[string]map[string]any{}
	for _, tool := range (&Tools{}).List(context.Background()) {
		props, _ := tool.InputSchema["properties"].(map[string]any)
		found[tool.Name] = props
	}
	if _, ok := found["work_edit"]; !ok {
		t.Error("work_edit is not listed")
	}
	if _, ok := found["work_submit"]["edits"]; !ok {
		t.Error("work_submit has no edits property")
	}
}
