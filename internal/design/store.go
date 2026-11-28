// Package design loads the design/ folder: parts, chapters and addressable
// sections. See design/specification/format.md.
package design

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/consoleillusion/maid/internal/frontmatter"
)

// Dir is the design folder name, relative to the project root.
const Dir = "design"

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)

// Store is a loaded design folder.
type Store struct {
	Root     string // project root
	Chapters []*Chapter
	byID     map[string]*Chapter
}

// Chapter is one Markdown file in the design.
type Chapter struct {
	ID       string // derived from the file name
	Title    string // derived from the level-1 heading
	Part     string // derived from the containing folder
	Path     string // relative to the project root, slash-separated
	FM       frontmatter.Map
	FMErr    error
	Body     string
	BodyLine int
	Sections []*Section
}

// Section is a heading and its content up to the next heading of the same
// or a higher level. A section is addressed by chapter ID and heading slug.
type Section struct {
	Anchor  string
	Heading string
	Level   int
	Line    int // 1-based line in the file
	Text    string
}

// Load reads every .md file under root/design.
func Load(root string) (*Store, error) {
	// Purpose: parse all design chapters into a Store.
	s := &Store{Root: root, byID: map[string]*Chapter{}}
	dir := filepath.Join(root, Dir)
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return s, nil // no design yet: the bootstrap phase
	}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") || p == filepath.Join(dir, IndexFile) {
			return nil // index.md is the reading order, not a chapter
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
	ch.Sections = splitSections(body, bodyLine)
	ch.ID = chapterID(rel)
	ch.Title = chapterTitle(ch.Sections)
	ch.Part = folderPart(ch.Path)
	return ch, nil
}

// chapterID derives the chapter ID from its file name.
func chapterID(p string) string {
	// Purpose: use the file name without its extension as the chapter ID.
	base := path.Base(p)
	return strings.TrimSuffix(base, path.Ext(base))
}

// chapterTitle derives the chapter title from its level-1 heading.
func chapterTitle(sections []*Section) string {
	// Purpose: use the first level-1 heading as the chapter title.
	for _, sec := range sections {
		if sec.Level == 1 {
			return strings.TrimSpace(sec.Heading)
		}
	}
	return ""
}

// folderPart derives the part from the folder that contains the chapter.
func folderPart(p string) string {
	// Purpose: return the immediate folder name under design/.
	dir := path.Dir(strings.TrimPrefix(p, Dir+"/"))
	if dir == "." {
		return Dir
	}
	if i := strings.LastIndex(dir, "/"); i >= 0 {
		return dir[i+1:]
	}
	return dir
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
		sec := &Section{Heading: h.text, Level: h.level, Line: firstLine + h.idx, Anchor: slug(h.text)}
		sec.Text = strings.TrimRight(strings.Join(lines[h.idx:end], "\n"), "\n")
		out = append(out, sec)
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
	// Purpose: look up a chapter by its file name.
	return s.byID[id]
}

// Resolves reports whether ref names a chapter or a chapter section.
func (s *Store) Resolves(ref string) bool {
	// Purpose: decide whether a link target exists.
	if _, ok := s.byID[ref]; ok {
		return true
	}
	chID, anchor, found := strings.Cut(ref, "#")
	if !found {
		return false
	}
	c := s.byID[chID]
	return c != nil && c.Section(anchor) != nil
}

// Section returns the section with the given anchor.
func (c *Chapter) Section(anchor string) *Section {
	// Purpose: look up a section of a chapter by its heading slug.
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
