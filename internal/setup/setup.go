// Package setup installs the repository files that maid owns in a project.
package setup

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// templates holds the files that maid writes into a new project.
//
//go:embed templates
var templates embed.FS

// File is one file that maid installs.
type File struct {
	Path string      // path relative to the project root
	Mode fs.FileMode // permission bits of a newly written file
	Data []byte
}

var specs = []struct {
	path     string
	template string
	mode     fs.FileMode
}{
	{".gitattributes", "gitattributes", 0o644},
	{"scripts/compile-pdf.sh", "compile-pdf.sh", 0o755},
	{".github/workflows/release-design-pdf.yml", "release-design-pdf.yml", 0o644},
}

// Files returns the files that maid installs, in install order.
func Files() ([]File, error) {
	// Purpose: read the embedded templates once for callers and tests.
	out := make([]File, 0, len(specs))
	for _, s := range specs {
		data, err := templates.ReadFile("templates/" + s.template)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: s.path, Mode: s.mode, Data: data})
	}
	return out, nil
}

// Ensure writes the files that are missing. It never overwrites a file that
// the project already has. It returns the paths that it wrote.
func Ensure(root string) ([]string, error) {
	// Purpose: give every installed project the same PDF and release assets.
	files, err := Files()
	if err != nil {
		return nil, err
	}
	var written []string
	for _, f := range files {
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if _, err := os.Lstat(p); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return written, err
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(p, f.Data, f.Mode); err != nil {
			return written, err
		}
		written = append(written, f.Path)
	}
	return written, nil
}
