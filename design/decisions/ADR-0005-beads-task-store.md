---
id: ADR-0005
title: Use Beads as the v1 task store, through its CLI, behind a port
part: decisions
status: accepted
style: ste
owner: paul
origin: agent
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["the bd CLI contract changes in a way that breaks the adapter", "claim latency limits the number of agents", "Beads offers a pure-Go embedded library"]
summary: >
  maid uses Beads for work items, dependencies, claims and leases in v1. maid
  calls the bd CLI with JSON output through a task port. maid does not link
  Beads as a library, because embedded Dolt needs cgo.
links: [task-store, work-items, constraints]
---

# ADR-0005 Use Beads as the v1 task store, through its CLI, behind a port

## Context and problem statement

maid needs work items with dependencies, a ready queue, atomic claims and leases for many agent processes. Should maid build this, or use an existing tool?

## Decision drivers

- Reuse existing tools (G-009).
- Claims must be atomic across many processes (C-007).
- maid must stay one static binary (C-001).

## Considered options

- Beads through its CLI.
- Beads as a linked Go library.
- A task store in maid's own SQLite state store.
- An external tracker (Windshift, Jira, Forgejo issues).

## Decision outcome

Chosen option: "Beads through its CLI". The spike showed that Beads already provides dependency-aware ready work, atomic claims, leases, reclaim and custom metadata. It also works with many processes at the same time. See `task-store` and `spikes/beads/README.md`.

### Consequences

- Good, because maid does not build a task tracker.
- Good, because the port lets maid replace Beads later.
- Bad, because users install `bd` (about 150 MB) as well as maid.
- Bad, because each operation starts a process. Claim latency is about 0.45 s and writes are serialised.

## Pros and cons of the options

### Beads as a linked Go library

- Good, because there is no extra process.
- Bad, because embedded Dolt needs cgo, which breaks the static binary.
- Bad, because the Beads module needs a newer Go version and a large dependency tree.

### Own SQLite task store

- Good, because there is no extra install.
- Bad, because maid builds and maintains a tracker (NG-001 risk).

### External tracker

- Bad, because it needs a network service (C-004) and is built for humans, not agents.
