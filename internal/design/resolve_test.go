package design

import (
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	// Purpose: accept chapter names, paths, titles, any case and heading text.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	cases := map[string]string{
		"alpha": "alpha", "Alpha": "alpha", "design/01-a/alpha.md": "alpha", "01-a/alpha.md": "alpha",
		"./design/01-a/alpha.md": "alpha", "01-a/alpha": "alpha",
		"alpha#Requirements": "alpha#requirements", "design/01-a/alpha.md#One": "alpha#one",
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
