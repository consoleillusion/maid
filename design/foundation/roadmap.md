---
status: accepted
style: ste
owner: paul
summary: >
  The milestones of maid, what each milestone delivers, and what is implemented
  now. Read this to see where the project is and what comes next.
links: [phases, work-items, acceptance, interfaces, stdlib-only, everything-is-a-task]
---

# Roadmap

Each area shows where the project is now and where it must be. The tasks for an area close the gap between the two columns. The design agent keeps the Now column true after each milestone.

## Milestone 1: maid builds maid

| Area | Now | Target | Sections |
|---|---|---|---|
| Design store, checks and writes | Done | Agents read, check and change the design through maid | `format` |
| Style validators | Length, EARS patterns, SBVR keywords, Gherkin structure | Same | `styles` |
| Task store | Beads through its CLI in stealth mode | Beads, committed and visible in git | `architecture` |
| Plans and tests items | Done, with stubs | Same | `work-items` |
| Claims, packets and worktree tools | Done | Same | `work-items` |
| Merge queue | Tests decide acceptance; `internal/score` computes a score that the queue does not use | Tests and score decide acceptance | `acceptance` |
| Guide | Phases, steps, ready-to-run calls and every ready task; no roles | Same | `phases` |
| MCP server | stdio and tools | Same | `interfaces` |
| Approval | Agents accept chapters directly; no approver is recorded | Same | `acceptance` |

## Milestone 2: score and trial

| Area | Now | Target | Sections |
|---|---|---|---|
| Score | `internal/score` computes a score; reports and the merge queue do not use it | Score in every report; minimum score rejects | `scoring` |
| Score history | Not built | Score kept for each agent and model | `scoring` |
| Calibration | Not built | Benchmark ranks good solutions above bad ones | `scoring` |
| Trial | Not run | Acceptance rate and cost of cheap agents measured | `vision` |

## Milestone 3: any project

| Area | Now | Target | Sections |
|---|---|---|---|
| Profiles | Specified only | Profiles offered as hints | `profiles` |
| New and existing projects | Bootstrap step only | Current and target state recorded for any project | `phases` |
| MCP extras | Not built | Resources, prompts, elicitation, HTTP | `interfaces` |
| Generated index and report | `design/index.md` and `design/index.typ` are generated; `maid status` reports state | Generated docs from code are in `docs/` where configured | `format` |
| Robustness | Serial merges, parallel attempts, lease renewal and a flaky list exist; automatic flaky detection is open | Lease expiry, parallel attempts, flaky tests | `work-items`, `acceptance` |

## Milestone 2b: everything is a task

| Area | Now | Target | Sections |
|---|---|---|---|
| Task list | `maid_next` returns a phase, a step, ready-to-run calls and every ready task | Same | `phases` |
| Plain prompts | Server instructions tell agents to call `maid_next` whenever the user asks for work | Same | `phases` |
| Seeded tasks | `maid init` creates no tasks | `maid init` seeds the design tasks | `everything-is-a-task` |
| Design tasks | Task kinds are `code`, `tests` and `docs`; design tasks are open | Design work is tasks of kind design, checked by the design check | `phases` |
| Metrics for any language | Go adapter implemented; manifest tests added, their code item open | Language-independent measures for every language; adapters for a few languages | `scoring` |


