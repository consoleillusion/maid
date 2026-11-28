---
id: phases
title: Phases and gates
part: method
status: draft
style: ears
owner: paul
summary: >
  The fixed progression that maid guides every project through, from bootstrap
  to verified code. Defines each phase, its outputs, its gate, the depth levels
  and how maid_next tells an agent what to do.
links: [vision, work-items, acceptance, mcp-interface]
---

# Phases and gates

## Overview

| # | Phase | Main agent | Outputs | Gate |
|---|---|---|---|---|
| 0 | Bootstrap | design | project mode, depth, profile, `AGENTS.md`, `design/index.md` | REQ-PH-010 |
| 1 | Intent | design + user | vision, glossary, constraints | REQ-PH-011 |
| 2 | Requirements | design | requirement chapters | REQ-PH-012 |
| 3 | Architecture | design | component chapters, decisions | REQ-PH-013 |
| 4 | Interfaces | design | contracts and schemas | REQ-PH-014 |
| 5 | Plan | design, then test | work items with acceptance criteria, then acceptance tests | REQ-PH-015 |
| 6 | Build | build | accepted submissions | REQ-PH-016 |
| 7 | Verify | design | updated item states, generated `docs/` where configured | REQ-PH-017 |

Phases 6 and 7 repeat for each work item. A change to the specification during phase 6 goes back to the phase that owns the changed item.

## Requirements

### REQ-PH-001 Single entry point
When an agent calls `maid_next`, maid shall return the current phase, the next step, the reason for the step, the IDs to read, the tools to use, the gate conditions and the open questions for the user.

### REQ-PH-002 Hints in every response
The MCP server shall end the response of each tool with a one-line hint that names the next tool to call.

### REQ-PH-003 Phase state
The MCP server shall derive the current phase from the design store, the task store and the state store, and shall not depend on memory of a previous session.

### REQ-PH-004 Tool exposure by phase
While the project is in a phase before phase 5, the MCP server shall refuse calls to the work item tools and shall return the step from maid_next.

### REQ-PH-005 Tool list notification
When the phase changes, the MCP server shall send a `tools/list_changed` notification to each connected client that supports it.

### REQ-PH-006 Deterministic gates
The MCP server shall evaluate each gate condition with a deterministic check, and shall not accept a statement from an agent as evidence that a condition is met.

### REQ-PH-007 Approval is the user's choice
Where the user asks to approve the design, maid shall ask for one approval that covers every chapter of the phase; otherwise the design agent shall decide and record its decisions.

### REQ-PH-022 Plan all work at once
When the design agent plans, the design agent shall create work items for every accepted requirement that is not implemented, in one plan.

### REQ-PH-023 Current and target state
The design agent shall record the current state and the target state of each area side by side in the roadmap.

### REQ-PH-020 No repeated questions
When the user has answered a question or given an approval, maid shall record it, and agents shall not ask it again.

### REQ-PH-021 One message for all questions
The MCP server shall instruct agents to ask the user every question of a step in one message.

### REQ-PH-008 Project mode
When maid starts in a project without a design, maid shall detect whether the project is new, has existing code, or has an existing design in another format.

### REQ-PH-009 Existing code
Where the project has existing code, maid shall instruct the design agent to write draft design items from the code, and from `docs/` where a generator exists, for human acceptance.

### REQ-PH-010 Gate 0
The bootstrap gate shall require a project mode, a depth level and a profile.

### REQ-PH-011 Gate 1
The intent gate shall require an accepted vision with at least one goal and at least one non-goal, and a glossary.

### REQ-PH-012 Gate 2
The requirements gate shall require that each goal links to at least one accepted requirement and that each requirement passes its style validator.

### REQ-PH-013 Gate 3
The architecture gate shall require that each accepted requirement links to at least one component.

### REQ-PH-014 Gate 4
The interfaces gate shall require a contract for each boundary between components.

### REQ-PH-015 Gate 5
The plan gate shall require that each accepted requirement is linked to a work item, and that each work item has acceptance tests that fail on the current main branch.

### REQ-PH-016 Gate 6
The build gate for a work item shall be the acceptance rule of the merge queue. See REQ-ACC-007.

### REQ-PH-017 Gate 7
The verify gate shall require that each requirement in the work item traces to code and to passing tests.

### REQ-PH-018 Questions do not block unrelated work
While a question is open, maid shall block only the items that link to the question.

## Depth levels

The user selects a depth level in phase 0. The depth level sets which sections each gate requires.

| Level | For | Requires |
|---|---|---|
| sketch | prototypes | vision with one goal, a few requirements, no NFRs, no contracts |
| standard | most projects | all core chapters, contracts at component boundaries |
| rigorous | high-risk work | standard, plus NFRs, a threat model, and human approval at each gate and for each proposal |

### REQ-PH-019 Raise the depth level
When the user raises the depth level, maid shall report the new gaps through `maid_next` and shall not discard existing items.

## Example: maid_next response

```yaml
phase: 2-requirements
step: Define requirements for the area "merge queue"
why: Vision accepted; 1 goal (G-005) has no requirements
read: [vision#G-005, glossary, acceptance]
tools: [design_write, design_question]
done_when:
  - each goal links to at least one accepted requirement
  - each requirement passes the EARS validator
ask_user:
  - Q-004: Should a human approve merges in the review band?
```

## Clients that do not reload the tool list

Some MCP clients load the tool list once and ignore `tools/list_changed`. So maid lists every tool in every phase. Tools that the current phase does not allow refuse the call and return the next step (REQ-PH-004).

## Phases as tasks

The phases in this chapter are the order of the tasks that `maid init` seeds. They guide agents. They do not block work. See ADR-0008.

### REQ-PH-024 Task list
When an agent calls maid_next, maid shall list every task that is ready for the agent, which can be none, one or several.

### REQ-PH-025 Plain prompts
The MCP server shall instruct agents to call maid_next whenever the user asks them to work on the project.

### REQ-PH-026 Seeded tasks
When maid initialises a project, maid shall create tasks to record the current state, record the target state, write the design, plan the tasks and set up the project.

### REQ-PH-027 Design tasks
When an agent submits a task of kind design, the merge queue shall accept it when the design check reports no errors and the chapters that the task names exist.
