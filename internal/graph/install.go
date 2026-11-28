package graph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/consoleillusion/maid/internal/config"
)

const packageName = "code-review-graph"

// Ensure installs code-review-graph in a project and records its command.
func Ensure(ctx context.Context, root string, cfg *config.Config) error {
	// Purpose: install and record the code-review-graph command.
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return errors.New("missing configuration")
	}
	if len(cfg.CodeReviewGraph.Command) == 0 {
		cfg.CodeReviewGraph.Command = []string{packageName}
	}
	if len(cfg.CodeReviewGraph.Args) == 0 {
		cfg.CodeReviewGraph.Args = []string{"serve"}
	}
	command, found, err := resolveCommand(cfg.CodeReviewGraph.Command)
	if err != nil {
		return err
	}
	if found {
		cfg.CodeReviewGraph.Command = []string{command}
	} else {
		command, err = installInProject(ctx, root)
		if err != nil {
			return err
		}
		cfg.CodeReviewGraph.Command = []string{command}
	}
	cfg.CodeReviewGraph.Enabled = true
	if err := config.Write(root, *cfg); err != nil {
		return fmt.Errorf("write %s: %w", config.File, err)
	}
	return buildGraph(ctx, command, root)
}

func resolveCommand(command []string) (string, bool, error) {
	// Purpose: find the configured code-review-graph executable.
	name := command[0]
	if strings.ContainsRune(name, os.PathSeparator) {
		info, err := os.Stat(name)
		if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
			return "", false, nil
		}
		return name, true, nil
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false, nil
	}
	return path, true, nil
}

func installInProject(ctx context.Context, root string) (string, error) {
	// Purpose: install code-review-graph in a project-local Python environment.
	python, err := exec.LookPath("python3")
	if err != nil {
		return "", fmt.Errorf("%s is not installed and Python 3 is unavailable; install Python 3 or add %s to PATH", packageName, packageName)
	}
	venv := filepath.Join(root, ".maid", "crg-venv")
	pip := filepath.Join(venv, "bin", "pip")
	if _, err := os.Stat(pip); errors.Is(err, os.ErrNotExist) {
		if err := run(ctx, python, "-m", "venv", venv); err != nil {
			return "", fmt.Errorf("create Python environment: %w", err)
		}
	}
	if err := run(ctx, pip, "install", packageName); err != nil {
		return "", fmt.Errorf("install %s: %w", packageName, err)
	}
	command := filepath.Join(venv, "bin", packageName)
	if _, err := os.Stat(command); err != nil {
		return "", fmt.Errorf("%s was not installed: %w", packageName, err)
	}
	return command, nil
}

func buildGraph(ctx context.Context, command, root string) error {
	// Purpose: build the graph for the project root.
	if err := run(ctx, command, "build", "--repo", root); err != nil {
		return fmt.Errorf("build code graph: %w", err)
	}
	return nil
}

func run(ctx context.Context, name string, args ...string) error {
	// Purpose: run one external command and include its output in errors.
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
