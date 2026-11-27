// Package design loads the design/ folder: parts, chapters, sections and
// items with stable IDs. See design/03-specification/01-chapter-format.md.
package design

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/frontmatter"
)

// Dir is the design folder name, relative to the project root.
const Dir = "design"

// ItemIDPattern matches the stable ID prefixes from the chapter format.
var ItemIDPattern = regexp.MustCompile(`\b(REQ-[A-Z]+-\d{3}|ADR-\d{4}|Q-\d{3}|G-\d{3}|NG-\d{3}|SM-\d{3}|C-\d{3}|A-\d{3})\b`)

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
var tableItemRe = regexp.MustCompile(`^\|\s*((?:SM|C|A|Q)-\d{3})\s*\|`)

// Store is a loaded design folder.
type Store struct {
	Root     string // project root
	Chapters []*Chapter
	byID     map[string]*Chapter
	items    map[string]ItemRef
}

// Chapter is one Markdown file in the design.
type Chapter struct {
	ID       string
	Path     string // relative to the project root, slash-separated
	FM       frontmatter.Map
	FMErr    error
	Body     string
	BodyLine int
	Sections []*Section
}

// Section is a heading and its content up to the next heading of the same
// or a higher level.
type Section struct {
	Anchor  string
	Heading string
	Level   int
	Line    int // 1-based line in the file
	ItemID  string
	Text    string
}

// ItemRef locates an item.
type ItemRef struct {
	Chapter *Chapter
	Section *Section // nil for table-row items
	Line    int
}

// Load reads every .md file under root/design.
func Load(root string) (*Store, error) {
	// Purpose: parse all design chapters into a Store.
	s := &Store{Root: root, byID: map[string]*Chapter{}, items: map[string]ItemRef{}}
	dir := filepath.Join(root, Dir)
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		ch, err := loadChapter(root, p)
		if err != nil {
			return err
		}
		s.Chapters = append(s.Chapters, ch)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load design: %w", err)
	}
	sort.Slice(s.Chapters, func(i, j int) bool { return s.Chapters[i].Path < s.Chapters[j].Path })
	for _, ch := range s.Chapters {
		if ch.ID != "" {
			if _, dup := s.byID[ch.ID]; !dup {
				s.byID[ch.ID] = ch
			}
		}
		for _, sec := range ch.Sections {
			if sec.ItemID != "" {
				if _, dup := s.items[sec.ItemID]; !dup {
					s.items[sec.ItemID] = ItemRef{Chapter: ch, Section: sec, Line: sec.Line}
				}
			}
		}
		for id, line := range tableItems(ch) {
			if _, dup := s.items[id]; !dup {
				s.items[id] = ItemRef{Chapter: ch, Line: line}
			}
		}
	}
	return s, nil
}

func loadChapter(root, p string) (*Chapter, error) {
	// Purpose: read one chapter file and split it into frontmatter and sections.
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(root, p)
	ch := &Chapter{Path: filepath.ToSlash(rel)}
	fmText, body, bodyLine, err := frontmatter.Split(string(b))
	ch.Body, ch.BodyLine = body, bodyLine
	if err != nil {
		ch.FMErr = err
	} else if ch.FM, err = frontmatter.Parse(fmText); err != nil {
		ch.FMErr = err
	}
	if ch.FM != nil {
		ch.ID = ch.FM.String("id")
	}
	ch.Sections = splitSections(body, bodyLine)
	return ch, nil
}

func splitSections(body string, firstLine int) []*Section {
	// Purpose: find headings outside code fences and cut the body into sections.
	lines := strings.Split(body, "\n")
	type head struct {
		idx, level int
		text       string
	}
	var heads []head
	inFence := false
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := headingRe.FindStringSubmatch(l); m != nil {
			heads = append(heads, head{i, len(m[1]), m[2]})
		}
	}
	var out []*Section
	for k, h := range heads {
		end := len(lines)
		for _, n := range heads[k+1:] {
			if n.level <= h.level {
				end = n.idx
				break
			}
		}
		sec := &Section{Heading: h.text, Level: h.level, Line: firstLine + h.idx}
		sec.Text = strings.TrimRight(strings.Join(lines[h.idx:end], "\n"), "\n")
		if id := leadingItemID(h.text); id != "" {
			sec.ItemID, sec.Anchor = id, id
		} else {
			sec.Anchor = slug(h.text)
		}
		out = append(out, sec)
	}
	return out
}

func leadingItemID(heading string) string {
	// Purpose: return the item ID that starts a heading, if any.
	loc := ItemIDPattern.FindStringIndex(heading)
	if loc == nil || loc[0] != 0 {
		return ""
	}
	return heading[loc[0]:loc[1]]
}

func tableItems(ch *Chapter) map[string]int {
	// Purpose: find items defined as the first cell of a table row.
	out := map[string]int{}
	for i, l := range strings.Split(ch.Body, "\n") {
		if m := tableItemRe.FindStringSubmatch(l); m != nil {
			if _, ok := out[m[1]]; !ok {
				out[m[1]] = ch.BodyLine + i
			}
		}
	}
	return out
}

func slug(s string) string {
	// Purpose: make a heading anchor in GitHub style.
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

// Chapter returns the chapter with the given ID.
func (s *Store) Chapter(id string) *Chapter {
	// Purpose: look up a chapter by its frontmatter ID.
	return s.byID[id]
}

// Item returns the location of an item ID.
func (s *Store) Item(id string) (ItemRef, bool) {
	// Purpose: look up a requirement, goal, question or other item.
	r, ok := s.items[id]
	return r, ok
}

// Resolves reports whether id names a chapter, an item, or chapter#anchor.
func (s *Store) Resolves(id string) bool {
	// Purpose: decide whether a link target exists.
	if _, ok := s.byID[id]; ok {
		return true
	}
	if _, ok := s.items[id]; ok {
		return true
	}
	ch, anchor, found := strings.Cut(id, "#")
	if !found {
		return false
	}
	c := s.byID[ch]
	if c == nil {
		return false
	}
	return c.Section(anchor) != nil || hasTableItem(c, anchor)
}

func hasTableItem(c *Chapter, id string) bool {
	// Purpose: check whether a table-row item is defined in a chapter.
	_, ok := tableItems(c)[id]
	return ok
}

// Section returns the section with the given anchor.
func (c *Chapter) Section(anchor string) *Section {
	// Purpose: look up a section of a chapter by anchor.
	for _, sec := range c.Sections {
		if sec.Anchor == anchor {
			return sec
		}
	}
	return nil
}

// Version returns a short content hash used for optimistic concurrency.
func Version(text string) string {
	// Purpose: hash text so writers can detect concurrent changes.
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:6])
}
