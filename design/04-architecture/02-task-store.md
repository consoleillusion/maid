---
id: task-store
title: Task store port and Beads adapter
part: 04-architecture
status: draft
style: ste
owner: paul
summary: >
  The interface that maid uses for work items, dependencies, claims and leases,
  and how the v1 adapter maps it onto the Beads CLI. Includes the findings from
  the Beads spike.
links: [work-items, architecture, ADR-0005]
---

# Task store port and Beads adapter

## Port

The task port is a Go interface. maid code uses only the port. The adapter is replaceable.

| Operation | Meaning |
|---|---|
| `Create(items, deps)` | Create work items and their dependencies in one operation. |
| `Ready(filter)` | List work items that have no open blocker. |
| `Claim(id, agent, ttl)` | Claim one work item atomically. Return a conflict if another agent holds it. |
| `ClaimNext(filter, agent, ttl)` | Claim the first ready work item that matches the filter, atomically. |
| `Heartbeat(id, agent)` | Renew the lease. |
| `Reclaim(olderThan)` | Return work items with expired leases to ready. |
| `Release(id, agent)` | Give up a claim. |
| `Close(id, reason)` | Close a work item. Dependents become ready. |
| `Get(id)` / `List(filter)` | Read work items. |

maid-specific fields (`implements`, `acceptance_tests`, `owns`, `reads`, `budget`, `kind`) are stored in a namespaced metadata object, `maid.*`.

## Beads adapter (v1)

The adapter calls the `bd` CLI with `--json` and parses the result. It does not link Beads as a Go library. See ADR-0005.

| Port operation | Beads command |
|---|---|
| `Create` | `bd create --graph <plan.json>` or `bd create … --deps … --metadata …` |
| `Ready` | `bd ready --json` |
| `Claim` | `bd update <id> --claim --json` |
| `ClaimNext` | `bd ready --claim --json` |
| `Heartbeat` | `bd heartbeat <id>` |
| `Reclaim` | `bd reclaim` |
| `Release` | `bd unclaim <id>` |
| `Close` | `bd close <id> --reason …` |
| `Get` | `bd show <id> --json` |

Native Beads fields that maid uses: `spec_id` (first linked requirement), `acceptance_criteria`, `metadata`, `assignee`, dependencies.

### Initialisation

maid runs `bd init --non-interactive --skip-agents --skip-hooks --quiet`. `--skip-agents` is required because `bd init` otherwise writes `AGENTS.md` and `CLAUDE.md`, and maid owns those files.

## Spike findings

Date: 2026-11-27. bd version 1.3.4, embedded Dolt mode. Reproduce with `spikes/beads/run.sh`.

| Test | Result |
|---|---|
| Create work items with dependencies, spec ID, acceptance text and custom metadata | Pass. All fields round-trip through `bd show --json`. |
| `bd ready` with a dependency graph | Pass. Only items without open blockers are listed. |
| Close an item | Pass. Its dependents become ready. |
| 6 processes claim the same item at the same time | Pass. Exactly 1 claim succeeds. 5 receive "already claimed by agentN" with exit code 1. |
| 20 processes claim 20 different items at the same time | Pass. 20 of 20 succeed. Total time about 9 s, so writes are serialised at about 0.45 s each. |
| 8 processes run `bd ready --claim` on a pool of 8 items | Pass. Each process gets a different item. |
| Concurrent readers during claims | Pass. No errors. |
| Leases | Present: `heartbeat`, `reclaim`, lease TTL. |

### Concerns

| Concern | Effect | Mitigation |
|---|---|---|
| Write latency about 0.45 s, serialised | Acceptable for tens of agents. Not for hundreds. | Measure again with server mode if needed. |
| The `bd` binary is about 150 MB | Larger install. | Install `bd` separately. maid checks for it and reports the install command. |
| The Go library needs cgo for embedded Dolt, and the module requires a newer Go than maid may use | Linking breaks the single static binary. | Use the CLI. |
| `bd close` refuses when the actor is not the assignee | The merge queue cannot close an item as itself. | The adapter sets `BEADS_ACTOR` to the claim owner when it closes an item. |
| Beads changes quickly | The CLI contract can change. | Pin a minimum `bd` version. Run adapter contract tests in CI. |
