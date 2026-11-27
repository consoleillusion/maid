// Package guide derives the current phase and builds the maid_next response.
// See design/02-method/01-phases.md.
package guide

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/design"
	"github.com/consoleillusion/maid/internal/style"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

// Next is the maid_next response (design/05-interfaces/01-mcp.md).
type Next struct {
	Phase    string   `json:"phase"`
	Step     string   `json:"step"`
	Why      string   `json:"why"`
	Read     []string `json:"read,omitempty"`
	Tools    []string `json:"tools"`
	DoneWhen []string `json:"done_when,omitempty"`
	AskUser  []string `json:"ask_user,omitempty"`
	Blocked  []string `json:"blocked,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Roles.
const (
	RoleDesign = "design"
	RoleBuild  = "build"
)

// Compute derives the next step for agent in role from the design, the task
// store and local state (REQ-PH-001, REQ-PH-003).
func Compute(ctx context.Context, s *work.Service, role, agent string) (Next, error) {
	// Purpose: choose the single next action for an agent.
	ds, err := s.Design()
	if err != nil {
		return Next{}, err
	}
	items, taskErr := s.Tasks.List(ctx, true)
	var warnings []string
	if taskErr != nil {
		// The design phases do not need the task store, so report the
		// problem instead of failing.
		warnings = append(warnings, "task store unavailable: "+taskErr.Error())
		if role == RoleBuild {
			return Next{Phase: "6-build", Why: "The task store cannot be read.", Tools: []string{"maid_status"}, Warnings: warnings}, nil
		}
	}
	questions := openQuestions(ds)
	if role == RoleBuild {
		return buildNext(ctx, s, agent, items), nil
	}
	n := designNext(ds, items)
	n.AskUser = questions
	n.Warnings = warnings
	if taskErr != nil && n.Phase >= "5" {
		n.Step = "Ask the user to fix the task store, then call maid_next again. " + n.Step
	}
	if n.Phase >= "6" {
		if b := buildNext(ctx, s, agent, items); b.Step != "" && role == "" {
			b.AskUser = questions
			return b, nil
		}
	}
	return n, nil
}

func designNext(ds *design.Store, items []tasks.WorkItem) Next {
	// Purpose: find the first unmet design gate.
	if len(ds.Chapters) == 0 {
		return Next{
			Phase: "0-bootstrap", Step: "Ask the user what the project is for, then create a vision chapter (01-foundation/01-vision.md, id vision, style ste) and a glossary chapter with design_create.",
			Why:   "The project has no design. The design is the source of truth.",
			Tools: []string{"design_create", "design_check"}, DoneWhen: []string{"design/ contains a vision chapter with at least one goal and one non-goal"},
		}
	}
	if errs := errorFindings(ds); len(errs) > 0 {
		return Next{
			Phase: phaseOf(ds, items), Step: fmt.Sprintf("Fix %d design check errors.", len(errs)),
			Why: "Each gate requires a design that passes its checks (REQ-PH-006).", Read: refsOf(errs),
			Tools: []string{"design_check", "design_read", "design_write"}, DoneWhen: []string{"design_check reports no errors"},
		}
	}
	if v := ds.Chapter("vision"); v == nil || v.FM.String("status") != "accepted" {
		return Next{
			Phase: "1-intent", Step: "Complete the vision chapter with the user. When the user explicitly accepts it, call design_set_status with status accepted and approved_by set to their name.",
			Why: "The intent gate requires an accepted vision (REQ-PH-011).", Read: []string{"vision", "glossary"},
			Tools: []string{"design_read", "design_write", "design_question", "design_set_status"}, DoneWhen: []string{"vision has status accepted"},
		}
	}
	if pending := notAccepted(ds); len(pending) > 0 && len(items) == 0 {
		return Next{
			Phase: "2-4-design", Step: "Finish these chapters and ask the user to accept them: " + strings.Join(pending, ", ") + ". Record each acceptance with design_set_status.",
			Why: "Work items can only implement accepted requirements.", Read: pending,
			Tools: []string{"design_read", "design_write", "design_append", "design_create", "design_question", "design_set_status"}, DoneWhen: []string{"each chapter that the plan needs has status accepted"},
		}
	}
	if len(items) == 0 {
		return Next{
			Phase: "5-plan", Step: "Write a work plan: small items with Gherkin criteria, named acceptance tests and owned paths. Submit it with work_plan.",
			Why: "The design is accepted and there are no work items.", Read: []string{"work-items"},
			Tools: []string{"work_plan"}, DoneWhen: []string{"each accepted requirement is linked to a work item"},
		}
	}
	if esc := escalated(items); len(esc) > 0 {
		return Next{
			Phase: "6-build", Step: "Resolve escalated work items: split them, clarify the requirements, or correct the acceptance tests.",
			Why: "These items used their attempt budget (REQ-WI-009, REQ-WI-010).", Blocked: esc,
			Tools: []string{"maid_status", "work_plan", "design_read"},
		}
	}
	if allClosed(items) {
		return Next{
			Phase: "7-verify", Step: "Check that each requirement traces to code and passing tests, then plan the next milestone.",
			Why: "All work items are closed.", Tools: []string{"maid_status", "design_check", "work_plan"},
		}
	}
	return Next{Phase: "6-build", Step: "Build agents are working. Wait, or plan more work.", Why: "Work items are open or in progress.",
		Tools: []string{"maid_status", "work_plan"}}
}

func buildNext(ctx context.Context, s *work.Service, agent string, items []tasks.WorkItem) Next {
	// Purpose: tell a build or test agent to continue, claim or wait.
	for _, it := range items {
		if it.Status == tasks.StatusInProgress && it.Assignee == agent && agent != "" {
			return Next{
				Phase: "6-build", Step: fmt.Sprintf("Continue work item %s (%s). Commit in your worktree, then call work_submit.", it.ID, it.Title),
				Why: "You hold the claim on this item.", Tools: []string{"work_claim", "work_heartbeat", "work_submit", "work_release"},
				DoneWhen: []string{"the merge queue accepts your submission"},
			}
		}
	}
	ready, err := s.Tasks.Ready(ctx)
	if err == nil {
		for _, it := range ready {
			if it.Maid.Kind != "" && !it.Maid.Escalated && it.Maid.TestsAuthor != agent {
				return Next{
					Phase: "6-build", Step: "Call work_claim to take the next ready work item. Its context packet has everything you need.",
					Why: "There is ready work.", Tools: []string{"work_claim"},
				}
			}
		}
	}
	return Next{Phase: "6-build", Step: "", Why: "No work item is ready for you. Try again later.", Tools: []string{"maid_status"}}
}

func phaseOf(ds *design.Store, items []tasks.WorkItem) string {
	// Purpose: name the phase for a step that does not move the gate.
	switch {
	case len(items) > 0:
		return "6-build"
	case ds.Chapter("vision") == nil || ds.Chapter("vision").FM.String("status") != "accepted":
		return "1-intent"
	}
	return "2-4-design"
}

func errorFindings(ds *design.Store) []design.Finding {
	// Purpose: keep only findings that fail a gate (REQ-STY-004).
	var out []design.Finding
	for _, f := range ds.Check() {
		if f.Severity == style.Error {
			out = append(out, f)
		}
	}
	return out
}

func refsOf(fs []design.Finding) []string {
	// Purpose: list the files that have findings.
	seen := map[string]bool{}
	var out []string
	for _, f := range fs {
		if !seen[f.Path] {
			seen[f.Path] = true
			out = append(out, f.Path)
		}
	}
	return out
}

func notAccepted(ds *design.Store) []string {
	// Purpose: list chapters in numbered parts that are not accepted yet.
	var out []string
	for _, ch := range ds.Chapters {
		part := ch.FM.String("part")
		if part == "" || part[0] < '0' || part[0] > '9' {
			continue // styles, decisions, index, questions
		}
		if st := ch.FM.String("status"); st != "accepted" && st != "superseded" {
			out = append(out, ch.ID)
		}
	}
	sort.Strings(out)
	return out
}

func openQuestions(ds *design.Store) []string {
	// Purpose: read open questions from the questions table.
	ch := ds.Chapter("questions")
	if ch == nil {
		return nil
	}
	var out []string
	for _, l := range strings.Split(ch.Body, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(l), "|"), "|")
		if len(cells) < 3 || !strings.HasPrefix(strings.TrimSpace(cells[0]), "Q-") {
			continue
		}
		if strings.TrimSpace(cells[len(cells)-1]) == "open" {
			out = append(out, strings.TrimSpace(cells[0])+": "+strings.TrimSpace(cells[1]))
		}
	}
	return out
}

func escalated(items []tasks.WorkItem) []string {
	// Purpose: list escalated items that are still open.
	var out []string
	for _, it := range items {
		if it.Maid.Escalated && it.Status != tasks.StatusClosed {
			out = append(out, it.ID)
		}
	}
	return out
}

func allClosed(items []tasks.WorkItem) bool {
	// Purpose: report whether every maid work item is closed.
	for _, it := range items {
		if it.Maid.Kind != "" && it.Status != tasks.StatusClosed {
			return false
		}
	}
	return true
}
