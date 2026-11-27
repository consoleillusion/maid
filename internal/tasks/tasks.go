// Package tasks defines the task port: work items, dependencies, claims and
// leases. See design/04-architecture/02-task-store.md.
package tasks

import (
	"context"
	"errors"
	"time"
)

// Status values used by maid. They map onto Beads statuses.
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

// Kinds of work item (REQ-WI-017).
const (
	KindCode  = "code"
	KindTests = "tests"
	KindDocs  = "docs"
)

// ErrConflict means another agent holds the claim (REQ-WI-004).
var ErrConflict = errors.New("work item is claimed by another agent")

// ErrNotFound means the work item does not exist.
var ErrNotFound = errors.New("work item not found")

// ErrNoReady means no work item is ready.
var ErrNoReady = errors.New("no ready work item")

// TestRef names acceptance tests in one file.
type TestRef struct {
	File  string   `json:"file"`
	Names []string `json:"names,omitempty"`
}

// Budget limits a work item (REQ-WI-001, REQ-WI-009).
type Budget struct {
	Attempts        int `json:"attempts,omitempty"`
	MaxChangedLines int `json:"max_changed_lines,omitempty"`
}

// Meta is the maid-specific data stored in the task store under "maid".
type Meta struct {
	Key             string    `json:"key,omitempty"` // plan key, unique within a plan
	Kind            string    `json:"kind"`
	Implements      []string  `json:"implements"`
	Criteria        string    `json:"criteria,omitempty"`
	AcceptanceTests []TestRef `json:"acceptance_tests,omitempty"`
	Owns            []string  `json:"owns"`
	Reads           []string  `json:"reads,omitempty"`
	TestsFor        string    `json:"tests_for,omitempty"`  // id of the code item
	TestsItem       string    `json:"tests_item,omitempty"` // id of its tests item
	TestsAuthor     string    `json:"tests_author,omitempty"`
	Budget          Budget    `json:"budget"`
	Attempts        int       `json:"attempts,omitempty"`
	Escalated       bool      `json:"escalated,omitempty"`
}

// WorkItem is a work item as maid sees it.
type WorkItem struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Assignee     string    `json:"assignee,omitempty"`
	LeaseExpires time.Time `json:"lease_expires,omitempty"`
	DependsOn    []string  `json:"depends_on,omitempty"`
	Maid         Meta      `json:"maid"`
}

// NewItem is the input to Create.
type NewItem struct {
	Title     string
	DependsOn []string // IDs of existing items
	Maid      Meta
}

// Port is the task store interface. maid code uses only this interface.
type Port interface {
	Create(ctx context.Context, item NewItem) (string, error)
	Get(ctx context.Context, id string) (WorkItem, error)
	List(ctx context.Context, all bool) ([]WorkItem, error)
	Ready(ctx context.Context) ([]WorkItem, error)
	Claim(ctx context.Context, id, agent string) (WorkItem, error)
	Heartbeat(ctx context.Context, id, agent string) error
	Release(ctx context.Context, id, agent string) error
	Close(ctx context.Context, id, reason string) error
	SetMeta(ctx context.Context, id string, m Meta) error
}
