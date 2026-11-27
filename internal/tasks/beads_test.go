package tasks

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"
)

// newBeads starts a Beads store in a temporary git repository.
// The test is skipped when bd is not installed.
func newBeads(t *testing.T) *Beads {
	// Purpose: set up a real bd database for contract tests.
	bin := os.Getenv("MAID_BD")
	if bin == "" {
		bin = "bd"
	}
	if _, err := exec.LookPath(bin); err != nil {
		t.Skip("bd not installed; set MAID_BD to run the Beads contract tests")
	}
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@example.invalid"}, {"config", "user.name", "t"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	b := &Beads{Dir: dir, Bin: bin}
	if err := b.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/AGENTS.md"); err == nil {
		t.Fatal("bd init wrote AGENTS.md")
	}
	return b
}

func contract(t *testing.T, p Port) {
	// Purpose: check the behaviour every task port adapter must have.
	ctx := context.Background()
	a, err := p.Create(ctx, NewItem{Title: "a", Maid: Meta{Kind: KindCode, Implements: []string{"REQ-WI-001"}, Owns: []string{"x/**"}}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Create(ctx, NewItem{Title: "b", DependsOn: []string{a}, Maid: Meta{Kind: KindCode}})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := p.Ready(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 || ready[0].ID != a {
		t.Fatalf("ready = %+v", ready)
	}
	if ready[0].Maid.Implements[0] != "REQ-WI-001" || ready[0].Maid.Owns[0] != "x/**" {
		t.Fatalf("metadata lost: %+v", ready[0].Maid)
	}
	// REQ-WI-004: exactly one of several concurrent claims wins.
	var wg sync.WaitGroup
	wins := make(chan string, 4)
	for _, agent := range []string{"w1", "w2", "w3", "w4"} {
		wg.Add(1)
		go func(agent string) {
			defer wg.Done()
			if _, err := p.Claim(ctx, a, agent); err == nil {
				wins <- agent
			} else if !errors.Is(err, ErrConflict) {
				t.Errorf("claim error: %v", err)
			}
		}(agent)
	}
	wg.Wait()
	close(wins)
	var winner string
	n := 0
	for w := range wins {
		winner = w
		n++
	}
	if n != 1 {
		t.Fatalf("%d claims won", n)
	}
	if err := p.Heartbeat(ctx, a, winner); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(ctx, a, "done"); err != nil {
		t.Fatal(err)
	}
	ready, _ = p.Ready(ctx)
	if len(ready) != 1 || ready[0].ID != b {
		t.Fatalf("after close, ready = %+v", ready)
	}
}

// REQ-WI-003 REQ-WI-004
func TestMemoryContract(t *testing.T) {
	// Purpose: the in-memory port obeys the port contract.
	contract(t, NewMemory())
}

// REQ-WI-003 REQ-WI-004
func TestBeadsContract(t *testing.T) {
	// Purpose: the Beads adapter obeys the port contract against real bd.
	contract(t, newBeads(t))
}
