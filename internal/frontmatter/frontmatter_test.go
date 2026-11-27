package frontmatter

import (
	"reflect"
	"testing"
)

// REQ-DOC-002
func TestSplitAndParse(t *testing.T) {
	// Purpose: parse every value form that design files use.
	src := "---\nid: vision\ntitle: \"Style card: STE\"\nstatus: draft # comment\nsummary: >\n  one\n  two\nlinks: [a, \"b c\", ADR-0001]\nrevisit-when: []\nprofile: null\nitems:\n  - x\n  - 'y'\n---\n# Body\n"
	fm, body, line, err := Split(src)
	if err != nil {
		t.Fatal(err)
	}
	if body != "# Body\n" || line != 15 {
		t.Fatalf("body=%q line=%d", body, line)
	}
	m, err := Parse(fm)
	if err != nil {
		t.Fatal(err)
	}
	want := Map{
		"id": "vision", "title": "Style card: STE", "status": "draft",
		"summary": "one two", "links": []string{"a", "b c", "ADR-0001"},
		"revisit-when": []string{}, "profile": nil, "items": []string{"x", "y"},
	}
	if !reflect.DeepEqual(m, want) {
		t.Fatalf("got %#v", m)
	}
}

func TestNoFrontmatter(t *testing.T) {
	// Purpose: report a missing frontmatter block.
	if _, _, _, err := Split("# title\n"); err != ErrNoFrontmatter {
		t.Fatalf("err = %v", err)
	}
}

func TestListAccessor(t *testing.T) {
	// Purpose: List treats a scalar as a one-item list and nil as empty.
	m := Map{"a": "x", "b": nil}
	if got := m.List("a"); !reflect.DeepEqual(got, []string{"x"}) {
		t.Fatalf("a = %v", got)
	}
	if got := m.List("b"); got != nil {
		t.Fatalf("b = %v", got)
	}
}
