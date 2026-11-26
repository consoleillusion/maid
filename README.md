# maid — Metric Assisted AI Driven Development

maid is an MCP server and CLI that you install in a software project. When you point an agent at it, maid guides the agent through a structured progression: design first, then small work items, then automatic acceptance of each commit.

- A large, capable agent writes the design and the acceptance tests.
- Smaller, cheaper agents implement work items, possibly in parallel.
- maid accepts a commit when its assigned tests pass and no other test breaks. A score based on tests, lines of code, dependencies and benchmarks ranks competing attempts.

## Status

maid is in the design phase. The specification is in [`design/`](design/index.md). Start with [`design/index.md`](design/index.md).

Agents: read [`AGENTS.md`](AGENTS.md).

## Repository layout

```text
AGENTS.md   instructions for agents
design/     specification (source of truth)
docs/       generated documentation of the code (not yet present)
spikes/     experiments that support design decisions
```

## License

MIT
