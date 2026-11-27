// Package frontmatter splits a Markdown file into YAML frontmatter and body,
// and parses the YAML subset that design files use.
//
// Supported YAML: `key: value`, quoted strings, `null`, inline lists
// `[a, b]`, block lists (`- a`), folded (`>`) and literal (`|`) block
// scalars, and `#` comments. Nested maps are not supported; design files
// do not use them.
package frontmatter

import (
	"errors"
	"fmt"
	"strings"
)

// Map is the parsed frontmatter. Values are string, []string or nil.
type Map map[string]any

// ErrNoFrontmatter is returned when a file does not start with `---`.
var ErrNoFrontmatter = errors.New("no frontmatter")

// Split returns the frontmatter text, the body and the 1-based line number
// where the body starts.
func Split(src string) (fm string, body string, bodyLine int, err error) {
	// Purpose: separate the leading `---` block from the Markdown body.
	src = strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(src, "---\n") {
		return "", src, 1, ErrNoFrontmatter
	}
	rest := src[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "\n---") {
			end = len(rest) - 4
		} else {
			return "", src, 1, fmt.Errorf("frontmatter not closed")
		}
	}
	fm = rest[:end]
	body = ""
	if end+5 <= len(rest) {
		body = rest[end+5:]
	}
	bodyLine = strings.Count(fm, "\n") + 4
	return fm, body, bodyLine, nil
}

// Parse parses the frontmatter text.
func Parse(fm string) (Map, error) {
	// Purpose: turn the supported YAML subset into a Map.
	out := Map{}
	lines := strings.Split(fm, "\n")
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return nil, fmt.Errorf("line %d: unexpected indentation", i+1)
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("line %d: expected key: value", i+1)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch {
		case val == ">" || val == "|" || val == ">-" || val == "|-":
			block, next := indentedBlock(lines, i+1)
			out[key] = joinBlock(block, val[0] == '>')
			i = next - 1
		case val == "":
			items, next := blockList(lines, i+1)
			if next == i+1 {
				out[key] = nil
			} else {
				out[key] = items
				i = next - 1
			}
		default:
			v, err := scalarOrList(val)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i+1, err)
			}
			out[key] = v
		}
	}
	return out, nil
}

// String returns the value of key as a string, or "" when absent or not a string.
func (m Map) String(key string) string {
	// Purpose: read a scalar field safely.
	s, _ := m[key].(string)
	return s
}

// List returns the value of key as a list. A scalar becomes a one-item list.
func (m Map) List(key string) []string {
	// Purpose: read a list field safely.
	switch v := m[key].(type) {
	case []string:
		return v
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	}
	return nil
}

func stripComment(line string) string {
	// Purpose: remove a `#` comment that is outside quotes.
	inS, inD := false, false
	for i, r := range line {
		switch r {
		case '\'':
			if !inD {
				inS = !inS
			}
		case '"':
			if !inS {
				inD = !inD
			}
		case '#':
			if !inS && !inD && (i == 0 || line[i-1] == ' ') {
				return strings.TrimRight(line[:i], " ")
			}
		}
	}
	return line
}

func indentedBlock(lines []string, start int) ([]string, int) {
	// Purpose: collect the indented lines of a block scalar.
	var block []string
	i := start
	for ; i < len(lines); i++ {
		l := lines[i]
		if strings.TrimSpace(l) == "" {
			block = append(block, "")
			continue
		}
		if !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			break
		}
		block = append(block, strings.TrimSpace(l))
	}
	return block, i
}

func joinBlock(block []string, folded bool) string {
	// Purpose: join block scalar lines as YAML folded or literal text.
	for len(block) > 0 && block[len(block)-1] == "" {
		block = block[:len(block)-1]
	}
	if !folded {
		return strings.Join(block, "\n")
	}
	var b strings.Builder
	for i, l := range block {
		if i > 0 {
			if l == "" || block[i-1] == "" {
				b.WriteString("\n")
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString(l)
	}
	return b.String()
}

func blockList(lines []string, start int) ([]string, int) {
	// Purpose: collect `- item` lines that follow a key with no value.
	var items []string
	i := start
	for ; i < len(lines); i++ {
		l := strings.TrimSpace(stripComment(lines[i]))
		if l == "" {
			continue
		}
		if !strings.HasPrefix(l, "- ") && l != "-" {
			break
		}
		items = append(items, unquote(strings.TrimSpace(strings.TrimPrefix(l, "-"))))
	}
	return items, i
}

func scalarOrList(val string) (any, error) {
	// Purpose: parse an inline value: null, an inline list, or a scalar.
	if val == "null" || val == "~" {
		return nil, nil
	}
	if strings.HasPrefix(val, "[") {
		if !strings.HasSuffix(val, "]") {
			return nil, fmt.Errorf("inline list not closed")
		}
		inner := strings.TrimSpace(val[1 : len(val)-1])
		if inner == "" {
			return []string{}, nil
		}
		var items []string
		for _, p := range splitList(inner) {
			items = append(items, unquote(strings.TrimSpace(p)))
		}
		return items, nil
	}
	return unquote(val), nil
}

func splitList(s string) []string {
	// Purpose: split an inline list on commas that are outside quotes.
	var parts []string
	inS, inD, start := false, false, 0
	for i, r := range s {
		switch r {
		case '\'':
			if !inD {
				inS = !inS
			}
		case '"':
			if !inS {
				inD = !inD
			}
		case ',':
			if !inS && !inD {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, s[start:])
}

func unquote(s string) string {
	// Purpose: remove matching single or double quotes.
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		inner := s[1 : len(s)-1]
		if s[0] == '"' {
			inner = strings.ReplaceAll(inner, `\"`, `"`)
		} else {
			inner = strings.ReplaceAll(inner, "''", "'")
		}
		return inner
	}
	return s
}
