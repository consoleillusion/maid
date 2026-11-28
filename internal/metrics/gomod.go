// Package metrics turns project files into score inputs.
// See design/method/scoring.md.
package metrics

import (
	"sort"
	"strings"
)

// GoDeps compares two go.mod files. Direct requirements count as runtime,
// requirements marked // indirect and tool directives count as dev.
func GoDeps(oldMod, newMod []byte) (runtimeAdded, devAdded int, added []string, err error) {
	// Purpose: count modules present only in the new manifest by dependency class.
	old := parseGoMod(oldMod)
	now := parseGoMod(newMod)
	for path, dev := range now {
		if _, exists := old[path]; exists {
			continue
		}
		added = append(added, path)
		if dev {
			devAdded++
		} else {
			runtimeAdded++
		}
	}
	sort.Strings(added)
	return runtimeAdded, devAdded, added, nil
}

// parseGoMod reads require lines and tool directives from a go.mod file. The
// boolean value reports whether the module is a dev dependency.
func parseGoMod(data []byte) map[string]bool {
	// Purpose: collect module paths and dev flags from one manifest.
	out := map[string]bool{}
	inBlock := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if inBlock {
			if strings.HasPrefix(line, ")") {
				inBlock = false
				continue
			}
			addRequire(out, line, true)
			continue
		}
		if line == "require (" || line == "require(" {
			inBlock = true
			continue
		}
		if strings.HasPrefix(line, "require ") || strings.HasPrefix(line, "require	") {
			addRequire(out, line, false)
			continue
		}
		if strings.HasPrefix(line, "tool ") || strings.HasPrefix(line, "tool	") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				out[fields[1]] = true
			}
		}
	}
	return out
}

// addRequire adds one require entry to a parsed manifest.
func addRequire(out map[string]bool, line string, inBlock bool) {
	// Purpose: treat indirect modules as dev and record their path.
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	path := fields[0]
	if !inBlock {
		path = fields[1]
	}
	out[path] = strings.Contains(line, "// indirect") || strings.Contains(line, "//indirect")
}
