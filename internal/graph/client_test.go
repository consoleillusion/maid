package graph

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

var upstreamTools = []string{
	"build_or_update_graph_tool",
	"run_postprocess_tool",
	"get_minimal_context_tool",
	"get_impact_radius_tool",
	"query_graph_tool",
	"get_review_context_tool",
	"semantic_search_nodes_tool",
	"embed_graph_tool",
	"list_graph_stats_tool",
	"get_docs_section_tool",
	"find_large_functions_tool",
	"list_flows_tool",
	"get_flow_tool",
	"get_affected_flows_tool",
	"list_communities_tool",
	"get_community_tool",
	"get_architecture_overview_tool",
	"detect_changes_tool",
	"refactor_tool",
	"apply_refactor_tool",
	"generate_wiki_tool",
	"get_wiki_page_tool",
	"get_hub_nodes_tool",
	"get_bridge_nodes_tool",
	"get_knowledge_gaps_tool",
	"get_surprising_connections_tool",
	"get_suggested_questions_tool",
	"traverse_graph_tool",
	"list_repos_tool",
	"cross_repo_search_tool",
}

const fakeServer = `import json, os, sys
log = os.environ["GRAPH_TEST_LOG"]
with open(log, "a") as f:
    f.write("cwd=" + os.getcwd() + "\n")
    f.write("argv=" + " ".join(sys.argv[1:]) + "\n")
pidfile = os.environ.get("GRAPH_TEST_PID")
if pidfile:
    with open(pidfile, "w") as f:
        f.write(str(os.getpid()))
tools = %s
for line in sys.stdin:
    try:
        message = json.loads(line)
    except Exception:
        continue
    if "id" not in message:
        continue
    method = message.get("method")
    params = message.get("params") or {}
    reply = {"jsonrpc": "2.0", "id": message["id"]}
    if method == "initialize":
        reply["result"] = {"protocolVersion": "2025-06-18", "capabilities": {"tools": {}}, "serverInfo": {"name": "code-review-graph", "version": "test"}}
    elif method == "tools/list":
        reply["result"] = {"tools": [{"name": n, "description": "upstream tool", "inputSchema": {"type": "object"}} for n in tools]}
    elif method == "tools/call":
        if params.get("name") == "failing_tool":
            reply["error"] = {"code": -32000, "message": "graph failed"}
        else:
            reply["result"] = {"content": [{"type": "text", "text": json.dumps(params.get("arguments") or {})}]}
    else:
        reply["result"] = {}
    print(json.dumps(reply), flush=True)
`

func startFakeServer(t *testing.T) (*Client, string, string, string) {
	// Purpose: start a scripted upstream MCP server for a test.
	t.Helper()
	root := t.TempDir()
	log := filepath.Join(root, "graph.log")
	pid := filepath.Join(root, "graph.pid")
	toolsJSON, err := json.Marshal(upstreamTools)
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(root, "fake.py")
	if err := os.WriteFile(script, []byte(strings.Replace(fakeServer, "%s", string(toolsJSON), 1)), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GRAPH_TEST_LOG", log)
	t.Setenv("GRAPH_TEST_PID", pid)
	client, err := Start(context.Background(), Options{Command: []string{"python3"}, Args: []string{script}, Repo: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client, log, pid, root
}

func toolNames(tools []Tool) map[string]bool {
	// Purpose: index tool definitions by name.
	out := map[string]bool{}
	for _, tool := range tools {
		out[tool.Name] = true
	}
	return out
}

// design section: code-review-graph#tool-list
// scenario: All graph tools appear in maid
func TestClientExposesUpstreamTools(t *testing.T) {
	// Purpose: the bridge lists every upstream tool under its upstream name.
	client, _, _, _ := startFakeServer(t)
	tools, err := client.Tools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := toolNames(tools)
	for _, name := range upstreamTools {
		if !got[name] {
			t.Errorf("tool %q is missing from %v", name, got)
		}
	}
}

// design section: code-review-graph#tool-calls
// scenario: A graph tool call is forwarded
func TestClientForwardsToolCall(t *testing.T) {
	// Purpose: call arguments reach the upstream server unchanged.
	client, _, _, _ := startFakeServer(t)
	args := json.RawMessage(`{"pattern":"callees","target":"F"}`)
	result, err := client.Call(context.Background(), "query_graph_tool", args)
	if err != nil {
		t.Fatal(err)
	}
	text := resultText(result)
	if !strings.Contains(text, "callees") || !strings.Contains(text, "F") {
		t.Fatalf("result text = %q; want the forwarded arguments", text)
	}
}

func resultText(result map[string]any) string {
	// Purpose: join the text parts of an MCP tool result.
	if result == nil {
		return ""
	}
	content, _ := result["content"].([]any)
	var out []string
	for _, part := range content {
		m, _ := part.(map[string]any)
		if s, ok := m["text"].(string); ok {
			out = append(out, s)
		}
	}
	return strings.Join(out, "\n")
}

// design section: code-review-graph#repository-root
// scenario: All graph tools appear in maid
func TestClientUsesRepositoryRoot(t *testing.T) {
	// Purpose: the upstream server receives the project root.
	client, log, _, root := startFakeServer(t)
	if _, err := client.Tools(context.Background()); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "cwd="+root) || !strings.Contains(text, "--repo "+root) {
		t.Fatalf("upstream start did not use %q: %q", root, text)
	}
}

// design section: code-review-graph#child-process
// scenario: The graph child process stops
func TestClientCloseStopsChild(t *testing.T) {
	// Purpose: the bridge stops the upstream child when it closes.
	client, _, pidPath, _ := startFakeServer(t)
	if _, err := client.Tools(context.Background()); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("upstream process %d is still alive", pid)
}

func processAlive(pid int) bool {
	// Purpose: report whether a process still exists.
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// design section: code-review-graph#tool-failure
// scenario: All graph tools appear in maid
func TestClientReturnsErrorResult(t *testing.T) {
	// Purpose: upstream errors reach the caller with their message.
	client, _, _, _ := startFakeServer(t)
	_, err := client.Call(context.Background(), "failing_tool", nil)
	if err == nil || !strings.Contains(err.Error(), "graph failed") {
		t.Fatalf("error = %v; want upstream message", err)
	}
}
