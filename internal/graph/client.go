package graph

import (
	"context"
	"encoding/json"
)

// Client is a bridge to the upstream code-review-graph MCP server.
type Client struct{}

// Start launches the upstream MCP server.
func Start(ctx context.Context, opts Options) (*Client, error) {
	// Purpose: start the upstream MCP server.
	return nil, nil
}

// Tools lists the upstream tools.
func (c *Client) Tools(ctx context.Context) ([]Tool, error) {
	// Purpose: list the upstream tools.
	return nil, nil
}

// Call forwards one tool call to the upstream server.
func (c *Client) Call(ctx context.Context, name string, args json.RawMessage) (map[string]any, error) {
	// Purpose: forward one tool call to the upstream server.
	return nil, nil
}

// Close stops the upstream server.
func (c *Client) Close() error {
	// Purpose: stop the upstream server.
	return nil
}
