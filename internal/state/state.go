// Package state keeps maid's local state in .maid/: attempts, the test
// baseline, the merge queue lock and the activity log.
//
// v1 uses JSON files with file locks instead of SQLite, because maid is
// built on the Go standard library only. The package is the
// state store port; a SQLite adapter can replace it.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

// Dir is the state folder name at the project root.
const Dir = ".maid"

// Store is the state folder of one project.
type Store struct {
	Root string // project root
}

// Path returns a path inside .maid.
func (s Store) Path(parts ...string) string {
	// Purpose: build a path under the state folder.
	return filepath.Join(append([]string{s.Root, Dir}, parts...)...)
}

// Ensure creates the state folders.
func (s Store) Ensure() error {
	// Purpose: create .maid and its subfolders.
	for _, d := range []string{"", "attempts", "worktrees", "queue"} {
		if err := os.MkdirAll(s.Path(d), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Lock is an exclusive file lock.
type Lock struct{ f *os.File }

// Lock blocks until it holds the named lock. It works across processes
// .
func (s Store) Lock(name string) (*Lock, error) {
	// Purpose: serialise work across maid processes with flock.
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(s.Path(name+".lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return &Lock{f}, nil
}

// Unlock releases the lock.
func (l *Lock) Unlock() {
	// Purpose: release a lock taken with Store.Lock.
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	_ = l.f.Close()
}

// Baseline is the test result on the main branch.
type Baseline struct {
	Commit string            `json:"commit"`
	Tests  map[string]string `json:"tests"`
	Flaky  []string          `json:"flaky,omitempty"`
}

// LoadBaseline reads the baseline. A missing file gives an empty baseline.
func (s Store) LoadBaseline() (Baseline, error) {
	// Purpose: read the last recorded main-branch test results.
	var b Baseline
	err := readJSON(s.Path("baseline.json"), &b)
	if errors.Is(err, os.ErrNotExist) {
		return Baseline{Tests: map[string]string{}}, nil
	}
	if b.Tests == nil {
		b.Tests = map[string]string{}
	}
	return b, err
}

// SaveBaseline writes the baseline.
func (s Store) SaveBaseline(b Baseline) error {
	// Purpose: record main-branch test results after a merge.
	return writeJSON(s.Path("baseline.json"), b)
}

// Attempt is one try by one agent on one work item.
type Attempt struct {
	WorkItem  string    `json:"work_item"`
	Number    int       `json:"number"`
	Agent     string    `json:"agent"`
	Branch    string    `json:"branch"`
	Worktree  string    `json:"worktree"`
	Base      string    `json:"base"`
	Started   time.Time `json:"started"`
	Submitted time.Time `json:"submitted,omitempty"`
	Model     string    `json:"model,omitempty"`
	Cost      string    `json:"cost,omitempty"`
	Result    *Report   `json:"result,omitempty"`
}

// Report is the merge queue result of an attempt.
type Report struct {
	Accepted     bool              `json:"accepted"`
	Reason       string            `json:"reason"`
	FailedCheck  string            `json:"failed_check,omitempty"`
	Checks       []Check           `json:"checks"`
	Regressions  []string          `json:"regressions,omitempty"`
	FailedTests  map[string]string `json:"failed_tests,omitempty"`
	Conflicts    []string          `json:"conflicts,omitempty"`
	OutsidePaths []string          `json:"outside_paths,omitempty"`
	ChangedLines int               `json:"changed_lines"`
	Commit       string            `json:"commit,omitempty"`
	MainBefore   string            `json:"main_before,omitempty"`
	Escalated    bool              `json:"escalated,omitempty"`
	Environment  bool              `json:"environment,omitempty"` // a check could not run
	Finished     time.Time         `json:"finished"`
}

// Check is one merge queue check.
type Check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// SaveAttempt writes the current attempt of a work item.
func (s Store) SaveAttempt(a Attempt) error {
	// Purpose: persist an attempt so any maid process can continue it.
	return writeJSON(s.Path("attempts", fmt.Sprintf("%s-%d.json", a.WorkItem, a.Number)), a)
}

// Attempts returns all attempts of a work item, oldest first.
func (s Store) Attempts(workItem string) ([]Attempt, error) {
	// Purpose: read the attempt history of a work item.
	files, err := filepath.Glob(s.Path("attempts", workItem+"-*.json"))
	if err != nil {
		return nil, err
	}
	var out []Attempt
	for _, f := range files {
		var a Attempt
		if err := readJSON(f, &a); err != nil {
			return nil, err
		}
		if a.WorkItem == workItem {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

// Current returns the latest unsubmitted attempt of a work item by agent.
func (s Store) Current(workItem, agent string) (Attempt, bool, error) {
	// Purpose: find the open attempt that a submit refers to.
	all, err := s.Attempts(workItem)
	if err != nil {
		return Attempt{}, false, err
	}
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Agent == agent && all[i].Result == nil {
			return all[i], true, nil
		}
	}
	return Attempt{}, false, nil
}

// Event is one line of the activity log.
type Event struct {
	Time     time.Time `json:"time"`
	Agent    string    `json:"agent,omitempty"`
	Action   string    `json:"action"`
	WorkItem string    `json:"work_item,omitempty"`
	Detail   string    `json:"detail,omitempty"`
}

// Log appends an event to the activity log.
func (s Store) Log(e Event) error {
	// Purpose: record what happened, for humans and for status reports.
	if err := s.Ensure(); err != nil {
		return err
	}
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	f, err := os.OpenFile(s.Path("activity.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(e)
	_, err = f.Write(append(b, '\n'))
	return err
}

// Recent returns the last n events.
func (s Store) Recent(n int) ([]Event, error) {
	// Purpose: read the tail of the activity log.
	b, err := os.ReadFile(s.Path("activity.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var all []Event
	for _, line := range splitLines(b) {
		var e Event
		if json.Unmarshal(line, &e) == nil {
			all = append(all, e)
		}
	}
	if len(all) > n {
		all = all[len(all)-n:]
	}
	return all, nil
}

func splitLines(b []byte) [][]byte {
	// Purpose: split file content into non-empty lines.
	var out [][]byte
	start := 0
	for i, c := range b {
		if c == '\n' {
			if i > start {
				out = append(out, b[start:i])
			}
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, b[start:])
	}
	return out
}

func readJSON(path string, v any) error {
	// Purpose: decode a JSON file.
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	// Purpose: write a JSON file atomically through a temporary file.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
