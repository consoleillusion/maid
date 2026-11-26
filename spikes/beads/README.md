# Beads spike

Purpose: check whether Beads can be maid's task store (ADR-0005, assumption A-003).

Run: `spikes/beads/run.sh [workdir]`. Requires `bd`, `git` and `python3`.

## Environment

- Date: 2026-11-27
- bd 1.3.4, installed with `npm install @beads/bd`
- Embedded Dolt mode (the default)
- Linux, amd64

## Results

| Check | Result |
|---|---|
| Dependency graph, `bd ready`, custom metadata, `spec_id`, `acceptance` | Pass |
| 6 processes claim the same item | Pass: 1 wins, 5 receive "already claimed" with exit code 1 |
| Close releases dependents | Pass |
| 20 processes claim 20 items | Pass: 20 of 20, about 9 s in total (writes serialised, about 0.45 s each) |
| 8 processes `bd ready --claim` on 8 items | Pass: no duplicates |
| Concurrent readers during claims | Pass |
| Leases | `bd heartbeat`, `bd reclaim` and lease TTLs exist (not load-tested) |

## Library check

The Beads module (`github.com/steveyegge/beads`) has a small public Go API at the module root (`Open`, `OpenFromConfig`, issue claimer and query interfaces). Its own documentation says that most extensions should use SQL against the database. Embedded Dolt needs cgo; non-cgo builds support only server mode. The module needs a newer Go than 1.24. Result: use the CLI, not the library.

## Notes

- `bd close` refuses when the actor is not the assignee. maid's merge queue closes items on behalf of the agent, so the adapter must set `BEADS_ACTOR` to the claim owner.
- `bd init` writes `AGENTS.md` and `CLAUDE.md` unless you pass `--skip-agents`.
- The `bd` binary is about 150 MB.
- Not tested: server mode, lease expiry under load, sync between machines.
