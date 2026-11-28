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

// ErrStale means the section changed after the writer read it.
var ErrStale = errors.New("section changed since you read it; read it again and retry")

// WriteResult is returned by every write.
type WriteResult struct {
	Path     string    `json:"path"`
	Ref      string    `json:"ref"`
	Version  string    `json:"version"`
	Findings []Finding `json:"findings"`       // check findings for the changed chapter
	Also     []string  `json:"also,omitempty"` // other files the write changed, such as design/index.md
}

// WriteSection replaces the text of one section, heading included.
// baseVersion must be the version that the writer read.
func (s *Store) WriteSection(ref, content, baseVersion string) (WriteResult, error) {
	// Purpose: change one section with an optimistic version check.
	ch, sec, err := s.locate(ref)
	if err != nil {
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
	// Purpose: add a new section to any chapter, including an accepted one.
	ch, err := s.ResolveChapter(chapterID)
	if err != nil {
		return WriteResult{}, err
	}
	chapterID = ch.ID
	content = strings.TrimRight(content, "\n")
	body := strings.TrimRight(ch.Body, "\n") + "\n\n" + content + "\n"
	return s.save(ch, body, chapterID, Version(content))
}

// Propose appends a proposal section to a chapter. It is optional: agents can
// change any chapter directly with WriteSection or AppendSection.
func (s *Store) Propose(chapterID, title, content, agent string) (WriteResult, error) {
	// Purpose: record an attributed change suggestion in a chapter.
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

// SetStatus changes the status of a chapter.
func (s *Store) SetStatus(chapterID, status string) (WriteResult, error) {
	// Purpose: record a status change.
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
	Path    string   `json:"path"`            // relative to design/, e.g. foundation/vision.md
	ID      string   `json:"id,omitempty"`    // ignored; derived from Path
	Title   string   `json:"title,omitempty"` // ignored; derived from Body
	Style   string   `json:"style"`
	Owner   string   `json:"owner"`
	Summary string   `json:"summary"`
	Links   []string `json:"links"`
	Body    string   `json:"body"`
}

// chapterPath is part/name.md; names start with a letter and carry no position numbers.
var chapterPath = regexp.MustCompile(`^([a-z][a-z0-9-]*/)?[a-z][a-z0-9._-]*\.md$`)

// CreateChapter writes a new chapter with status draft.
func (s *Store) CreateChapter(c ChapterSpec) (WriteResult, error) {
	// Purpose: start a new chapter with values derived from path and body.
	c.Path = path.Clean(strings.TrimPrefix(c.Path, "design/"))
	if !chapterPath.MatchString(c.Path) || strings.Contains(c.Path, "..") {
		return WriteResult{}, fmt.Errorf("invalid chapter path %q: use part/name.md, lowercase words joined by hyphens, with no numbers in front", c.Path)
	}
	if c.Style == "" || c.Summary == "" {
		return WriteResult{}, errors.New("style and summary are required")
	}
	id := chapterID(path.Join(Dir, c.Path))
	if s.byID[id] != nil {
		return WriteResult{}, fmt.Errorf("chapter id %q exists", id)
	}
	full := filepath.Join(s.Root, Dir, filepath.FromSlash(c.Path))
	if _, err := os.Stat(full); err == nil {
		return WriteResult{}, fmt.Errorf("%s exists", c.Path)
	}
	owner := c.Owner
	if owner == "" {
		owner = "user"
	}
	links := "[]"
	if len(c.Links) > 0 {
		links = "[" + strings.Join(c.Links, ", ") + "]"
	}
	text := fmt.Sprintf("---\nstatus: draft\nstyle: %s\nowner: %s\nsummary: >\n  %s\nlinks: %s\n---\n\n%s\n",
		c.Style, owner, strings.ReplaceAll(strings.TrimSpace(c.Summary), "\n", "\n  "),
		links, strings.TrimRight(c.Body, "\n"))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return WriteResult{}, err
	}
	if err := atomicWrite(full, text); err != nil {
		return WriteResult{}, err
	}
	var also []string
	if PartOf(path.Join(Dir, c.Path)) != "" {
		var err error
		if also, err = s.addToIndex(c.Path); err != nil {
			return WriteResult{}, err
		}
	}
	res, err := s.result(path.Join(Dir, c.Path), id, Version(text))
	res.Also = also
	return res, err
}

func (s *Store) locate(ref string) (*Chapter, *Section, error) {
	// Purpose: find the chapter and section that a ref names.
	if canonical, err := s.Resolve(ref); err == nil {
		ref = canonical
	} else {
		return nil, nil, err
	}
	chID, anchor, found := strings.Cut(ref, "#")
	if found {
		if ch := s.byID[chID]; ch != nil {
			if sec := ch.Section(anchor); sec != nil {
				return ch, sec, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("unknown section %q: use chapter#anchor", ref)
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
