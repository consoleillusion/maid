package graph

// Tool is one upstream MCP tool definition.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Options starts one upstream code-review-graph MCP server.
type Options struct {
	Command []string
	Args    []string
	Repo    string
}
