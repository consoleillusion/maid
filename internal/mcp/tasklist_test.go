package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/guide"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

const tasklistIndex = `---
id: index
title: Design
status: draft
style: ste
owner: paul
summary: Reading order.
links: []
---

# Design

The design.

## Foundation

1. [Vision](foundation/vision.md)
`

const tasklistVision = `---
id: vision
title: Vision
part: foundation
status: accepted
style: ste
owner: paul
summary: Why.
links: []
---

# Vision

### Goal
The tool adds numbers.

### No floats
The tool does not add floats.
`

func writeTaskDesign(t *testing.T, root, rel, content string) {
	// Purpose: write one design file into a temporary project.
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tasklistProject(t *testing.T) (*work.Service, *tasks.Memory) {
	// Purpose: build a project with an accepted requirement and an empty task store.
	t.Helper()
	root := t.TempDir()
	writeTaskDesign(t, root, "design/index.md", tasklistIndex)
	writeTaskDesign(t, root, "design/foundation/vision.md", tasklistVision)
	port := tasks.NewMemory()
	return work.New(root, config.Default(), port), port
}

// scenario: several ready tasks
func TestNextListsReadyTasks(t *testing.T) {
	// Purpose: maid_next lists every ready work item with its id, title and kind.
	svc, port := tasklistProject(t)
	ctx := context.Background()
	for _, it := range []struct{ title, file string }{
		{"First tests", "first/first_test.go"},
		{"Second tests", "second/second_test.go"},
	} {
		if _, err := port.Create(ctx, tasks.NewItem{Title: it.title, Maid: tasks.Meta{
			Kind: tasks.KindTests, Owns: []string{it.file},
			AcceptanceTests: []tasks.TestRef{{File: it.file, Names: []string{"TestOne"}}},
		}}); err != nil {
			t.Fatal(err)
		}
	}
	n, err := guide.Compute(ctx, svc, "", "builder")
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Tasks) != 2 {
		t.Fatalf("Tasks = %+v, want two ready items", n.Tasks)
	}
	ids := map[string]bool{}
	for _, task := range n.Tasks {
		if task.ID == "" || task.Title == "" {
			t.Errorf("task has no id or title: %+v", task)
		}
		if task.Kind != tasks.KindTests {
			t.Errorf("task %q kind = %q, want tests", task.Title, task.Kind)
		}
		ids[task.ID] = true
	}
	if len(ids) != 2 {
		t.Errorf("task ids are not unique: %+v", n.Tasks)
	}
}

// scenario: no ready tasks
func TestNextListsNoTasks(t *testing.T) {
	// Purpose: tasks is an empty list, not null, when nothing is ready.
	svc, _ := tasklistProject(t)
	n, err := guide.Compute(context.Background(), svc, "", "builder")
	if err != nil {
		t.Fatal(err)
	}
	if n.Tasks == nil {
		t.Fatal("Tasks is null, want an empty list")
	}
	if len(n.Tasks) != 0 {
		t.Fatalf("Tasks = %+v, want no ready items", n.Tasks)
	}
}

// scenario: plain prompts
func TestInstructionsForPlainPrompts(t *testing.T) {
	// Purpose: initialize tells agents to work on the tasks that maid_next lists.
	const sentence = "Whenever the user asks you to work on this project, call maid_next and work on the tasks it lists."
	out := session(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"test"},"capabilities":{}}}`)
	if len(out) != 1 {
		t.Fatalf("got %d replies: %v", len(out), out)
	}
	res, ok := out[0]["result"].(map[string]any)
	if !ok {
		t.Fatalf("no initialize result: %v", out[0])
	}
	instructions, _ := res["instructions"].(string)
	if !strings.Contains(instructions, sentence) {
		t.Fatalf("instructions = %q, want it to contain %q", instructions, sentence)
	}
}
