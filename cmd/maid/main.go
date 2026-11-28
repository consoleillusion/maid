// Command maid is the maid CLI and MCP server.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/consoleillusion/maid/internal/config"
	"github.com/consoleillusion/maid/internal/design"
	"github.com/consoleillusion/maid/internal/guide"
	"github.com/consoleillusion/maid/internal/mcp"
	"github.com/consoleillusion/maid/internal/setup"
	"github.com/consoleillusion/maid/internal/state"
	"github.com/consoleillusion/maid/internal/style"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

const usage = `maid — Metric Assisted AI Driven Development

Usage:
  maid init                         set up maid.json, Beads and .maid/ in this repository
  maid mcp                          run the MCP server on stdio
  maid next   [-agent ID]
  maid status [-detail] [-json]
  maid check  [-json]               check the design; exit 1 on errors
  maid index  [-json]
  maid read   [-summary] REF...     REF is a chapter ID, item ID or chapter#anchor
  maid accept CHAPTER...            accept chapters
  maid accept -pending              accept every chapter that is not accepted
  maid plan   FILE.json             create work items from a plan
  maid claim  -agent ID [WORK_ITEM]
  maid submit -agent ID [-model M] WORK_ITEM
  maid release -agent ID -reason TEXT WORK_ITEM
  maid version
`

func main() {
	// Purpose: dispatch the subcommand and set the exit code.
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	code, err := run(context.Background(), os.Args[1], os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "maid:", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

func run(ctx context.Context, cmd string, args []string) (int, error) {
	// Purpose: run one subcommand.
	switch cmd {
	case "version", "-v", "--version":
		fmt.Println("maid", mcp.Version)
		return 0, nil
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0, nil
	}
	root, err := projectRoot()
	if err != nil {
		return 1, err
	}
	if cmd == "init" {
		return 0, initProject(ctx, root)
	}
	svc, err := service(root)
	if err != nil {
		return 1, err
	}
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	agent := fs.String("agent", os.Getenv("MAID_AGENT"), "agent ID")
	detail := fs.Bool("detail", false, "include each work item")
	summary := fs.Bool("summary", false, "chapter summaries only")
	model := fs.String("model", "", "model name")
	reason := fs.String("reason", "", "reason")
	pending := fs.Bool("pending", false, "accept every chapter that is not accepted")
	if err := fs.Parse(args); err != nil {
		return 2, err
	}
	rest := fs.Args()
	switch cmd {
	case "mcp":
		return 0, mcp.NewServer(&mcp.Tools{Svc: svc}).Serve(ctx, os.Stdin, os.Stdout)
	case "check":
		return check(svc, *asJSON)
	case "index":
		ds, err := svc.Design()
		if err != nil {
			return 1, err
		}
		idx := ds.Index(design.Filter{})
		if *asJSON {
			return 0, printJSON(idx)
		}
		for _, e := range idx {
			fmt.Printf("%-22s %-9s %-8s %s\n", e.ID, e.Status, e.Style, e.Title)
		}
		return 0, nil
	case "read":
		ds, err := svc.Design()
		if err != nil {
			return 1, err
		}
		ex, err := ds.Read(rest, *summary)
		if err != nil {
			return 1, err
		}
		for _, e := range ex {
			fmt.Printf("== %s (%s:%d, %s, version %s)\n%s\n\n", e.Ref, e.Path, e.Line, e.Status, e.Version, e.Text)
		}
		return 0, nil
	case "next":
		n, err := guide.Compute(ctx, svc, "", *agent)
		if err != nil {
			return 1, err
		}
		return 0, printJSON(n)
	case "status":
		st, err := guide.Report(ctx, svc, *detail)
		if err != nil {
			return 1, err
		}
		if *asJSON {
			return 0, printJSON(st)
		}
		printStatus(st)
		return 0, nil
	case "accept":
		ids := rest
		if *pending {
			ds, err := svc.Design()
			if err != nil {
				return 1, err
			}
			ids = guide.Pending(ds)
		}
		if len(ids) == 0 {
			return 2, errors.New("usage: maid accept CHAPTER... or maid accept -pending")
		}
		out, _ := exec.Command("git", "-C", root, "config", "user.name").Output()
		res, err := svc.SetStatusBatch(ctx, strings.TrimSpace(string(out)), ids, "accepted")
		if err != nil {
			return 1, err
		}
		failed := 0
		for _, r := range res {
			if r.OK {
				fmt.Printf("accepted  %s\n", r.ID)
			} else {
				failed++
				fmt.Printf("refused   %s: %s\n", r.ID, r.Error)
			}
		}
		if failed > 0 {
			return 1, nil
		}
		return 0, nil
	case "plan":
		if len(rest) != 1 {
			return 2, errors.New("usage: maid plan FILE.json")
		}
		b, err := os.ReadFile(rest[0])
		if err != nil {
			return 1, err
		}
		var p work.Plan
		if err := json.Unmarshal(b, &p); err != nil {
			return 1, err
		}
		out, err := svc.CreatePlan(ctx, p)
		if err != nil {
			return 1, err
		}
		return 0, printJSON(out)
	case "claim":
		id := ""
		if len(rest) > 0 {
			id = rest[0]
		}
		p, err := svc.Claim(ctx, *agent, id)
		if err != nil {
			return 1, err
		}
		return 0, printJSON(p)
	case "submit":
		if len(rest) != 1 {
			return 2, errors.New("usage: maid submit -agent ID WORK_ITEM")
		}
		rep, err := svc.Submit(ctx, work.Submission{Agent: *agent, ID: rest[0], Model: *model})
		if err != nil {
			return 1, err
		}
		if err := printJSON(rep); err != nil {
			return 1, err
		}
		if !rep.Accepted {
			return 3, nil
		}
		return 0, nil
	case "release":
		if len(rest) != 1 {
			return 2, errors.New("usage: maid release -agent ID -reason TEXT WORK_ITEM")
		}
		return 0, svc.Release(ctx, *agent, rest[0], *reason)
	}
	fmt.Fprint(os.Stderr, usage)
	return 2, fmt.Errorf("unknown command %q", cmd)
}

func projectRoot() (string, error) {
	// Purpose: find the git repository root of the current folder.
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", errors.New("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func service(root string) (*work.Service, error) {
	// Purpose: build the work service with the Beads task store.
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	return work.New(root, cfg, &tasks.Beads{Dir: root, Bin: cfg.Bd}), nil
}

func initProject(ctx context.Context, root string) error {
	// Purpose: create maid.json, .maid/, the Beads database and the project assets.
	if _, err := os.Stat(root + "/" + config.File); errors.Is(err, os.ErrNotExist) {
		if err := config.Write(root, config.Default()); err != nil {
			return err
		}
		fmt.Println("wrote", config.File, "(go-cli defaults; edit for other stacks)")
	}
	if err := (state.Store{Root: root}).Ensure(); err != nil {
		return err
	}
	if err := ensureIgnored(root, "/.maid/"); err != nil {
		return err
	}
	created, err := design.EnsureIndexFiles(root)
	if err != nil {
		return err
	}
	assets, err := setup.Ensure(root)
	if err != nil {
		return err
	}
	for _, p := range append(created, assets...) {
		fmt.Println("wrote", p)
	}
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}
	cleanAttemptBranches(ctx, root)
	if err := work.New(root, cfg, nil).OnMaidBranch(ctx); err != nil {
		fmt.Println("maid branch:", err)
	} else {
		fmt.Println("maid commits to branch", cfg.Branch)
	}
	b := &tasks.Beads{Dir: root, Bin: cfg.Bd}
	if !b.Initialised() {
		if err := b.Init(ctx); err != nil {
			return fmt.Errorf("%w (install bd: npm install -g @beads/bd)", err)
		}
		fmt.Println("initialised Beads in .beads/")
	}
	fmt.Println("maid is ready. Point your agent at `maid mcp` and tell it to call maid_next.")
	return nil
}

func cleanAttemptBranches(ctx context.Context, root string) {
	// Purpose: remove the maid/* branches and their worktrees from the old layout.
	out, err := exec.CommandContext(ctx, "git", "-C", root, "worktree", "list", "--porcelain").Output()
	if err == nil {
		path := ""
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "worktree ") {
				path = strings.TrimPrefix(l, "worktree ")
			}
			if strings.HasPrefix(l, "branch refs/heads/maid/") {
				_ = exec.CommandContext(ctx, "git", "-C", root, "worktree", "remove", "--force", path).Run()
			}
		}
	}
	_ = exec.CommandContext(ctx, "git", "-C", root, "worktree", "prune").Run()
	refs, err := exec.CommandContext(ctx, "git", "-C", root, "for-each-ref", "--format=%(refname:short)", "refs/heads/maid/").Output()
	if err != nil {
		return
	}
	for _, b := range strings.Fields(string(refs)) {
		if exec.CommandContext(ctx, "git", "-C", root, "branch", "-D", b).Run() == nil {
			fmt.Println("deleted branch", b)
		}
	}
}

func ensureIgnored(root, entry string) error {
	// Purpose: add an entry to .gitignore if it is missing.
	p := root + "/.gitignore"
	b, err := os.ReadFile(p)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) == entry {
			return nil
		}
	}
	if len(b) > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	return os.WriteFile(p, append(b, []byte(entry+"\n")...), 0o644)
}

func check(svc *work.Service, asJSON bool) (int, error) {
	// Purpose: print design findings and fail on errors.
	ds, err := svc.Design()
	if err != nil {
		return 1, err
	}
	fs := ds.Check()
	errs := 0
	for _, f := range fs {
		if f.Severity == style.Error {
			errs++
		}
	}
	if asJSON {
		if err := printJSON(fs); err != nil {
			return 1, err
		}
	} else {
		for _, f := range fs {
			fmt.Printf("%s:%d: %s [%s] %s\n", f.Path, f.Line, f.Severity, f.Rule, f.Message)
		}
		fmt.Printf("%d chapters, %d findings, %d errors\n", len(ds.Chapters), len(fs), errs)
	}
	if errs > 0 {
		return 1, nil
	}
	return 0, nil
}

func printStatus(st guide.Status) {
	// Purpose: print a short human status report.
	fmt.Printf("Phase:      %s\nNext step:  %s\n", st.Next.Phase, st.Next.Step)
	fmt.Printf("Chapters:   %v\nWork items: %v\n", st.Chapters, st.WorkItems)
	if len(st.Escalated) > 0 {
		fmt.Printf("Escalated:  %s\n", strings.Join(st.Escalated, ", "))
	}
	fmt.Printf("Tests on main (%s): %d pass, %d fail, %d skip\n", short(st.Baseline.Commit), st.Baseline.Pass, st.Baseline.Fail, st.Baseline.Skip)
	for _, it := range st.Items {
		fmt.Printf("  %-14s %-6s %-12s %-12s %s\n", it.ID, it.Kind, it.Status, it.Assignee, it.Title)
	}
	if len(st.Recent) > 0 {
		fmt.Println("Recent:")
		for _, e := range st.Recent {
			fmt.Printf("  %s %-8s %-14s %s %s\n", e.Time.Format("01-02 15:04"), e.Action, e.WorkItem, e.Agent, e.Detail)
		}
	}
}

func short(s string) string {
	// Purpose: shorten a commit ID.
	if len(s) > 10 {
		return s[:10]
	}
	if s == "" {
		return "none"
	}
	return s
}

func printJSON(v any) error {
	// Purpose: print indented JSON to stdout.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
