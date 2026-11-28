---
status: accepted
style: ears
owner: paul
summary: >
  The fixed progression that maid guides every project through, from bootstrap
  to verified code. Defines each phase, its outputs, its gate, the depth levels
  and how maid_next tells an agent what to do.
links: [vision, work-items, acceptance, mcp]
---

# Phases and gates

## Overview

| Phase | Main agent | Outputs | Gate |
|---|---|---|---|
| Bootstrap | design | project mode, depth, profile, `AGENTS.md`, `design/index.md` | Bootstrap gate |
| Intent | design + user | vision, glossary, constraints | Intent gate |
| Requirements | design | requirement chapters | Requirements gate |
| Architecture | design | component chapters, decisions | Architecture gate |
| Interfaces | design | contracts and schemas | Interfaces gate |
| Plan | design, then test | work items with acceptance criteria, then acceptance tests | Plan gate |
| Build | build | accepted submissions | Build gate |
| Verify | design | updated item states, generated `docs/` where configured | Verify gate |

Build and verify repeat for each work item. A change to the specification during build goes back to the phase that owns the changed item.

## Requirements

### Single entry point
When an agent calls `maid_next`, maid shall return the current phase, the next step, the reason for the step, the sections to read, the tools to use, the gate conditions and the questions for the user.

### Hints in every response
The MCP server shall end the response of each tool with a one-line hint that names the next tool to call.

### Phase state
The MCP server shall derive the current phase from the design store, the task store and the state store, and shall not depend on memory of a previous session.

### Tool exposure by phase
While the project is in a phase before plan, the MCP server shall refuse calls to the work item tools and shall return the step from maid_next.

### Tool list notification
When the phase changes, the MCP server shall send a `tools/list_changed` notification to each connected client that supports it.

### Deterministic gates
The MCP server shall evaluate each gate condition with a deterministic check, and shall not accept a statement from an agent as evidence that a condition is met.

### Accepting chapters
When an agent sets a chapter to accepted, maid shall accept the chapter if it has no check errors, and shall not ask for or record an approver.

### Plan all work at once
When the design agent plans, the design agent shall create work items for every accepted requirement that is not implemented, in one plan.

### Current and target state
The design agent shall record the current state and the target state of each area side by side in the roadmap.

### No repeated questions
When the user has answered a question or given an approval, maid shall record the answer in the design, and agents shall not ask it again.

### One message for all questions
The MCP server shall instruct agents to ask the user every question of a step in one message.

### Project mode
When maid starts in a project without a design, maid shall detect whether the project is new, has existing code, or has an existing design in another format.

### Existing code
Where the project has existing code, maid shall instruct the design agent to write draft design sections from the code, and from `docs/` where a generator exists, for human acceptance.

### Bootstrap gate
The bootstrap gate shall require a project mode, a depth level and a profile.

### Intent gate
The intent gate shall require an accepted vision with at least one goal and at least one non-goal, and a glossary.

### Requirements gate
The requirements gate shall require that each goal links to at least one accepted requirement and that each requirement passes its style validator.

### Architecture gate
The architecture gate shall require that each accepted requirement links to at least one component.

### Interfaces gate
The interfaces gate shall require a contract for each boundary between components.

### Plan gate
The plan gate shall require that each accepted requirement is linked to a work item, and that each work item has acceptance tests that fail on the current main branch.

### Build gate
The build gate for a work item shall be the acceptance rule of the merge queue.

### Verify gate
The verify gate shall require that each requirement in the work item traces to code and to passing tests.

### Missing information
If information is missing, then the agent shall ask the user, shall record the answer in the design, and shall continue with work that does not need the answer.

## Depth levels

The user selects a depth level during bootstrap. The depth level sets which sections each gate requires.

| Level | For | Requires |
|---|---|---|
| sketch | prototypes | vision with one goal, a few requirements, no non-functional requirements, no contracts |
| standard | most projects | all core chapters, contracts at component boundaries |
| rigorous | high-risk work | standard, plus non-functional requirements, a threat model, and human approval at each gate and for each proposal |

### Raise the depth level
When the user raises the depth level, maid shall report the new gaps through `maid_next` and shall not discard existing items.

## Example: maid_next response

```yaml
phase: requirements
step: Define requirements for the area "merge queue"
why: Vision accepted; the goal "mechanical acceptance" has no requirements
read: [vision, glossary, acceptance]
tools: [design_write, design_question]
done_when:
  - each goal links to at least one accepted requirement
  - each requirement passes the EARS validator
ask_user:
  - Should a human approve merges in the review band?
```

## Clients that do not reload the tool list

Some MCP clients load the tool list once and ignore `tools/list_changed`. So maid lists every tool in every phase. Tools that the current phase does not allow refuse the call and return the next step.

## Phases as tasks

The phases in this chapter are the order of the tasks that `maid init` seeds. They guide agents. They do not block work. See the decision to treat everything as a task.

### Task list
When an agent calls maid_next, maid shall list every task that is ready for the agent, which can be none, one or several.

### Plain prompts
The MCP server shall instruct agents to call maid_next whenever the user asks them to work on the project.

### Seeded tasks
When maid initialises a project, maid shall create tasks to record the current state, record the target state, write the design, plan the tasks and set up the project.

### Design tasks
When an agent submits a task of kind design, the merge queue shall accept it when the design check reports no errors and the chapters that the task names exist.
