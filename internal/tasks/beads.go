package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Beads is the task port adapter for the bd CLI (ADR-0005).
type Beads struct {
	Dir string // project root
	Bin string // path to bd; "bd" when empty
}

// Init initialises Beads in the project without touching AGENTS.md. Stealth
// mode keeps the database local: no Dolt remote, and .beads/ is excluded
// from git through .git/info/exclude (C-004).
func (b *Beads) Init(ctx context.Context) error {
	// Purpose: run bd init with the flags maid requires.
	_, err := b.run(ctx, "", "init", "--non-interactive", "--skip-agents", "--skip-hooks", "--stealth", "--quiet")
	return err
}

// Initialised reports whether the project has a Beads database.
func (b *Beads) Initialised() bool {
	// Purpose: check for the .beads folder.
	st, err := os.Stat(b.Dir + "/.beads")
	return err == nil && st.IsDir()
}

type bdIssue struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	Assignee     string          `json:"assignee"`
	LeaseExpires *time.Time      `json:"lease_expires_at"`
	Metadata     json.RawMessage `json:"metadata"`
	Dependencies []struct {
		DependsOnID string `json:"depends_on_id"`
		Type        string `json:"type"`
	} `json:"dependencies"`
}

func (i bdIssue) item() WorkItem {
	// Purpose: convert a bd JSON issue into a WorkItem.
	w := WorkItem{ID: i.ID, Title: i.Title, Status: i.Status, Assignee: i.Assignee}
	if i.LeaseExpires != nil {
		w.LeaseExpires = *i.LeaseExpires
	}
	for _, d := range i.Dependencies {
		if d.Type == "" || d.Type == "blocks" {
			w.DependsOn = append(w.DependsOn, d.DependsOnID)
		}
	}
	var md struct {
		Maid Meta `json:"maid"`
	}
	if len(i.Metadata) > 0 {
		_ = json.Unmarshal(i.Metadata, &md)
	}
	w.Maid = md.Maid
	return w
}

// Create adds a work item and returns its ID.
func (b *Beads) Create(ctx context.Context, n NewItem) (string, error) {
	// Purpose: create one bd issue with maid metadata and dependencies.
	md, err := json.Marshal(map[string]Meta{"maid": n.Maid})
	if err != nil {
		return "", err
	}
	args := []string{"create", n.Title, "--silent", "--metadata", string(md)}
	if len(n.Maid.Implements) > 0 {
		args = append(args, "--spec-id", n.Maid.Implements[0])
	}
	if n.Maid.Criteria != "" {
		args = append(args, "--acceptance", n.Maid.Criteria)
	}
	if len(n.DependsOn) > 0 {
		args = append(args, "--deps", strings.Join(n.DependsOn, ","))
	}
	out, err := b.run(ctx, "", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Get returns one work item.
func (b *Beads) Get(ctx context.Context, id string) (WorkItem, error) {
	// Purpose: read one issue with bd show.
	out, err := b.run(ctx, "", "show", id, "--json")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return WorkItem{}, ErrNotFound
		}
		return WorkItem{}, err
	}
	items, err := decode(out)
	if err != nil {
		return WorkItem{}, err
	}
	if len(items) == 0 {
		return WorkItem{}, ErrNotFound
	}
	return items[0], nil
}

// List returns open and in-progress items, or all items.
func (b *Beads) List(ctx context.Context, all bool) ([]WorkItem, error) {
	// Purpose: list issues with bd list.
	args := []string{"list", "--json", "--limit", "0"}
	if all {
		args = append(args, "--all")
	}
	out, err := b.run(ctx, "", args...)
	if err != nil {
		return nil, err
	}
	return decode(out)
}

// Ready returns items with no open blocker.
func (b *Beads) Ready(ctx context.Context) ([]WorkItem, error) {
	// Purpose: list ready work with bd ready.
	out, err := b.run(ctx, "", "ready", "--json", "--limit", "0")
	if err != nil {
		return nil, err
	}
	return decode(out)
}

// Claim atomically claims an item for agent.
func (b *Beads) Claim(ctx context.Context, id, agent string) (WorkItem, error) {
	// Purpose: claim with bd update --claim and map a lost race to ErrConflict.
	out, err := b.run(ctx, agent, "update", id, "--claim", "--json")
	if err != nil {
		if strings.Contains(err.Error(), "already claimed") {
			return WorkItem{}, ErrConflict
		}
		return WorkItem{}, err
	}
	items, err := decode(out)
	if err != nil || len(items) == 0 {
		return b.Get(ctx, id)
	}
	return items[0], nil
}

// Heartbeat renews the lease.
func (b *Beads) Heartbeat(ctx context.Context, id, agent string) error {
	// Purpose: renew the claim lease as the claiming agent.
	_, err := b.run(ctx, agent, "heartbeat", id)
	if err != nil && strings.Contains(err.Error(), "claimed by") {
		return ErrConflict
	}
	return err
}

// Release gives up a claim.
func (b *Beads) Release(ctx context.Context, id, agent string) error {
	// Purpose: unclaim an item as the claiming agent.
	_, err := b.run(ctx, agent, "unclaim", id)
	return err
}

// Close closes an item as its current assignee, so bd accepts the close.
func (b *Beads) Close(ctx context.Context, id, reason string) error {
	// Purpose: close an item on behalf of the agent that holds it.
	it, err := b.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = b.run(ctx, it.Assignee, "close", id, "--reason", reason)
	return err
}

// SetMeta replaces the maid metadata of an item.
func (b *Beads) SetMeta(ctx context.Context, id string, m Meta) error {
	// Purpose: store updated maid metadata, such as attempt counts.
	md, err := json.Marshal(map[string]Meta{"maid": m})
	if err != nil {
		return err
	}
	it, err := b.Get(ctx, id)
	if err != nil {
		return err
	}
	_, err = b.run(ctx, it.Assignee, "update", id, "--metadata", string(md))
	return err
}

func (b *Beads) run(ctx context.Context, actor string, args ...string) ([]byte, error) {
	// Purpose: run bd in the project folder as the given actor.
	bin := b.Bin
	if bin == "" {
		bin = "bd"
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = b.Dir
	cmd.Env = append(os.Environ(), "BD_NON_INTERACTIVE=1")
	if actor != "" {
		cmd.Env = append(cmd.Env, "BEADS_ACTOR="+actor)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return nil, fmt.Errorf("bd %s: %s", args[0], msg)
	}
	return stdout.Bytes(), nil
}

func decode(out []byte) ([]WorkItem, error) {
	// Purpose: decode a bd JSON array or object into work items.
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	var issues []bdIssue
	if out[0] == '{' {
		var one bdIssue
		if err := json.Unmarshal(out, &one); err != nil {
			return nil, err
		}
		issues = []bdIssue{one}
	} else if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("decode bd output: %w", err)
	}
	items := make([]WorkItem, 0, len(issues))
	for _, i := range issues {
		items = append(items, i.item())
	}
	return items, nil
}
