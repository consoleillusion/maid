// Package score calculates the 0-100 score of an attempt.
// See design/02-method/04-scoring.md.
package score

// Inputs are the measured values of one attempt.
type Inputs struct {
	AcceptancePassed, AcceptanceTotal int
	Added, Deleted, Budget            int
	RuntimeDepsAdded, DevDepsAdded    int
	ComplexityAdded                   int
	BenchDeltaPct                     float64 // positive means slower
	HasBench, AllTestsPass            bool
}

// Weights are the component weights.
type Weights struct{ Size, Deps, Complexity, Reduction, Bench float64 }

// Component is one line of the score report.
type Component struct {
	Name                    string // acceptance, size, deps, complexity, reduction, bench
	Raw, Normalised, Weight float64
	Contribution            float64 // signed term added to P; 0 for acceptance
}

// Report is the score of an attempt with its components.
type Report struct {
	Score      float64
	Components []Component
}

// Balanced returns the balanced weights (REQ-SC-014).
func Balanced() Weights {
	// Purpose: stub; the code work item implements it.
	return Weights{}
}

// Compute calculates the score (REQ-SC-013).
func Compute(in Inputs, w Weights) Report {
	// Purpose: stub; the code work item implements it.
	return Report{}
}
