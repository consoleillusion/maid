---
status: accepted
style: ears
owner: paul
summary: >
  Specifies installation of code-review-graph and its full tool surface inside maid's MCP server.
links: [interfaces, architecture]
---

# Code review graph

maid exposes the local code graph from code-review-graph through its MCP server. The external tool supplies the tools. maid adds no second MCP server to the client. A bridge child process keeps the upstream schemas unchanged.

## Requirements: install

### Installed by init
When init runs, maid shall install code-review-graph and shall build the graph in the project root.

### Existing command
If code-review-graph is already on PATH, then maid shall record that command and shall use it.

### Failed install
If no installer can install code-review-graph, then maid shall return an actionable environment error.

### Configured command
maid shall keep the command and arguments that start code-review-graph in maid.json.

## Requirements: MCP bridge

### Tool list
When the MCP server starts, the MCP server shall fetch the upstream tool list and expose every tool under its upstream name.

### Tool calls
When an agent calls an upstream tool, the MCP server shall forward the call arguments and shall return the upstream result.

### Repository root
The MCP server shall start code-review-graph with the project root as its repository root.

### Child process
When the MCP server stops, the MCP server shall terminate the code-review-graph child process.

### Tool failure
If code-review-graph returns an error, then the MCP server shall return an MCP error result with that message.

## Scenarios

```gherkin
Scenario: All graph tools appear in maid
  Given code-review-graph is installed
  When a client lists MCP tools
  Then every upstream graph tool appears under its upstream name

Scenario: A graph tool call is forwarded
  Given a graph tool call with arguments
  When the client calls the tool through maid
  Then code-review-graph receives the same arguments
  And the client receives the upstream result
```

## Requirements: diagrams and traceability

### Diagram commands
maid shall record that `code-review-graph visualize` generates static SVG, JSON, GraphML, Cypher and Obsidian exports from the project graph.

### External diagram generation
The MCP bridge shall not expose `visualize` as a tool.

### Mermaid prompt target
Where a client supports prompts, maid shall expose the upstream `architecture_map` prompt as a target for Mermaid architecture diagrams.

### Markdown indexing limit
The code graph shall not index `design/*.md` as document sections.

### Traceability source
Where an agent needs the code for a design section, maid shall use the work item `implements` path, its acceptance tests and the verify gate until the code graph has a design edge.
