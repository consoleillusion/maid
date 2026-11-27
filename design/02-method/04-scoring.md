---
id: scoring
title: Score
part: 02-method
status: accepted
style: ears
owner: paul
summary: >
  How maid calculates the 0-100 score of an attempt from tests, lines of code,
  dependencies, complexity and benchmarks. The score ranks passing attempts and
  sets the review band. It never replaces the acceptance rule.
links: [acceptance, profiles, ADR-0003]
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

### REQ-SC-001 Range
maid shall calculate the score of an attempt as a number from 0 to 100.

### REQ-SC-002 Transparency
The score report shall list each component with its raw value, its normalised value, its weight and its contribution.

### REQ-SC-003 Weights from the profile
maid shall read the component weights and the review band from the profile decision of the project.

### REQ-SC-004 Score does not accept
maid shall not accept a submission because of its score. The acceptance rule in REQ-ACC-007 is the only condition for acceptance.

### REQ-SC-005 Relative size
maid shall normalise the lines-of-code penalty by the changed-line budget of the work item, so that a work item with a larger budget is not penalised for the same relative size.

### REQ-SC-006 Denied dependencies
If a submission adds a dependency that is on the deny list of the project, then the merge queue shall reject the submission.

### REQ-SC-007 Language independence
maid shall receive metrics from a metrics adapter in a common format, and shall not contain language-specific metric logic.

### REQ-SC-008 Docs work items
Where a work item has kind `docs`, maid shall score it with the docs component set from the profile, without lines-of-code and dependency penalties.

### REQ-SC-009 History
maid shall store the score of each attempt with the agent, the model name reported by the agent, and the work item.

### REQ-SC-010 Agent comparison
When a human asks for agent performance, maid shall report acceptance rate, mean score and mean attempts for each agent and model.

### REQ-SC-011 Calibration benchmark
The maid repository shall contain a benchmark set of work items, each with known good and known bad solutions.

### REQ-SC-012 Calibrated weights
Each built-in weight set shall rank every known good solution of the benchmark set above every known bad solution of the same work item.
