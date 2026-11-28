// Package config reads maid.json, the project's tool configuration.
//
// In v1 the tools are configured here. Later the design decisions select
// them and maid.json is generated from those decisions.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// File is the configuration file name at the project root.
const File = "maid.json"

// Test configures the test runner adapter.
type Test struct {
	Adapter string   `json:"adapter"` // "go" or "exit"
	Command []string `json:"command"`
	Serial  bool     `json:"serial,omitempty"`
}

// Budget holds work item defaults.
type Budget struct {
	Attempts        int `json:"attempts"`
	MaxChangedLines int `json:"max_changed_lines"`
}

// CodeReviewGraph holds the command that starts the graph MCP server.
type CodeReviewGraph struct {
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Enabled bool     `json:"enabled"`
}

// Config is the content of maid.json.
type Config struct {
	Branch          string          `json:"branch"` // every maid commit goes here
	Test            Test            `json:"test"`
	FormatCheck     []string        `json:"format_check,omitempty"` // must exit 0 and print nothing
	Lint            []string        `json:"lint,omitempty"`         // must exit 0
	Budget          Budget          `json:"budget"`
	Bd              string          `json:"bd,omitempty"`
	PortRange       int             `json:"port_range,omitempty"` // ports per attempt
	CodeReviewGraph CodeReviewGraph `json:"code_review_graph,omitempty"`
}

// Default returns the configuration for a Go project.
func Default() Config {
	// Purpose: give the go-cli profile defaults.
	return Config{
		Branch:      "maid-auto",
		Test:        Test{Adapter: "go", Command: []string{"go", "test", "-json", "./..."}},
		FormatCheck: []string{"gofmt", "-l", "."},
		Lint:        []string{"go", "vet", "./..."},
		Budget:      Budget{Attempts: 2, MaxChangedLines: 300},
		PortRange:   100,
		CodeReviewGraph: CodeReviewGraph{
			Command: []string{"code-review-graph"},
			Args:    []string{"serve"},
			Enabled: true,
		},
	}
}

// Load reads maid.json from root. A missing file gives the defaults.
func Load(root string) (Config, error) {
	// Purpose: read the project configuration over the defaults.
	c := Default()
	b, err := os.ReadFile(filepath.Join(root, File))
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("%s: %w", File, err)
	}
	if c.Branch == "" {
		c.Branch = "maid-auto"
	}
	if c.Budget.Attempts <= 0 {
		c.Budget.Attempts = 2
	}
	if c.PortRange <= 0 {
		c.PortRange = 100
	}
	if len(c.CodeReviewGraph.Command) == 0 {
		c.CodeReviewGraph.Command = []string{"code-review-graph"}
	}
	if len(c.CodeReviewGraph.Args) == 0 {
		c.CodeReviewGraph.Args = []string{"serve"}
	}
	return c, nil
}

// Write saves c to root/maid.json.
func Write(root string, c Config) error {
	// Purpose: write the configuration file with stable formatting.
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, File), append(b, '\n'), 0o644)
}
