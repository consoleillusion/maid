# maid — Metric Assisted AI Driven Development

maid is an MCP server and CLI that you install in a software project. When you point an agent at it, maid guides the agent through a structured progression: design first, then small work items, then automatic acceptance of each commit.

- A large, capable agent writes the design and the acceptance tests.
- Smaller, cheaper agents implement work items, possibly in parallel.
- maid accepts a commit when its assigned tests pass and no other test breaks. A score based on tests, lines of code, dependencies and benchmarks ranks competing attempts.

## Status

Milestone 1 is implemented: the design store and checks, plans with separate tests items, claims with isolated worktrees and context packets, the merge queue, `maid_next`, and an MCP server (stdio, tools). See [`design/07-delivery/01-roadmap.md`](design/07-delivery/01-roadmap.md).

The specification is in [`design/`](design/index.md). Start with [`design/index.md`](design/index.md).

## Install and run

With Nix:

```sh
nix develop                 # go, gopls, golangci-lint, git, python3 and bd on PATH
go install ./cmd/maid       # installs to .bin/, which the dev shell puts on PATH
cd your-project && maid init
```

Or run maid without installing it: `nix run github:consoleillusion/maid -- init`.

Without Nix, you need Go, git and [Beads](https://github.com/gastownhall/beads). `go install ./cmd/maid` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`). Add that folder to your `PATH`, or run maid with its full path.

```sh
maid init          # maid.json, .maid/, Beads
maid check         # check design/
maid status
```

Add maid to your MCP client as a stdio server with the command `maid mcp`, run in the project root. Then tell the agent to call `maid_next`.

The agent can write and accept design chapters through maid's tools alone. To accept a chapter yourself, run `maid accept CHAPTER`. Build and test agents also need to edit files and run `git commit` in their worktree, so use an agent with file and shell access, such as Claude Code, or the Claude desktop app with the project folder added.

## Development

```sh
go test ./...                       # set MAID_BD=/path/to/bd to include the Beads contract test
go build -o maid ./cmd/maid
test/e2e/run.sh ./maid              # full loop with the real binary and Beads
```

Agents: read [`AGENTS.md`](AGENTS.md).

## Repository layout

```text
AGENTS.md   instructions for agents
cmd/maid/   CLI and MCP server entry point
internal/   implementation, one package for each architecture component
design/     specification (source of truth)
docs/       generated documentation of the code (optional)
spikes/     experiments that support design decisions
test/e2e/   end-to-end test
```

## License

MIT
