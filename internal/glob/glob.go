// Package glob matches slash-separated paths against patterns with `**`.
package glob

import (
	"path"
	"strings"
)

// Match reports whether p matches pattern. `*` matches within one path
// segment; `**` matches any number of segments, including none.
func Match(pattern, p string) bool {
	// Purpose: match owned-path patterns such as "internal/store/**".
	return matchSegs(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

// Any reports whether p matches any of the patterns.
func Any(patterns []string, p string) bool {
	// Purpose: check a path against a list of patterns.
	for _, pat := range patterns {
		if Match(pat, p) {
			return true
		}
	}
	return false
}

func matchSegs(pat, segs []string) bool {
	// Purpose: match pattern segments against path segments recursively.
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			for i := 0; i <= len(segs); i++ {
				if matchSegs(rest, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, _ := path.Match(pat[0], segs[0]); !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// Overlap reports whether two pattern lists can match a common path. It is
// conservative: it compares the literal prefixes of the patterns.
func Overlap(a, b []string) bool {
	// Purpose: detect work items that could change the same files (REQ-WI-003).
	for _, x := range a {
		for _, y := range b {
			px, py := literalPrefix(x), literalPrefix(y)
			if strings.HasPrefix(px, py) || strings.HasPrefix(py, px) {
				return true
			}
		}
	}
	return false
}

func literalPrefix(pattern string) string {
	// Purpose: return the part of a pattern before its first wildcard.
	if i := strings.IndexAny(pattern, "*?["); i >= 0 {
		return pattern[:i]
	}
	return pattern
}
