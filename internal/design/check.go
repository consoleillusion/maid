package design

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/consoleillusion/maid/internal/style"
)

// Finding is one problem in the design.
type Finding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// RequiredFields are the frontmatter fields every chapter must have.
// ID, title and part come from the path and the level-1 heading, not frontmatter.
var RequiredFields = []string{"status", "style", "owner", "summary", "links"}

// Statuses are the allowed status values.
var Statuses = map[string]bool{"stub": true, "draft": true, "review": true, "accepted": true, "superseded": true}

const maxSummaryWords = 60

var mdLink = regexp.MustCompile(`\]\(([^)\s#:]+\.md)(#[^)]*)?\)`)

// Check validates the whole design and returns all findings.
func (s *Store) Check() []Finding {
	// Purpose: run structure, link and style checks over every chapter.
	var out []Finding
	seen := map[string]string{}
	for _, ch := range s.Chapters {
		if isTemplate(ch) {
			continue
		}
		out = append(out, checkFrontmatter(ch)...)
		if ch.ID != "" {
			if other, dup := seen[ch.ID]; dup {
				out = append(out, Finding{ch.Path, 1, "duplicate-chapter", style.Error, "duplicate chapter id " + ch.ID + " (also in " + other + ")"})
			}
			seen[ch.ID] = ch.Path
		}
		out = append(out, checkHeadings(ch)...)
		out = append(out, s.checkLinks(ch)...)
		out = append(out, checkStyle(ch)...)
	}
	return append(out, s.checkIndex()...)
}

var numberedHeading = regexp.MustCompile(`^#{1,6}\s+\d`)

func checkHeadings(ch *Chapter) []Finding {
	// Purpose: report headings that start with a position number.
	var out []Finding
	fence := false
	for i, l := range strings.Split(ch.Body, "\n") {
		if strings.HasPrefix(l, "```") {
			fence = !fence
		}
		if !fence && numberedHeading.MatchString(l) {
			out = append(out, Finding{ch.Path, ch.BodyLine + i, "heading-numbered", style.Error, "heading starts with a number; index.md holds the order: " + l})
		}
	}
	return out
}

func (s *Store) checkIndex() []Finding {
	// Purpose: compare design/index.md with the chapters in part folders.
	idx := path.Join(Dir, IndexFile)
	if !s.HasIndex() {
		if len(s.Chapters) == 0 {
			return nil
		}
		return []Finding{{idx, 1, "index", style.Advice, "design/index.md is missing; create it with the chapters in reading order"}}
	}
	var out []Finding
	listed := map[string]bool{}
	for _, e := range s.Order() {
		listed[e.Path] = true
		if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(e.Path))); err != nil {
			out = append(out, Finding{idx, 1, "index", style.Error, "index entry does not exist: " + e.Path})
		}
	}
	for _, ch := range s.Chapters {
		if PartOf(ch.Path) != "" && !isTemplate(ch) && !listed[ch.Path] {
			out = append(out, Finding{ch.Path, 1, "index", style.Error, "chapter is not listed in design/index.md"})
		}
	}
	return out
}

func isTemplate(ch *Chapter) bool {
	// Purpose: skip template files, which contain placeholders on purpose.
	return strings.HasSuffix(ch.Path, "/template.md")
}

func checkFrontmatter(ch *Chapter) []Finding {
	// Purpose: check required fields, status values, style and summary length.
	if ch.FMErr != nil {
		return []Finding{{ch.Path, 1, "frontmatter", style.Error, "frontmatter: " + ch.FMErr.Error()}}
	}
	var out []Finding
	for _, f := range RequiredFields {
		if _, ok := ch.FM[f]; !ok {
			out = append(out, Finding{ch.Path, 1, "frontmatter", style.Error, "missing frontmatter field " + f})
		}
	}
	if st := ch.FM.String("status"); st != "" && !Statuses[st] {
		out = append(out, Finding{ch.Path, 1, "status", style.Error, "invalid status " + st})
	}
	if sty := ch.FM.String("style"); sty != "" {
		if _, ok := style.Registry[sty]; !ok {
			out = append(out, Finding{ch.Path, 1, "style-unknown", style.Error, "unregistered style " + sty})
		}
	}
	if n := len(strings.Fields(ch.FM.String("summary"))); n > maxSummaryWords {
		out = append(out, Finding{ch.Path, 1, "summary-length", style.Error, fmt.Sprintf("summary has %d words (max %d)", n, maxSummaryWords)})
	}
	return out
}

func (s *Store) checkLinks(ch *Chapter) []Finding {
	// Purpose: check frontmatter links and relative Markdown links.
	var out []Finding
	for _, l := range ch.FM.List("links") {
		if !s.Resolves(l) {
			out = append(out, Finding{ch.Path, 1, "link", style.Error, "link does not resolve: " + l})
		}
	}
	for i, line := range strings.Split(ch.Body, "\n") {
		for _, m := range mdLink.FindAllStringSubmatch(line, -1) {
			target := path.Join(path.Dir(ch.Path), m[1])
			if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(target))); err != nil {
				out = append(out, Finding{ch.Path, ch.BodyLine + i, "link", style.Error, "broken link: " + m[1]})
			}
		}
	}
	return out
}

func checkStyle(ch *Chapter) []Finding {
	// Purpose: run the chapter style on the body and embedded styles on their blocks.
	var out []Finding
	add := func(fs []style.Finding, offset int, enforcement string) {
		for _, f := range fs {
			sev := f.Severity
			if enforcement == style.Advice {
				sev = style.Advice
			}
			out = append(out, Finding{ch.Path, ch.BodyLine + offset + f.Line - 1, f.Rule, sev, f.Message})
		}
	}
	if st, ok := style.Registry[ch.FM.String("style")]; ok && st.Validate != nil {
		add(st.Validate(ch.Body), 0, st.Enforcement)
	}
	_, blocks := style.Prose(ch.Body)
	for _, b := range blocks {
		if st, ok := style.Registry[b.Lang]; ok && st.Validate != nil {
			add(st.Validate(b.Text), b.Line, st.Enforcement)
		}
	}
	return out
}
