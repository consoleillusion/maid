---
id: index
title: maid design index
part: design
status: draft
style: ste
owner: paul
summary: >
  The catalog of the maid design: every part and chapter with its status and
  summary, and the reading path for each role. maid will generate this file
  (REQ-DOC-014). Until then, keep it up to date by hand.
links: []
---

# maid design index

`design/` is the source of truth for maid. Read the summary of a chapter before you read the chapter. Read only what your task needs.

## Reading paths

| Role | Read first | Then |
|---|---|---|
| Human, first time | vision, phases | questions |
| Design agent | `AGENTS.md`, vision, glossary, chapter-format, the style card of the chapter you write | the chapters of your phase |
| Build agent | `AGENTS.md`, your context packet | conventions |
| Reviewer | the proposal, the item it changes, the linked decisions | — |

## Parts and chapters

### 01 Foundation

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [vision](01-foundation/01-vision.md) | draft | ste | Why maid exists, its goals, non-goals and success measures. |
| [glossary](01-foundation/02-glossary.md) | draft | sbvr-se | Defined terms and business rules. |
| [constraints](01-foundation/03-constraints.md) | draft | ste | Fixed constraints and testable assumptions. |

### 02 Method

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [phases](02-method/01-phases.md) | draft | ears | The progression from bootstrap to verified code, gates, depth levels, `maid_next`. |
| [work-items](02-method/02-work-items.md) | draft | ears | Work item fields, claims, leases, escalation, context packets. |
| [acceptance](02-method/03-acceptance.md) | draft | ears | The merge queue and the acceptance rule. |
| [scoring](02-method/04-scoring.md) | draft | ears | The 0–100 score that ranks passing attempts. |
| [presets](02-method/05-presets.md) | draft | ears | Versioned defaults, and default decisions that the user can change. |

### 03 Specification

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [chapter-format](03-specification/01-chapter-format.md) | draft | ears | Parts, chapters, sections, frontmatter, IDs, status and write rules. |
| [style-registry](03-specification/02-style-registry.md) | draft | ears | Registered styles, style cards, validators and enforcement. |

### 04 Architecture

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [architecture](04-architecture/01-overview.md) | draft | ste | Components, process model and state locations. |
| [task-store](04-architecture/02-task-store.md) | draft | ste | The task port, the Beads adapter and the spike findings. |

### 05 Interfaces

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [mcp-interface](05-interfaces/01-mcp.md) | draft | ears | Tools, resources, prompts and the `maid_next` contract. |

### 06 Engineering

| Chapter | Status | Style | Summary |
|---|---|---|---|
| [conventions](06-engineering/01-conventions.md) | draft | ste | How maid's own code is written, tested and committed. |

### Styles

| Card | Status | Summary |
|---|---|---|
| [ste](styles/ste.md) | draft | Simplified Technical English for prose. |
| [ears](styles/ears.md) | draft | EARS patterns for requirements. |
| [sbvr-se](styles/sbvr-se.md) | draft | SBVR Structured English for vocabulary and rules. |
| [gherkin](styles/gherkin.md) | draft | Gherkin for acceptance criteria. |
| [ace](styles/ace.md) | stub | Attempto Controlled English. Advice only. |

### Decisions

| ID | Status | Origin | Title |
|---|---|---|---|
| [ADR-0001](decisions/ADR-0001-go.md) | review | user | Write maid in Go and restart from the design |
| [ADR-0002](decisions/ADR-0002-pull-model.md) | review | user | Agents pull work from maid; maid does not start agents |
| [ADR-0003](decisions/ADR-0003-merge-queue.md) | review | user | Accept submissions automatically through a local merge queue |
| [ADR-0004](decisions/ADR-0004-acceptance-tests-first.md) | review | agent | The design agent writes complete acceptance tests before build work |
| [ADR-0005](decisions/ADR-0005-beads-task-store.md) | review | agent | Use Beads as the v1 task store, through its CLI, behind a port |
| [ADR-0006](decisions/ADR-0006-chapters-and-styles.md) | review | user | Organise the design as parts and chapters, each with a declared style |

New decisions use [the template](decisions/template.md).

### Open questions

See [questions.md](questions.md). 13 open.
