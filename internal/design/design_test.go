package design

import (
	"os"
	"path/filepath"
	"testing"
)

func writeDesign(t *testing.T, files map[string]string) string {
	// Purpose: create a temporary project with the given design files.
	root := t.TempDir()
	for p, c := range files {
		full := filepath.Join(root, Dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const goodChapter = `---
status: draft
style: ears
owner: paul
summary: Alpha chapter.
links: [beta]
---

# Alpha

## Requirements

### REQ-AL-001 One
The design store shall load chapters.

### REQ-AL-002 Two
When a chapter changes, maid shall regenerate the index.
`

const betaChapter = `---
status: accepted
style: ste
owner: paul
summary: Beta chapter.
links: []
---

# Beta

This links to REQ-AL-001 and to [alpha](alpha.md).

| ID | Thing |
|---|---|
| Q-001 | A question |
`

// REQ-DOC-008
func TestLoadIndexRead(t *testing.T) {
	// Purpose: load chapters, list them, and read chapters, items and anchors.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	idx := s.Index(Filter{})
	if len(idx) != 2 || idx[0].ID != "alpha" || idx[0].Items != 2 {
		t.Fatalf("index = %+v", idx)
	}
	if got := s.Index(Filter{Status: "accepted"}); len(got) != 1 || got[0].ID != "beta" {
		t.Fatalf("filtered = %+v", got)
	}
	ex, err := s.Read([]string{"REQ-AL-002", "alpha#requirements", "Q-001", "beta"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if ex[0].Text != "### REQ-AL-002 Two\nWhen a chapter changes, maid shall regenerate the index." {
		t.Fatalf("item text = %q", ex[0].Text)
	}
	if ex[1].Line != 11 {
		t.Fatalf("section line = %d", ex[1].Line)
	}
	if ex[2].Text != "| Q-001 | A question |" || ex[3].Text != "Beta chapter." {
		t.Fatalf("got %q / %q", ex[2].Text, ex[3].Text)
	}
	if _, err := s.Read([]string{"nope"}, false); err == nil {
		t.Fatal("unknown ref did not fail")
	}
}

// REQ-DOC-009 REQ-DOC-006 REQ-DOC-015 REQ-STY-001
func TestCheckFindsProblems(t *testing.T) {
	// Purpose: report bad status, bad style, duplicate items and broken references.
	bad := "---\nstatus: wip\nstyle: klingon\nsummary: s\nlinks: [missing]\n---\n\n### REQ-AL-001 Dup\nThe store shall x.\n\nSee REQ-ZZ-999 and [x](nope.md).\n"
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/bad.md": bad})
	s, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, f := range s.Check() {
		got[f.Rule]++
	}
	want := map[string]int{"REQ-DOC-009": 1, "REQ-STY-001": 1, "REQ-DOC-006": 1, "REQ-DOC-015": 4}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %d, want %d (all %v)", k, got[k], v, got)
		}
	}
}

// REQ-DOC-015
func TestOwnDesignIsClean(t *testing.T) {
	// Purpose: maid's own design must pass maid's own checks.
	s, err := Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range s.Check() {
		t.Errorf("%s:%d %s %s", f.Path, f.Line, f.Rule, f.Message)
	}
}
