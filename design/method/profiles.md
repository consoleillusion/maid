---
id: profiles
title: Profiles and default decisions
part: method
status: accepted
style: ears
owner: paul
summary: >
  How maid fills decisions that the user does not make. Profiles are versioned
  sets of recommended defaults for stack, architecture, code style, tools,
  templates and score weights. The design agent can accept, customise or ignore
  a profile. Each choice becomes a decision record that the user can change.
links: [scoring, chapter-format, style-registry, ADR-0006]
---

# Profiles and default decisions

## Purpose

A profile is only a hint. The design agent decides every choice with the user, or alone when the user delegates the choice. The agent can use a profile, change it, or ignore it. Each choice is recorded as a decision.

## Layers

Defaults come from three layers. A more specific layer replaces a less specific layer.

1. maid built-in profiles.
2. User or organisation profiles.
3. Decisions in the project design.

## Profile contents

| Area | Examples |
|---|---|
| Kind | web service, CLI tool, library, script |
| Stack | runtime, language, storage, test framework |
| Architecture | style, declared boundaries |
| Code | code style, formatter, linter |
| Metrics | metrics adapter for each language |
| Documentation | depth level, chapter templates, styles for each chapter type, optional `docs/` generator |
| Commits | size budget, message convention |
| Work items | changed-line budget, attempts before escalation, parallel attempts |
| Score | weights, review band |

## Example

```yaml
profile: web-local
version: 1
kind: web
stack: { runtime: bun, language: typescript, storage: sqlite, test: bun-test }
architecture: { style: modular-monolith, boundaries: [storage, http, auth] }
code: { style: terse-functional, formatter: biome, linter: biome-recommended }
metrics: { typescript: fta }
docs:
  depth: standard
  styles: { vision: ste, glossary: sbvr-se, requirements: ears, acceptance: gherkin }
work_items: { max_changed_lines: 200, attempts: 2, parallel_attempts: 1 }
score: { weights: balanced, review_band: [50, 70] }
docs_generator: typedoc        # optional
```

The values in this example are illustrations.

## Built-in profiles

maid v1 ships several built-in profiles. The proposed set:

| Profile | Kind | Main defaults |
|---|---|---|
| `go-cli` | CLI tool | Go, gofmt, golangci-lint, `go test`, gomarkdoc |
| `web-local` | web service | Bun, TypeScript, SQLite, Biome, `bun test`, TypeDoc |
| `rust-perf` | performance-critical service, library or tool | Rust, rustfmt, clippy, `cargo test`, criterion benchmarks, rustdoc |
| `script-python` | script or tool | Python, uv, pytest, ruff, no docs generator |
| `generic` | any | No stack defaults. The design agent makes every decision. |

The final list and values are decided when the profiles are implemented. Each profile is a draft item until then.

## Requirements

### REQ-PR-001 Versioned profiles
Each profile shall have a name and an integer version, and a project shall reference a profile by name and version.

### REQ-PR-002 Default decision record
When maid applies a default for a choice that the user did not make, maid shall write a decision record with `origin: default`, the profile name and version, the reason, and the conditions to revisit the decision.

### REQ-PR-003 Origin
Each decision record shall have an `origin` field with one of the values `user`, `default` or `agent`.

### REQ-PR-004 Change of a default
When an agent wants to change a decision with `origin: default`, the agent shall write a new decision that supersedes it.

### REQ-PR-005 Change of a user decision
When an agent wants to change a decision with `origin: user`, the agent shall write a proposal and shall not change the decision.

### REQ-PR-006 Ports
Where a decision selects a component that the design marks as replaceable, the decision shall name the port in `design/` that the component implements.

### REQ-PR-007 Revisit conditions
When a requirement changes in a way that meets a revisit condition of a decision, maid shall report the decision through `maid_next`.

### REQ-PR-008 Metrics adapter selection
The design agent shall select a metrics adapter for each language in the project from the adapters that the profile allows, and shall record the selection as a decision.

### REQ-PR-009 Enforcement
maid shall run the formatter and the linter that the decisions select, as checks in the merge queue.

### REQ-PR-010 Recommendation
When the user does not give enough information for a decision, maid shall recommend the profile that best matches the project kind, with the reason for the match.

### REQ-PR-011 Agent choice
The design agent shall make each choice freely, with the user or alone when the user delegates it, and shall record each choice as a decision.

### REQ-PR-012 Origin of a customised value
If the design agent changes a profile value, then the decision for that value shall have `origin: agent` and shall name the profile value that it replaces.

### REQ-PR-013 No matching profile
If no profile matches the project, then maid shall recommend the `generic` profile.

### REQ-PR-014 Optional docs generator
Where the selected decisions name a `docs/` generator, maid shall run it after each accepted merge; otherwise maid shall not require `docs/`.
