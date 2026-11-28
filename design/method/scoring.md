---
status: accepted
style: ears
owner: paul
summary: >
  How maid calculates the 0-100 score of an attempt from tests, lines of code,
  dependencies, complexity and benchmarks. The score ranks passing attempts and
  sets the review band. It never replaces the acceptance rule.
links: [acceptance, profiles, merge-queue]
---

# Score

## Purpose

Tests decide whether a submission is acceptable. The score decides which acceptable submission is best. Without size and dependency measures, a 1000-line solution and a 1-line solution that pass the same tests look the same. The score makes the difference visible.

## Components

| Component | Direction | Measure |
|---|---|---|
| Acceptance tests | reward | Fraction of the work item's acceptance tests that pass. |
| Regressions | penalty | Number of tests that passed on the baseline and fail now. Any value above 0 also fails acceptance. |
| Lines of code | penalty | Net changed lines, relative to the budget of the work item. |
| Dependencies | penalty | Added dependencies, weighted by scope (runtime more than dev). |
| Complexity | penalty | Change in complexity measures from the metrics adapter. |
| Benchmarks | reward or penalty | Change in configured benchmarks against the baseline. |
| Reduction | reward | Net removal of lines or dependencies while all tests pass. |
| Untraced code | penalty | Changed code with no link to a requirement. |

## Requirements

### Range
maid shall calculate the score of an attempt as a number from 0 to 100.

### Transparency
The score report shall list each component with its raw value, its normalised value, its weight and its contribution.

### Weights from the profile
maid shall read the component weights and the review band from the profile decision of the project.

### Score does not accept
maid shall not accept a submission because of its score. A submission is accepted only when it meets the acceptance rule in the acceptance chapter and its score is not below the minimum score of the profile.

### Minimum score
If the score of a submission that meets the acceptance rule is below the minimum score of the profile, then the merge queue shall reject the submission.

### Relative size
maid shall normalise the lines-of-code penalty by the changed-line budget of the work item, so that a work item with a larger budget is not penalised for the same relative size.

### Denied dependencies
If a submission adds a dependency that is on the deny list of the project, then the merge queue shall reject the submission.

### Language independence
maid shall receive metrics from a metrics adapter in a common format, and shall not contain language-specific metric logic.

### Docs work items
Where a work item has kind `docs`, maid shall score it with the docs component set from the profile, without lines-of-code and dependency penalties.

### History
maid shall store the score of each attempt with the agent, the model name reported by the agent, and the work item.

### Agent comparison
When a human asks for agent performance, maid shall report acceptance rate, mean score and mean attempts for each agent and model.

### Calibration benchmark
The maid repository shall contain a benchmark set of work items, each with known good and known bad solutions.

### Calibrated weights
Each built-in weight set shall rank every known good solution of the benchmark set above every known bad solution of the same work item.

## Formula

### Formula
The scorer shall calculate the score with the formula in this section.

```text
score = round1(100 × A × (1 − P))
P     = clamp(ws·size + wd·deps + wc·complexity − wr·reduction − wb·bench, 0, 1)

A          = acceptance tests passed / acceptance tests total; 1 when total is 0
size       = min(1, (added + deleted) / budget); 0 when budget ≤ 0
deps       = min(1, (2 × runtime deps added + dev deps added) / 4)
complexity = clamp(complexity added / 10, 0, 1)
reduction  = min(1, (deleted − added) / budget) when all tests pass and deleted > added; else 0
bench      = clamp(−bench delta % / 20, −1, 1) when benchmarks ran; else 0
round1     = round to one decimal place
```

### Balanced weights
The scorer shall use the balanced weights ws = 0.4, wd = 0.3, wc = 0.3, wr = 0.2 and wb = 0.2 when the project configures no weights.

### Examples
The scorer shall give the scores in the examples table for the given inputs.

| Passed / total | Added | Deleted | Budget | Runtime deps | All pass | Score |
|---|---|---|---|---|---|---|
| 2 / 2 | 10 | 0 | 100 | 0 | yes | 96.0 |
| 1 / 2 | 50 | 0 | 100 | 1 | no | 32.5 |
| 2 / 2 | 0 | 30 | 100 | 0 | yes | 94.0 |
| 0 / 0 | 0 | 0 | 0 | 0 | yes | 100.0 |

## Go interface

Package `internal/score` has no dependencies outside the standard library.

```go
type Inputs struct {
	AcceptancePassed, AcceptanceTotal int
	Added, Deleted, Budget            int
	RuntimeDepsAdded, DevDepsAdded    int
	ComplexityAdded                   int
	BenchDeltaPct                     float64 // positive means slower
	HasBench, AllTestsPass            bool
}
type Weights struct{ Size, Deps, Complexity, Reduction, Bench float64 }
type Component struct {
	Name                     string // acceptance, size, deps, complexity, reduction, bench
	Raw, Normalised, Weight  float64
	Contribution             float64 // signed term added to P; 0 for acceptance
}
type Report struct {
	Score      float64
	Components []Component // in the order of Name above
}
func Balanced() Weights
func Compute(in Inputs, w Weights) Report
```

Package `internal/metrics` gives the Go inputs:

```go
// GoDeps compares two go.mod files. Direct requirements count as runtime,
// requirements marked // indirect and tool directives count as dev.
func GoDeps(oldMod, newMod []byte) (runtimeAdded, devAdded int, added []string, err error)
// GoComplexity returns the cyclomatic complexity of each function in a Go
// file: 1, plus 1 for each if, for, range, case (not default), comm clause, && and ||.
func GoComplexity(src []byte) (map[string]int, error)
// ComplexityAdded sums new minus old for every function in either map.
func ComplexityAdded(oldFuncs, newFuncs map[string]int) int
```

`maid.json` configures the score:

```json
"score": { "weights": {"size": 0.4, "deps": 0.3, "complexity": 0.3, "reduction": 0.2, "bench": 0.2},
           "review_band": [50, 70], "deny_deps": ["example.com/banned"] }
```

## Any language

### Language-independent measures
Where no metrics adapter exists for a language, the scorer shall use the changed lines from git and the dependency changes from the manifest files go.mod, package.json, Cargo.toml, pyproject.toml and requirements.txt.
