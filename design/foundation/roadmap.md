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
| Task store | Beads through its CLI, local only | Beads, committed and visible in git | `architecture` |
| Plans and tests items | Done, with stubs | Same | `work-items` |
| Claims, packets and worktree tools | Done | Same | `work-items` |
| Merge queue | Tests decide acceptance | Tests and score decide acceptance | `acceptance` |
| Guide | Phases, steps and ready calls; asks for a role | No roles; steps follow from the project state | `phases` |
| MCP server | stdio, tools only | Same | `interfaces` |
| Approval | Every acceptance needs an approver | The user decides whether to approve or to delegate | `acceptance` |

## Milestone 2: score and trial

| Area | Now | Target | Sections |
|---|---|---|---|
| Score | Not built; eight work items planned | Score in every report; minimum score rejects | `scoring` |
| Score history | Not built | Score kept for each agent and model | `scoring` |
| Calibration | Not built | Benchmark ranks good solutions above bad ones | `scoring` |
| Trial | Not run | Acceptance rate and cost of cheap agents measured | `vision` |

## Milestone 3: any project

| Area | Now | Target | Sections |
|---|---|---|---|
| Profiles | Specified only | Profiles offered as hints | `profiles` |
| New and existing projects | Bootstrap step only | Current and target state recorded for any project | `phases` |
| MCP extras | Not built | Resources, prompts, elicitation, HTTP | `interfaces` |
| Generated index and report | Not built | `design/index.md` generated; HTML report | `format` |
| Robustness | Not built | Lease expiry, parallel attempts, flaky tests | `work-items`, `acceptance` |

## Milestone 2b: everything is a task

| Area | Now | Target | Sections |
|---|---|---|---|
| Task list | `maid_next` gives one step | `maid_next` lists every ready task | `phases` |
| Plain prompts | Server instructions say to call `maid_next` first | Agents call `maid_next` whenever the user asks them to work | `phases` |
| Seeded tasks | `maid init` creates no tasks | `maid init` seeds the design tasks | `everything-is-a-task` |
| Design tasks | Design work happens outside tasks | Design work is tasks of kind design, checked by the design check | `phases` |
| Metrics for any language | Go adapter planned | Language-independent measures for every language; adapters for a few languages | `scoring` |

## Proposal: Update current state to match code

status: proposed · by opencode-978a9f · 2026-11-28

### Milestone 1 current state

| Area | Now |
|---|---|
| Design store, checks and writes | Done through MCP tools and the CLI. |
| Style validators | Length, EARS patterns, SBVR keywords and Gherkin structure are checked. |
| Task store | Beads through its CLI in stealth mode. Export to `.beads/issues.jsonl` and commit are open. |
| Plans and tests items | Done, with separate tests items. |
| Claims, packets and worktree tools | Done. |
| Merge queue | Tests decide acceptance. `internal/score` calculates scores, but the queue does not use them. |
| Guide | No roles. `maid_next` returns a phase, a step and ready-to-run calls. Listing every ready task and plain-prompt instructions are open. |
| MCP server | stdio and tools, plus all 30 code-review-graph tools. |
| Approval | Agents accept chapters directly. No approver is recorded. |

### Milestone 2 current state

| Area | Now |
|---|---|
| Score | `internal/score` computes a score. Minimum-score rejection, score in reports, score history and `maid.json` weights are open. |
| Metrics | The Go adapter is stubbed. |

### Milestone 2b current state

| Area | Now |
|---|---|
| Task list | `maid_next` returns one `work_claim` call. Listing every ready task is open. |
| Plain prompts | Server instructions say to call `maid_next` first. Plain-prompt guidance is open. |
| Seeded tasks | `maid init` creates no tasks. |
| Design tasks | Task kinds are `code`, `tests` and `docs`. Design tasks are open. |

### Milestone 3 current state

| Area | Now |
|---|---|
| Generated index | `design/index.md` and `design/index.typ` are generated. The HTML report is open. |
| Robustness | Serial merges, parallel attempts, lease renewal and a flaky list exist. Automatic flaky detection is open. |
