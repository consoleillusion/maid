package work

import (
	"context"
	"fmt"
	"strings"

	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/tasks"
)

// Packet is the context packet of a claimed work item (REQ-WI-013).
type Packet struct {
	WorkItem     tasks.WorkItem    `json:"work_item"`
	Attempt      int               `json:"attempt"`
	Worktree     string            `json:"worktree"`
	Branch       string            `json:"branch"`
	Instructions []string          `json:"instructions"`
	Requirements []string          `json:"requirements"`
	Criteria     string            `json:"criteria,omitempty"`
	TestFiles    map[string]string `json:"test_files,omitempty"`
	Owns         []string          `json:"owns"`
	Reads        []string          `json:"reads,omitempty"`
	Conventions  string            `json:"conventions,omitempty"`
	Commands     map[string]string `json:"commands"`
	Tokens       int               `json:"tokens"` // estimate (REQ-WI-014)
}

func (s *Service) packet(ctx context.Context, it tasks.WorkItem, att state.Attempt) (Packet, error) {
	// Purpose: collect exactly what a build or test agent needs for one item.
	p := Packet{
		WorkItem: it, Attempt: att.Number, Worktree: att.Worktree, Branch: att.Branch,
		Criteria: it.Maid.Criteria, Owns: it.Maid.Owns, Reads: it.Maid.Reads,
		Commands: map[string]string{"test": strings.Join(s.Cfg.Test.Command, " ")},
	}
	if len(s.Cfg.FormatCheck) > 0 {
		p.Commands["format_check"] = strings.Join(s.Cfg.FormatCheck, " ")
	}
	if len(s.Cfg.Lint) > 0 {
		p.Commands["lint"] = strings.Join(s.Cfg.Lint, " ")
	}
	ds, err := s.Design()
	if err != nil {
		return p, err
	}
	if ex, err := ds.Read(it.Maid.Implements, false); err == nil {
		for _, e := range ex {
			p.Requirements = append(p.Requirements, e.Text)
		}
	}
	if conv, err := ds.Read([]string{"conventions"}, false); err == nil {
		p.Conventions = conv[0].Text
	}
	switch it.Maid.Kind {
	case tasks.KindTests:
		p.Instructions = testsInstructions(it, att, s.Cfg.MainBranch)
	default:
		p.TestFiles = map[string]string{}
		for _, t := range it.Maid.AcceptanceTests {
			if c, err := s.Git.Show(ctx, att.Base, t.File); err == nil {
				p.TestFiles[t.File] = c
			}
		}
		p.Instructions = codeInstructions(it, att)
	}
	p.Tokens = estimateTokens(p)
	return p, nil
}

func codeInstructions(it tasks.WorkItem, att state.Attempt) []string {
	// Purpose: tell a build agent how to complete a code work item.
	var names []string
	for _, t := range it.Maid.AcceptanceTests {
		names = append(names, t.Names...)
	}
	return []string{
		"Work only in the worktree " + att.Worktree + " on branch " + att.Branch + ".",
		"Make these acceptance tests pass: " + strings.Join(names, ", ") + ".",
		"Change only files that match owns: " + strings.Join(it.Maid.Owns, ", ") + ".",
		"Do not change, delete or skip the acceptance test files.",
		"Do not break any test that passes now.",
		fmt.Sprintf("Keep the change under %d changed lines.", it.Maid.Budget.MaxChangedLines),
		"Commit your work on the branch, then call work_submit with id " + it.ID + ".",
		"Call work_heartbeat if you work for longer than a few minutes.",
	}
}

func testsInstructions(it tasks.WorkItem, att state.Attempt, mainBranch string) []string {
	// Purpose: tell a test agent how to write acceptance tests from the criteria.
	var want []string
	for _, t := range it.Maid.AcceptanceTests {
		want = append(want, fmt.Sprintf("%s: %s", t.File, strings.Join(t.Names, ", ")))
	}
	return []string{
		"Work only in the worktree " + att.Worktree + " on branch " + att.Branch + ".",
		"Write acceptance tests from the criteria. Create exactly these tests: " + strings.Join(want, "; ") + ".",
		"Each test must compile and must fail on " + mainBranch + ", because the feature does not exist yet.",
		"Tag each test with the requirement and scenario it covers, in a comment above the test.",
		"Change only the test files listed above. Do not implement the feature.",
		"Commit your work on the branch, then call work_submit with id " + it.ID + ".",
	}
}

func estimateTokens(p Packet) int {
	// Purpose: estimate packet size as characters divided by four.
	n := len(p.Criteria) + len(p.Conventions)
	for _, r := range p.Requirements {
		n += len(r)
	}
	for _, c := range p.TestFiles {
		n += len(c)
	}
	for _, i := range p.Instructions {
		n += len(i)
	}
	return n / 4
}
