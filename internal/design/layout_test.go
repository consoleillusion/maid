package design

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/style"
)

const indexHead = "---\nstatus: draft\nstyle: ste\nowner: paul\nsummary: Reading order.\nlinks: []\n---\n\n# Design\n\nThe design.\n\n"

func chapterText(id, part string) string {
	// Purpose: build a minimal valid chapter; id and part come from the path.
	_ = id
	_ = part
	return "---\nstatus: draft\nstyle: ste\nowner: paul\nsummary: S.\nlinks: []\n---\n\n# T\n\nText.\n"
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

// scenario: new chapters go into the index
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
	if want := "# Design\n\n- foundation/vision.md\n- foundation/glossary.md\n- method/phases.md\n"; string(b) != want {
		t.Fatalf("index.md = %q, want %q", b, want)
	}
	if _, err := os.Stat(filepath.Join(root, "design/index.typ")); err != nil {
		t.Error("design/index.typ was not created")
	}
	s, _ := Load(root)
	if p := s.Chapter("vision").Part; p != "foundation" {
		t.Errorf("part = %q, want foundation", p)
	}
	if n := countFindings(s.Check(), "index"); n != 0 {
		t.Errorf("%d index errors after creating chapters", n)
	}
}

// scenario: numbered path
func TestCreateChapterRejectsNumberedPath(t *testing.T) {
	// Purpose: part and chapter names carry no position numbers.
	root := t.TempDir()
	for _, p := range []string{"01-foundation/01-vision.md", "foundation/01-vision.md", "01-foundation/vision.md"} {
		if err := createChapter(t, root, p, "vision", "Vision"); err == nil {
			t.Errorf("CreateChapter(%s) succeeded", p)
		}
	}
}

// index scenario: incomplete index
func TestCheckIndexComplete(t *testing.T) {
	// Purpose: unlisted chapters and missing entries are errors.
	root := writeDesign(t, map[string]string{
		"index.md":             indexHead + "## Method\n\n1. [Missing](method/missing.md)\n",
		"foundation/vision.md": chapterText("vision", "foundation"),
	})
	s, _ := Load(root)
	if n := countFindings(s.Check(), "index"); n != 2 {
		t.Errorf("%d index errors, want 2: %+v", n, s.Check())
	}
}

// scenario: part comes from the folder
func TestPartComesFromFolder(t *testing.T) {
	// Purpose: the part is the folder that contains the chapter file.
	root := writeDesign(t, map[string]string{
		"index.md":             indexHead + "## Foundation\n\n1. [Vision](foundation/vision.md)\n",
		"foundation/vision.md": chapterText("vision", "method"),
	})
	s, _ := Load(root)
	if got := s.Chapter("vision").Part; got != "foundation" {
		t.Errorf("part = %q, want foundation", got)
	}
}

// scenario: order
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

// scenario: PDF template
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
	for _, want := range []string{`read("index.md")`, "@preview/cmarker", "@preview/merman", "render-with-metadata", "show-mermaid-blocks", `it.lang == "mermaid"`, "bottom + left", "Sans", "font:", "lower(title)", "lower(breadcrumb-part.get())"} {
		if !strings.Contains(typ, want) {
			t.Errorf("index.typ does not contain %q", want)
		}
	}
}
