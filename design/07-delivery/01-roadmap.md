---
id: roadmap
title: Roadmap
part: 07-delivery
status: draft
style: ste
owner: paul
summary: >
  The milestones of maid, the requirements that each milestone delivers, and
  what is implemented now. Read this to see where the project is and what
  comes next.
links: [phases, work-items, acceptance, mcp-interface, ADR-0007]
---

# Roadmap

## Milestone 1: maid builds maid

Goal: the smallest maid that can manage its own build with cheap agents.

| Area | Requirements | State |
|---|---|---|
| Design store, checks and writes | REQ-DOC-001 to REQ-DOC-013, REQ-DOC-015, REQ-STY-001, REQ-STY-003, REQ-STY-004, REQ-STY-006 | implemented |
| Style validators | STE length rules, EARS patterns, SBVR rule keywords, Gherkin structure | implemented |
| Task port and Beads adapter | REQ-WI-003, REQ-WI-004, REQ-WI-005 | implemented |
| Plans with tests items | REQ-WI-001, REQ-WI-017, REQ-WI-018 | implemented |
| Claims and context packets | REQ-WI-002, REQ-WI-007, REQ-WI-013, REQ-WI-014, REQ-WI-019 | implemented |
| Merge queue | REQ-ACC-001 to REQ-ACC-011, REQ-ACC-017, REQ-WI-008, REQ-WI-009, REQ-WI-020 | implemented |
| Guide | REQ-PH-001, REQ-PH-002, REQ-PH-003, REQ-PH-004, REQ-PH-005 | basic version |
| MCP server | REQ-MCP-001, REQ-MCP-002, REQ-MCP-005 | stdio, tools only |
| Design tools | REQ-DOC-010 to REQ-DOC-013, REQ-PH-007 (approval recorded with design_set_status or `maid accept`) | implemented |
| CLI and status | G-007 | implemented |

Not in milestone 1: REQ-DOC-014 (generated index), REQ-WI-006 (lease expiry handling), REQ-WI-011 (parallel attempts), REQ-ACC-012 (flaky tests), REQ-ACC-013, REQ-ACC-014, REQ-MCP-003, REQ-MCP-004, REQ-MCP-006.

## Milestone 2: trial and score

1. Run the Haiku trial on maid's own work items. Record the acceptance rate and the cost (SM-001, SM-002).
2. Implement the score (REQ-SC-001 to REQ-SC-010) and the Go metrics adapter.
3. Build the calibration benchmark (REQ-SC-011, REQ-SC-012).
4. Implement gates with human approval (REQ-PH-006, REQ-PH-007).

## Milestone 3: any project

1. Profiles and default decisions (REQ-PR-001 to REQ-PR-014).
2. Design scaffolding for new projects, and the existing-code path (REQ-PH-008, REQ-PH-009).
3. MCP resources, prompts, elicitation and HTTP (REQ-MCP-003, REQ-MCP-006).
4. Review of proposals (design_review) and the HTML report.
