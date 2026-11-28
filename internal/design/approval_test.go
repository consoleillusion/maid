package design

import "testing"

// scenario: accept without an approver
func TestAcceptWithoutApprover(t *testing.T) {
	// Purpose: the agent can accept a chapter when the user delegated approval.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetStatus("alpha", "accepted"); err != nil {
		t.Fatalf("SetStatus without approver: %v", err)
	}
	if got := s.Chapter("alpha").FM.String("status"); got != "accepted" {
		t.Errorf("status = %q, want accepted", got)
	}
}
