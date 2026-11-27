package guide

import (
	"context"

	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

// Status is the maid_status response: what humans need to see progress (G-007).
type Status struct {
	Next      Next            `json:"next"`
	Chapters  map[string]int  `json:"chapters_by_status"`
	WorkItems map[string]int  `json:"work_items_by_status"`
	Escalated []string        `json:"escalated,omitempty"`
	Items     []ItemStatus    `json:"items,omitempty"`
	Baseline  BaselineSummary `json:"baseline"`
	Recent    []state.Event   `json:"recent_activity"`
	Warnings  []string        `json:"warnings,omitempty"`
}

// ItemStatus is one work item in the status report.
type ItemStatus struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Kind       string   `json:"kind"`
	Status     string   `json:"status"`
	Assignee   string   `json:"assignee,omitempty"`
	Implements []string `json:"implements"`
	Attempts   int      `json:"attempts"`
}

// BaselineSummary counts main-branch test results.
type BaselineSummary struct {
	Commit string `json:"commit,omitempty"`
	Pass   int    `json:"pass"`
	Fail   int    `json:"fail"`
	Skip   int    `json:"skip"`
}

// Report builds the status report.
func Report(ctx context.Context, s *work.Service, detail bool) (Status, error) {
	// Purpose: summarise design, work items, tests and recent activity.
	st := Status{Chapters: map[string]int{}, WorkItems: map[string]int{}}
	n, err := Compute(ctx, s, RoleDesign, "")
	if err != nil {
		return st, err
	}
	st.Next = n
	ds, err := s.Design()
	if err != nil {
		return st, err
	}
	for _, ch := range ds.Chapters {
		st.Chapters[ch.FM.String("status")]++
	}
	st.Warnings = n.Warnings
	items, err := s.Tasks.List(ctx, true)
	if err != nil {
		items = nil // reported in Warnings
	}
	for _, it := range items {
		if it.Maid.Kind == "" {
			continue
		}
		key := it.Status
		if it.Maid.Escalated && it.Status != tasks.StatusClosed {
			key = "escalated"
			st.Escalated = append(st.Escalated, it.ID)
		}
		st.WorkItems[key]++
		if detail {
			st.Items = append(st.Items, ItemStatus{ID: it.ID, Title: it.Title, Kind: it.Maid.Kind, Status: key,
				Assignee: it.Assignee, Implements: it.Maid.Implements, Attempts: it.Maid.Attempts})
		}
	}
	b, err := s.State.LoadBaseline()
	if err == nil {
		st.Baseline.Commit = b.Commit
		for _, r := range b.Tests {
			switch r {
			case "pass":
				st.Baseline.Pass++
			case "fail":
				st.Baseline.Fail++
			case "skip":
				st.Baseline.Skip++
			}
		}
	}
	st.Recent, _ = s.State.Recent(15)
	return st, nil
}
