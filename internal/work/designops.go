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
	if changed, _ := gitx.Run(ctx, s.Root, "status", "--porcelain", "--", res.Path); changed == "" {
		return res, nil // nothing changed, for example a status that was already set
	}
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

// StatusResult is the outcome for one chapter in a batch status change.
type StatusResult struct {
	ID    string `json:"id"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// SetStatusBatch changes the status of many chapters under one lock and
// records them in one commit, so a user can approve a set at once.
func (s *Service) SetStatusBatch(ctx context.Context, agent string, ids []string, status, approvedBy string) ([]StatusResult, error) {
	// Purpose: accept or change many chapters with one call and one commit.
	lock, err := s.State.Lock("design")
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()
	ds, err := s.Design()
	if err != nil {
		return nil, err
	}
	var out []StatusResult
	var paths, done []string
	for _, id := range ids {
		res, err := ds.SetStatus(id, status, approvedBy)
		if err != nil {
			out = append(out, StatusResult{ID: id, Error: err.Error()})
			continue
		}
		out = append(out, StatusResult{ID: res.Ref, OK: true})
		if changed, _ := gitx.Run(ctx, s.Root, "status", "--porcelain", "--", res.Path); changed == "" {
			continue // already had this status
		}
		paths = append(paths, res.Path)
		done = append(done, res.Ref)
	}
	if len(paths) == 0 {
		return out, nil
	}
	// The commit names who approved the change, or the agent that made it
	// when the user delegated approval (REQ-PH-007).
	who := " (by " + agent + ")"
	if approvedBy != "" {
		who = " (approved by " + approvedBy + ")"
	}
	subject := fmt.Sprintf("design: %s %s%s", status, strings.Join(done, ", "), who)
	if len(done) > 3 {
		subject = fmt.Sprintf("design: %s %d chapters%s", status, len(done), who)
	}
	msg := firstLine(subject) + "\n\n" + strings.Join(done, "\n") + "\n\n" + "status by " + agent + " through maid."
	args := append([]string{"add", "--"}, paths...)
	if _, err := gitx.Run(ctx, s.Root, args...); err == nil {
		args = append([]string{"commit", "-q", "-m", msg, "--"}, paths...)
		if _, err := gitx.Run(ctx, s.Root, args...); err != nil {
			out = append(out, StatusResult{ID: "git", Error: "saved but not committed: " + err.Error()})
		}
	}
	_ = s.State.Log(state.Event{Agent: agent, Action: "status", Detail: subject})
	return out, nil
}
