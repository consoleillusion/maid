// Package score calculates the 0-100 score of an attempt.
// See design/method/scoring.md.
package score

import "math"

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

// Balanced returns the balanced weights.
func Balanced() Weights {
	// Purpose: give the default weights used when a project sets none.
	return Weights{Size: 0.4, Deps: 0.3, Complexity: 0.3, Reduction: 0.2, Bench: 0.2}
}

// Compute calculates the score.
func Compute(in Inputs, w Weights) Report {
	// Purpose: combine the components into a score from 0 to 100.
	acc := acceptanceComponent(in)
	comps := []Component{
		acc,
		sizeComponent(in, w.Size),
		depsComponent(in, w.Deps),
		complexityComponent(in, w.Complexity),
		reductionComponent(in, w.Reduction),
		benchComponent(in, w.Bench),
	}
	p := 0.0
	for _, c := range comps {
		p += c.Contribution
	}
	p = clamp(p, 0, 1)
	return Report{Score: round1(100 * acc.Normalised * (1 - p)), Components: comps}
}

// acceptanceComponent reports the fraction of acceptance tests that pass.
func acceptanceComponent(in Inputs) Component {
	// Purpose: report the acceptance test fraction, 1 when there are none.
	a := 1.0
	if in.AcceptanceTotal > 0 {
		a = float64(in.AcceptancePassed) / float64(in.AcceptanceTotal)
	}
	return Component{Name: "acceptance", Raw: a, Normalised: a}
}

// sizeComponent penalises changed lines relative to the work item budget.
func sizeComponent(in Inputs, weight float64) Component {
	// Purpose: build the size penalty from changed lines and budget.
	raw := 0.0
	if in.Budget > 0 {
		raw = float64(in.Added+in.Deleted) / float64(in.Budget)
	}
	return penalty("size", raw, math.Min(1, raw), weight)
}

// depsComponent penalises added dependencies, runtime counting double.
func depsComponent(in Inputs, weight float64) Component {
	// Purpose: build the dependency penalty.
	raw := float64(2*in.RuntimeDepsAdded + in.DevDepsAdded)
	return penalty("deps", raw, math.Min(1, raw/4), weight)
}

// complexityComponent penalises the change in cyclomatic complexity.
func complexityComponent(in Inputs, weight float64) Component {
	// Purpose: build the complexity penalty.
	raw := float64(in.ComplexityAdded)
	return penalty("complexity", raw, clamp(raw/10, 0, 1), weight)
}

// reductionComponent rewards net removal of lines when all tests pass.
func reductionComponent(in Inputs, weight float64) Component {
	// Purpose: build the reduction reward.
	raw := float64(in.Deleted - in.Added)
	norm := 0.0
	if in.AllTestsPass && in.Deleted > in.Added && in.Budget > 0 {
		norm = math.Min(1, raw/float64(in.Budget))
	}
	return reward("reduction", raw, norm, weight)
}

// benchComponent rewards a faster benchmark and penalises a slower one.
func benchComponent(in Inputs, weight float64) Component {
	// Purpose: build the benchmark term, zero when no benchmark ran.
	norm := 0.0
	if in.HasBench {
		norm = clamp(-in.BenchDeltaPct/20, -1, 1)
	}
	return reward("bench", in.BenchDeltaPct, norm, weight)
}

// penalty builds a component whose normalised value adds to P.
func penalty(name string, raw, norm, weight float64) Component {
	// Purpose: build a component that adds weight times norm to P.
	return Component{Name: name, Raw: raw, Normalised: norm, Weight: weight, Contribution: weight * norm}
}

// reward builds a component whose normalised value subtracts from P.
func reward(name string, raw, norm, weight float64) Component {
	// Purpose: build a component that subtracts weight times norm from P.
	return Component{Name: name, Raw: raw, Normalised: norm, Weight: weight, Contribution: -weight * norm}
}

// clamp keeps v between lo and hi.
func clamp(v, lo, hi float64) float64 {
	// Purpose: limit v to the range lo to hi.
	return math.Max(lo, math.Min(hi, v))
}

// round1 rounds v to one decimal place.
func round1(v float64) float64 {
	// Purpose: round v to one decimal place.
	return math.Round(v*10) / 10
}
