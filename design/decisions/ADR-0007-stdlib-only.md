---
status: accepted
style: ste
owner: paul
origin: agent
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["maid needs streamable HTTP or elicitation (REQ-MCP-003, REQ-MCP-006)", "the state store needs queries across many attempts", "maid must run on Windows"]
summary: >
  Milestone 1 uses no third-party Go modules. maid has its own small MCP
  server for stdio tools, its own parser for the frontmatter YAML subset, and
  a JSON-file state store with file locks instead of SQLite. Tools are
  configured in maid.json until decisions select them.
links: [overview, conventions, mcp, ADR-0001]
---

# ADR-0007 Build milestone 1 on the Go standard library only

## Context and problem statement

The design names the official Go MCP SDK, a YAML library and a pure-Go SQLite driver. The environment that built milestone 1 cannot download Go modules, because the module checksum database is not reachable. Should milestone 1 wait for dependencies, or use only the standard library?

## Decision drivers

- maid scores agents lower when they add dependencies. maid itself should follow the same rule.
- Milestone 1 needs only stdio MCP tools, simple frontmatter and simple local state.
- Checksum verification must stay on.

## Considered options

- Standard library only, with small replaceable components.
- Wait until dependencies can be downloaded and verified.

## Decision outcome

Chosen option: "Standard library only, with small replaceable components."

- `internal/mcp` implements JSON-RPC 2.0 over stdio with tools only. It does not implement resources, prompts, elicitation or HTTP yet.
- `internal/frontmatter` parses the YAML subset that design files use: scalars, quoted strings, inline and block lists, and block scalars.
- `internal/state` stores attempts, the test baseline and the activity log as JSON files in `.maid/`. A `flock` lock serialises the merge queue across processes. `flock` works on Linux and macOS (C-005).
- `maid.json` configures the formatter, linter, test adapter and budgets until design decisions select them (REQ-PR-009).

### Consequences

- Good, because maid has no third-party dependencies and builds anywhere Go runs.
- Good, because each component is behind a package boundary and can be replaced.
- Bad, because the MCP server lacks resources, prompts, elicitation and HTTP.
- Bad, because JSON files do not support queries. Status reports read all attempt files.
