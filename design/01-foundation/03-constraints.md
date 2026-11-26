---
id: constraints
title: Constraints and assumptions
part: 01-foundation
status: draft
style: ste
owner: paul
summary: >
  Fixed constraints on how maid is built and run, and the assumptions that the
  design depends on. Each assumption names the evidence that would make it
  false.
links: [vision, ADR-0001, ADR-0002, ADR-0005]
---

# Constraints and assumptions

## Constraints

| ID | Constraint | Source |
|---|---|---|
| C-001 | maid is written in Go and ships as one binary for each platform. | ADR-0001 |
| C-002 | maid uses the MCP pull model. maid does not start agents. | ADR-0002 |
| C-003 | maid must work with any MCP client that supports tools. Prompts, resources and elicitation are optional enhancements. | G-008 |
| C-004 | maid must not require a network service to run. All state is in the project folder. | G-009, NG-002 |
| C-005 | maid runs on Linux, macOS and Windows. | TBD (Q-002) |
| C-006 | maid's own license is MIT. Linked dependencies must have licenses that are compatible with MIT. | README |
| C-007 | Each MCP client can start its own maid process. All shared state must be safe for many processes at the same time. | ADR-0002 |

## Assumptions

| ID | Assumption | False if |
|---|---|---|
| A-001 | A small model can complete a work item reliably when it has a context packet and failing acceptance tests. | SM-001 is below its target in trials. |
| A-002 | Acceptance tests that a large model writes are cheaper than the implementation and catch most wrong implementations. | Accepted work items are later found wrong more often than TBD (Q-001). |
| A-003 | Beads can be the task store for v1. | Concurrent claims fail, or Beads changes its CLI contract often. See `spikes/beads/README.md`. |
| A-004 | git worktrees give enough isolation for parallel attempts on one machine. | Test suites of target projects need shared resources such as fixed ports or one database. |
| A-005 | MCP clients pass the user's answers to maid's questions without change. | Clients without elicitation change or skip questions. |
