package work

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/glob"
)

// Edit is one exact text replacement in a file.
type Edit struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

// EditFile applies exact text replacements to owned files in the agent's
// worktree. Either every edit is applied or none.
func (s *Service) EditFile(ctx context.Context, agent, id string, edits []Edit) error {
	// Purpose: change parts of owned files without sending whole files.
	it, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return err
	}
	next, err := planEdits(att.Worktree, it.ID, it.Maid.Owns, edits)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(next))
	for p := range next {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		full := filepath.Join(att.Worktree, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(next[p]), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// planEdits computes the new content of each edited file without writing.
func planEdits(worktree, id string, owns []string, edits []Edit) (map[string]string, error) {
	// Purpose: check every edit before any file changes.
	if len(edits) == 0 {
		return nil, errors.New(`edits is empty; pass [{"path": "...", "old": "exact text", "new": "replacement"}]`)
	}
	next := map[string]string{}
	for i, e := range edits {
		p, err := cleanPath(e.Path)
		if err != nil {
			return nil, fmt.Errorf("edit %d: %w", i, err)
		}
		if !glob.Any(owns, p) {
			return nil, fmt.Errorf("edit %d: %s is not owned by %s; you may change only: %s", i, p, id, strings.Join(owns, ", "))
		}
		cur, ok := next[p]
		if !ok {
			b, err := os.ReadFile(filepath.Join(worktree, filepath.FromSlash(p)))
			switch {
			case errors.Is(err, fs.ErrNotExist) && e.Old == "":
				next[p] = e.New
				continue
			case errors.Is(err, fs.ErrNotExist):
				return nil, fmt.Errorf("edit %d: %s does not exist; use an empty old text to create it", i, p)
			case err != nil:
				return nil, err
			}
			cur = string(b)
		}
		if e.Old == "" {
			return nil, fmt.Errorf("edit %d: old text is empty but %s exists; give the exact text to replace", i, p)
		}
		if n := strings.Count(cur, e.Old); n != 1 {
			return nil, fmt.Errorf("edit %d: old text has %d matches in %s; it must match exactly once, so include more surrounding text. No file was changed", i, n, p)
		}
		cur = strings.Replace(cur, e.Old, e.New, 1)
		if len(cur) > maxFileBytes {
			return nil, fmt.Errorf("edit %d: %s would be %d bytes; the limit is %d", i, p, len(cur), maxFileBytes)
		}
		next[p] = cur
	}
	return next, nil
}
