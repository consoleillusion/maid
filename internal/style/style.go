// Package style holds the registered writing styles and their validators.
// See design/specification/styles.md.
package style

import (
	"regexp"
	"strings"
)

// Severity of a finding.
const (
	Error  = "error"
	Advice = "advice"
)

// Finding is one problem in a text. Line is relative to the text start (1-based).
type Finding struct {
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// Style is a registered writing style.
type Style struct {
	ID          string
	Card        string // path of the style card, relative to design/
	Enforcement string
	Validate    func(text string) []Finding // chapter body or block content
}

// Registry lists the registered styles by ID.
var Registry = map[string]Style{
	"ste":     {ID: "ste", Card: "specification/styles.md#simplified-technical-english", Enforcement: Error, Validate: ValidateSTE},
	"ears":    {ID: "ears", Card: "specification/styles.md#ears-requirements", Enforcement: Error, Validate: ValidateEARS},
	"sbvr-se": {ID: "sbvr-se", Card: "specification/styles.md#sbvr-structured-english", Enforcement: Error, Validate: ValidateSBVR},
	"gherkin": {ID: "gherkin", Card: "specification/styles.md#gherkin", Enforcement: Error, Validate: ValidateGherkin},
	"ace":     {ID: "ace", Card: "specification/styles.md#attempto-controlled-english", Enforcement: Advice, Validate: nil},
}

// Block is a fenced code block found in a body.
type Block struct {
	Lang string
	Line int // line of the opening fence, relative to body start (1-based)
	Text string
}

// Prose returns the body with fenced blocks, tables, headings and HTML
// comments blanked out, so line numbers stay correct. It also returns the
// fenced blocks.
func Prose(body string) (string, []Block) {
	// Purpose: separate checkable prose from code blocks and tables.
	lines := strings.Split(body, "\n")
	var blocks []Block
	var cur *Block
	var buf []string
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "```") {
			if cur == nil {
				cur = &Block{Lang: strings.TrimSpace(strings.TrimPrefix(t, "```")), Line: i + 1}
				buf = nil
			} else {
				cur.Text = strings.Join(buf, "\n")
				blocks = append(blocks, *cur)
				cur = nil
			}
			lines[i] = ""
			continue
		}
		if cur != nil {
			buf = append(buf, l)
			lines[i] = ""
			continue
		}
		if strings.HasPrefix(t, "|") || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "<!--") {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n"), blocks
}

var sentenceEnd = regexp.MustCompile(`[.!?](\s|$)`)

// sentences returns the sentences of prose with their starting line.
func sentences(prose string) []struct {
	Line int
	Text string
} {
	// Purpose: split prose into sentences; paragraphs and list items end a sentence.
	var out []struct {
		Line int
		Text string
	}
	var cur []string
	start := 0
	flush := func() {
		s := strings.TrimSpace(strings.Join(cur, " "))
		if s != "" {
			out = append(out, struct {
				Line int
				Text string
			}{start, s})
		}
		cur = nil
	}
	for i, l := range strings.Split(prose, "\n") {
		t := strings.TrimSpace(l)
		if t == "" || listItem(t) {
			flush()
			if t == "" {
				continue
			}
			t = strings.TrimSpace(listMarker.ReplaceAllString(t, ""))
		}
		for t != "" {
			if len(cur) == 0 {
				start = i + 1
			}
			loc := sentenceEnd.FindStringIndex(t)
			if loc == nil {
				cur = append(cur, t)
				break
			}
			cur = append(cur, t[:loc[0]+1])
			flush()
			t = strings.TrimSpace(t[loc[1]:])
		}
	}
	flush()
	return out
}

var listMarker = regexp.MustCompile(`^([-*+]|\d+\.)\s+`)

func listItem(t string) bool {
	// Purpose: detect a Markdown list item.
	return listMarker.MatchString(t)
}

func words(s string) int {
	// Purpose: count words, ignoring inline code spans as single words.
	return len(strings.Fields(s))
}
