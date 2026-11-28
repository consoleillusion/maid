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
		Key: "add", Title: "Add two integers", Implements: []string{"calc#add"}, Criteria: crit,
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

// scenario: chapter path in the criteria
func TestPacketIncludesReferencedChapter(t *testing.T) {
	// Purpose: a design file path in the criteria puts the chapter text in the packet.
	pk := claimWithCriteria(t, "Follow design/01-a/calc.md\n\n"+criteria)
	got := pk.References["design/01-a/calc.md"]
	if !strings.Contains(got, "the calc package shall return their sum") {
		t.Errorf("References[design/01-a/calc.md] = %q, want the text of calc#add", got)
	}
}

// scenario: section ref in the criteria
func TestPacketIncludesReferencedSection(t *testing.T) {
	// Purpose: a chapter#anchor ref in the criteria puts the section text in the packet.
	pk := claimWithCriteria(t, "Follow calc#rounding.\n\n"+criteria)
	got := pk.References["calc#rounding"]
	if !strings.HasPrefix(got, "### Rounding") {
		t.Errorf("References[calc#rounding] = %q, want it to start with ### Rounding", got)
	}
}
