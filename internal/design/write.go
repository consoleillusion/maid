package design

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ErrAccepted means a direct write hit an accepted chapter (REQ-DOC-011).
var ErrAccepted = errors.New("chapter is accepted; write a proposal instead")

// ErrStale means the section changed after the writer read it (REQ-DOC-012).
var ErrStale = errors.New("section changed since you read it; read it again and retry")

// WriteResult is returned by every write.
type WriteResult struct {
	Path     string    `json:"path"`
	Ref      string    `json:"ref"`
	Version  string    `json:"version"`
	Findings []Finding `json:"findings"` // check findings for the changed chapter
}

// writable reports whether direct writes to ch are allowed (REQ-DOC-010, REQ-DOC-011).
func writable(ch *Chapter) error {
	// Purpose: block direct writes to accepted and superseded chapters.
	switch ch.FM.String("status") {
	case "accepted", "superseded":
		return ErrAccepted
	}
	return nil
}

// WriteSection replaces the text of one section, heading included.
// baseVersion must be the version that the writer read.
func (s *Store) WriteSection(ref, content, baseVersion string) (WriteResult, error) {
	// Purpose: change one section with an optimistic version check.
	ch, sec, err := s.locate(ref)
	if err != nil {
		return WriteResult{}, err
	}
	if err := writable(ch); err != nil {
		return WriteResult{}, err
	}
	if baseVersion != Version(sec.Text) {
		return WriteResult{}, fmt.Errorf("%w (current version %s)", ErrStale, Version(sec.Text))
	}
	content = strings.TrimRight(content, "\n")
	if !headingRe.MatchString(strings.SplitN(content, "\n", 2)[0]) {
		return WriteResult{}, errors.New("content must start with the section heading")
	}
	body := strings.Replace(ch.Body, sec.Text, content, 1)
	return s.save(ch, body, ref, Version(content))
}

// AppendSection adds content at the end of a chapter.
func (s *Store) AppendSection(chapterID, content string) (WriteResult, error) {
	// Purpose: add a new section to a chapter that is not accepted.
	ch, err := s.ResolveChapter(chapterID)
	if err != nil {
		return WriteResult{}, err
	}
	chapterID = ch.ID
	if err := writable(ch); err != nil {
		return WriteResult{}, err
	}
	content = strings.TrimRight(content, "\n")
	body := strings.TrimRight(ch.Body, "\n") + "\n\n" + content + "\n"
	return s.save(ch, body, chapterID, Version(content))
}

// Propose appends a proposal to a chapter, which is allowed even when the
// chapter is accepted (REQ-DOC-011).
func (s *Store) Propose(chapterID, title, content, agent string) (WriteResult, error) {
	// Purpose: record a proposed change without changing accepted text.
	ch, err := s.ResolveChapter(chapterID)
	if err != nil {
		return WriteResult{}, err
	}
	chapterID = ch.ID
	block := fmt.Sprintf("## Proposal: %s\n\nstatus: proposed · by %s · %s\n\n%s",
		strings.TrimSpace(title), agent, time.Now().UTC().Format("2006-01-02"), strings.TrimSpace(content))
	body := strings.TrimRight(ch.Body, "\n") + "\n\n" + block + "\n"
	return s.save(ch, body, chapterID+"#"+slug("Proposal: "+title), Version(block))
}

// statusLine matches the status line of a frontmatter block.
var statusLine = regexp.MustCompile(`(?m)^status:[^\n]*$`)

// SetStatus changes the status of a chapter. approvedBy names the person who
// approved it; it may be empty when an agent accepts on the user's behalf (REQ-PH-007).
func (s *Store) SetStatus(chapterID, status, approvedBy string) (WriteResult, error) {
	// Purpose: record a status change, such as a human or delegated acceptance.
	ch, err := s.ResolveChapter(chapterID)
	if err != nil {
		return WriteResult{}, err
	}
	chapterID = ch.ID
	if !Statuses[status] {
		return WriteResult{}, fmt.Errorf("invalid status %q", status)
	}
	if status == "accepted" {
		var errs []string
		for _, f := range s.Check() {
			if f.Path == ch.Path && f.Severity == "error" {
				errs = append(errs, fmt.Sprintf("line %d: %s", f.Line, f.Message))
			}
		}
		if len(errs) > 0 {
			return WriteResult{}, fmt.Errorf("fix the check errors before accepting:\n- %s", strings.Join(errs, "\n- "))
		}
	}
	raw, err := os.ReadFile(filepath.Join(s.Root, ch.Path))
	if err != nil {
		return WriteResult{}, err
	}
	text := string(raw)
	end := strings.Index(text[4:], "\n---") + 4
	fm := text[:end]
	if !statusLine.MatchString(fm) {
		return WriteResult{}, errors.New("frontmatter has no status line")
	}
	text = statusLine.ReplaceAllString(fm, "status: "+status) + text[end:]
	if err := atomicWrite(filepath.Join(s.Root, ch.Path), text); err != nil {
		return WriteResult{}, err
	}
	return s.result(ch.Path, chapterID, Version(text))
}

// ChapterSpec is the input of CreateChapter.
type ChapterSpec struct {
	Path    string   `json:"path"` // relative to design/, e.g. 01-foundation/01-vision.md
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Style   string   `json:"style"`
	Owner   string   `json:"owner"`
	Summary string   `json:"summary"`
	Links   []string `json:"links"`
	Body    string   `json:"body"`
}

var safePath = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)*\.md$`)

// CreateChapter writes a new chapter with status draft.
func (s *Store) CreateChapter(c ChapterSpec) (WriteResult, error) {
	// Purpose: start a new chapter with valid frontmatter (REQ-DOC-001, REQ-DOC-002).
	c.Path = path.Clean(strings.TrimPrefix(c.Path, "design/"))
	if !safePath.MatchString(c.Path) || strings.Contains(c.Path, "..") {
		return WriteResult{}, fmt.Errorf("invalid chapter path %q: use part/NN-name.md in lowercase", c.Path)
	}
	if c.ID == "" || c.Title == "" || c.Style == "" || c.Summary == "" {
		return WriteResult{}, errors.New("id, title, style and summary are required")
	}
	if s.byID[c.ID] != nil {
		return WriteResult{}, fmt.Errorf("chapter id %q exists", c.ID)
	}
	full := filepath.Join(s.Root, Dir, filepath.FromSlash(c.Path))
	if _, err := os.Stat(full); err == nil {
		return WriteResult{}, fmt.Errorf("%s exists", c.Path)
	}
	part := path.Dir(c.Path)
	if part == "." {
		part = "design"
	}
	owner := c.Owner
	if owner == "" {
		owner = "user"
	}
	links := "[]"
	if len(c.Links) > 0 {
		links = "[" + strings.Join(c.Links, ", ") + "]"
	}
	text := fmt.Sprintf("---\nid: %s\ntitle: %s\npart: %s\nstatus: draft\nstyle: %s\nowner: %s\nsummary: >\n  %s\nlinks: %s\n---\n\n%s\n",
		c.ID, quoteIfNeeded(c.Title), part, c.Style, owner, strings.ReplaceAll(strings.TrimSpace(c.Summary), "\n", "\n  "),
		links, strings.TrimRight(c.Body, "\n"))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return WriteResult{}, err
	}
	if err := atomicWrite(full, text); err != nil {
		return WriteResult{}, err
	}
	return s.result(path.Join(Dir, c.Path), c.ID, Version(text))
}

// AddQuestion appends a row to the questions table and returns its ID.
func (s *Store) AddQuestion(question, blocks, owner string) (string, WriteResult, error) {
	// Purpose: record missing information instead of a guess (REQ-DOC-013).
	ch := s.byID["questions"]
	if ch == nil {
		return "", WriteResult{}, errors.New("design has no questions chapter (id: questions)")
	}
	max := 0
	for id := range tableItems(ch) {
		var n int
		if _, err := fmt.Sscanf(id, "Q-%d", &n); err == nil && n > max {
			max = n
		}
	}
	id := fmt.Sprintf("Q-%03d", max+1)
	if owner == "" {
		owner = "user"
	}
	row := fmt.Sprintf("| %s | %s | | %s | open |", id, cell(question), cell(blocks))
	lines := strings.Split(strings.TrimRight(ch.Body, "\n"), "\n")
	last := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "|") {
			last = i
		}
	}
	if last < 0 {
		lines = append(lines, "", "| ID | Question | Answer | Updated items | Status |", "|---|---|---|---|---|", row)
	} else {
		lines = append(lines[:last+1], append([]string{row}, lines[last+1:]...)...)
	}
	res, err := s.save(ch, strings.Join(lines, "\n")+"\n", id, Version(row))
	return id, res, err
}

func cell(s string) string {
	// Purpose: keep text safe inside a Markdown table cell.
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "/"), "\n", " ")
}

func quoteIfNeeded(s string) string {
	// Purpose: quote a YAML scalar that contains a colon or a hash.
	if strings.ContainsAny(s, ":#") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func (s *Store) locate(ref string) (*Chapter, *Section, error) {
	// Purpose: find the chapter and section that a ref names.
	if canonical, err := s.Resolve(ref); err == nil {
		ref = canonical
	} else {
		return nil, nil, err
	}
	if it, ok := s.items[ref]; ok && it.Section != nil {
		return it.Chapter, it.Section, nil
	}
	chID, anchor, found := strings.Cut(ref, "#")
	if found {
		if ch := s.byID[chID]; ch != nil {
			if sec := ch.Section(anchor); sec != nil {
				return ch, sec, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("unknown section %q: use an item ID or chapter#anchor", ref)
}

func (s *Store) save(ch *Chapter, body, ref, version string) (WriteResult, error) {
	// Purpose: write a chapter back with its original frontmatter.
	full := filepath.Join(s.Root, ch.Path)
	raw, err := os.ReadFile(full)
	if err != nil {
		return WriteResult{}, err
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	end := strings.Index(text[4:], "\n---\n")
	if !strings.HasPrefix(text, "---\n") || end < 0 {
		return WriteResult{}, errors.New("chapter has no frontmatter")
	}
	head := text[:end+4+5]
	if err := atomicWrite(full, head+body); err != nil {
		return WriteResult{}, err
	}
	return s.result(ch.Path, ref, version)
}

func (s *Store) result(relPath, ref, version string) (WriteResult, error) {
	// Purpose: reload the design and return the findings for the changed file.
	fresh, err := Load(s.Root)
	if err != nil {
		return WriteResult{}, err
	}
	*s = *fresh
	res := WriteResult{Path: relPath, Ref: ref, Version: version, Findings: []Finding{}}
	for _, f := range s.Check() {
		if f.Path == relPath {
			res.Findings = append(res.Findings, f)
		}
	}
	return res, nil
}

func atomicWrite(p, text string) error {
	// Purpose: replace a file through a temporary file.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(text), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
