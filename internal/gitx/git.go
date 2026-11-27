// Package gitx runs the git CLI for worktrees, rebases and merges.
package gitx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Repo is a git repository at Dir.
type Repo struct {
	Dir string
}

// Run runs git in dir and returns trimmed stdout.
func Run(ctx context.Context, dir string, args ...string) (string, error) {
	// Purpose: run one git command and return a readable error.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()), fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(errb.String()+" "+out.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// Head returns the commit of a ref.
func (r Repo) Head(ctx context.Context, ref string) (string, error) {
	// Purpose: resolve a ref to a commit ID.
	return Run(ctx, r.Dir, "rev-parse", "--verify", ref+"^{commit}")
}

// AddWorktree creates a worktree at path on a new branch from base.
func (r Repo) AddWorktree(ctx context.Context, path, branch, base string) error {
	// Purpose: give an attempt its own isolated checkout (REQ-WI-007).
	_, err := Run(ctx, r.Dir, "worktree", "add", "-q", "-b", branch, path, base)
	return err
}

// AddDetached creates a detached worktree at path on commit.
func (r Repo) AddDetached(ctx context.Context, path, commit string) error {
	// Purpose: create a scratch checkout for the merge queue.
	_, err := Run(ctx, r.Dir, "worktree", "add", "-q", "--detach", path, commit)
	return err
}

// RemoveWorktree removes a worktree and prunes its metadata.
func (r Repo) RemoveWorktree(ctx context.Context, path string) error {
	// Purpose: clean up a finished worktree.
	_, err := Run(ctx, r.Dir, "worktree", "remove", "--force", path)
	_, _ = Run(ctx, r.Dir, "worktree", "prune")
	return err
}

// Rebase rebases the detached HEAD in worktree onto onto. On conflict it
// aborts and returns the conflicting paths.
func Rebase(ctx context.Context, worktree, onto string) (conflicts []string, err error) {
	// Purpose: replay an attempt on the latest main branch (REQ-ACC-004, REQ-ACC-005).
	if _, err := Run(ctx, worktree, "rebase", "-q", onto); err != nil {
		files, _ := Run(ctx, worktree, "diff", "--name-only", "--diff-filter=U")
		_, _ = Run(ctx, worktree, "rebase", "--abort")
		if files == "" {
			return nil, err
		}
		return strings.Split(files, "\n"), nil
	}
	return nil, nil
}

// Change is one changed file with line counts.
type Change struct {
	Path    string
	Added   int
	Deleted int
	Status  string // A, M, D, R...
}

// Diff lists files changed between base and head.
func (r Repo) Diff(ctx context.Context, base, head string) ([]Change, error) {
	// Purpose: list changed paths and line counts for path and size checks.
	num, err := Run(ctx, r.Dir, "diff", "--numstat", "--no-renames", base, head)
	if err != nil {
		return nil, err
	}
	st, err := Run(ctx, r.Dir, "diff", "--name-status", "--no-renames", base, head)
	if err != nil {
		return nil, err
	}
	status := map[string]string{}
	for _, l := range lines(st) {
		if f := strings.SplitN(l, "\t", 2); len(f) == 2 {
			status[f[1]] = f[0]
		}
	}
	var out []Change
	for _, l := range lines(num) {
		f := strings.SplitN(l, "\t", 3)
		if len(f) != 3 {
			continue
		}
		a, _ := strconv.Atoi(f[0]) // "-" for binary gives 0
		d, _ := strconv.Atoi(f[1])
		out = append(out, Change{Path: f[2], Added: a, Deleted: d, Status: status[f[2]]})
	}
	return out, nil
}

// FastForward moves branch to commit if commit descends from it.
func (r Repo) FastForward(ctx context.Context, branch, commit string) error {
	// Purpose: merge an accepted attempt without a merge commit (REQ-ACC-007).
	old, err := r.Head(ctx, branch)
	if err != nil {
		return err
	}
	if _, err := Run(ctx, r.Dir, "merge-base", "--is-ancestor", old, commit); err != nil {
		return fmt.Errorf("%s is not an ancestor of %s", branch, commit)
	}
	cur, _ := Run(ctx, r.Dir, "symbolic-ref", "--short", "-q", "HEAD")
	if cur == branch {
		_, err = Run(ctx, r.Dir, "merge", "--ff-only", "-q", commit)
		return err
	}
	_, err = Run(ctx, r.Dir, "update-ref", "refs/heads/"+branch, commit, old)
	return err
}

// FileExists reports whether path exists in commit.
func (r Repo) FileExists(ctx context.Context, commit, path string) bool {
	// Purpose: check that acceptance test files exist on the main branch.
	_, err := Run(ctx, r.Dir, "cat-file", "-e", commit+":"+path)
	return err == nil
}

// Show returns the content of path at commit.
func (r Repo) Show(ctx context.Context, commit, path string) (string, error) {
	// Purpose: read a file from a commit without a checkout.
	return Run(ctx, r.Dir, "show", commit+":"+path)
}

func lines(s string) []string {
	// Purpose: split command output into non-empty lines.
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
