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
	References   map[string]string `json:"references,omitempty"` // design text named in the criteria (REQ-WI-030)
	Tokens       int               `json:"tokens"`                // estimate (REQ-WI-014)
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

	// Design paths, chapter#anchor refs and item IDs in the criteria are
	// resolved; plain words are not, so a word such as "index" does not pull
	// in a chapter. Refs that do not resolve, and requirements that the
	// packet already holds, are skipped. Each key is the ref as written.
	const refPunct = "`\"'()[]{}<>,;:.!?"
	p.References = map[string]string{}
	for _, tok := range strings.Fields(it.Maid.Criteria) {
		ref := strings.Trim(tok, refPunct)
		if ref == "" || !isDesignRef(ref) || contains(it.Maid.Implements, ref) {
			continue
		}
		if _, seen := p.References[ref]; seen {
			continue
		}
		if ex, err := ds.Read([]string{ref}, false); err == nil && len(ex) > 0 {
			p.References[ref] = ex[0].Text
		}
	}
	if len(p.References) == 0 {
		p.References = nil
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
		"If you have no file access, use work_files and work_read to read, work_edit to change existing files with exact replacements, and work_write only for new files and full rewrites; then work_run and work_commit. Or send edits or a patch with work_submit.",
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
		"If the code the tests call does not exist, add a stub in the owned non-test files: the exact names and signatures from the criteria, with bodies that return zero values. The stub must not implement the behaviour.",
		"Tag each test with the requirement and scenario it covers, in a comment above the test.",
		"Change only the owned files: " + strings.Join(it.Maid.Owns, ", ") + ".",
		"If you have no file access, use work_files and work_read to read, work_edit to change existing files with exact replacements, and work_write only for new files and full rewrites; then work_run and work_commit. Or send edits or a patch with work_submit.",
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
	for _, r := range p.References {
		n += len(r)
	}
	for _, i := range p.Instructions {
		n += len(i)
	}
	return n / 4
}

func isDesignRef(ref string) bool {
	// Purpose: accept design file paths, chapter#anchor refs and item IDs, not plain words.
	if strings.HasPrefix(ref, "design/") && strings.HasSuffix(ref, ".md") {
		return true
	}
	if i := strings.Index(ref, "#"); i > 0 && i < len(ref)-1 {
		return true
	}
	return isItemID(ref)
}

func isItemID(ref string) bool {
	// Purpose: match item IDs such as REQ-WI-030, ADR-0008 and Q-001.
	parts := strings.Split(ref, "-")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts[:len(parts)-1] {
		if p == "" || strings.ToUpper(p) != p || strings.ToLower(p) == p {
			return false
		}
	}
	last := parts[len(parts)-1]
	if last == "" {
		return false
	}
	for _, c := range last {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func contains(list []string, s string) bool {
	// Purpose: report whether list holds s.
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}