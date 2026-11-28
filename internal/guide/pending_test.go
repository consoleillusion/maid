package guide

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/consoleillusion/maid/internal/design"
)

func chapter(id, part string) string {
	// Purpose: build a minimal draft chapter.
	return "---\nid: " + id + "\ntitle: T\npart: " + part + "\nstatus: draft\nstyle: ste\nowner: paul\nsummary: S.\nlinks: []\n---\n\n# T\n\nText.\n"
}

// REQ-DOC-014 scenario: pending in index order
func TestPendingFollowsIndex(t *testing.T) {
	// Purpose: chapters wait for approval in reading order, and only listed chapters count.
	root := t.TempDir()
	files := map[string]string{
		"index.md":               "---\nid: index\ntitle: Design\nstatus: draft\nstyle: ste\nowner: paul\nsummary: Order.\nlinks: []\n---\n\n# Design\n\n## Method\n\n1. [Phases](method/phases.md)\n\n## Foundation\n\n1. [Glossary](foundation/glossary.md)\n",
		"method/phases.md":       chapter("phases", "method"),
		"foundation/glossary.md": chapter("glossary", "foundation"),
		"other/extra.md":         chapter("extra", "other"),
	}
	for p, c := range files {
		full := filepath.Join(root, "design", p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ds, err := design.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := Pending(ds); !reflect.DeepEqual(got, []string{"phases", "glossary"}) {
		t.Errorf("Pending = %v, want [phases glossary]", got)
	}
}
