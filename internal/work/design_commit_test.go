package work

import (
	"context"
	"testing"
)

// scenario: index files are committed
func TestDesignCommitIncludesIndex(t *testing.T) {
	// Purpose: a chapter created through maid is committed with index.md and index.typ.
	s, root := project(t)
	if _, err := s.DesignOp(context.Background(), "a", "create", "add extra", createOp("b/extra.md")); err != nil {
		t.Fatal(err)
	}
	if st := sh(t, root, "git", "status", "--porcelain"); st != "" {
		t.Errorf("uncommitted changes after a design change:\n%s", st)
	}
}
