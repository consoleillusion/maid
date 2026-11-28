// Package guide derives the current phase and builds the maid_next response.
// See design/method/phases.md.
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

// Next is the maid_next response (design/implementation/interfaces.md).
type Next struct {
	Phase    string        `json:"phase"`
	Step     string        `json:"step"`
	Why      string        `json:"why"`
	Read     []string      `json:"read,omitempty"`
	Tools    []string      `json:"tools"`
	DoneWhen []string      `json:"done_when,omitempty"`
	AskUser  []string      `json:"ask_user,omitempty"`
	Blocked  []string      `json:"blocked,omitempty"`
	Warnings []string      `json:"warnings,omitempty"`
	Calls    []Call        `json:"calls,omitempty"` // ready-to-run tool calls for this step
	Tasks    []TaskSummary `json:"tasks"`           // work items the agent may claim
}

// TaskSummary is one ready work item in maid_next.
type TaskSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

// Call is a tool call that the agent can make as it is.
type Call struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
	When      string         `json:"when,omitempty"`
}

// Roles are ignored: every agent gets the step that follows from the
// project state. The constants remain for callers.
const (
	RoleDesign = "design"
	RoleBuild  = "build"
)

// Compute derives the next step for agent from the design, the task
// store and local state.
func Compute(ctx context.Context, s *work.Service, _, agent string) (Next, error) {
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
	}
	n := designNext(ds, items)
	n.Warnings = warnings
	if taskErr != nil && n.Phase >= "5" {
		n.Step = "Ask the user to fix the task store, then call maid_next again. " + n.Step
	}
	if n.Phase >= "6" {
		if b := buildNext(ctx, s, agent, items); b.Step != "" {
			return b, nil
		}
	}
	return n, nil
}

func designNext(ds *design.Store, items []tasks.WorkItem) Next {
	// Purpose: find the first unmet design gate and give ready-to-run calls.
	if len(ds.Chapters) == 0 {
		return Next{
			Phase: "0-bootstrap", Step: "Ask the user what the project is for and what it must not do. Then create the vision chapter from the answers.",
			Why:   "The project has no design. The design is the source of truth.",
			Tools: []string{"design_create", "design_check"}, DoneWhen: []string{"design/ contains a vision chapter with at least one goal and one non-goal"},
			AskUser: []string{"What is this project for, and what is out of scope?"},
			Calls: []Call{{Tool: "design_create", When: "after the user answers; replace the goal and non-goal text with theirs",
				Arguments: map[string]any{"path": "foundation/vision.md", "style": "ste",
					"summary": "Why the project exists and what it must and must not do.",
					"body":    "# Vision\n\n## Goals\n\n### <goal name>\n<one or two sentences>\n\n## Non-goals\n\n### <non-goal name>\n<one or two sentences>\n"}}},
		}
	}
	if errs := errorFindings(ds); len(errs) > 0 {
		f := errs[0]
		return Next{
			Phase: phaseOf(ds, items), Step: fmt.Sprintf("Fix %d design check errors. The first is %s line %d: %s", len(errs), f.Path, f.Line, f.Message),
			Why: "Each gate requires a design that passes its checks.", Read: refsOf(errs),
			Tools: []string{"design_check", "design_read", "design_write"}, DoneWhen: []string{"design_check reports no errors"},
			Calls: []Call{{Tool: "design_check", Arguments: map[string]any{}, When: "to see every finding"},
				{Tool: "design_read", Arguments: map[string]any{"refs": []string{f.Path}}, When: "to get the text and version before design_write"}},
		}
	}
	visionOK := ds.Chapter("vision") != nil && ds.Chapter("vision").FM.String("status") == "accepted"
	if pending := Pending(ds); (!visionOK || len(pending) > 0) && len(items) == 0 {
		phase, why := "2-4-design", "Work items can only implement accepted requirements."
		if !visionOK {
			phase, why = "1-intent", "The intent gate requires an accepted vision."
		}
		if len(pending) == 0 {
			pending = []string{"vision"}
		}
		return Next{
			Phase: phase,
			Step: fmt.Sprintf("%d chapters are not accepted yet: %s. Read them, fix any problem with design_write, "+
				"then accept them all with ONE design_set_status call.", len(pending), strings.Join(pending, ", ")),
			Why:      why,
			Tools:    []string{"design_read", "design_write", "design_set_status", "design_check"},
			DoneWhen: []string{"every chapter that the plan needs has status accepted"},
			Calls: []Call{{Tool: "design_set_status", Arguments: map[string]any{"ids": pending, "status": "accepted"},
				When: "after any fixes"}},
		}
	}
	if len(items) == 0 {
		return Next{
			Phase: "5-plan", Step: "Write a work plan: small items with Gherkin criteria, named acceptance tests and owned paths. The test files must not be inside owns.",
			Why: "The design is accepted and there are no work items.", Read: []string{"work-items"},
			Tools: []string{"work_plan"}, DoneWhen: []string{"each accepted requirement is linked to a work item"},
			Calls: []Call{{Tool: "work_plan", When: "with one item for each small piece of work",
				Arguments: map[string]any{"plan": map[string]any{"items": []map[string]any{{
					"key": "short-key", "title": "What the item delivers", "implements": []string{"acceptance#atomic-claim"},
					"criteria":         "Scenario: <behaviour>\n  Given <state>\n  When <action>\n  Then <result>",
					"acceptance_tests": []map[string]any{{"file": "pkg/thing_test.go", "names": []string{"TestThing"}}},
					"owns":             []string{"pkg/thing.go"}}}}}}},
		}
	}
	if esc := escalated(items); len(esc) > 0 {
		return Next{
			Phase: "6-build", Step: "Resolve escalated work items: split them, clarify the requirements, or correct the acceptance tests.",
			Why: "These items used their attempt budget.", Blocked: esc,
			Tools: []string{"maid_status", "work_result", "work_plan", "design_read"},
			Calls: []Call{{Tool: "work_result", Arguments: map[string]any{"id": esc[0]}, When: "to see why the attempts were rejected"}},
		}
	}
	if allClosed(items) {
		return Next{
			Phase: "7-verify", Step: "Check that each requirement traces to code and passing tests, then plan the next milestone.",
			Why: "All work items are closed.", Tools: []string{"maid_status", "design_check", "work_plan"},
			Calls: []Call{{Tool: "maid_status", Arguments: map[string]any{"detail": true}}},
		}
	}
	return Next{Phase: "6-build", Step: "Build agents are working. Wait, or plan more work.", Why: "Work items are open or in progress.",
		Tools: []string{"maid_status", "work_plan"}, Calls: []Call{{Tool: "maid_status", Arguments: map[string]any{"detail": true}}}}
}

func buildNext(ctx context.Context, s *work.Service, agent string, items []tasks.WorkItem) Next {
	// Purpose: tell a build or test agent to continue, claim or wait.
	for _, it := range items {
		if it.Status == tasks.StatusInProgress && it.Assignee == agent && agent != "" {
			return Next{
				Phase: "6-build", Step: fmt.Sprintf("Continue work item %s (%s). Commit in your worktree, then call work_submit.", it.ID, it.Title),
				Why: "You hold the claim on this item.", Tools: []string{"work_claim", "work_heartbeat", "work_submit", "work_release"},
				DoneWhen: []string{"the merge queue accepts your submission"},
				Calls: []Call{{Tool: "work_claim", Arguments: map[string]any{"id": it.ID}, When: "to get your context packet and worktree again"},
					{Tool: "work_files", Arguments: map[string]any{"id": it.ID}, When: "if you have no file access: then work_read, work_edit, work_run and work_commit"},
					{Tool: "work_submit", Arguments: map[string]any{"id": it.ID}, When: "after you commit"}},
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
					Calls: []Call{{Tool: "work_claim", Arguments: map[string]any{}}},
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
	// Purpose: keep only findings that fail a gate.
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

// Pending lists the chapters that wait for approval, vision first.
func Pending(ds *design.Store) []string {
	// Purpose: give the CLI and tools the same list that maid_next asks about.
	var rest []string
	for _, id := range notAccepted(ds) {
		if id != "vision" {
			rest = append(rest, id)
		}
	}
	if v := ds.Chapter("vision"); v != nil && v.FM.String("status") != "accepted" {
		return append([]string{"vision"}, rest...)
	}
	return rest
}

func notAccepted(ds *design.Store) []string {
	// Purpose: list part chapters that are not accepted yet, in reading order.
	open := func(ch *design.Chapter) bool {
		st := ch.FM.String("status")
		return design.PartOf(ch.Path) != "" && st != "accepted" && st != "superseded"
	}
	var out []string
	if ds.HasIndex() {
		byPath := map[string]*design.Chapter{}
		for _, ch := range ds.Chapters {
			byPath[ch.Path] = ch
		}
		for _, e := range ds.Order() {
			if ch := byPath[e.Path]; ch != nil && open(ch) {
				out = append(out, ch.ID)
			}
		}
		return out
	}
	for _, ch := range ds.Chapters {
		if open(ch) {
			out = append(out, ch.ID)
		}
	}
	sort.Strings(out)
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
