package metrics

import (
	"reflect"
	"testing"
)

// design section: scoring#language-independent-measures
// scenario: package.json
func TestPackageJSONDeps(t *testing.T) {
	// Purpose: dependencies are runtime and devDependencies are dev.
	oldFile := []byte(`{"dependencies":{"left-pad":"1.0.0"}}`)
	newFile := []byte(`{"dependencies":{"left-pad":"1.0.0","lodash":"4.0.0"},"devDependencies":{"vitest":"1.0.0"}}`)
	runtime, dev, added, err := ManifestDeps("package.json", oldFile, newFile)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 1 || dev != 1 || !reflect.DeepEqual(added, []string{"lodash", "vitest"}) {
		t.Errorf("package.json deps = %d, %d, %v; want 1, 1, [lodash vitest]", runtime, dev, added)
	}
}

// design section: scoring#language-independent-measures
// scenario: Cargo.toml
func TestCargoDeps(t *testing.T) {
	// Purpose: dependencies and dev-dependencies are counted separately.
	newFile := []byte(`[package]
name = "x"

[dependencies]
serde = "1"

[dev-dependencies]
criterion = "0.5"
`)
	runtime, dev, _, err := ManifestDeps("Cargo.toml", nil, newFile)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 1 || dev != 1 {
		t.Errorf("Cargo.toml deps = %d, %d; want 1, 1", runtime, dev)
	}
}

// design section: scoring#language-independent-measures
// scenario: requirements.txt
func TestRequirementsDeps(t *testing.T) {
	// Purpose: comments and include lines are ignored; new packages are runtime.
	oldFile := []byte("requests==2.0\n")
	newFile := []byte("# comment\nrequests==2.0\nflask>=3\n-r other.txt\n")
	runtime, _, added, err := ManifestDeps("requirements.txt", oldFile, newFile)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 1 || !reflect.DeepEqual(added, []string{"flask"}) {
		t.Errorf("requirements deps = %d, %v; want 1, [flask]", runtime, added)
	}
}

// design section: scoring#language-independent-measures
// scenario: pyproject.toml
func TestPyprojectDeps(t *testing.T) {
	// Purpose: project dependencies are runtime and optional groups are dev.
	newFile := []byte(`[project]
dependencies = ["httpx"]

[project.optional-dependencies]
dev = ["pytest"]
`)
	runtime, dev, _, err := ManifestDeps("pyproject.toml", nil, newFile)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 1 || dev != 1 {
		t.Errorf("pyproject deps = %d, %d; want 1, 1", runtime, dev)
	}
}

// design section: scoring#language-independent-measures
// scenario: unknown file
func TestUnknownManifest(t *testing.T) {
	// Purpose: an unsupported file name returns zero changes and no error.
	runtime, dev, added, err := ManifestDeps("build.gradle", nil, []byte("dependencies {}"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 0 || dev != 0 || len(added) != 0 {
		t.Errorf("unknown manifest = %d, %d, %v; want zeros", runtime, dev, added)
	}
}
