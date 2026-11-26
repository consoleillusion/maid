---
id: presets
title: Presets and default decisions
part: 02-method
status: draft
style: ears
owner: paul
summary: >
  How maid fills decisions that the user does not make. Presets are versioned
  sets of defaults for stack, architecture, code style, tools, templates and
  score weights. Each default becomes a decision record that the user can
  change.
links: [scoring, chapter-format, style-registry, ADR-0006]
---

# Presets and default decisions

## Layers

Defaults come from three layers. A more specific layer replaces a less specific layer.

1. maid built-in presets.
2. User or organisation presets.
3. Decisions in the project design.

## Preset contents

| Area | Examples |
|---|---|
| Kind | web service, CLI tool, library, script |
| Stack | runtime, language, storage, test framework |
| Architecture | style, declared boundaries |
| Code | code style, formatter, linter |
| Metrics | metrics adapter for each language |
| Documentation | depth level, chapter templates, styles for each chapter type |
| Commits | size budget, message convention |
| Work items | changed-line budget, attempts before escalation, parallel attempts |
| Score | weights, review band |

## Example

```yaml
preset: web-local
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
score: { preset: balanced, review_band: [50, 70] }
```

The values in this example are illustrations. The built-in preset list is TBD (Q-006).

## Requirements

### REQ-PR-001 Versioned presets
Each preset shall have a name and an integer version, and a project shall reference a preset by name and version.

### REQ-PR-002 Default decision record
When maid applies a default for a choice that the user did not make, maid shall write a decision record with `origin: default`, the preset name and version, the reason, and the conditions to revisit the decision.

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
The design agent shall select a metrics adapter for each language in the project from the adapters that the preset allows, and shall record the selection as a decision.

### REQ-PR-009 Enforcement
maid shall run the formatter and the linter that the decisions select, as checks in the merge queue.
