package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

func session(t *testing.T, lines ...string) []map[string]any {
	// Purpose: run the server over in-memory pipes and collect its replies.
	root, _ := filepath.Abs(filepath.Join("..", ".."))
	svc := work.New(root, config.Default(), tasks.NewMemory())
	srv := NewServer(&Tools{Svc: svc})
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	go func() {
		_ = srv.Serve(context.Background(), inR, outW)
		outW.Close()
	}()
	go func() {
		for _, l := range lines {
			_, _ = inW.Write([]byte(l + "\n"))
		}
		inW.Close()
	}()
	var out []map[string]any
	sc := bufio.NewScanner(outR)
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("bad JSON from server: %s", sc.Text())
		}
		out = append(out, m)
	}
	return out
}

func text(t *testing.T, reply map[string]any) string {
	// Purpose: return the text content of a tools/call reply.
	res, ok := reply["result"].(map[string]any)
	if !ok {
		t.Fatalf("no result: %v", reply)
	}
	c := res["content"].([]any)[0].(map[string]any)
	return c["text"].(string)
}

// REQ-MCP-001 REQ-MCP-002 REQ-PH-002 REQ-PH-004
func TestProtocol(t *testing.T) {
	// Purpose: initialize, list tools for the phase, and call tools with hints.
	out := session(t,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"test"},"capabilities":{}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"design_read","arguments":{"refs":["REQ-WI-004"]}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"maid_next","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"nope"}`,
	)
	if len(out) != 5 {
		t.Fatalf("got %d replies: %v", len(out), out)
	}
	init := out[0]["result"].(map[string]any)
	if init["protocolVersion"] != "2025-06-18" || !strings.Contains(init["instructions"].(string), "maid_next") {
		t.Fatalf("initialize = %v", init)
	}
	var names []string
	for _, tl := range out[1]["result"].(map[string]any)["tools"].([]any) {
		names = append(names, tl.(map[string]any)["name"].(string))
	}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "maid_next") || strings.Contains(joined, "work_claim") {
		t.Fatalf("tools before phase 5 = %s", joined)
	}
	if read := text(t, out[2]); !strings.Contains(read, "REQ-WI-004 Atomic claim") || !strings.Contains(read, "\nnext: ") {
		t.Fatalf("design_read = %s", read)
	}
	if next := text(t, out[3]); !strings.Contains(next, `"phase"`) {
		t.Fatalf("maid_next = %s", next)
	}
	if out[4]["error"] == nil {
		t.Fatalf("unknown method did not error: %v", out[4])
	}
}
