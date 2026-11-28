package design

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// REQ-DOC-010 REQ-DOC-012
func TestWriteSectionVersionCheck(t *testing.T) {
	// Purpose: a write with the read version succeeds; a stale version fails.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	ex, _ := s.Read([]string{"REQ-AL-001"}, false)
	res, err := s.WriteSection("REQ-AL-001", "### REQ-AL-001 One\nThe design store shall load every chapter.", ex[0].Version)
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
	if _, err := s.WriteSection("REQ-AL-001", "### REQ-AL-001 One\nThe design store shall x.", ex[0].Version); !errors.Is(err, ErrStale) {
		t.Fatalf("stale write: err = %v", err)
	}
}

// REQ-DOC-011
func TestAcceptedChapterNeedsProposal(t *testing.T) {
	// Purpose: accepted chapters reject direct writes but take proposals.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	if _, err := s.AppendSection("beta", "## New\nText."); !errors.Is(err, ErrAccepted) {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.Propose("beta", "Add a rule", "Add a rule because ...", "agent-1"); err != nil {
		t.Fatal(err)
	}
	if s.Chapter("beta").Section("proposal-add-a-rule") == nil {
		t.Fatal("proposal section missing")
	}
}

// REQ-PH-007
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

// REQ-DOC-001 REQ-DOC-002 REQ-DOC-013
func TestCreateChapterAndQuestion(t *testing.T) {
	// Purpose: new chapters pass the checks, and questions get the next ID.
	root := writeDesign(t, map[string]string{"01-a/alpha.md": goodChapter, "01-a/beta.md": betaChapter})
	s, _ := Load(root)
	res, err := s.CreateChapter(ChapterSpec{Path: "foundation/vision.md", ID: "vision", Title: "Vision", Style: "ste",
		Summary: "Why the project exists.", Body: "# Vision\n\n### G-001 Goal\nThe tool adds numbers.\n"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings: %+v", res.Findings)
	}
	if _, err := s.CreateChapter(ChapterSpec{Path: "../x.md", ID: "x", Title: "x", Style: "ste", Summary: "x"}); err == nil {
		t.Fatal("path outside design accepted")
	}
	if _, err := s.CreateChapter(ChapterSpec{Path: "questions.md", ID: "questions", Title: "Questions", Style: "ste", Summary: "Open questions.", Body: "# Questions\n"}); err != nil {
		t.Fatal(err)
	}
	id, _, err := s.AddQuestion("How big | can it be?", "vision#G-001", "")
	if err != nil || id != "Q-001" {
		t.Fatalf("id=%s err=%v", id, err)
	}
	if id, _, _ = s.AddQuestion("Second?", "", ""); id != "Q-002" {
		t.Fatalf("second id = %s", id)
	}
	if _, ok := s.Item("Q-002"); !ok {
		t.Fatal("Q-002 not indexed")
	}
}
