package design

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSectionVersionCheck(t *testing.T) {
	// Purpose: a write with the read version succeeds; a stale version fails.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	ex, _ := s.Read([]string{"alpha#one"}, false)
	res, err := s.WriteSection("alpha#one", "### One\nThe design store shall load every chapter.", ex[0].Version)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings: %+v", res.Findings)
	}
	b, _ := os.ReadFile(filepath.Join(root, "design/01-a/alpha.md"))
	if !strings.Contains(string(b), "load every chapter") || !strings.HasPrefix(string(b), "---\nstatus: draft\n") {
		t.Fatalf("file = %s", b)
	}
	if _, err := s.WriteSection("alpha#one", "### One\nThe design store shall x.", ex[0].Version); !errors.Is(err, ErrStale) {
		t.Fatalf("stale write: err = %v", err)
	}
}

func TestAcceptedChapterAcceptsDirectWrites(t *testing.T) {
	// Purpose: accepted chapters take direct writes; proposals stay optional.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	if _, err := s.AppendSection("beta", "## New\nText."); err != nil {
		t.Fatalf("append: %v", err)
	}
	if s.Chapter("beta").Section("new") == nil {
		t.Fatal("section missing")
	}
	if _, err := s.Propose("beta", "Add a rule", "Add a rule because ...", "agent-1"); err != nil {
		t.Fatal(err)
	}
	if s.Chapter("beta").Section("proposal-add-a-rule") == nil {
		t.Fatal("proposal section missing")
	}
}

func TestSetStatus(t *testing.T) {
	// Purpose: a clean chapter can be accepted.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	if _, err := s.SetStatus("alpha", "accepted"); err != nil {
		t.Fatal(err)
	}
	if s.Chapter("alpha").FM.String("status") != "accepted" {
		t.Fatalf("status = %s", s.Chapter("alpha").FM.String("status"))
	}
}

func TestCreateChapter(t *testing.T) {
	// Purpose: a new chapter passes the checks and invalid paths are refused.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	res, err := s.CreateChapter(ChapterSpec{Path: "foundation/vision.md", Style: "ste",
		Summary: "Why the project exists.", Body: "# Vision\n\n### Goal\nThe tool is useful.\n"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings: %+v", res.Findings)
	}
	if _, err := s.CreateChapter(ChapterSpec{Path: "../x.md", Style: "ste", Summary: "x"}); err == nil {
		t.Fatal("path outside design accepted")
	}
}
