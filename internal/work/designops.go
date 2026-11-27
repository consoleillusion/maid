package work

import (
	"context"
	"fmt"
	"strings"

	"github.com/consoleillusion/maid/internal/design"
	"github.com/consoleillusion/maid/internal/gitx"
	"github.com/consoleillusion/maid/internal/state"
)

// DesignOp runs one change to design/ under the design lock, then commits
// the changed file so the change is recorded and visible to new worktrees.
func (s *Service) DesignOp(ctx context.Context, agent, action, reason string, op func(*design.Store) (design.WriteResult, error)) (design.WriteResult, error) {
	// Purpose: serialise design writes across processes and commit each one.
	lock, err := s.State.Lock("design")
	if err != nil {
		return design.WriteResult{}, err
	}
	defer lock.Unlock()
	ds, err := s.Design()
	if err != nil {
		return design.WriteResult{}, err
	}
	res, err := op(ds)
	if err != nil {
		return res, err
	}
	if strings.TrimSpace(reason) == "" {
		reason = action + " " + res.Ref
	}
	msg := fmt.Sprintf("design(%s): %s\n\n%s by %s through maid.", res.Ref, firstLine(reason), action, agent)
	if _, err := gitx.Run(ctx, s.Root, "add", "--", res.Path); err == nil {
		if _, err := gitx.Run(ctx, s.Root, "commit", "-q", "-m", msg, "--", res.Path); err != nil {
			res.Findings = append(res.Findings, design.Finding{Path: res.Path, Rule: "git", Severity: "advice",
				Message: "the change is saved but not committed: " + err.Error()})
		}
	}
	_ = s.State.Log(state.Event{Agent: agent, Action: action, Detail: res.Ref + ": " + firstLine(reason)})
	return res, nil
}

func firstLine(s string) string {
	// Purpose: keep a commit subject to one line.
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 72 {
		s = s[:69] + "..."
	}
	return s
}
