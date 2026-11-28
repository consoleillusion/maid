package work

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/gitx"
	"github.com/consoleillusion/maid/internal/glob"
	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/testrun"
)

// These operations let an agent without file or shell access complete a
// work item through maid alone. Each one needs the
// agent's claim and works only in the agent's own worktree.

// maxFileBytes limits one read or write.
const maxFileBytes = 512 * 1024

// attempt returns the open attempt of agent on id, checking the claim.
func (s *Service) attempt(ctx context.Context, agent, id string) (tasks.WorkItem, state.Attempt, error) {
	// Purpose: make sure the agent holds the claim before touching its worktree.
	it, err := s.Tasks.Get(ctx, id)
	if err != nil {
		return it, state.Attempt{}, err
	}
	if it.Status != tasks.StatusInProgress || it.Assignee != agent {
		return it, state.Attempt{}, fmt.Errorf("%s does not hold the claim on %s; call work_claim first", agent, id)
	}
	att, ok, err := s.State.Current(id, agent)
	if err != nil {
		return it, att, err
	}
	if !ok {
		return it, att, fmt.Errorf("no open attempt of %s on %s; call work_claim", agent, id)
	}
	return it, att, nil
}

// cleanPath checks that p is a relative path inside the worktree.
func cleanPath(p string) (string, error) {
	// Purpose: reject absolute paths, parent references and git internals.
	p = strings.TrimPrefix(path.Clean("/"+filepath.ToSlash(strings.TrimSpace(p))), "/")
	if p == "" || p == "." {
		return "", errors.New("path is empty")
	}
	if p == ".git" || strings.HasPrefix(p, ".git/") || strings.HasPrefix(p, ".maid/") || strings.HasPrefix(p, ".beads/") {
		return "", fmt.Errorf("%s is not a project file", p)
	}
	return p, nil
}

// FileEntry is one file in a worktree listing.
type FileEntry struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
	Owned bool   `json:"owned"` // the agent may change it
}

// Files lists the files in the agent's worktree, marking the ones it owns.
func (s *Service) Files(ctx context.Context, agent, id string) ([]FileEntry, error) {
	// Purpose: show an agent what it can read and what it can change.
	it, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return nil, err
	}
	var out []FileEntry
	err = filepath.WalkDir(att.Worktree, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(att.Worktree, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || rel == ".maid" || rel == ".beads" || rel == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == ".git" {
			return nil
		}
		info, _ := d.Info()
		var size int64
		if info != nil {
			size = info.Size()
		}
		out = append(out, FileEntry{Path: rel, Bytes: size, Owned: glob.Any(it.Maid.Owns, rel)})
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

// ReadFile returns a file from the agent's worktree.
func (s *Service) ReadFile(ctx context.Context, agent, id, p string) (string, error) {
	// Purpose: let an agent read any project file in its worktree.
	_, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return "", err
	}
	p, err = cleanPath(p)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(filepath.Join(att.Worktree, filepath.FromSlash(p)))
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("%s does not exist; call work_files to list files", p)
	}
	if err != nil {
		return "", err
	}
	if len(b) > maxFileBytes {
		return "", fmt.Errorf("%s is %d bytes; the limit is %d", p, len(b), maxFileBytes)
	}
	return string(b), nil
}

// WriteFile creates or replaces an owned file in the agent's worktree.
// Empty content with remove set deletes the file.
func (s *Service) WriteFile(ctx context.Context, agent, id, p, content string, remove bool) error {
	// Purpose: let an agent change only the files its work item owns.
	it, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return err
	}
	p, err = cleanPath(p)
	if err != nil {
		return err
	}
	if !glob.Any(it.Maid.Owns, p) {
		return fmt.Errorf("%s is not owned by %s; you may change only: %s", p, id, strings.Join(it.Maid.Owns, ", "))
	}
	if len(content) > maxFileBytes {
		return fmt.Errorf("content is %d bytes; the limit is %d", len(content), maxFileBytes)
	}
	full := filepath.Join(att.Worktree, filepath.FromSlash(p))
	if remove {
		if err := os.Remove(full); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

// RunResult is the result of a check run in a worktree.
type RunResult struct {
	Check    string            `json:"check"`
	Passed   bool              `json:"passed"`
	Output   string            `json:"output,omitempty"`
	Tests    map[string]string `json:"tests,omitempty"`
	Failures map[string]string `json:"failures,omitempty"`
	Build    []string          `json:"build_failed,omitempty"`
}

// Run runs the configured test, format or lint command in the worktree.
func (s *Service) Run(ctx context.Context, agent, id, check string) (RunResult, error) {
	// Purpose: let an agent check its work before it submits.
	it, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return RunResult{}, err
	}
	switch check {
	case "", "test":
		res, err := testrun.Run(ctx, att.Worktree, s.Cfg.Test.Adapter, s.Cfg.Test.Command, s.attemptEnv(it.ID, att.Number))
		if err != nil {
			return RunResult{}, err
		}
		ok := res.ExitCode == 0 && len(res.BuildFailed) == 0
		return RunResult{Check: "test", Passed: ok, Tests: res.Tests, Failures: res.Failures, Build: res.BuildFailed, Output: res.Output}, nil
	case "format":
		msg, ok := runCheck(ctx, att.Worktree, s.Cfg.FormatCheck, true)
		return RunResult{Check: check, Passed: ok, Output: msg}, nil
	case "lint":
		msg, ok := runCheck(ctx, att.Worktree, s.Cfg.Lint, false)
		return RunResult{Check: check, Passed: ok, Output: msg}, nil
	}
	return RunResult{}, fmt.Errorf("unknown check %q: use test, format or lint", check)
}

// Commit commits every change in the agent's worktree.
func (s *Service) Commit(ctx context.Context, agent, id, message string) (string, error) {
	// Purpose: record the agent's changes on its attempt branch.
	_, att, err := s.attempt(ctx, agent, id)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(message) == "" {
		return "", errors.New("message is empty; describe the change in one line")
	}
	if _, err := gitx.Run(ctx, att.Worktree, "add", "-A"); err != nil {
		return "", err
	}
	if st, _ := gitx.Run(ctx, att.Worktree, "status", "--porcelain"); st == "" {
		return "", errors.New("nothing to commit; change files with work_edit or work_write first")
	}
	if _, err := gitx.Run(ctx, att.Worktree, "-c", "user.name="+agent, "-c", "user.email="+agent+"@maid.invalid",
		"commit", "-q", "-m", message); err != nil {
		return "", err
	}
	return gitx.Run(ctx, att.Worktree, "rev-parse", "HEAD")
}

// applyChanges writes the files, applies the edits and the patch of a submission in
// the agent's worktree, then commits them.
func (s *Service) applyChanges(ctx context.Context, sub Submission) error {
	// Purpose: turn a patch or a set of files into a commit on the attempt branch.
	it, att, err := s.attempt(ctx, sub.Agent, sub.ID)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(sub.Files))
	for p := range sub.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if err := s.WriteFile(ctx, sub.Agent, sub.ID, p, sub.Files[p], false); err != nil {
			return err
		}
	}
	if len(sub.Edits) > 0 {
		if err := s.EditFile(ctx, sub.Agent, sub.ID, sub.Edits); err != nil {
			return err
		}
	}
	if strings.TrimSpace(sub.Patch) != "" {
		if err := s.applyPatch(ctx, it, att.Worktree, sub.Patch); err != nil {
			return err
		}
	}
	msg := sub.Message
	if strings.TrimSpace(msg) == "" {
		msg = it.Title
	}
	_, err = s.Commit(ctx, sub.Agent, sub.ID, msg)
	return err
}

func (s *Service) applyPatch(ctx context.Context, it tasks.WorkItem, worktree, patch string) error {
	// Purpose: apply a unified diff after checking that it touches only owned paths.
	if !strings.HasSuffix(patch, "\n") {
		patch += "\n"
	}
	tmp, err := os.CreateTemp("", "maid-patch-*.diff")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(patch); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()
	names, err := gitx.Run(ctx, worktree, "apply", "--numstat", "--recount", "--ignore-whitespace", tmp.Name())
	if err != nil {
		return fmt.Errorf("the patch is not a valid unified diff: %w", err)
	}
	var outside []string
	for _, l := range strings.Split(names, "\n") {
		f := strings.SplitN(l, "\t", 3)
		if len(f) == 3 && !glob.Any(it.Maid.Owns, f[2]) {
			outside = append(outside, f[2])
		}
	}
	if len(outside) > 0 {
		return fmt.Errorf("the patch changes files that %s does not own: %s; you may change only: %s",
			it.ID, strings.Join(outside, ", "), strings.Join(it.Maid.Owns, ", "))
	}
	if _, err := gitx.Run(ctx, worktree, "apply", "--recount", "--ignore-whitespace", "--whitespace=nowarn", tmp.Name()); err != nil {
		return fmt.Errorf("the patch does not apply to the current files; read them with work_read and send the patch again: %w", err)
	}
	return nil
}
