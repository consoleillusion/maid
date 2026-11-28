package design

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// NotFound describes a ref that did not resolve, with suggestions.
type NotFound struct {
	Ref        string   `json:"ref"`
	DidYouMean []string `json:"did_you_mean,omitempty"`
}

func (n NotFound) Error() string {
	// Purpose: format a NotFound as an error message with suggestions.
	msg := fmt.Sprintf("nothing in the design matches %q", n.Ref)
	if len(n.DidYouMean) > 0 {
		msg += "; did you mean: " + strings.Join(n.DidYouMean, ", ")
	}
	return msg + ". Call design_index to list chapters."
}

// Resolve turns what an agent typed into a canonical ref: a chapter ID or a
// chapter#anchor. It accepts IDs in any letter case, file paths (with or
// without design/), chapter titles, and heading text after #.
func (s *Store) Resolve(ref string) (string, error) {
	// Purpose: accept every reasonable way to name a chapter or section.
	raw := strings.TrimSpace(ref)
	if raw == "" {
		return "", NotFound{Ref: ref}
	}
	base, anchor, hasAnchor := strings.Cut(raw, "#")
	ch := s.findChapter(base)
	if ch == nil {
		return "", NotFound{Ref: ref, DidYouMean: s.suggest(base)}
	}
	if !hasAnchor || strings.TrimSpace(anchor) == "" {
		return ch.ID, nil
	}
	anchor = strings.TrimSpace(anchor)
	for _, cand := range []string{anchor, slug(anchor)} {
		if ch.Section(cand) != nil {
			return ch.ID + "#" + cand, nil
		}
	}
	var anchors []string
	for _, sec := range ch.Sections {
		anchors = append(anchors, ch.ID+"#"+sec.Anchor)
	}
	return "", NotFound{Ref: ref, DidYouMean: closest(anchor, anchors, 5)}
}

// ResolveChapter resolves ref and requires it to name a chapter.
func (s *Store) ResolveChapter(ref string) (*Chapter, error) {
	// Purpose: find the chapter for tools that act on whole chapters.
	id, err := s.Resolve(ref)
	if err != nil {
		return nil, err
	}
	if ch := s.byID[id]; ch != nil {
		return ch, nil
	}
	chID, _, _ := strings.Cut(id, "#")
	return nil, fmt.Errorf("%s is a section; pass the chapter ID %q", id, chID)
}

func (s *Store) findChapter(q string) *Chapter {
	// Purpose: match a chapter by ID, path or title, ignoring case.
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	if ch := s.byID[q]; ch != nil {
		return ch
	}
	lq := strings.ToLower(q)
	p := strings.TrimPrefix(strings.TrimPrefix(path.Clean(strings.ReplaceAll(lq, "\\", "/")), "./"), "/")
	for _, ch := range s.Chapters {
		lp := strings.ToLower(ch.Path)
		rel := strings.TrimPrefix(lp, Dir+"/")
		switch {
		case strings.ToLower(ch.ID) == lq, lp == p, rel == p,
			strings.TrimSuffix(rel, ".md") == p, strings.HasSuffix(lp, "/"+p) && strings.Contains(p, "/"),
			strings.ToLower(ch.Title) == lq:
			if ch.ID != "" {
				return ch
			}
		}
	}
	return nil
}

func (s *Store) suggest(q string) []string {
	// Purpose: list the chapter IDs closest to an unknown ref.
	var all []string
	for id := range s.byID {
		all = append(all, id)
	}
	return closest(q, all, 5)
}

func closest(q string, cands []string, n int) []string {
	// Purpose: rank candidates by substring match, then by edit distance.
	lq := strings.ToLower(q)
	if i := strings.LastIndex(lq, "/"); i >= 0 {
		lq = lq[i+1:]
	}
	lq = strings.TrimSuffix(lq, ".md")
	type scored struct {
		s string
		d int
	}
	var out []scored
	for _, c := range cands {
		lc := strings.ToLower(c)
		d := distance(lq, lc)
		if strings.Contains(lc, lq) || strings.Contains(lq, lc) {
			d = 0
		}
		if d <= max(3, len(lq)/2) {
			out = append(out, scored{c, d})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].d != out[j].d {
			return out[i].d < out[j].d
		}
		return out[i].s < out[j].s
	})
	var res []string
	for i := 0; i < len(out) && i < n; i++ {
		res = append(res, out[i].s)
	}
	return res
}

func distance(a, b string) int {
	// Purpose: compute the Levenshtein distance between two strings.
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
