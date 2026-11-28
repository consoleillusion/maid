---
status: accepted
style: ears
owner: paul
summary: >
  The tools and server instructions that maid exposes to agents, and the
  contract of maid_next. Tools are the baseline. Resources, prompts,
  elicitation and streamable HTTP are targets.
links: [phases, work-items, acceptance, format]
---

# MCP interface

## Tools

### Always available

maid has no roles. Any agent can call any tool and take any work item. `maid_next` returns the next step from the state of the project: design work while the design or the plan is incomplete, and work items when tasks are ready. The documentation recommends a capable agent for the first session, which records the current state, the target state and the tasks, but does not require one. The server also exposes the code-review-graph tools alongside its own.

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
| `design_create` | `path`, `style`, `summary`, `body`, `links?`, `owner?` | Created draft chapter and index entry. |
| `design_write` | `ref` (`chapter#heading`), `content`, `base_version`, `reason` | New version, or a conflict with the current text. Any chapter can be written. |
| `design_append` | `chapter`, `content`, `reason` | Added section. Any chapter can be changed. |
| `design_set_status` | `ids`, `status` | Changed statuses in one call. |
| `design_propose` | `chapter`, `title`, `content`, `agent?` | Optional proposal section. |
| `work_plan` | work items and dependencies | Created work items. |

### Build and verify phases

| Tool | Input | Output |
|---|---|---|
| `work_claim` | `id?`, `agent?` | The claim, the lease, the worktree path and the context packet. |
| `work_files` | `id` | Files in the worktree; owned files are marked. |
| `work_read` | `id`, `path` | File content. |
| `work_write` | `id`, `path`, `content`, `remove?` | Writes an owned file. |
| `work_edit` | `id`, `edits` | Applies exact text replacements to owned files. |
| `work_run` | `id`, `check` (`test`, `format`, `lint`) | Check result. |
| `work_commit` | `id`, `message` | Commit ID. |
| `work_heartbeat` | `id` | New lease expiry. |
| `work_release` | `id`, `reason` | Result. |
| `work_submit` | `id`, `edits?`, `model?`, `cost?` | maid commits the worktree, runs the merge queue, and returns accepted or the rejection report. |
| `work_result` | `id` | Attempts and their reports. |

## Resources

Resources are a target. The implemented server exposes tools over stdio only.

| URI | Content |
|---|---|
| `maid://design/index` | Same as `design_index` with no filter. |
| `maid://design/{id}` | A chapter. |
| `maid://design/{id}#{section}` | A section. |
| `maid://work/{id}/packet` | The context packet of a claimed work item. |

## Prompts

Prompts are a target. The implemented server exposes tools over stdio only.

| Prompt | Use |
|---|---|
| `scaffold-design` | Start the bootstrap and intent phases with the user. |
| `pick-up-work` | Claim a work item and follow the build loop. |
| `review-proposals` | Review open proposals. |

## Requirements

### Tools baseline
The MCP server shall make every implemented operation available as a tool.

### Graph tools
The MCP server shall expose the code-review-graph tools alongside its own tools.

### Resource and prompt targets
Where resources or prompts are added, they shall only repeat operations that tools provide.

### Server instructions
When a client connects, the MCP server shall send instructions that state the current phase and tell the agent to call `maid_next` first.

### Elicitation
Where the client supports elicitation, the MCP server shall ask the user its questions directly; otherwise the MCP server shall return the questions in the `ask_user` field of `maid_next`.

### Output formats
Each tool that returns structured data shall return JSON. `toon` output shall be a target.

### Agent identity
Each tool call that changes state shall record an agent ID that the client supplies or that maid derives from the session. Recording the agent ID for `work_plan` shall be a target.

### Transports
The MCP server shall support stdio. Streamable HTTP shall be a target.

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
| `calls` | list | Ready-to-run tool calls for this step. |
| `tasks` | list | Every ready work item that the caller may claim. |
| `warnings` | list | Warnings about the project state. |
