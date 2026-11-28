// Package score calculates the 0-100 score of an attempt.
// See design/02-method/04-scoring.md.
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
	Name                    string
	Raw, Normalised, Weight float64
	Contribution            float64
}

// Report is the score of an attempt with its components.
type Report struct {
	Score      float64
	Components []Component
}

// Balanced returns the balanced weights (REQ-SC-014).
func Balanced() Weights {
	return Weights{Size: 0.4, Deps: 0.3, Complexity: 0.3, Reduction: 0.2, Bench: 0.2}
}

// Compute calculates the score (REQ-SC-013).
func Compute(in Inputs, w Weights) Report {
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

func acceptanceComponent(in Inputs) Component {
	a := 1.0
	if in.AcceptanceTotal > 0 {
		a = float64(in.AcceptancePassed) / float64(in.AcceptanceTotal)
	}
	return Component{Name: "acceptance", Raw: a, Normalised: a}
}

func sizeComponent(in Inputs, weight float64) Component {
	raw := 0.0
	if in.Budget > 0 {
		raw = float64(in.Added+in.Deleted) / float64(in.Budget)
	}
	return penalty("size", raw, math.Min(1, raw), weight)
}

func depsComponent(in Inputs, weight float64) Component {
	raw := float64(2*in.RuntimeDepsAdded + in.DevDepsAdded)
	return penalty("deps", raw, math.Min(1, raw/4), weight)
}

func complexityComponent(in Inputs, weight float64) Component {
	raw := float64(in.ComplexityAdded)
	return penalty("complexity", raw, clamp(raw/10, 0, 1), weight)
}

func reductionComponent(in Inputs, weight float64) Component {
	raw := float64(in.Deleted - in.Added)
	norm := 0.0
	if in.AllTestsPass && in.Deleted > in.Added && in.Budget > 0 {
		norm = math.Min(1, raw/float64(in.Budget))
	}
	return reward("reduction", raw, norm, weight)
}

func benchComponent(in Inputs, weight float64) Component {
	norm := 0.0
	if in.HasBench {
		norm = clamp(-in.BenchDeltaPct/20, -1, 1)
	}
	return reward("bench", in.BenchDeltaPct, norm, weight)
}

func penalty(name string, raw, norm, weight float64) Component {
	return Component{Name: name, Raw: raw, Normalised: norm, Weight: weight, Contribution: weight * norm}
}

func reward(name string, raw, norm, weight float64) Component {
	return Component{Name: name, Raw: raw, Normalised: norm, Weight: weight, Contribution: -weight * norm}
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}