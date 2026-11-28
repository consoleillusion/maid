package work

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/gitx"
	"github.com/consoleillusion/maid/internal/glob"
	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/testrun"
)

// Submission is the input of work_submit.
type Submission struct {
	Agent string
	ID    string
	Model string // reported by the agent
	Cost  string // reported by the agent

	// Optional changes that maid applies and commits before the merge
	// queue runs, for agents without file access.
	Patch   string            // unified diff, relative to the repository root
	Files   map[string]string // path -> complete new content
	Edits   []Edit            // exact text replacements
	Message string            // commit message for the applied changes
}

// Submit runs the merge queue for one submission and returns its report.
// Submissions are processed one at a time across all maid processes.
func (s *Service) Submit(ctx context.Context, sub Submission) (state.Report, error) {
	// Purpose: rebase, check, test and merge or reject one attempt.
	if sub.Patch != "" || len(sub.Files) > 0 || len(sub.Edits) > 0 {
		if err := s.applyChanges(ctx, sub); err != nil {
			return state.Report{}, err
		}
	}
	lock, err := s.State.Lock("queue")
	if err != nil {
		return state.Report{}, err
	}
	defer lock.Unlock()

	it, err := s.Tasks.Get(ctx, sub.ID)
	if err != nil {
		return state.Report{}, err
	}
	if it.Assignee != sub.Agent || it.Status != tasks.StatusInProgress {
		return state.Report{}, fmt.Errorf("%s does not hold the claim on %s", sub.Agent, sub.ID)
	}
	att, ok, err := s.State.Current(sub.ID, sub.Agent)
	if err != nil {
		return state.Report{}, err
	}
	if !ok {
		return state.Report{}, fmt.Errorf("no open attempt of %s on %s", sub.Agent, sub.ID)
	}
	att.Submitted, att.Model, att.Cost = s.Now().UTC(), sub.Model, sub.Cost
	_ = s.State.Log(state.Event{Agent: sub.Agent, Action: "submit", WorkItem: sub.ID})

	rep, newHead := s.evaluate(ctx, it, att)
	if rep.Accepted {
		if err := s.Git.FastForward(ctx, s.Cfg.Branch, newHead); err != nil {
			rep.Accepted, rep.FailedCheck, rep.Reason = false, "merge", "fast-forward failed: "+err.Error()
		}
	}
	rep.Finished = s.Now().UTC()
	if rep.Accepted {
		rep.Commit = newHead
		err = s.accept(ctx, it, att, &rep)
	} else {
		err = s.reject(ctx, it, att, &rep)
	}
	att.Result = &rep
	if serr := s.State.SaveAttempt(att); serr != nil && err == nil {
		err = serr
	}
	_ = s.Git.RemoveWorktree(ctx, att.Worktree)
	action := "reject"
	if rep.Accepted {
		action = "accept"
	}
	_ = s.State.Log(state.Event{Agent: sub.Agent, Action: action, WorkItem: sub.ID, Detail: rep.Reason})
	return rep, err
}

// evaluate runs every check on a rebased copy of the attempt.
func (s *Service) evaluate(ctx context.Context, it tasks.WorkItem, att state.Attempt) (state.Report, string) {
	// Purpose: apply the acceptance rule without changing the main branch.
	rep := state.Report{}
	fail := func(check, reason string) (state.Report, string) {
		rep.Checks = append(rep.Checks, state.Check{Name: check, Passed: false, Detail: reason})
		rep.FailedCheck, rep.Reason = check, reason
		return rep, ""
	}
	pass := func(check, detail string) {
		rep.Checks = append(rep.Checks, state.Check{Name: check, Passed: true, Detail: detail})
	}
	head, err := gitx.Run(ctx, att.Worktree, "rev-parse", "HEAD")
	if err != nil {
		return fail("branch", err.Error())
	}
	if head == att.Base {
		return fail("branch", "the attempt has no commits; commit in the worktree or send edits, a patch or files")
	}
	main, err := s.Git.Head(ctx, s.Cfg.Branch)
	if err != nil {
		return fail("branch", err.Error())
	}
	rep.MainBefore = main

	scratch := s.State.Path("queue", fmt.Sprintf("%s-%d", it.ID, att.Number))
	_ = s.Git.RemoveWorktree(ctx, scratch)
	if err := s.Git.AddDetached(ctx, scratch, head); err != nil {
		return fail("rebase", err.Error())
	}
	defer s.Git.RemoveWorktree(ctx, scratch)
	conflicts, err := gitx.Rebase(ctx, scratch, main)
	if err != nil || len(conflicts) > 0 {
		rep.Conflicts = conflicts
		return fail("rebase", "rebase onto "+s.Cfg.Branch+" failed: "+errString(err, conflicts))
	}
	newHead, err := squash(ctx, scratch, main, it.ID, att.Agent)
	if err != nil {
		return fail("rebase", err.Error())
	}
	pass("rebase", "rebased onto "+short(main))

	changes, err := s.Git.Diff(ctx, main, newHead)
	if err != nil {
		return fail("owned-paths", err.Error())
	}
	for _, c := range changes {
		rep.ChangedLines += c.Added + c.Deleted
		if !glob.Any(it.Maid.Owns, c.Path) {
			rep.OutsidePaths = append(rep.OutsidePaths, c.Path)
		}
	}
	if len(rep.OutsidePaths) > 0 {
		return fail("owned-paths", "changed files outside owns: "+strings.Join(rep.OutsidePaths, ", "))
	}
	pass("owned-paths", fmt.Sprintf("%d files, %d changed lines", len(changes), rep.ChangedLines))

	if it.Maid.Kind == tasks.KindCode {
		var touched []string
		for _, c := range changes {
			for _, t := range it.Maid.AcceptanceTests {
				if c.Path == t.File {
					touched = append(touched, c.Path)
				}
			}
		}
		if len(touched) > 0 {
			return fail("protected-tests", "acceptance test files changed: "+strings.Join(touched, ", "))
		}
		pass("protected-tests", "")
	}

	if msg, ok := runCheck(ctx, scratch, s.Cfg.FormatCheck, true); !ok {
		return fail("format", msg)
	}
	pass("format", "")
	if msg, ok := runCheck(ctx, scratch, s.Cfg.Lint, false); !ok {
		return fail("lint", msg)
	}
	pass("lint", "")

	base, err := s.baseline(ctx, main)
	if err != nil {
		return fail("baseline", err.Error())
	}
	res, err := testrun.Run(ctx, scratch, s.Cfg.Test.Adapter, s.Cfg.Test.Command, s.attemptEnv(it.ID, att.Number))
	if err != nil {
		return fail("tests", err.Error())
	}
	if len(res.BuildFailed) > 0 {
		rep.FailedTests = res.Failures
		msg := "does not compile: " + strings.Join(res.BuildFailed, ", ")
		if it.Maid.Kind == tasks.KindTests {
			msg += ". Add stubs with the names and signatures from the criteria in the owned non-test files, so the tests compile and fail"
		}
		return fail("build", msg)
	}
	rep.Regressions = regressions(base, res)
	if it.Maid.Kind == tasks.KindTests {
		rep.Regressions = withoutAcceptance(rep.Regressions, it)
	}
	if len(rep.Regressions) > 0 {
		rep.FailedTests = pick(res.Failures, rep.Regressions)
		return fail("regressions", fmt.Sprintf("%d previously passing tests fail", len(rep.Regressions)))
	}
	pass("regressions", "none")

	if reason := acceptanceResult(it, res); reason != "" {
		rep.FailedTests = res.Failures
		return fail("acceptance-tests", reason)
	}
	pass("acceptance-tests", "")
	rep.Accepted, rep.Reason = true, "all checks passed"
	s.pendingBaseline = &state.Baseline{Commit: newHead, Tests: res.Tests, Flaky: base.Flaky}
	return rep, newHead
}

// withoutAcceptance drops the item's own acceptance tests from regressions:
// a tests item may redefine an existing test, which then fails.
func withoutAcceptance(ids []string, it tasks.WorkItem) []string {
	// Purpose: let a tests item change what an existing test expects.
	own := map[string]bool{}
	for _, t := range it.Maid.AcceptanceTests {
		for _, n := range t.Names {
			own[n] = true
		}
	}
	var out []string
	for _, id := range ids {
		if i := strings.LastIndex(id, "#"); i >= 0 && own[id[i+1:]] {
			continue
		}
		out = append(out, id)
	}
	return out
}

// acceptanceResult checks the acceptance tests for the kind of item.
func acceptanceResult(it tasks.WorkItem, res testrun.Result) string {
	// Purpose: code items need passing acceptance tests; tests items need failing ones.
	for _, t := range it.Maid.AcceptanceTests {
		for _, name := range t.Names {
			ids := res.Matching([]string{name})
			if len(ids) == 0 {
				return "acceptance test " + name + " was not found in the test run"
			}
			for _, id := range ids {
				st := res.Tests[id]
				if it.Maid.Kind == tasks.KindTests && st == testrun.Pass && !strings.Contains(id, "/") {
					return "acceptance test " + name + " passes before the feature exists; it must fail on main"
				}
				if it.Maid.Kind == tasks.KindCode && st != testrun.Pass {
					return "acceptance test " + id + " " + st
				}
			}
		}
	}
	return ""
}

// baseline returns test results for the main commit, running them if needed.
func (s *Service) baseline(ctx context.Context, main string) (state.Baseline, error) {
	// Purpose: keep the record of which tests pass on main.
	b, err := s.State.LoadBaseline()
	if err != nil {
		return b, err
	}
	if b.Commit == main {
		return b, nil
	}
	dir := s.State.Path("queue", "baseline-"+short(main))
	_ = s.Git.RemoveWorktree(ctx, dir)
	if err := s.Git.AddDetached(ctx, dir, main); err != nil {
		return b, err
	}
	defer s.Git.RemoveWorktree(ctx, dir)
	res, err := testrun.Run(ctx, dir, s.Cfg.Test.Adapter, s.Cfg.Test.Command, s.attemptEnv("baseline", 0))
	if err != nil {
		return b, err
	}
	nb := state.Baseline{Commit: main, Tests: res.Tests, Flaky: b.Flaky}
	return nb, s.State.SaveBaseline(nb)
}

func regressions(base state.Baseline, res testrun.Result) []string {
	// Purpose: list tests that passed on main and do not pass now.
	flaky := map[string]bool{}
	for _, f := range base.Flaky {
		flaky[f] = true
	}
	var out []string
	for id, st := range base.Tests {
		if st == testrun.Pass && !flaky[id] && res.Tests[id] != testrun.Pass {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func (s *Service) accept(ctx context.Context, it tasks.WorkItem, att state.Attempt, rep *state.Report) error {
	// Purpose: record the merge, update the baseline and close the work item.
	if s.pendingBaseline != nil {
		_ = s.State.SaveBaseline(*s.pendingBaseline)
		s.pendingBaseline = nil
	}
	if it.Maid.Kind == tasks.KindTests && it.Maid.TestsFor != "" {
		if code, err := s.Tasks.Get(ctx, it.Maid.TestsFor); err == nil {
			code.Maid.TestsAuthor = att.Agent
			_ = s.Tasks.SetMeta(ctx, code.ID, code.Maid)
		}
	}
	return s.Tasks.Close(ctx, it.ID, "accepted by merge queue at "+short(rep.Commit))
}

func (s *Service) reject(ctx context.Context, it tasks.WorkItem, att state.Attempt, rep *state.Report) error {
	// Purpose: count the attempt, escalate when the budget is used, and release the claim.
	s.pendingBaseline = nil
	it.Maid.Attempts++
	budget := it.Maid.Budget.Attempts
	if budget <= 0 {
		budget = s.Cfg.Budget.Attempts
	}
	if it.Maid.Attempts >= budget {
		it.Maid.Escalated, rep.Escalated = true, true
	}
	if err := s.Tasks.SetMeta(ctx, it.ID, it.Maid); err != nil {
		return err
	}
	return s.Tasks.Release(ctx, it.ID, att.Agent)
}

// attemptEnv gives each attempt unique resources.
func (s *Service) attemptEnv(id string, n int) []string {
	// Purpose: build MAID_* variables for one test run.
	key := fmt.Sprintf("%s-%d", id, n)
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	slots := 40000 / s.Cfg.PortRange
	port := 20000 + int(h.Sum32()%uint32(slots))*s.Cfg.PortRange
	tmp := s.State.Path("tmp", key)
	_ = os.MkdirAll(tmp, 0o755)
	return []string{"MAID_ATTEMPT_ID=" + key, fmt.Sprintf("MAID_PORT_BASE=%d", port), "MAID_TMPDIR=" + tmp}
}

func runCheck(ctx context.Context, dir string, command []string, mustBeSilent bool) (string, bool) {
	// Purpose: run a format or lint command and decide whether it passed.
	if len(command) == 0 {
		return "", true
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	text := strings.TrimSpace(out.String())
	if err != nil {
		return strings.Join(command, " ") + " failed: " + clip(text), false
	}
	if mustBeSilent && text != "" {
		return strings.Join(command, " ") + " reported: " + clip(text), false
	}
	return "", true
}

func pick(m map[string]string, keys []string) map[string]string {
	// Purpose: keep only the failure output of the given tests.
	out := map[string]string{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out
}

func errString(err error, conflicts []string) string {
	// Purpose: describe a rebase failure.
	if len(conflicts) > 0 {
		return "conflicts in " + strings.Join(conflicts, ", ")
	}
	if err != nil {
		return err.Error()
	}
	return "unknown error"
}

func short(commit string) string {
	// Purpose: shorten a commit ID for messages.
	if len(commit) > 10 {
		return commit[:10]
	}
	return commit
}

func clip(s string) string {
	// Purpose: keep report text short.
	if len(s) > 1500 {
		return s[:1500] + "..."
	}
	return s
}

// squash turns the rebased commits in scratch into one commit on base with
// a Maid-Task trailer, and returns it.
func squash(ctx context.Context, scratch, base, id, agent string) (string, error) {
	// Purpose: keep one atomic, task-based commit for each accepted attempt.
	msg, err := gitx.Run(ctx, scratch, "log", "-1", "--format=%B", "HEAD")
	if err != nil {
		return "", err
	}
	if !strings.Contains(msg, "Maid-Task:") {
		msg = strings.TrimSpace(msg) + "\n\nMaid-Task: " + id
	}
	if _, err := gitx.Run(ctx, scratch, "reset", "-q", "--soft", base); err != nil {
		return "", err
	}
	if st, _ := gitx.Run(ctx, scratch, "status", "--porcelain"); st == "" {
		return "", errors.New("the attempt makes no change to the maid branch")
	}
	if _, err := gitx.Run(ctx, scratch, "-c", "user.name="+agent, "-c", "user.email="+agent+"@maid.invalid", "commit", "-q", "-m", msg); err != nil {
		return "", err
	}
	return gitx.Run(ctx, scratch, "rev-parse", "HEAD")
}
