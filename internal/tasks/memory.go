package tasks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Memory is an in-memory Port for tests. It follows the same claim rules as Beads.
type Memory struct {
	mu    sync.Mutex
	next  int
	items map[string]*WorkItem
	Lease time.Duration
}

// NewMemory returns an empty in-memory task store.
func NewMemory() *Memory {
	// Purpose: build a Memory port with a default lease.
	return &Memory{items: map[string]*WorkItem{}, Lease: 5 * time.Minute}
}

// Create adds an item.
func (m *Memory) Create(_ context.Context, n NewItem) (string, error) {
	// Purpose: store a new open item with a sequential ID.
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	id := fmt.Sprintf("mem-%d", m.next)
	m.items[id] = &WorkItem{ID: id, Title: n.Title, Status: StatusOpen, DependsOn: n.DependsOn, Maid: n.Maid}
	return id, nil
}

// Get returns a copy of an item.
func (m *Memory) Get(_ context.Context, id string) (WorkItem, error) {
	// Purpose: read one item.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return WorkItem{}, ErrNotFound
	}
	return *it, nil
}

// List returns items in ID order.
func (m *Memory) List(_ context.Context, all bool) ([]WorkItem, error) {
	// Purpose: list open items, or all items.
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []WorkItem
	for _, it := range m.items {
		if all || it.Status != StatusClosed {
			out = append(out, *it)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Ready returns open items whose dependencies are all closed.
func (m *Memory) Ready(ctx context.Context) ([]WorkItem, error) {
	// Purpose: apply the ready rule of the task store.
	all, _ := m.List(ctx, true)
	closed := map[string]bool{}
	for _, it := range all {
		if it.Status == StatusClosed {
			closed[it.ID] = true
		}
	}
	var out []WorkItem
	for _, it := range all {
		if it.Status != StatusOpen {
			continue
		}
		ok := true
		for _, d := range it.DependsOn {
			ok = ok && closed[d]
		}
		if ok {
			out = append(out, it)
		}
	}
	return out, nil
}

// Claim claims an item atomically.
func (m *Memory) Claim(_ context.Context, id, agent string) (WorkItem, error) {
	// Purpose: give the claim to exactly one agent.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return WorkItem{}, ErrNotFound
	}
	if it.Status == StatusInProgress && it.Assignee != agent {
		return WorkItem{}, ErrConflict
	}
	it.Status, it.Assignee, it.LeaseExpires = StatusInProgress, agent, time.Now().Add(m.Lease)
	return *it, nil
}

// Heartbeat renews the lease.
func (m *Memory) Heartbeat(_ context.Context, id, agent string) error {
	// Purpose: extend the lease of the holder.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return ErrNotFound
	}
	if it.Assignee != agent {
		return ErrConflict
	}
	it.LeaseExpires = time.Now().Add(m.Lease)
	return nil
}

// Release clears the claim.
func (m *Memory) Release(_ context.Context, id, agent string) error {
	// Purpose: return an item to open.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return ErrNotFound
	}
	if it.Assignee != agent {
		return ErrConflict
	}
	it.Status, it.Assignee = StatusOpen, ""
	return nil
}

// Close closes an item.
func (m *Memory) Close(_ context.Context, id, _ string) error {
	// Purpose: mark an item closed.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return ErrNotFound
	}
	it.Status = StatusClosed
	return nil
}

// SetMeta replaces maid metadata.
func (m *Memory) SetMeta(_ context.Context, id string, meta Meta) error {
	// Purpose: store updated metadata.
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return ErrNotFound
	}
	it.Maid = meta
	return nil
}
