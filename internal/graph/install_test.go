package graph

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consoleillusion/maid/internal/config"
)

func writeExec(t *testing.T, path, body string) {
	// Purpose: write an executable shell helper for a test.
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func testBin(t *testing.T) string {
	// Purpose: create an isolated bin directory for command lookup.
	t.Helper()
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mkdir", "chmod"} {
		p, err := exec.LookPath(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(p, filepath.Join(bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	return bin
}

func fakePython() string {
	// Purpose: return a stub python that creates a project-local virtualenv.
	return `#!/bin/sh
printf '%s\n' "$*" >> "$GRAPH_TEST_LOG"
if [ "$1" = "-m" ] && [ "$2" = "venv" ]; then
  venv="$3"
  mkdir -p "$venv/bin"
  for name in python3 pip code-review-graph; do
    printf '#!/bin/sh\nprintf "%%s\\n" "$*" >> "$GRAPH_TEST_LOG"\nexit 0\n' > "$venv/bin/$name"
    chmod +x "$venv/bin/$name"
  done
fi
exit 0
`
}

// design section: code-review-graph#existing-command
// scenario: Existing command is recorded
func TestEnsureUsesExistingCommand(t *testing.T) {
	// Purpose: maid records an installed command instead of installing another copy.
	root := t.TempDir()
	bin := testBin(t)
	log := filepath.Join(root, "commands.log")
	t.Setenv("GRAPH_TEST_LOG", log)
	want := filepath.Join(bin, "code-review-graph")
	writeExec(t, want, "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GRAPH_TEST_LOG\"\nexit 0\n")

	cfg := config.Default()
	if err := Ensure(context.Background(), root, &cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.CodeReviewGraph.Enabled || len(cfg.CodeReviewGraph.Command) == 0 || cfg.CodeReviewGraph.Command[0] != want {
		t.Fatalf("command = %+v, enabled = %v; want %q", cfg.CodeReviewGraph.Command, cfg.CodeReviewGraph.Enabled, want)
	}
	got, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CodeReviewGraph.Command) == 0 || got.CodeReviewGraph.Command[0] != want {
		t.Fatalf("saved command = %v; want %q", got.CodeReviewGraph.Command, want)
	}
}

// design section: code-review-graph#installed-by-init
// scenario: Missing command is installed
func TestEnsureInstallsIntoProjectEnvironment(t *testing.T) {
	// Purpose: maid installs code-review-graph in the project environment.
	root := t.TempDir()
	bin := testBin(t)
	log := filepath.Join(root, "commands.log")
	t.Setenv("GRAPH_TEST_LOG", log)
	writeExec(t, filepath.Join(bin, "python3"), fakePython())

	cfg := config.Default()
	if err := Ensure(context.Background(), root, &cfg); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(".maid", "crg-venv", "bin", "code-review-graph")
	if len(cfg.CodeReviewGraph.Command) == 0 || cfg.CodeReviewGraph.Command[0] != want {
		t.Fatalf("command = %v; want %q", cfg.CodeReviewGraph.Command, want)
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "-m venv") || !strings.Contains(text, "install code-review-graph") {
		t.Fatalf("install commands missing from log: %q", text)
	}
	if _, err := os.Stat(filepath.Join(root, want)); err != nil {
		t.Fatalf("installed command is missing: %v", err)
	}
}

// design section: code-review-graph#installed-by-init
// scenario: Missing command is installed
func TestEnsureBuildsGraph(t *testing.T) {
	// Purpose: init builds the graph after the command is available.
	root := t.TempDir()
	bin := testBin(t)
	log := filepath.Join(root, "commands.log")
	t.Setenv("GRAPH_TEST_LOG", log)
	writeExec(t, filepath.Join(bin, "code-review-graph"), "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GRAPH_TEST_LOG\"\nexit 0\n")

	cfg := config.Default()
	if err := Ensure(context.Background(), root, &cfg); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "build") || !strings.Contains(string(b), root) {
		t.Fatalf("build command missing from log: %q", b)
	}
}

// design section: code-review-graph#failed-install
// scenario: No installer is available
func TestEnsureReportsMissingPython(t *testing.T) {
	// Purpose: init reports an actionable error when no installer exists.
	root := t.TempDir()
	_ = testBin(t)

	err := Ensure(context.Background(), root, &config.Config{})
	if err == nil {
		t.Fatal("Ensure() succeeded without an installer")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "python") {
		t.Fatalf("error = %q; want a Python hint", err)
	}
}
