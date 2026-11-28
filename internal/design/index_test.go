package design

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// scenario: bulleted index on create
func TestIndexBullets(t *testing.T) {
	// Purpose: index.md holds the title and bulleted paths, grouped by part folder.
	root := t.TempDir()
	for _, c := range [][3]string{{"foundation/vision.md", "vision", "Vision"}, {"method/phases.md", "phases", "Phases"}, {"foundation/glossary.md", "glossary", "Glossary"}} {
		if err := createChapter(t, root, c[0], c[1], c[2]); err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "design/index.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := "# Design\n\n- foundation/vision.md\n- foundation/glossary.md\n- method/phases.md\n"
	if string(b) != want {
		t.Fatalf("index.md = %q, want %q", b, want)
	}
}

// scenario: index is not a chapter
func TestIndexNotLoaded(t *testing.T) {
	// Purpose: index.md has no frontmatter and is not checked as a chapter.
	root := writeDesign(t, map[string]string{
		"index.md":             "# D\n\n- foundation/vision.md\n",
		"foundation/vision.md": chapterText("vision", "foundation"),
	})
	s, _ := Load(root)
	for _, c := range s.Chapters {
		if c.Path == "design/index.md" {
			t.Error("index.md is loaded as a chapter")
		}
	}
	for _, f := range s.Check() {
		if f.Path == "design/index.md" {
			t.Errorf("finding for index.md: %+v", f)
		}
	}
}

// scenario: bulleted paths give the order
func TestOrderBullets(t *testing.T) {
	// Purpose: parts come from folders and titles from the chapters.
	root := writeDesign(t, map[string]string{
		"index.md":             "# D\n\n- method/phases.md\n- foundation/vision.md\n",
		"method/phases.md":     chapterText("phases", "method"),
		"foundation/vision.md": chapterText("vision", "foundation"),
	})
	s, _ := Load(root)
	want := []IndexEntry{{Part: "Method", Title: "T", Path: "design/method/phases.md"}, {Part: "Foundation", Title: "T", Path: "design/foundation/vision.md"}}
	if got := s.Order(); !reflect.DeepEqual(got, want) {
		t.Errorf("Order() = %+v, want %+v", got, want)
	}
}

// heading-numbered scenario: numbered heading
func TestCheckNumberedHeading(t *testing.T) {
	// Purpose: a heading that starts with a position number is an error.
	ch := strings.Replace(chapterText("vision", "foundation"), "# T\n", "# T\n\n## 1. Scope\n\nText.\n", 1)
	root := writeDesign(t, map[string]string{"index.md": "# D\n\n- foundation/vision.md\n", "foundation/vision.md": ch})
	s, _ := Load(root)
	if n := countFindings(s.Check(), "heading-numbered"); n != 1 {
		t.Errorf("%d heading-numbered errors, want 1", n)
	}
}

// scenario: only parts are numbered, in roman numerals
func TestIndexTypRomanParts(t *testing.T) {
	// Purpose: the PDF numbers parts with roman numerals and nothing else.
	root := t.TempDir()
	if err := createChapter(t, root, "foundation/vision.md", "vision", "Vision"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "design/index.typ"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#set heading(numbering: none)", `numbering: "I"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("index.typ does not contain %q", want)
		}
	}
}

// scenario: the PDF template supports both modes, a cover image and a lowercase footer
func TestIndexTypModes(t *testing.T) {
	// Purpose: the installed template builds a light and a dark book.
	root := t.TempDir()
	written, err := EnsureIndexFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"design/index.md", "design/index.typ"}
	if !reflect.DeepEqual(written, want) {
		t.Fatalf("EnsureIndexFiles() wrote %v, want %v", written, want)
	}
	b, err := os.ReadFile(filepath.Join(root, "design/index.typ"))
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{`sys.inputs.at("mode"`, `rgb("#222222")`, `cover_image.png`, "bottom + left", "width: 100%", "height: auto", `dx: 2.3cm`, `dy: -2.5cm`, "lower(title)", "lower(breadcrumb-part.get())", "lower(breadcrumb-chapter.get())", "leading: 0.85em", "spacing: 1.4em", "v(1.6em)", "v(0.8em)", "spacing: 1em"} {
		if !strings.Contains(string(b), part) {
			t.Errorf("index.typ does not contain %q", part)
		}
	}
	if again, err := EnsureIndexFiles(root); err != nil || len(again) != 0 {
		t.Fatalf("second EnsureIndexFiles() = %v, %v; want no writes", again, err)
	}
}

// heading-numbered scenario: maid's own design has no position numbers
func TestOwnIndexAndHeadings(t *testing.T) {
	// Purpose: maid's index.md is bulleted paths, and no heading in its design is numbered.
	root := filepath.Join("..", "..", Dir)
	b, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	idx := string(b)
	if strings.HasPrefix(idx, "---") || !strings.Contains(idx, "\n- foundation/vision.md\n") || regexp.MustCompile(`(?m)^\d+\.`).MatchString(idx) {
		t.Errorf("design/index.md is not a title and bulleted paths:\n%s", idx)
	}
	numbered := regexp.MustCompile(`^#{1,6}\s+\d`)
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		c, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fence := false
		for i, l := range strings.Split(string(c), "\n") {
			if strings.HasPrefix(l, "```") {
				fence = !fence
			}
			if !fence && numbered.MatchString(l) {
				t.Errorf("%s:%d numbered heading: %s", p, i+1, l)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
