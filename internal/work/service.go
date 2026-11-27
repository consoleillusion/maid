// Package work implements planning, claims, context packets and the merge
// queue. See design/02-method/02-work-items.md and 03-acceptance.md.
package work

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/design"
	"github.com/consoleillusion/maid/internal/gitx"
	"github.com/consoleillusion/maid/internal/glob"
	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/tasks"
)

// Service ties together the task port, git, local state and the design.
type Service struct {
	Root   string
	Cfg    config.Config
	Tasks  tasks.Port
	Git    gitx.Repo
	State  state.Store
	Design func() (*design.Store, error)
	Now    func() time.Time

	pendingBaseline *state.Baseline // set by evaluate, used by accept
}

// New builds a Service for the project at root.
func New(root string, cfg config.Config, port tasks.Port) *Service {
	// Purpose: create a Service with default dependencies.
	return &Service{
		Root: root, Cfg: cfg, Tasks: port,
		Git:    gitx.Repo{Dir: root},
		State:  state.Store{Root: root},
		Design: func() (*design.Store, error) { return design.Load(root) },
		Now:    time.Now,
	}
}

// PlanItem is one entry of a work plan.
type PlanItem struct {
	Key             string          `json:"key"`
	Title           string          `json:"title"`
	Kind            string          `json:"kind,omitempty"` // code (default) or docs
	Implements      []string        `json:"implements"`
	Criteria        string          `json:"criteria"`
	AcceptanceTests []tasks.TestRef `json:"acceptance_tests"`
	Owns            []string        `json:"owns"`
	Reads           []string        `json:"reads,omitempty"`
	DependsOn       []string        `json:"depends_on,omitempty"` // keys in this plan or existing IDs
	Budget          tasks.Budget    `json:"budget,omitempty"`
}

// Plan is the input of work_plan.
type Plan struct {
	Items []PlanItem `json:"items"`
}

// Planned maps plan keys to created work item IDs.
type Planned struct {
	Code  map[string]string `json:"code"`
	Tests map[string]string `json:"tests"`
}

// CreatePlan validates a plan and creates a tests item and a code item for
// each entry (REQ-WI-017, REQ-WI-018).
func (s *Service) CreatePlan(ctx context.Context, p Plan) (Planned, error) {
	// Purpose: turn a plan into work items in dependency order.
	order, err := s.validatePlan(p)
	if err != nil {
		return Planned{}, err
	}
	out := Planned{Code: map[string]string{}, Tests: map[string]string{}}
	for _, it := range order {
		budget := it.Budget
		if budget.Attempts == 0 {
			budget.Attempts = s.Cfg.Budget.Attempts
		}
		if budget.MaxChangedLines == 0 {
			budget.MaxChangedLines = s.Cfg.Budget.MaxChangedLines
		}
		kind := it.Kind
		if kind == "" {
			kind = tasks.KindCode
		}
		var deps []string
		for _, d := range it.DependsOn {
			if id, ok := out.Code[d]; ok {
				deps = append(deps, id)
			} else {
				deps = append(deps, d)
			}
		}
		var testsID string
		if kind == tasks.KindCode {
			var testFiles []string
			for _, t := range it.AcceptanceTests {
				testFiles = append(testFiles, t.File)
			}
			testsID, err = s.Tasks.Create(ctx, tasks.NewItem{
				Title: "Tests: " + it.Title, DependsOn: deps,
				Maid: tasks.Meta{Key: it.Key + ".tests", Kind: tasks.KindTests, Implements: it.Implements,
					Criteria: it.Criteria, AcceptanceTests: it.AcceptanceTests, Owns: testFiles,
					Reads: append(append([]string{}, it.Owns...), it.Reads...), Budget: budget},
			})
			if err != nil {
				return out, err
			}
			out.Tests[it.Key] = testsID
			deps = append(deps, testsID)
		}
		id, err := s.Tasks.Create(ctx, tasks.NewItem{
			Title: it.Title, DependsOn: deps,
			Maid: tasks.Meta{Key: it.Key, Kind: kind, Implements: it.Implements, Criteria: it.Criteria,
				AcceptanceTests: it.AcceptanceTests, Owns: it.Owns, Reads: it.Reads, TestsItem: testsID, Budget: budget},
		})
		if err != nil {
			return out, err
		}
		out.Code[it.Key] = id
		if testsID != "" {
			m, err := s.Tasks.Get(ctx, testsID)
			if err == nil {
				m.Maid.TestsFor = id
				_ = s.Tasks.SetMeta(ctx, testsID, m.Maid)
			}
		}
	}
	_ = s.State.Log(state.Event{Action: "plan", Detail: fmt.Sprintf("%d items", len(order))})
	return out, nil
}

func (s *Service) validatePlan(p Plan) ([]PlanItem, error) {
	// Purpose: check required fields and requirement links, then sort by dependencies.
	if len(p.Items) == 0 {
		return nil, errors.New("plan has no items")
	}
	ds, err := s.Design()
	if err != nil {
		return nil, err
	}
	byKey := map[string]PlanItem{}
	var problems []string
	for _, it := range p.Items {
		switch {
		case it.Key == "" || it.Title == "":
			problems = append(problems, "item without key or title")
		case byKey[it.Key].Key != "":
			problems = append(problems, "duplicate key "+it.Key)
		case len(it.Implements) == 0:
			problems = append(problems, it.Key+": implements no requirement")
		case len(it.Owns) == 0:
			problems = append(problems, it.Key+": owns no paths")
		case (it.Kind == "" || it.Kind == tasks.KindCode) && (len(it.AcceptanceTests) == 0 || strings.TrimSpace(it.Criteria) == ""):
			problems = append(problems, it.Key+": code items need criteria and acceptance_tests")
		}
		for _, r := range it.Implements {
			if _, ok := ds.Item(r); !ok {
				problems = append(problems, it.Key+": unknown requirement "+r)
			}
		}
		for _, t := range it.AcceptanceTests {
			if len(t.Names) == 0 {
				problems = append(problems, it.Key+": acceptance test file "+t.File+" lists no test names")
			}
			if glob.Any(it.Owns, t.File) {
				problems = append(problems, it.Key+": acceptance test "+t.File+" is inside owns; build agents must not own their tests")
			}
		}
		byKey[it.Key] = it
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid plan:\n- %s", strings.Join(problems, "\n- "))
	}
	return topo(p.Items, byKey)
}

func topo(items []PlanItem, byKey map[string]PlanItem) ([]PlanItem, error) {
	// Purpose: order plan items so that dependencies come first.
	var out []PlanItem
	state := map[string]int{} // 1 visiting, 2 done
	var visit func(k string) error
	visit = func(k string) error {
		switch state[k] {
		case 1:
			return fmt.Errorf("dependency cycle at %s", k)
		case 2:
			return nil
		}
		state[k] = 1
		for _, d := range byKey[k].DependsOn {
			if _, inPlan := byKey[d]; inPlan {
				if err := visit(d); err != nil {
					return err
				}
			}
		}
		state[k] = 2
		out = append(out, byKey[k])
		return nil
	}
	for _, it := range items {
		if err := visit(it.Key); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Claim gives agent a work item and an isolated worktree. With id empty it
// claims the first ready item that the agent may take.
func (s *Service) Claim(ctx context.Context, agent, id string) (Packet, error) {
	// Purpose: apply the ready rules, claim atomically and build the context packet.
	if agent == "" {
		return Packet{}, errors.New("agent id is required")
	}
	candidates, err := s.claimable(ctx, agent)
	if err != nil {
		return Packet{}, err
	}
	if id != "" {
		var found []tasks.WorkItem
		for _, c := range candidates {
			if c.ID == id {
				found = append(found, c)
			}
		}
		if len(found) == 0 {
			return Packet{}, fmt.Errorf("work item %s is not claimable by %s: %s", id, agent, s.whyNot(ctx, agent, id))
		}
		candidates = found
	}
	for _, c := range candidates {
		it, err := s.Tasks.Claim(ctx, c.ID, agent)
		if errors.Is(err, tasks.ErrConflict) {
			continue
		}
		if err != nil {
			return Packet{}, err
		}
		att, err := s.startAttempt(ctx, it, agent)
		if err != nil {
			_ = s.Tasks.Release(ctx, it.ID, agent)
			return Packet{}, err
		}
		_ = s.State.Log(state.Event{Agent: agent, Action: "claim", WorkItem: it.ID, Detail: att.Branch})
		return s.packet(ctx, it, att)
	}
	return Packet{}, tasks.ErrNoReady
}

// claimable lists ready items that agent may claim now.
func (s *Service) claimable(ctx context.Context, agent string) ([]tasks.WorkItem, error) {
	// Purpose: filter ready items by escalation, independence, tests on main and path overlap.
	ready, err := s.Tasks.Ready(ctx)
	if err != nil {
		return nil, err
	}
	active, err := s.Tasks.List(ctx, false)
	if err != nil {
		return nil, err
	}
	main, err := s.Git.Head(ctx, s.Cfg.MainBranch)
	if err != nil {
		return nil, err
	}
	var out []tasks.WorkItem
	for _, it := range ready {
		if reason := s.blocked(ctx, it, agent, active, main); reason == "" {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// blocked returns why agent cannot claim it, or "".
func (s *Service) blocked(ctx context.Context, it tasks.WorkItem, agent string, active []tasks.WorkItem, main string) string {
	// Purpose: apply REQ-WI-002, REQ-WI-003, REQ-WI-009 and REQ-WI-019 to one item.
	if it.Maid.Kind == "" {
		return "not a maid work item"
	}
	if it.Maid.Escalated {
		return "escalated to a design agent"
	}
	if it.Maid.Kind == tasks.KindCode {
		if it.Maid.TestsAuthor == agent {
			return "agent wrote the acceptance tests of this item"
		}
		for _, t := range it.Maid.AcceptanceTests {
			if !s.Git.FileExists(ctx, main, t.File) {
				return "acceptance test " + t.File + " is not on " + s.Cfg.MainBranch
			}
		}
	}
	for _, a := range active {
		if a.ID != it.ID && a.Status == tasks.StatusInProgress && glob.Overlap(a.Maid.Owns, it.Maid.Owns) {
			return "owned paths overlap with claimed item " + a.ID
		}
	}
	return ""
}

func (s *Service) whyNot(ctx context.Context, agent, id string) string {
	// Purpose: explain a refused claim.
	it, err := s.Tasks.Get(ctx, id)
	if err != nil {
		return err.Error()
	}
	if it.Status != tasks.StatusOpen {
		return "status is " + it.Status
	}
	ready, _ := s.Tasks.Ready(ctx)
	isReady := false
	for _, r := range ready {
		isReady = isReady || r.ID == id
	}
	if !isReady {
		return "it has open dependencies"
	}
	active, _ := s.Tasks.List(ctx, false)
	main, _ := s.Git.Head(ctx, s.Cfg.MainBranch)
	return s.blocked(ctx, it, agent, active, main)
}

func (s *Service) startAttempt(ctx context.Context, it tasks.WorkItem, agent string) (state.Attempt, error) {
	// Purpose: create the attempt record, branch and worktree (REQ-WI-007).
	if att, ok, err := s.State.Current(it.ID, agent); err != nil {
		return att, err
	} else if ok {
		return att, nil // the same agent claimed again: continue its attempt
	}
	prev, err := s.State.Attempts(it.ID)
	if err != nil {
		return state.Attempt{}, err
	}
	n := len(prev) + 1
	base, err := s.Git.Head(ctx, s.Cfg.MainBranch)
	if err != nil {
		return state.Attempt{}, err
	}
	if err := s.State.Ensure(); err != nil {
		return state.Attempt{}, err
	}
	att := state.Attempt{
		WorkItem: it.ID, Number: n, Agent: agent, Base: base, Started: s.Now().UTC(),
		Branch:   fmt.Sprintf("maid/%s/%d", it.ID, n),
		Worktree: s.State.Path("worktrees", fmt.Sprintf("%s-%d", it.ID, n)),
	}
	if err := s.Git.AddWorktree(ctx, att.Worktree, att.Branch, base); err != nil {
		return state.Attempt{}, err
	}
	return att, s.State.SaveAttempt(att)
}

// Heartbeat renews the claim of agent on id.
func (s *Service) Heartbeat(ctx context.Context, agent, id string) error {
	// Purpose: keep a claim alive (REQ-WI-005).
	return s.Tasks.Heartbeat(ctx, id, agent)
}

// Release gives up a claim and removes the attempt worktree.
func (s *Service) Release(ctx context.Context, agent, id, reason string) error {
	// Purpose: return an item to ready and close the open attempt as abandoned.
	if att, ok, _ := s.State.Current(id, agent); ok {
		att.Result = &state.Report{Reason: "released: " + reason, Finished: s.Now().UTC()}
		_ = s.State.SaveAttempt(att)
		_ = s.Git.RemoveWorktree(ctx, att.Worktree)
	}
	_ = s.State.Log(state.Event{Agent: agent, Action: "release", WorkItem: id, Detail: reason})
	return s.Tasks.Release(ctx, id, agent)
}
