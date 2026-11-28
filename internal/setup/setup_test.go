package setup

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// scenario: init installs the missing project assets
func TestEnsureWritesMissingFiles(t *testing.T) {
	// Purpose: a new project gets every asset, once.
	root := t.TempDir()
	want := []string{".gitattributes", "scripts/compile-pdf.sh", ".github/workflows/release-design-pdf.yml"}
	written, err := Ensure(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(written, want) {
		t.Fatalf("Ensure() wrote %v, want %v", written, want)
	}
	files, err := Files()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, f.Data) {
			t.Errorf("%s does not match the embedded template", f.Path)
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != f.Mode.Perm() {
			t.Errorf("%s mode = %v, want %v", f.Path, info.Mode().Perm(), f.Mode.Perm())
		}
	}

	// An existing file belongs to the project and is not overwritten.
	custom := "# local attributes\n"
	if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if again, err := Ensure(root); err != nil || len(again) != 0 {
		t.Fatalf("second Ensure() = %v, %v; want no writes", again, err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	if err != nil || string(b) != custom {
		t.Fatalf(".gitattributes was overwritten: %q, %v", b, err)
	}
}

// scenario: embedded assets match the repository assets
func TestTemplatesMatchRepository(t *testing.T) {
	// Purpose: maid installs the same assets that its own repository uses.
	files, err := Files()
	if err != nil {
		t.Fatal(err)
	}
	repo := map[string]string{
		".gitattributes":                           filepath.Join("..", "..", ".gitattributes"),
		"scripts/compile-pdf.sh":                   filepath.Join("..", "..", "scripts", "compile-pdf.sh"),
		".github/workflows/release-design-pdf.yml": filepath.Join("..", "..", ".github", "workflows", "release-design-pdf.yml"),
	}
	for _, f := range files {
		b, err := os.ReadFile(repo[f.Path])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, f.Data) {
			t.Errorf("%s differs from its embedded template", f.Path)
		}
	}
}
