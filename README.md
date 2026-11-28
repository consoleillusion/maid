
# maid — Metric Assisted AI Driven Development

maid is an MCP server and CLI that you install in a software project. When you point an agent at it, maid guides the agent through a structured progression: design first, then small work items, then automatic acceptance of each commit.

- A large, capable agent writes the design and the acceptance tests.
- Smaller, cheaper agents implement work items, possibly in parallel.
- maid accepts a commit when its assigned tests pass and no other test breaks. A score based on tests, lines of code, dependencies and benchmarks ranks competing attempts.

## Status

Milestone 1 is implemented: the design store and checks, plans with separate tests items, claims with isolated worktrees and context packets, the merge queue, `maid_next`, and an MCP server (stdio, tools). See [`design/delivery/roadmap.md`](design/delivery/roadmap.md).

maid commits all its work to one branch, `maid-auto`: one commit for each accepted task, with a `Maid-Task:` trailer, and one commit for each design change. Merge `maid-auto` into your main branch when you want.

The specification is in [`design/`](design/index.md). Start with [`design/index.md`](design/index.md). To build it as two PDFs, run `scripts/compile-pdf.sh` (Typst 0.15 or later; the first build downloads the cmarker and merman packages). It writes a light and a dark PDF named after the design title. The title page shows the title at the bottom left and uses `design/cover_image.png` above it when that file exists. `scripts/compile-pdf.sh --dark-only` writes only the dark PDF. Mermaid code fences render as diagrams. `maid init` installs this script, the PDF template, a release workflow and a `.gitattributes` file in any project. The workflow attaches the dark PDF to published major and minor releases (`vX.Y.0`), not to patches.

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
maid init          # maid.json, .maid/, Beads, code graph
maid check         # check design/
maid status
```

Add maid to your MCP client as a stdio server with the command `maid mcp`, run in the project root. With Nix, `nix shell /path/to/maid#maid /path/to/maid#bd -c maid mcp` builds maid and bd from the flake, so the server always matches the checked-out commit. Keep the MCP command's stdout clean: anything else printed there breaks the protocol. Then tell the agent to call `maid_next`.

## Code review graph

`maid init` installs [code-review-graph](https://code-review-graph.com/) in the project environment at `.maid/crg-venv` and builds the graph for the repository. Python 3 is required at init time. If `code-review-graph` is already on `PATH`, maid records that command instead.

The `maid` MCP server exposes all 30 code-review-graph tools under their upstream names, such as `get_minimal_context_tool`, `query_graph_tool` and `get_impact_radius_tool`. Agents do not need a second MCP server. `maid.json` stores the command, its `serve` arguments and an `enabled` flag under `code_review_graph`. Set `enabled` to `false` to turn the bridge off. The generated graph is local to the project.

The agent can write and accept design chapters through maid's tools alone. To accept a chapter yourself, run `maid accept CHAPTER`. Build and test agents can work through maid's tools alone: `work_read`, `work_edit`, `work_run`, then `work_submit`. Agents with file and shell access can also edit and commit in their worktree directly.

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
