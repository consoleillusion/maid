package work

import (
	"context"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/tasks"
)

// claimWithCriteria plans one item with the given criteria and claims its tests item.
func claimWithCriteria(t *testing.T, crit string) Packet {
	// Purpose: build the packet that an agent gets for criteria that name design refs.
	t.Helper()
	s, _ := project(t)
	ctx := context.Background()
	p, err := s.CreatePlan(ctx, Plan{Items: []PlanItem{{
		Key: "add", Title: "Add two integers", Implements: []string{"REQ-CA-001"}, Criteria: crit,
		AcceptanceTests: []tasks.TestRef{{File: "calc/add_test.go", Names: []string{"TestAdd"}}},
		Owns:            []string{"calc/calc.go"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	pk, err := s.Claim(ctx, "tester", p.Tests["add"])
	if err != nil {
		t.Fatal(err)
	}
	return pk
}

// REQ-WI-030 scenario: chapter path in the criteria
func TestPacketIncludesReferencedChapter(t *testing.T) {
	// Purpose: a design file path in the criteria puts the chapter text in the packet.
	pk := claimWithCriteria(t, "Follow design/01-a/calc.md\n\n"+criteria)
	got := pk.References["design/01-a/calc.md"]
	if !strings.Contains(got, "the calc package shall return their sum") {
		t.Errorf("References[design/01-a/calc.md] = %q, want the text of REQ-CA-001", got)
	}
}

// REQ-WI-030 scenario: section ref in the criteria
func TestPacketIncludesReferencedSection(t *testing.T) {
	// Purpose: a chapter#anchor ref in the criteria puts the section text in the packet.
	pk := claimWithCriteria(t, "Follow calc#REQ-CA-001.\n\n"+criteria)
	got := pk.References["calc#REQ-CA-001"]
	if !strings.HasPrefix(got, "### REQ-CA-001") {
		t.Errorf("References[calc#REQ-CA-001] = %q, want it to start with ### REQ-CA-001", got)
	}
}
