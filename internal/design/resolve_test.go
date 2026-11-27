package design

import (
	"strings"
	"testing"
)

// REQ-DOC-008
func TestResolve(t *testing.T) {
	// Purpose: accept IDs, paths, titles, any case and heading text.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	cases := map[string]string{
		"alpha": "alpha", "Alpha": "alpha", "design/01-a/alpha.md": "alpha", "01-a/alpha.md": "alpha",
		"./design/01-a/alpha.md": "alpha", "01-a/alpha": "alpha", "req-al-002": "REQ-AL-002",
		"alpha#Requirements": "alpha#requirements", "design/01-a/alpha.md#REQ-AL-001": "alpha#REQ-AL-001",
		"beta#q-001": "beta#Q-001",
	}
	for in, want := range cases {
		got, err := s.Resolve(in)
		if err != nil || got != want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	_, err := s.Resolve("alpah")
	if err == nil || !strings.Contains(err.Error(), "did you mean: alpha") {
		t.Fatalf("typo error = %v", err)
	}
	ex, missing := s.ReadAll([]string{"design/01-a/alpha.md", "nope-chapter"}, true)
	if len(ex) != 1 || ex[0].Ref != "alpha" || len(missing) != 1 || missing[0].Ref != "nope-chapter" {
		t.Fatalf("ReadAll = %+v / %+v", ex, missing)
	}
}
