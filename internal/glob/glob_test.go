package glob

import "testing"

func TestMatch(t *testing.T) {
	// Purpose: cover single-segment and multi-segment wildcards.
	cases := []struct {
		pat, p string
		want   bool
	}{
		{"internal/store/**", "internal/store/a.go", true},
		{"internal/store/**", "internal/store/x/y.go", true},
		{"internal/store/**", "internal/storex/a.go", false},
		{"internal/*/a.go", "internal/store/a.go", true},
		{"internal/*/a.go", "internal/x/y/a.go", false},
		{"**/*_test.go", "a/b/c_test.go", true},
		{"**/*_test.go", "c_test.go", true},
		{"go.mod", "go.mod", true},
		{"go.mod", "go.sum", false},
	}
	for _, c := range cases {
		if got := Match(c.pat, c.p); got != c.want {
			t.Errorf("Match(%q, %q) = %v", c.pat, c.p, got)
		}
	}
}
func TestOverlap(t *testing.T) {
	// Purpose: detect overlapping owned paths conservatively.
	if !Overlap([]string{"internal/store/**"}, []string{"internal/store/a.go"}) {
		t.Error("expected overlap")
	}
	if Overlap([]string{"internal/store/**"}, []string{"internal/queue/**"}) {
		t.Error("unexpected overlap")
	}
}
