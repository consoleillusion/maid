package design

import (
	_ "embed" // index.typ template
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// indexTyp builds the design as one PDF from index.md (REQ-DOC-018).
//
//go:embed index.typ
var indexTyp string

// IndexFile is the front page and reading order of the design (REQ-DOC-014).
const IndexFile = "index.md"

const indexStart = "---\nid: index\ntitle: Design\nstatus: draft\nstyle: ste\nowner: user\nsummary: Front page of the design and the reading order of its chapters.\nlinks: []\n---\n\n# Design\n\nThis page lists the chapters of the design in reading order.\n"

// IndexEntry is one chapter in the reading order of design/index.md.
type IndexEntry struct {
	Part  string // the level-2 heading above the entry
	Title string // the link text
	Path  string // relative to the project root, for example design/foundation/vision.md
}

var (
	indexPartRe = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	indexItemRe = regexp.MustCompile(`^\s*\d+[.)]\s+\[([^\]]+)\]\(([^)\s]+\.md)\)`)
)

// Order returns the chapters in the order that design/index.md lists them
// (REQ-DOC-014).
func (s *Store) Order() []IndexEntry {
	// Purpose: read the reading order of the design.
	b, err := os.ReadFile(filepath.Join(s.Root, Dir, IndexFile))
	if err != nil {
		return nil
	}
	var out []IndexEntry
	part := ""
	for _, l := range strings.Split(bodyOf(string(b)), "\n") {
		if m := indexPartRe.FindStringSubmatch(l); m != nil {
			part = m[1]
		} else if m := indexItemRe.FindStringSubmatch(l); m != nil {
			out = append(out, IndexEntry{Part: part, Title: m[1], Path: path.Join(Dir, path.Clean(m[2]))})
		}
	}
	return out
}

// HasIndex reports whether design/index.md exists.
func (s *Store) HasIndex() bool {
	// Purpose: tell whether the reading order is defined.
	_, err := os.Stat(filepath.Join(s.Root, Dir, IndexFile))
	return err == nil
}

func bodyOf(text string) string {
	// Purpose: drop a leading frontmatter block.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.HasPrefix(text, "---\n") {
		if i := strings.Index(text[4:], "\n---\n"); i >= 0 {
			return text[4+i+5:]
		}
	}
	return text
}

// PartOf returns the part folder of a chapter path such as
// design/foundation/vision.md, or "" when the file is not in a part folder.
func PartOf(p string) string {
	// Purpose: tell part chapters from root files, styles and decisions.
	dir := path.Dir(strings.TrimPrefix(p, Dir+"/"))
	if dir == "." || strings.Contains(dir, "/") || dir == "styles" || dir == "decisions" {
		return ""
	}
	return dir
}

func partTitle(folder string) string {
	// Purpose: turn a part folder name into its index heading.
	w := strings.ReplaceAll(folder, "-", " ")
	return strings.ToUpper(w[:1]) + w[1:]
}

// addToIndex lists a new chapter in index.md and creates index.md and
// index.typ when they are missing (REQ-DOC-017, REQ-DOC-019). It returns the
// changed paths, relative to the project root.
func (s *Store) addToIndex(rel, title string) ([]string, error) {
	// Purpose: keep the reading order complete when a chapter is created.
	var changed []string
	dir := filepath.Join(s.Root, Dir)
	if typ := filepath.Join(dir, "index.typ"); !exists(typ) {
		if err := atomicWrite(typ, indexTyp); err != nil {
			return nil, err
		}
		changed = append(changed, path.Join(Dir, "index.typ"))
	}
	idx := filepath.Join(dir, IndexFile)
	text := indexStart
	if b, err := os.ReadFile(idx); err == nil {
		text = strings.ReplaceAll(string(b), "\r\n", "\n")
	}
	if err := atomicWrite(idx, insertEntry(text, path.Dir(rel), title, rel)); err != nil {
		return nil, err
	}
	return append(changed, path.Join(Dir, IndexFile)), nil
}

func exists(p string) bool {
	// Purpose: report whether a file exists.
	_, err := os.Stat(p)
	return err == nil
}

// insertEntry appends "N. [title](rel)" to the list of the part that holds
// folder, or adds a heading for the part at the end.
func insertEntry(text, folder, title, rel string) string {
	// Purpose: place a new chapter at the end of its part in index.md.
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	start := partHeading(lines, folder)
	if start < 0 {
		return strings.Join(lines, "\n") + "\n\n## " + partTitle(folder) + "\n\n1. [" + title + "](" + rel + ")\n"
	}
	n, last := 0, start
	for j := start + 1; j < len(lines) && !strings.HasPrefix(lines[j], "#"); j++ {
		if indexItemRe.MatchString(lines[j]) {
			n, last = n+1, j
		}
	}
	entry := fmt.Sprintf("%d. [%s](%s)", n+1, title, rel)
	add := []string{entry}
	if n == 0 {
		add = []string{"", entry}
	}
	out := append(append(append([]string{}, lines[:last+1]...), add...), lines[last+1:]...)
	return strings.Join(out, "\n") + "\n"
}

func partHeading(lines []string, folder string) int {
	// Purpose: find the index line of the part heading for a folder.
	head := -1
	for i, l := range lines {
		if indexPartRe.MatchString(l) {
			head = i
			if strings.EqualFold(indexPartRe.FindStringSubmatch(l)[1], partTitle(folder)) {
				return i
			}
		} else if m := indexItemRe.FindStringSubmatch(l); m != nil && head >= 0 && path.Dir(path.Clean(m[2])) == folder {
			return head
		}
	}
	return -1
}
