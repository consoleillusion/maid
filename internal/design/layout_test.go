package design

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/style"
)

const indexHead = "---\nid: index\ntitle: Design\nstatus: draft\nstyle: ste\nowner: paul\nsummary: Reading order.\nlinks: []\n---\n\n# Design\n\nThe design.\n\n"

func chapterText(id, part string) string {
	// Purpose: build a minimal valid chapter with the given part field.
	return "---\nid: " + id + "\ntitle: T\npart: " + part + "\nstatus: draft\nstyle: ste\nowner: paul\nsummary: S.\nlinks: []\n---\n\n# T\n\nText.\n"
}

func createChapter(t *testing.T, root, path, id, title string) error {
	// Purpose: create one chapter through a freshly loaded store.
	t.Helper()
	s, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateChapter(ChapterSpec{Path: path, ID: id, Title: title, Style: "ste", Summary: title + ".", Body: "# " + title + "\n\nText.\n"})
	return err
}

func countFindings(fs []Finding, rule string) int {
	// Purpose: count error findings of one rule.
	n := 0
	for _, f := range fs {
		if f.Rule == rule && f.Severity == style.Error {
			n++
		}
	}
	return n
}

// REQ-DOC-017 REQ-DOC-019 scenario: new chapters go into the index
func TestCreateChapterAddsToIndex(t *testing.T) {
	// Purpose: created chapters are listed under their part in creation order.
	root := t.TempDir()
	for _, c := range [][3]string{{"foundation/vision.md", "vision", "Vision"}, {"foundation/glossary.md", "glossary", "Glossary"}, {"method/phases.md", "phases", "Phases"}} {
		if err := createChapter(t, root, c[0], c[1], c[2]); err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "design/index.md"))
	if err != nil {
		t.Fatal(err)
	}
	idx := string(b)
	found := "## Foundation\n\n1. [Vision](foundation/vision.md)\n2. [Glossary](foundation/glossary.md)\n"
	method := "## Method\n\n1. [Phases](method/phases.md)\n"
	if !strings.Contains(idx, found) || !strings.Contains(idx, method) || strings.Index(idx, found) > strings.Index(idx, method) {
		t.Fatalf("index.md = %q", idx)
	}
	if _, err := os.Stat(filepath.Join(root, "design/index.typ")); err != nil {
		t.Error("design/index.typ was not created")
	}
	s, _ := Load(root)
	if p := s.Chapter("vision").FM.String("part"); p != "foundation" {
		t.Errorf("part = %q, want foundation", p)
	}
	if n := countFindings(s.Check(), "REQ-DOC-016"); n != 0 {
		t.Errorf("%d REQ-DOC-016 errors after creating chapters", n)
	}
}

// REQ-DOC-001 scenario: numbered path
func TestCreateChapterRejectsNumberedPath(t *testing.T) {
	// Purpose: part and chapter names carry no position numbers.
	root := t.TempDir()
	for _, p := range []string{"01-foundation/01-vision.md", "foundation/01-vision.md", "01-foundation/vision.md"} {
		if err := createChapter(t, root, p, "vision", "Vision"); err == nil {
			t.Errorf("CreateChapter(%s) succeeded", p)
		}
	}
}

// REQ-DOC-016 scenario: incomplete index
func TestCheckIndexComplete(t *testing.T) {
	// Purpose: unlisted chapters and missing entries are errors.
	root := writeDesign(t, map[string]string{
		"index.md":             indexHead + "## Method\n\n1. [Missing](method/missing.md)\n",
		"foundation/vision.md": chapterText("vision", "foundation"),
	})
	s, _ := Load(root)
	if n := countFindings(s.Check(), "REQ-DOC-016"); n != 2 {
		t.Errorf("%d REQ-DOC-016 errors, want 2: %+v", n, s.Check())
	}
}

// REQ-DOC-020 scenario: wrong part field
func TestCheckPartField(t *testing.T) {
	// Purpose: the part field must name the folder.
	root := writeDesign(t, map[string]string{
		"index.md":             indexHead + "## Foundation\n\n1. [Vision](foundation/vision.md)\n",
		"foundation/vision.md": chapterText("vision", "method"),
	})
	s, _ := Load(root)
	if n := countFindings(s.Check(), "REQ-DOC-020"); n != 1 {
		t.Errorf("%d REQ-DOC-020 errors, want 1", n)
	}
}

// REQ-DOC-014 scenario: order
func TestOrder(t *testing.T) {
	// Purpose: the reading order is the order of index.md.
	root := writeDesign(t, map[string]string{
		"index.md":             indexHead + "## Method\n\n1. [Phases](method/phases.md)\n\n## Foundation\n\n1. [Vision](foundation/vision.md)\n",
		"method/phases.md":     chapterText("phases", "method"),
		"foundation/vision.md": chapterText("vision", "foundation"),
	})
	s, _ := Load(root)
	want := []IndexEntry{{Part: "Method", Title: "Phases", Path: "design/method/phases.md"}, {Part: "Foundation", Title: "Vision", Path: "design/foundation/vision.md"}}
	if got := s.Order(); !reflect.DeepEqual(got, want) {
		t.Errorf("Order() = %+v, want %+v", got, want)
	}
}

// REQ-DOC-018 scenario: PDF template
func TestIndexTypTemplate(t *testing.T) {
	// Purpose: index.typ builds the PDF from index.md with cmarker.
	root := t.TempDir()
	if err := createChapter(t, root, "foundation/vision.md", "vision", "Vision"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "design/index.typ"))
	if err != nil {
		t.Fatal(err)
	}
	typ := string(b)
	for _, want := range []string{`read("index.md")`, "@preview/cmarker", "render-with-metadata", "bottom + left", "Sans", "font:"} {
		if !strings.Contains(typ, want) {
			t.Errorf("index.typ does not contain %q", want)
		}
	}
}
