---
id: ADR-0001
title: Write maid in Go and restart from the design
part: decisions
status: accepted
style: ste
owner: paul
origin: user
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["the official Go MCP SDK stops supporting a protocol version that clients need"]
summary: >
  maid is written in Go and ships as one static binary for each platform. The
  existing TypeScript implementation is removed. Implementation restarts from
  design/.
links: [constraints, conventions, architecture]
---

# ADR-0001 Write maid in Go and restart from the design

## Context and problem statement

maid must be easy to install in any project, in any language. Which language gives the simplest install and fits the tools that maid uses?

## Decision drivers

- "Install an MCP in a project" must be one binary and one configuration entry.
- maid must work for projects in any language, so it must not need the runtime of the target project.
- maid runs many processes and many tool runs in parallel.
- Related tools (Beads, Dolt, Forgejo, Windshift) are written in Go.

## Considered options

- Go
- TypeScript on Bun
- Rust

## Decision outcome

Chosen option: "Go", because it gives one static binary with no runtime. Go also has an official MCP SDK that supports the current protocol, and simple concurrency.

The existing TypeScript implementation is removed. Implementation starts again from `design/`. The git history keeps the old implementation for reference.

### Consequences

- Good, because users install one file.
- Good, because the official Go MCP SDK supports stdio, streamable HTTP, elicitation and list-changed notifications.
- Bad, because the existing TypeScript code and tests are not reused.
- Bad, because some validators (for example APE for ACE) are not Go and run as external programs.

### Confirmation

The build produces static binaries for Linux, macOS and Windows with `CGO_ENABLED=0`.

## Pros and cons of the options

### TypeScript on Bun

- Good, because existing code could be reused.
- Bad, because users need Bun or a large compiled binary.

### Rust

- Good, because it gives a small static binary.
- Bad, because development is slower, and the related tools are not Rust.
