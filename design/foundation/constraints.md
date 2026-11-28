---
status: accepted
style: ste
owner: paul
summary: >
  Fixed constraints on how maid is built and run, and the assumptions that the
  design depends on. Each assumption names the evidence that would make it
  false.
links: [vision, go, pull-model, beads-task-store]
---

# Constraints and assumptions

## Constraints

| Constraint | Source |
|---|---|
| maid is written in Go and ships as one binary for each platform. | Decision to write maid in Go. |
| maid uses the MCP pull model. maid does not start agents. | Decision that agents pull work. |
| maid must work with any MCP client that supports tools. Prompts, resources and elicitation are optional enhancements. | Goal: any agent, any language. |
| maid must not require a network service to run. All state is in the project repository. Tasks are shared through the project's git remote. | Goals: reuse existing tools; no network service. |
| maid v1 runs on Linux and macOS. The design must not prevent Windows support later. | Answer recorded in this chapter. |
| maid's own license is MIT. Linked dependencies must have licenses that are compatible with MIT. | README. |
| Each MCP client can start its own maid process. All shared state must be safe for many processes at the same time. | Decision that agents pull work. |

## Assumptions

| Assumption | False if |
|---|---|
| A small model can complete a work item reliably when it has a context packet and failing acceptance tests. | The small-agent acceptance rate is below its target in trials. |
| Acceptance tests that a large model writes are cheaper than the implementation and catch most wrong implementations. | Accepted work items are later found wrong more often than the rate set after the calibration benchmark. |
| Beads can be the task store for v1. | Concurrent claims fail, or Beads changes its CLI contract often. See `spikes/beads/README.md`. |
| git worktrees give enough isolation for parallel attempts on one machine. | Test suites need shared resources that the attempt environment and serial test rules do not cover. |
| MCP clients pass the user's answers to maid without change. | Clients without elicitation change or skip answers. |
