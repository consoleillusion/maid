package work

import "context"

// Edit is one exact text replacement in a file (REQ-WI-031).
type Edit struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

// EditFile applies exact text replacements to owned files in the agent's
// worktree (REQ-WI-031, REQ-WI-032).
func (s *Service) EditFile(ctx context.Context, agent, id string, edits []Edit) error {
	// Purpose: stub; the code work item implements it.
	return nil
}
