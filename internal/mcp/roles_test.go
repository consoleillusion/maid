package mcp

import (
	"context"
	"testing"
)

// scenario: no role in the schema
func TestNextHasNoRole(t *testing.T) {
	// Purpose: maid_next does not ask the agent for a role.
	tl := &Tools{}
	for _, tool := range tl.List(context.Background()) {
		if tool.Name != "maid_next" {
			continue
		}
		props, _ := tool.InputSchema["properties"].(map[string]any)
		if _, ok := props["role"]; ok {
			t.Fatal("maid_next still has a role property")
		}
		return
	}
	t.Fatal("maid_next not listed")
}
