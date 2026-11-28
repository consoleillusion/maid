---
status: accepted
style: ste
owner: paul
summary: >
  How maid's own code is written, tested and committed. These are the rules for
  agents that build maid. They are not the defaults that maid applies to other
  projects; those are in profiles.
links: [ADR-0001, ADR-0007, profiles]
---

# Engineering conventions for maid

## Language and tools

| Item | Rule |
|---|---|
| Language | Go. maid builds and passes its tests with the two newest Go releases, as Go's release policy supports. |
| Module path | `github.com/consoleillusion/maid` |
| Formatter | `gofmt` |
| Linter | `golangci-lint` with the configuration in the repository |
| Tests | `go test ./...` |
| Dependencies | The Go standard library only, until a decision adds a module (ADR-0007). |
| Git | Call the `git` CLI. Do not use a Go git library. |
| MCP | `internal/mcp`, stdio and tools only (ADR-0007). The official Go SDK replaces it when HTTP or elicitation is needed. |

## Code

1. Start each function body with a comment line `// Purpose: <one sentence>`.
2. Keep functions small. One function does one thing.
3. Return errors. Do not panic, except for programming errors at start-up.
4. Put each component from the architecture in its own package under `internal/`.
5. Access the task store, git and external tools only through their ports.
6. Link each test to a requirement with a comment `// REQ-<AREA>-NNN` above the test.

Example:

```go
func ReadyItems(ctx context.Context, store TaskPort) ([]WorkItem, error) {
	// Purpose: list work items that have no open blocker.
	return store.Ready(ctx, Filter{})
}
```

## Commits

Commits follow Conventional Commits.

```text
<type>(<optional scope>): <description>

[optional body]

[optional footer]
```

1. Make each commit atomic. One logical change in each commit.
2. Every commit on the main branch builds and passes its tests.
3. Use the imperative mood: "add", "fix", "update".
4. Keep the subject at 50 characters or less if you can, and never more than 72.
5. Do not end the subject with a period.
6. Put a blank line between the subject and the body.
7. Wrap the body at 72 characters. Explain why the change was made.
8. Mark a breaking change with `!` after the type, or with a `BREAKING CHANGE:` footer.

| Type | Use |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation or design only |
| `style` | Formatting, no change to behaviour |
| `refactor` | Structure change, no change to behaviour |
| `perf` | Performance |
| `test` | Tests |
| `build` / `ci` | Build system or CI |
| `chore` | Maintenance and dependencies |
