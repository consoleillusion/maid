---
status: accepted
style: ears
owner: paul
summary: >
  The tools, resources, prompts and server instructions that maid exposes to
  agents, and the contract of maid_next. Tools are the baseline. Resources,
  prompts and elicitation are optional because not every client supports them.
links: [phases, work-items, acceptance, format]
---

# MCP interface

## Tools

### Always available

maid has no roles. Any agent can call any tool. `maid_next` returns the next step from the state of the project: design work while the design or the plan is incomplete, and work items when tasks are ready. The documentation recommends a capable agent for the first session, which records the current state, the target state and the tasks.

| Tool | Input | Output |
|---|---|---|
| `maid_next` | none | The next step. See the contract below. |
| `design_index` | `part?`, `status?`, `style?`, `link?` | Chapters with summaries only. |
| `design_read` | `refs[]`, `detail?` (`summary`, `full`) | Text and version of each requested section. |
| `design_check` | none | Findings: style, links, required sections. |
| `maid_status` | `detail?` | Phase, chapter and work item states, test baseline, recent activity. |

### Design phases and review

| Tool | Input | Output |
|---|---|---|
| `design_scaffold` | `depth`, `profile?`, `answers?` (free text or fields) | Created stubs and the questions that remain. |
| `design_write` | `ref` (`chapter#heading`), `content`, `base_version`, `reason` | New version, or a conflict with the current text. |
| `design_propose` | `chapter`, `title`, `content`, `reason` | Result of the proposal write. |
| `design_review` | `proposal_id`, `decision` (`accept`, `reject`), `reason` | Result. |
| `design_decide` | decision fields | New decision chapter. |
| `work_plan` | work items and dependencies | Created work items. |

### Build and verify phases

| Tool | Input | Output |
|---|---|---|
| `work_claim` | `id?`, `filter?` | The claim, the lease, the worktree path and the context packet. |
| `work_files` | `id` | Files in the worktree; owned files are marked. |
| `work_read` | `id`, `path` | File content. |
| `work_write` | `id`, `path`, `content`, `remove?` | Writes an owned file. |
| `work_run` | `id`, `check` (`test`, `format`, `lint`) | Check result. |
| `work_commit` | `id`, `message` | Commit ID. |
| `work_heartbeat` | `id` | New lease expiry. |
| `work_release` | `id`, `reason` | Result. |
| `work_submit` | `id`, `patch?` (unified diff), `files?` (path → content), `message?`, `model?`, `cost?` | maid applies the patch or files, commits, runs the merge queue, and returns accepted or the rejection report. |
| `work_result` | `id`, `attempt?` | Attempts and their reports. |

## Resources

| URI | Content |
|---|---|
| `maid://design/index` | Same as `design_index` with no filter. |
| `maid://design/{id}` | A chapter. |
| `maid://design/{id}#{section}` | A section. |
| `maid://work/{id}/packet` | The context packet of a claimed work item. |

## Prompts

| Prompt | Use |
|---|---|
| `scaffold-design` | Start the bootstrap and intent phases with the user. |
| `pick-up-work` | Claim a work item and follow the build loop. |
| `review-proposals` | Review open proposals. |

## Requirements

### Tools baseline
The MCP server shall make every operation available as a tool, and resources and prompts shall only repeat operations that tools provide.

### Server instructions
When a client connects, the MCP server shall send instructions that state the current phase and tell the agent to call `maid_next` first.

### Elicitation
Where the client supports elicitation, the MCP server shall ask the user its questions directly; otherwise the MCP server shall return the questions in the `ask_user` field of `maid_next`.

### Output formats
Each tool that returns structured data shall accept `format` with the values `json` and `toon`, and the default shall be `toon`.

### Agent identity
Each tool call that changes state shall record an agent ID that the client supplies or that maid derives from the session.

### Transports
The MCP server shall support stdio and streamable HTTP.

### Local HTTP only
While maid runs in HTTP mode, the MCP server shall accept connections only on the loopback interface.

## maid_next contract

| Field | Type | Meaning |
|---|---|---|
| `phase` | string | Current phase name. |
| `step` | string | One sentence. The next action. |
| `why` | string | One sentence. The reason. |
| `read` | list of section paths | The minimum sections to read. |
| `tools` | list | The tools for this step. |
| `done_when` | list | The gate conditions for this step. |
| `ask_user` | list | Questions for the user, in plain words. |
| `blocked` | list | Work items that cannot proceed. |

## Proposal: Align MCP interface with implemented tools

status: proposed · by opencode-978a9f · 2026-11-28

### Implemented tool set
The MCP server shall expose `maid_next`, `maid_status`, `design_index`, `design_read`, `design_check`, `design_write`, `design_append`, `design_create`, `design_set_status`, `design_propose`, `work_plan`, `work_claim`, `work_files`, `work_read`, `work_write`, `work_edit`, `work_run`, `work_commit`, `work_heartbeat`, `work_release`, `work_submit` and `work_result`.

### Graph tool set
The MCP server shall expose the 30 code-review-graph tools alongside its own tools.

### Implemented tool inputs
`work_claim` shall accept `id` and `agent`. `work_result` shall accept `id`. `design_propose` shall accept `chapter`, `title`, `content` and `agent`. `work_submit` shall accept `edits`.

### maid_next response fields
The `maid_next` response shall include `calls`, `tasks` and `warnings`.

### Future MCP extras
Where resources, prompts, elicitation, streamable HTTP or `toon` output are not implemented, the MCP server shall report them as targets and shall not advertise them as current behaviour.

### Agent identity
The MCP server shall record an agent ID for each state-changing tool except `work_plan`, which shall be a target.
