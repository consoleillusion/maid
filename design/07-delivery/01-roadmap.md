---
id: roadmap
title: Roadmap
part: 07-delivery
status: accepted
style: ste
owner: paul
summary: >
  The milestones of maid, the requirements that each milestone delivers, and
  what is implemented now. Read this to see where the project is and what
  comes next.
links: [phases, work-items, acceptance, mcp-interface, ADR-0007]
---

# Roadmap

Each area shows where the project is now and where it must be. The tasks for an area close the gap between the two columns. The design agent keeps the Now column true after each milestone.

## Milestone 1: maid builds maid

| Area | Now | Target | Requirements |
|---|---|---|---|
| Design store, checks and writes | Done | Agents read, check and change the design through maid | REQ-DOC-001 to REQ-DOC-013, REQ-DOC-015, REQ-STY-001, REQ-STY-003, REQ-STY-004, REQ-STY-006 |
| Style validators | STE length, EARS patterns, SBVR keywords, Gherkin structure | Same | REQ-STY-003 |
| Task store | Beads through its CLI, local only | Beads, committed and visible in git | REQ-WI-003, REQ-WI-004, REQ-WI-005, REQ-WI-029 |
| Plans and tests items | Done, with stubs | Same | REQ-WI-001, REQ-WI-017, REQ-WI-018, REQ-WI-022, REQ-WI-023 |
| Claims, packets and worktree tools | Done | Same | REQ-WI-002, REQ-WI-007, REQ-WI-013, REQ-WI-014, REQ-WI-019, REQ-WI-024 to REQ-WI-028 |
| Merge queue | Tests decide acceptance | Tests and score decide acceptance | REQ-ACC-001 to REQ-ACC-011, REQ-ACC-017, REQ-WI-008, REQ-WI-009, REQ-WI-020 |
| Guide | Phases, steps and ready calls; asks for a role | No roles; steps follow from the project state | REQ-PH-001 to REQ-PH-005, REQ-PH-021 |
| MCP server | stdio, tools only | Same | REQ-MCP-001, REQ-MCP-002, REQ-MCP-005 |
| Approval | Every acceptance needs approved_by | The user decides whether to approve or to delegate | REQ-PH-007 |

## Milestone 2: score and trial

| Area | Now | Target | Requirements |
|---|---|---|---|
| Score | Not built; 8 work items planned | Score in every report; minimum score rejects | REQ-SC-001 to REQ-SC-007, REQ-SC-013 to REQ-SC-016 |
| Score history | Not built | Score kept for each agent and model | REQ-SC-009, REQ-SC-010 |
| Calibration | Not built | Benchmark ranks good solutions above bad ones | REQ-SC-011, REQ-SC-012 |
| Trial | Not run | Acceptance rate and cost of cheap agents measured | SM-001, SM-002 |

## Milestone 3: any project

| Area | Now | Target | Requirements |
|---|---|---|---|
| Profiles | Specified only | Profiles offered as hints | REQ-PR-001 to REQ-PR-014 |
| New and existing projects | Bootstrap step only | Current and target state recorded for any project | REQ-PH-008, REQ-PH-009 |
| MCP extras | Not built | Resources, prompts, elicitation, HTTP | REQ-MCP-003, REQ-MCP-006 |
| Generated index and report | Not built | design/index.md generated; HTML report | REQ-DOC-014 |
| Robustness | Not built | Lease expiry, parallel attempts, flaky tests | REQ-WI-006, REQ-WI-011, REQ-ACC-012 to REQ-ACC-014 |
