package design

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scenario: maid's own design uses the layout
func TestOwnDesignLayout(t *testing.T) {
	// Purpose: maid's own design has no position numbers and has index.md and index.typ.
	root := filepath.Join("..", "..", Dir)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if p != root && name[0] >= '0' && name[0] <= '9' {
			t.Errorf("%s starts with a digit", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "\n- foundation/vision.md\n") {
		t.Error("design/index.md does not list - foundation/vision.md")
	}
	if _, err := os.Stat(filepath.Join(root, "index.typ")); err != nil {
		t.Error("design/index.typ does not exist")
	}
}
