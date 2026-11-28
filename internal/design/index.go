package design

import (
	_ "embed" // index.typ template
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

const indexStart = "# Design\n"

// IndexEntry is one chapter in the reading order of design/index.md.
type IndexEntry struct {
	Part  string // the part folder name, with hyphens as spaces and a capital first letter
	Title string // the title of the chapter
	Path  string // relative to the project root, for example design/foundation/vision.md
}

var (
	indexPartRe = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	// indexItemRe matches "- path.md", "- [Title](path.md)" and the same
	// with a number instead of the bullet.
	indexItemRe = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+(?:\[([^\]]+)\]\(([^)\s]+\.md)\)|([^\s\[]+\.md))\s*$`)
)

// indexItem returns the title and path of an index list line.
func indexItem(l string) (title, p string, ok bool) {
	// Purpose: read one entry in either the plain or the link form.
	m := indexItemRe.FindStringSubmatch(l)
	if m == nil {
		return "", "", false
	}
	if m[2] != "" {
		return m[1], path.Clean(m[2]), true
	}
	return "", path.Clean(m[3]), true
}

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
			continue
		}
		title, p, ok := indexItem(l)
		if !ok {
			continue
		}
		e := IndexEntry{Part: part, Title: title, Path: path.Join(Dir, p)}
		if f := PartOf(e.Path); f != "" && e.Part == "" || f != "" && title == "" {
			e.Part = partTitle(f)
		}
		if e.Title == "" {
			for _, ch := range s.Chapters {
				if ch.Path == e.Path {
					e.Title = ch.FM.String("title")
				}
			}
		}
		out = append(out, e)
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
func (s *Store) addToIndex(rel string) ([]string, error) {
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
	if err := atomicWrite(idx, insertEntry(text, path.Dir(rel), rel)); err != nil {
		return nil, err
	}
	return append(changed, path.Join(Dir, IndexFile)), nil
}

func exists(p string) bool {
	// Purpose: report whether a file exists.
	_, err := os.Stat(p)
	return err == nil
}

// insertEntry puts rel after the last path in the same folder, or at the
// end of the list (REQ-DOC-017).
func insertEntry(text, folder, rel string) string {
	// Purpose: place a new chapter in the reading order.
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	at, last := -1, -1
	for i, l := range lines {
		if _, p, ok := indexItem(l); ok {
			last = i
			if path.Dir(p) == folder {
				at = i
			}
		}
	}
	if at < 0 {
		at = last
	}
	entry := "- " + rel
	if at < 0 {
		lines = append(lines, "", entry)
	} else {
		lines = append(lines[:at+1], append([]string{entry}, lines[at+1:]...)...)
	}
	return strings.Join(lines, "\n") + "\n"
}
