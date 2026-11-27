package design

import (
	"fmt"
	"strings"
)

// Entry is one row of the design index.
type Entry struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Part    string   `json:"part"`
	Status  string   `json:"status"`
	Style   string   `json:"style"`
	Path    string   `json:"path"`
	Summary string   `json:"summary"`
	Links   []string `json:"links,omitempty"`
	Items   int      `json:"items"`
}

// Filter selects index entries. Empty fields match everything.
type Filter struct {
	Part   string
	Status string
	Style  string
	Link   string // entries that link to this ID
}

// Index lists chapters with summaries only.
func (s *Store) Index(f Filter) []Entry {
	// Purpose: return the catalog of chapters that match the filter.
	var out []Entry
	for _, ch := range s.Chapters {
		if ch.FM == nil {
			continue
		}
		e := Entry{
			ID: ch.ID, Title: ch.FM.String("title"), Part: ch.FM.String("part"),
			Status: ch.FM.String("status"), Style: ch.FM.String("style"),
			Path: ch.Path, Summary: ch.FM.String("summary"), Links: ch.FM.List("links"),
		}
		for _, sec := range ch.Sections {
			if sec.ItemID != "" {
				e.Items++
			}
		}
		if !match(f, e) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func match(f Filter, e Entry) bool {
	// Purpose: apply a Filter to one index entry.
	if f.Part != "" && e.Part != f.Part || f.Status != "" && e.Status != f.Status || f.Style != "" && e.Style != f.Style {
		return false
	}
	if f.Link != "" {
		for _, l := range e.Links {
			if l == f.Link {
				return true
			}
		}
		return false
	}
	return true
}

// Excerpt is the text of one requested chapter or section.
type Excerpt struct {
	Ref     string `json:"ref"`
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Style   string `json:"style"`
	Status  string `json:"status"`
	Version string `json:"version"`
	Text    string `json:"text"`
}

// Read returns excerpts for refs. A ref is a chapter ID, an item ID, or
// chapter#anchor. With summaryOnly, chapters return only their summary.
func (s *Store) Read(refs []string, summaryOnly bool) ([]Excerpt, error) {
	// Purpose: resolve each ref to the smallest text that answers it.
	var out []Excerpt
	for _, ref := range refs {
		ex, err := s.read(ref, summaryOnly)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}

// ReadAll is like Read but never fails as a whole: refs that do not
// resolve are returned in the second result with suggestions.
func (s *Store) ReadAll(refs []string, summaryOnly bool) ([]Excerpt, []NotFound) {
	// Purpose: give agents every excerpt it can, and hints for the rest.
	var out []Excerpt
	var missing []NotFound
	for _, ref := range refs {
		ex, err := s.read(ref, summaryOnly)
		var nf NotFound
		switch {
		case err == nil:
			out = append(out, ex)
		case asNotFound(err, &nf):
			missing = append(missing, nf)
		default:
			missing = append(missing, NotFound{Ref: ref})
		}
	}
	return out, missing
}

func asNotFound(err error, nf *NotFound) bool {
	// Purpose: extract a NotFound from an error.
	v, ok := err.(NotFound)
	if ok {
		*nf = v
	}
	return ok
}

func (s *Store) read(ref string, summaryOnly bool) (Excerpt, error) {
	// Purpose: resolve one ref to an excerpt.
	canonical, err := s.Resolve(ref)
	if err != nil {
		return Excerpt{}, err
	}
	ref = canonical
	if ch := s.byID[ref]; ch != nil {
		text := strings.TrimSpace(ch.Body)
		if summaryOnly {
			text = ch.FM.String("summary")
		}
		return excerpt(ref, ch, ch.BodyLine, text), nil
	}
	if it, ok := s.items[ref]; ok {
		if it.Section != nil {
			return excerpt(ref, it.Chapter, it.Section.Line, it.Section.Text), nil
		}
		return excerpt(ref, it.Chapter, it.Line, lineAt(it.Chapter, it.Line)), nil
	}
	chID, anchor, found := strings.Cut(ref, "#")
	if found {
		if ch := s.byID[chID]; ch != nil {
			if sec := ch.Section(anchor); sec != nil {
				return excerpt(ref, ch, sec.Line, sec.Text), nil
			}
			if line, ok := tableItems(ch)[anchor]; ok {
				return excerpt(ref, ch, line, lineAt(ch, line)), nil
			}
		}
	}
	return Excerpt{}, fmt.Errorf("unknown design ref %q", ref)
}

func excerpt(ref string, ch *Chapter, line int, text string) Excerpt {
	// Purpose: build an Excerpt with chapter metadata and a version hash.
	return Excerpt{
		Ref: ref, Path: ch.Path, Line: line, Text: text, Version: Version(text),
		Style: ch.FM.String("style"), Status: ch.FM.String("status"),
	}
}

func lineAt(ch *Chapter, line int) string {
	// Purpose: return one body line by its file line number.
	lines := strings.Split(ch.Body, "\n")
	i := line - ch.BodyLine
	if i < 0 || i >= len(lines) {
		return ""
	}
	return lines[i]
}
