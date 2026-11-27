package score

import (
	"math"
	"testing"
)

const eps = 1e-9

func near(a, b float64) bool {
	// Purpose: compare floating point values.
	return math.Abs(a-b) < eps
}

// REQ-SC-015 REQ-SC-013 scenario: examples table
func TestComputeExamples(t *testing.T) {
	// Purpose: every row of the examples table gives its score.
	cases := []struct {
		name string
		in   Inputs
		want float64
	}{
		{"small change", Inputs{AcceptancePassed: 2, AcceptanceTotal: 2, Added: 10, Budget: 100, AllTestsPass: true}, 96.0},
		{"half passing with a dependency", Inputs{AcceptancePassed: 1, AcceptanceTotal: 2, Added: 50, Budget: 100, RuntimeDepsAdded: 1}, 32.5},
		{"reduction", Inputs{AcceptancePassed: 2, AcceptanceTotal: 2, Deleted: 30, Budget: 100, AllTestsPass: true}, 94.0},
		{"no tests, no budget", Inputs{AllTestsPass: true}, 100.0},
	}
	for _, c := range cases {
		if got := Compute(c.in, Balanced()).Score; !near(got, c.want) {
			t.Errorf("%s: score = %v, want %v", c.name, got, c.want)
		}
	}
}

// REQ-SC-002 REQ-SC-005 scenario: components are reported
func TestComputeComponents(t *testing.T) {
	// Purpose: the report lists six components in order with their values.
	r := Compute(Inputs{AcceptancePassed: 2, AcceptanceTotal: 2, Added: 10, Budget: 100, AllTestsPass: true}, Balanced())
	names := []string{"acceptance", "size", "deps", "complexity", "reduction", "bench"}
	if len(r.Components) != len(names) {
		t.Fatalf("got %d components, want %d", len(r.Components), len(names))
	}
	for i, n := range names {
		if r.Components[i].Name != n {
			t.Errorf("component %d = %q, want %q", i, r.Components[i].Name, n)
		}
	}
	size := r.Components[1]
	if !near(size.Raw, 0.1) || !near(size.Normalised, 0.1) || !near(size.Weight, 0.4) || !near(size.Contribution, 0.04) {
		t.Errorf("size component = %+v", size)
	}
}

// REQ-SC-001 REQ-SC-013 scenario: penalty and reward limits
func TestComputeLimits(t *testing.T) {
	// Purpose: penalties are capped, so the score never goes below 0.
	in := Inputs{AcceptancePassed: 1, AcceptanceTotal: 1, Added: 1000, Budget: 100, RuntimeDepsAdded: 5, ComplexityAdded: 50}
	r := Compute(in, Balanced())
	if !near(r.Score, 0) || len(r.Components) != 6 {
		t.Errorf("score = %v with %d components, want 0 with 6", r.Score, len(r.Components))
	}
}

// REQ-SC-014 scenario: balanced weights
func TestBalanced(t *testing.T) {
	// Purpose: the balanced weights have the values in the design.
	want := Weights{Size: 0.4, Deps: 0.3, Complexity: 0.3, Reduction: 0.2, Bench: 0.2}
	if got := Balanced(); got != want {
		t.Errorf("Balanced() = %+v, want %+v", got, want)
	}
}
