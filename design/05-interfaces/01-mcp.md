---
id: mcp-interface
title: MCP interface
part: 05-interfaces
status: accepted
style: ears
owner: paul
summary: >
  The tools, resources, prompts and server instructions that maid exposes to
  agents, and the contract of maid_next. Tools are the baseline. Resources,
  prompts and elicitation are optional because not every client supports them.
links: [phases, work-items, acceptance, chapter-format]
---

# MCP interface

## Tools

### Always available

| Tool | Input | Output |
|---|---|---|
| `maid_next` | `role?` (`design`, `build`, `review`) | The next step. See the contract below. |
| `design_index` | `type?`, `status?`, `part?`, `link?`, `path?` | Chapters and items with summaries only. |
| `design_read` | `ids[]`, `sections[]?`, `detail` (`summary`, `full`) | Text and version of each requested section. |
| `design_question` | `action` (`add`, `answer`, `resolve`), question fields | The question and the items that it blocks or unblocks. |
| `design_check` | `ids[]?` | Findings: style, links, required sections, stale items. |
| `maid_status` | `detail` | Phase, item states, work item states, queue, recent activity. |

### Design phases (0–5) and review

| Tool | Input | Output |
|---|---|---|
| `design_scaffold` | `depth`, `profile?`, `answers?` (free text or fields) | Created stubs and the questions that remain. |
| `design_write` | `id`, `section`, `content`, `base_version`, `reason` | New version, or a conflict with the current text. |
| `design_propose` | `id`, `section`, `content`, `base_version`, `reason` | Proposal ID. |
| `design_review` | `proposal_id`, `decision` (`accept`, `reject`), `reason` | Result. |
| `design_decide` | decision fields | New decision ID. |
| `work_plan` | work items and dependencies | Created work item IDs. |

### Build phase (6–7)

| Tool | Input | Output |
|---|---|---|
| `work_claim` | `id?`, `filter?` | The claim, the lease, the worktree path and the context packet. |
| `work_heartbeat` | `id` | New lease expiry. |
| `work_release` | `id`, `reason` | Result. |
| `work_submit` | `id`, `model?`, `cost?` | Queue position. |
| `work_result` | `id`, `attempt?` | Accepted, or the rejection report. |

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
| `scaffold-design` | Start phase 0 and 1 with the user. |
| `pick-up-work` | Claim a work item and follow the build loop. |
| `review-proposals` | Review open proposals. |

## Requirements

### REQ-MCP-001 Tools baseline
The MCP server shall make every operation available as a tool, and resources and prompts shall only repeat operations that tools provide.

### REQ-MCP-002 Server instructions
When a client connects, the MCP server shall send instructions that state the current phase and tell the agent to call `maid_next` first.

### REQ-MCP-003 Elicitation
Where the client supports elicitation, the MCP server shall ask the user its questions directly; otherwise the MCP server shall return the questions in the `ask_user` field of `maid_next`.

### REQ-MCP-004 Output formats
Each tool that returns structured data shall accept `format` with the values `json` and `toon`, and the default shall be `toon`.

### REQ-MCP-005 Agent identity
Each tool call that changes state shall record an agent ID that the client supplies or that maid derives from the session.

### REQ-MCP-006 Transports
The MCP server shall support stdio and streamable HTTP.

### REQ-MCP-007 Local HTTP only
While maid runs in HTTP mode, the MCP server shall accept connections only on the loopback interface.

## maid_next contract

| Field | Type | Meaning |
|---|---|---|
| `phase` | string | Current phase ID. |
| `step` | string | One sentence. The next action. |
| `why` | string | One sentence. The reason. |
| `read` | list of IDs | The minimum sections to read. |
| `tools` | list | The tools for this step. |
| `done_when` | list | The gate conditions for this step. |
| `ask_user` | list | Questions for the user, with IDs. |
| `blocked` | list | Items that open questions block. |
