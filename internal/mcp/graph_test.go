package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/work"
)

const graphServer = `import json, sys
for line in sys.stdin:
    try:
        message = json.loads(line)
    except Exception:
        continue
    if "id" not in message:
        continue
    params = message.get("params") or {}
    reply = {"jsonrpc": "2.0", "id": message["id"]}
    if message.get("method") == "initialize":
        reply["result"] = {"protocolVersion": "2025-06-18", "capabilities": {"tools": {}}, "serverInfo": {"name": "code-review-graph", "version": "test"}}
    elif message.get("method") == "tools/list":
        reply["result"] = {"tools": [{"name": "get_minimal_context_tool", "description": "graph context", "inputSchema": {"type": "object"}}]}
    elif message.get("method") == "tools/call":
        reply["result"] = {"content": [{"type": "text", "text": json.dumps(params.get("arguments") or {})}]}
    else:
        reply["result"] = {}
    print(json.dumps(reply), flush=True)
`

func graphTools(t *testing.T) *Tools {
	// Purpose: build tools with a scripted upstream graph server.
	t.Helper()
	root := t.TempDir()
	script := filepath.Join(root, "graph.py")
	if err := os.WriteFile(script, []byte(graphServer), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.CodeReviewGraph.Command = []string{"python3"}
	cfg.CodeReviewGraph.Args = []string{script}
	cfg.CodeReviewGraph.Enabled = true
	if err := config.Write(root, cfg); err != nil {
		t.Fatal(err)
	}
	return &Tools{Svc: work.New(root, cfg, nil)}
}

func toolNames(t *testing.T, tools []Tool) map[string]bool {
	// Purpose: index MCP tools by name.
	t.Helper()
	out := map[string]bool{}
	for _, tool := range tools {
		out[tool.Name] = true
	}
	return out
}

// design section: code-review-graph#tool-list
// scenario: All graph tools appear in maid
func TestGraphToolsAreListed(t *testing.T) {
	// Purpose: maid lists graph tools alongside its own tools.
	tools := graphTools(t)
	if !toolNames(t, tools.List(context.Background()))["get_minimal_context_tool"] {
		t.Fatal("get_minimal_context_tool is missing from tools.List")
	}
}

// design section: code-review-graph#tool-calls
// scenario: A graph tool call is forwarded
func TestGraphToolCallIsForwarded(t *testing.T) {
	// Purpose: maid forwards a graph tool call and returns its result.
	tools := graphTools(t)
	result := tools.Call(context.Background(), "get_minimal_context_tool", json.RawMessage(`{"task":"review"}`), "agent")
	content, _ := result["content"].([]any)
	if len(content) == 0 {
		t.Fatalf("result = %v; want upstream content", result)
	}
	part, _ := content[0].(map[string]any)
	text, _ := part["text"].(string)
	if !strings.Contains(text, "review") {
		t.Fatalf("result text = %q; want forwarded arguments", text)
	}
}
