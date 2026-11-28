package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/guide"
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
	if !strings.Contains(joined, "maid_next") || !strings.Contains(joined, "work_claim") {
		t.Fatalf("every tool must be listed in every phase: %s", joined)
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

// REQ-PH-007 REQ-DOC-010
func TestDesignFlowOverMCP(t *testing.T) {
	// Purpose: an agent with no file access can create, accept and advance the design.
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.email", "t@example.invalid"}, {"config", "user.name", "t"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	svc := work.New(root, config.Default(), tasks.NewMemory())
	tl := &Tools{Svc: svc}
	ctx := context.Background()
	call := func(name, args string) (string, bool) {
		r := tl.Call(ctx, name, json.RawMessage(args), "haiku-1")
		txt := r["content"].([]map[string]any)[0]["text"].(string)
		return txt, r["isError"] == true
	}
	if txt, _ := call("maid_next", `{}`); !strings.Contains(txt, "0-bootstrap") {
		t.Fatalf("maid_next = %s", txt)
	}
	vision := `{"path":"foundation/vision.md","id":"vision","title":"Vision","style":"ste","summary":"Why.","body":"# Vision\n\n### G-001 Goal\nThe tool adds numbers.\n\n### NG-001 No floats\nThe tool does not add floats.\n"}`
	if txt, isErr := call("design_create", vision); isErr {
		t.Fatalf("design_create: %s", txt)
	}
	if txt, _ := call("maid_next", `{}`); !strings.Contains(txt, "1-intent") || !strings.Contains(txt, "design_set_status") {
		t.Fatalf("maid_next after create = %s", txt)
	}
	if txt, isErr := call("design_set_status", `{"id":"vision","status":"accepted","approved_by":"paul"}`); isErr {
		t.Fatalf("design_set_status: %s", txt)
	}
	if txt, _ := call("maid_next", `{}`); !strings.Contains(txt, "5-plan") {
		t.Fatalf("maid_next after accept = %s", txt)
	}
	log, _ := exec.Command("git", "-C", root, "log", "--format=%s").Output()
	if !strings.Contains(string(log), "design: accepted vision (approved by paul)") {
		t.Fatalf("git log = %s", log)
	}
}

type brokenTasks struct{ tasks.Port }

func (brokenTasks) List(context.Context, bool) ([]tasks.WorkItem, error) {
	// Purpose: simulate a task store that cannot be read.
	return nil, errors.New("schema v66, bd expects v74")
}

// REQ-PH-003
func TestNextSurvivesTaskStoreError(t *testing.T) {
	// Purpose: design phases still work when the task store fails.
	root, _ := filepath.Abs(filepath.Join("..", ".."))
	svc := work.New(root, config.Default(), brokenTasks{tasks.NewMemory()})
	r := (&Tools{Svc: svc}).Call(context.Background(), "maid_next", json.RawMessage(`{}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if r["isError"] == true || !strings.Contains(txt, "1-intent") || !strings.Contains(txt, "task store unavailable") {
		t.Fatalf("maid_next = %s", txt)
	}
}

// G-007
func TestStatusSurvivesTaskStoreError(t *testing.T) {
	// Purpose: maid_status reports a task store error instead of failing.
	root, _ := filepath.Abs(filepath.Join("..", ".."))
	svc := work.New(root, config.Default(), brokenTasks{tasks.NewMemory()})
	r := (&Tools{Svc: svc}).Call(context.Background(), "maid_status", json.RawMessage(`{}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if r["isError"] == true || !strings.Contains(txt, "task store unavailable") {
		t.Fatalf("maid_status = %s", txt)
	}
}

// REQ-PH-001
func TestNextGivesReadyCalls(t *testing.T) {
	// Purpose: maid_next returns calls with arguments that work as they are.
	root, _ := filepath.Abs(filepath.Join("..", ".."))
	tl := &Tools{Svc: work.New(root, config.Default(), tasks.NewMemory())}
	ctx := context.Background()
	n, err := guide.Compute(ctx, tl.Svc, "", "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Calls) == 0 || len(n.AskUser) == 0 {
		t.Fatalf("no calls or questions: %+v", n)
	}
	first, _ := json.Marshal(n.Calls[0].Arguments)
	r := tl.Call(ctx, n.Calls[0].Tool, first, "a")
	if r["isError"] == true {
		t.Fatalf("suggested call %s %s failed: %v", n.Calls[0].Tool, first, r)
	}
	r = tl.Call(ctx, "design_read", json.RawMessage(`{"refs":["design/foundation/vision.md","visoin"]}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if !strings.Contains(txt, "G-001") || !strings.Contains(txt, `"did_you_mean"`) || !strings.Contains(txt, "vision") {
		t.Fatalf("design_read with a path and a typo = %s", txt)
	}
}

// REQ-PH-007
func TestBatchAccept(t *testing.T) {
	// Purpose: one call accepts many chapters in one commit and reports refusals.
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.email", "t@example.invalid"}, {"config", "user.name", "t"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	tl := &Tools{Svc: work.New(root, config.Default(), tasks.NewMemory())}
	ctx := context.Background()
	for _, c := range []string{
		`{"path":"a/vision.md","id":"vision","title":"Vision","style":"ste","summary":"Why.","body":"# Vision\n\n### G-001 Goal\nIt adds.\n"}`,
		`{"path":"a/glossary.md","id":"glossary","title":"Glossary","style":"sbvr-se","summary":"Terms.","body":"# Glossary\n\n## Terms\n\n**sum**: a total.\n"}`,
	} {
		if r := tl.Call(ctx, "design_create", json.RawMessage(c), "a"); r["isError"] == true {
			t.Fatalf("create: %v", r)
		}
	}
	n, _ := guide.Compute(ctx, tl.Svc, "", "a")
	if len(n.AskUser) != 1 || !strings.Contains(n.AskUser[0], "vision, glossary") {
		t.Fatalf("expected one question listing both chapters: %+v", n.AskUser)
	}
	r := tl.Call(ctx, "design_set_status", json.RawMessage(`{"ids":["vision","glossary","nope"],"status":"accepted","approved_by":"paul"}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if strings.Count(txt, `"ok": true`) != 2 || !strings.Contains(txt, `nothing in the design matches \"nope\"`) {
		t.Fatalf("batch result = %s", txt)
	}
	log, _ := exec.Command("git", "-C", root, "log", "-1", "--format=%s").Output()
	if !strings.Contains(string(log), "accepted vision, glossary (approved by paul)") {
		t.Fatalf("last commit = %s", log)
	}
	if n, _ = guide.Compute(ctx, tl.Svc, "", "a"); n.Phase != "5-plan" {
		t.Fatalf("phase = %s", n.Phase)
	}
}

// REQ-PH-004
func TestWorkToolsRefusedBeforePlan(t *testing.T) {
	// Purpose: work tools are listed but refuse calls before phase 5, with the next step.
	root, _ := filepath.Abs(filepath.Join("..", ".."))
	tl := &Tools{Svc: work.New(root, config.Default(), tasks.NewMemory())}
	ctx := context.Background()
	n, _ := guide.Compute(ctx, tl.Svc, guide.RoleDesign, "a")
	if n.Phase >= "5" {
		t.Skip("maid's own design is past phase 4")
	}
	r := tl.Call(ctx, "work_claim", json.RawMessage(`{}`), "a")
	txt := r["content"].([]map[string]any)[0]["text"].(string)
	if r["isError"] != true || !strings.Contains(txt, "next step:") {
		t.Fatalf("work_claim before phase 5 = %s", txt)
	}
}
