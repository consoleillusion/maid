// Package testrun runs a project's test suite through an adapter and returns
// results in one common format.
package testrun

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Test status values.
const (
	Pass = "pass"
	Fail = "fail"
	Skip = "skip"
)

// Result is the common test result format.
type Result struct {
	Tests       map[string]string `json:"tests"`        // test ID ("package#Test/sub") -> status
	BuildFailed []string          `json:"build_failed"` // packages or units that did not compile
	ExitCode    int               `json:"exit_code"`
	Output      string            `json:"output"`   // truncated
	Failures    map[string]string `json:"failures"` // test ID -> truncated output
}

// MaxOutput is the number of bytes of output kept for reports.
const MaxOutput = 4000

// Run runs command in dir with env and parses the output with adapter.
func Run(ctx context.Context, dir, adapter string, command []string, env []string) (Result, error) {
	// Purpose: execute the test command and convert its output.
	if len(command) == 0 {
		return Result{}, fmt.Errorf("no test command configured")
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		return Result{}, err
	}
	switch adapter {
	case "go":
		r := parseGo(out.Bytes())
		r.ExitCode = code
		return r, nil
	case "exit", "":
		st := Pass
		if code != 0 {
			st = Fail
		}
		r := Result{Tests: map[string]string{"suite": st}, ExitCode: code, Output: trunc(out.String()), Failures: map[string]string{}}
		if st == Fail {
			r.Failures["suite"] = r.Output
		}
		return r, nil
	}
	return Result{}, fmt.Errorf("unknown test adapter %q", adapter)
}

type goEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

func parseGo(b []byte) Result {
	// Purpose: convert `go test -json` events into the common format.
	r := Result{Tests: map[string]string{}, Failures: map[string]string{}}
	outputs := map[string]*strings.Builder{}
	pkgHasTests := map[string]bool{}
	pkgFailed := map[string]bool{}
	var plain strings.Builder
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		var e goEvent
		if len(line) == 0 || line[0] != '{' || json.Unmarshal(line, &e) != nil {
			plain.Write(line)
			plain.WriteByte('\n')
			continue
		}
		id := e.Package
		if e.Test != "" {
			id = e.Package + "#" + e.Test
			pkgHasTests[e.Package] = true
		}
		switch e.Action {
		case "output":
			if outputs[id] == nil {
				outputs[id] = &strings.Builder{}
			}
			outputs[id].WriteString(e.Output)
		case "pass", "fail", "skip":
			if e.Test != "" {
				r.Tests[id] = e.Action
			} else if e.Action == "fail" {
				pkgFailed[e.Package] = true
			}
		case "build-fail":
			pkgFailed[e.Package] = true
		}
	}
	for pkg := range pkgFailed {
		if !pkgHasTests[pkg] {
			r.BuildFailed = append(r.BuildFailed, pkg)
		}
	}
	sort.Strings(r.BuildFailed)
	for id, st := range r.Tests {
		if st == Fail && outputs[id] != nil {
			r.Failures[id] = trunc(outputs[id].String())
		}
	}
	for _, pkg := range r.BuildFailed {
		if outputs[pkg] != nil {
			r.Failures[pkg] = trunc(outputs[pkg].String())
		}
	}
	r.Output = trunc(plain.String())
	return r
}

// Matching returns the IDs of tests whose name is in names, in any package.
// A name matches the test itself and its subtests.
func (r Result) Matching(names []string) []string {
	// Purpose: find acceptance tests in a result by name.
	var out []string
	for id := range r.Tests {
		_, short, found := strings.Cut(id, "#")
		if !found {
			short = id
		}
		for _, n := range names {
			if short == n || strings.HasPrefix(short, n+"/") {
				out = append(out, id)
			}
		}
	}
	sort.Strings(out)
	return out
}

func trunc(s string) string {
	// Purpose: keep the end of long output, where failures usually are.
	if len(s) <= MaxOutput {
		return s
	}
	return "...\n" + s[len(s)-MaxOutput:]
}
