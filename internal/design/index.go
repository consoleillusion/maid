package design

// IndexEntry is one chapter in the reading order of design/index.md.
type IndexEntry struct {
	Part  string // the level-2 heading above the entry
	Title string // the link text
	Path  string // relative to the project root, for example design/foundation/vision.md
}

// Order returns the chapters in the order that design/index.md lists them
// (REQ-DOC-014).
func (s *Store) Order() []IndexEntry {
	// Purpose: stub; the code work item implements it.
	return nil
}
