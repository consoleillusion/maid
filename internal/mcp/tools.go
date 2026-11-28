package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/consoleillusion/maid/internal/design"
	"github.com/consoleillusion/maid/internal/guide"
	"github.com/consoleillusion/maid/internal/tasks"
	"github.com/consoleillusion/maid/internal/work"
)

// Tools implements maid's MCP tools on top of a work.Service.
type Tools struct {
	Svc *work.Service
}

// Tool is a tool definition for tools/list.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type def struct {
	tool    Tool
	buildOK bool // listed only from phase 5 (REQ-PH-004)
	run     func(ctx context.Context, t *Tools, args map[string]any, agent string) (any, string, error)
}

func obj(props map[string]any, required ...string) map[string]any {
	// Purpose: build a JSON Schema object for tool input.
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func str(desc string) map[string]any {
	// Purpose: build a JSON Schema string property.
	return map[string]any{"type": "string", "description": desc}
}

var agentProp = str("Your agent ID. Defaults to the client name plus a session ID.")

var defs = []def{
	{tool: Tool{"maid_next", "The next step for you: phase, step, what to read, which tools to use, and when the step is done. Call this first.",
		obj(map[string]any{"agent": agentProp})},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			n, err := guide.Compute(ctx, t.Svc, "", agent)
			hint := "follow the step"
			if len(n.Calls) > 0 {
				hint = "make the calls in \"calls\" in order; their arguments are ready to use"
			} else if len(n.Tools) > 0 {
				hint = "call " + n.Tools[0]
			}
			return n, hint, err
		}},
	{tool: Tool{"maid_status", "Project progress for humans and agents: chapters by status, work items, test baseline, escalations and recent activity.",
		obj(map[string]any{"detail": map[string]any{"type": "boolean", "description": "include each work item"}})},
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			st, err := guide.Report(ctx, t.Svc, bArg(a, "detail"))
			return st, "call maid_next", err
		}},
	{tool: Tool{"design_index", "List design chapters with summaries only. Filter by part, status, style or link.",
		obj(map[string]any{"part": str("part folder"), "status": str("stub, draft, review, accepted or superseded"), "style": str("style ID"), "link": str("only chapters that link to this ID")})},
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			ds, err := t.Svc.Design()
			if err != nil {
				return nil, "", err
			}
			return ds.Index(design.Filter{Part: sArg(a, "part"), Status: sArg(a, "status"), Style: sArg(a, "style"), Link: sArg(a, "link")}),
				"call design_read with the IDs you need", nil
		}},
	{tool: Tool{"design_read", "Read design text. Each ref can be a chapter ID (vision), a file path (design/01-foundation/01-vision.md), a chapter title, an item ID (REQ-WI-004, any case), or chapter#heading. Returns the smallest text for each ref with its version; refs that match nothing come back in not_found with suggestions. Example: {\"refs\": [\"vision\", \"REQ-WI-004\"]}",
		obj(map[string]any{"refs": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "chapter IDs, paths, titles, item IDs or chapter#heading"},
			"detail": map[string]any{"type": "string", "enum": []string{"summary", "full"}, "description": "summary returns only chapter summaries"}}, "refs")},
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			ds, err := t.Svc.Design()
			if err != nil {
				return nil, "", err
			}
			refs := lArg(a, "refs")
			if len(refs) == 0 {
				return nil, "", errors.New(`refs is empty; pass chapter IDs or paths, for example {"refs": ["vision"]}. Call design_index to list chapters`)
			}
			ex, missing := ds.ReadAll(refs, sArg(a, "detail") == "summary")
			hint := "call maid_next when you have what you need"
			if len(missing) > 0 {
				hint = "some refs matched nothing; use a suggestion from not_found, or call design_index"
			}
			return map[string]any{"excerpts": ex, "not_found": missing}, hint, nil
		}},
	{tool: Tool{"design_check", "Check the design: frontmatter, IDs, links and writing styles. Errors fail gates; advice does not.", obj(map[string]any{})},
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			ds, err := t.Svc.Design()
			if err != nil {
				return nil, "", err
			}
			f := ds.Check()
			hint := "design is clean; call maid_next"
			if len(f) > 0 {
				hint = fmt.Sprintf("fix %d findings, then call design_check again", len(f))
			}
			return map[string]any{"findings": f, "count": len(f)}, hint, nil
		}},
	{tool: Tool{"design_write", "Replace one design section (heading included). Pass base_version from design_read; the write fails if the section changed since. Accepted chapters cannot be written; use design_propose.",
		obj(map[string]any{"ref": str("item ID or chapter#anchor"), "content": str("the new section text, starting with its heading"),
			"base_version": str("version from design_read"), "reason": str("why, one line"), "agent": agentProp}, "ref", "content", "base_version", "reason")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			res, err := t.Svc.DesignOp(ctx, agent, "write", sArg(a, "reason"), func(ds *design.Store) (design.WriteResult, error) {
				return ds.WriteSection(sArg(a, "ref"), rawArg(a, "content"), sArg(a, "base_version"))
			})
			return res, writeHint(res), err
		}},
	{tool: Tool{"design_append", "Add a new section at the end of a design chapter that is not accepted.",
		obj(map[string]any{"chapter": str("chapter ID"), "content": str("section text, starting with a heading"), "reason": str("why, one line"), "agent": agentProp}, "chapter", "content")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			res, err := t.Svc.DesignOp(ctx, agent, "append", sArg(a, "reason"), func(ds *design.Store) (design.WriteResult, error) {
				return ds.AppendSection(sArg(a, "chapter"), rawArg(a, "content"))
			})
			return res, writeHint(res), err
		}},
	{tool: Tool{"design_create", "Create a new design chapter with status draft. path is relative to design/, for example 01-foundation/01-vision.md. Read the style card in design/styles/ before you write the body.",
		obj(map[string]any{"path": str("part/NN-name.md"), "id": str("chapter ID, lowercase"), "title": str("title"),
			"style":   map[string]any{"type": "string", "enum": []string{"ste", "ears", "sbvr-se", "gherkin", "ace"}},
			"summary": str("at most 60 words"), "links": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"body": str("Markdown body, starting with '# Title'"), "owner": str("person who accepts changes"), "agent": agentProp}, "path", "id", "title", "style", "summary", "body")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			spec := design.ChapterSpec{Path: sArg(a, "path"), ID: sArg(a, "id"), Title: sArg(a, "title"), Style: sArg(a, "style"),
				Owner: sArg(a, "owner"), Summary: sArg(a, "summary"), Links: lArg(a, "links"), Body: rawArg(a, "body")}
			res, err := t.Svc.DesignOp(ctx, agent, "create", "create chapter "+spec.ID, func(ds *design.Store) (design.WriteResult, error) {
				return ds.CreateChapter(spec)
			})
			return res, writeHint(res), err
		}},
	{tool: Tool{"design_set_status", "Change the status of one or more chapters in one call and one commit. Pass ids (a list of chapter IDs or file paths). Use status accepted only for chapters that the user approved. Pass approved_by with the name of the person who approved them; leave it out when the user delegated approval to you. Chapters with check errors are refused and listed; the others still change. Example: {\"ids\": [\"vision\", \"glossary\"], \"status\": \"accepted\", \"approved_by\": \"Paul\"}",
		obj(map[string]any{"ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "chapter IDs or paths"},
			"id": str("one chapter ID; use ids for several"), "status": map[string]any{"type": "string", "enum": []string{"stub", "draft", "review", "accepted", "superseded"}},
			"approved_by": str("the person who approved; leave out when the user delegated approval to you"), "agent": agentProp}, "status")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			ids := lArg(a, "ids")
			if id := sArg(a, "id"); id != "" {
				ids = append(ids, id)
			}
			if len(ids) == 0 {
				return nil, "", errors.New(`pass ids, for example {"ids": ["vision"], "status": "accepted", "approved_by": "Paul"}`)
			}
			res, err := t.Svc.SetStatusBatch(ctx, agent, ids, sArg(a, "status"), sArg(a, "approved_by"))
			hint := "call maid_next"
			for _, r := range res {
				if !r.OK {
					hint = "some chapters did not change; read their errors, fix them, and retry only those"
				}
			}
			return map[string]any{"results": res}, hint, err
		}},
	{tool: Tool{"design_propose", "Propose a change to a chapter, including an accepted one. The proposal is appended as a section for the owner to review.",
		obj(map[string]any{"chapter": str("chapter ID"), "title": str("short title"), "content": str("the proposed change and its reason"), "agent": agentProp}, "chapter", "title", "content")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			res, err := t.Svc.DesignOp(ctx, agent, "propose", sArg(a, "title"), func(ds *design.Store) (design.WriteResult, error) {
				return ds.Propose(sArg(a, "chapter"), sArg(a, "title"), rawArg(a, "content"), agent)
			})
			return res, "tell the user that a proposal is waiting for review", err
		}},
	{tool: Tool{"design_question", "Record a question for the user instead of guessing. Returns the question ID; write TBD (Q-NNN) where the answer is needed.",
		obj(map[string]any{"question": str("the question"), "blocks": str("IDs of the items that need the answer"), "agent": agentProp}, "question")},
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			var qid string
			res, err := t.Svc.DesignOp(ctx, agent, "question", sArg(a, "question"), func(ds *design.Store) (design.WriteResult, error) {
				id, r, err := ds.AddQuestion(sArg(a, "question"), sArg(a, "blocks"), "")
				qid = id
				return r, err
			})
			return map[string]any{"id": qid, "result": res}, "ask the user, or continue with unblocked work", err
		}},
	{tool: Tool{"work_plan", "Create work items from a plan. For each code item maid also creates a tests item that a different agent completes first. Each item needs key, title, implements (requirement IDs), criteria (Gherkin), acceptance_tests ([{file, names}]) and owns (path globs that must not include the test files).",
		obj(map[string]any{"plan": map[string]any{"type": "object", "description": "{\"items\": [ ... ]}"}}, "plan")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			var p work.Plan
			if err := remarshal(a["plan"], &p); err != nil {
				return nil, "", fmt.Errorf("plan: %w", err)
			}
			out, err := t.Svc.CreatePlan(ctx, p)
			return out, "build agents can now call work_claim", err
		}},
	{tool: Tool{"work_claim", "Claim a ready work item (or the given one). Returns your worktree, branch and context packet: everything you need to complete the item.",
		obj(map[string]any{"id": str("work item ID; empty for the next ready item"), "agent": agentProp})},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			p, err := t.Svc.Claim(ctx, agent, sArg(a, "id"))
			if errors.Is(err, tasks.ErrNoReady) {
				return map[string]any{"claimed": false, "reason": "no ready work item for you"}, "call maid_next later", nil
			}
			return p, "work in the worktree, commit, then call work_submit", err
		}},
	{tool: Tool{"work_files", "List the files in your worktree for a claimed work item. owned=true marks the files you may change.",
		obj(map[string]any{"id": str("work item ID"), "agent": agentProp}, "id")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			f, err := t.Svc.Files(ctx, agent, sArg(a, "id"))
			return f, "read files with work_read, change owned files with work_write", err
		}},
	{tool: Tool{"work_read", "Read a file from your worktree. path is relative to the repository root, for example internal/score/score.go.",
		obj(map[string]any{"id": str("work item ID"), "path": str("file path"), "agent": agentProp}, "id", "path")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			c, err := t.Svc.ReadFile(ctx, agent, sArg(a, "id"), sArg(a, "path"))
			return map[string]any{"path": sArg(a, "path"), "content": c}, "change owned files with work_write", err
		}},
	{tool: Tool{"work_write", "Create or replace a file in your worktree with the full new content. Only files your work item owns can be written. Set remove to delete the file.",
		obj(map[string]any{"id": str("work item ID"), "path": str("file path"), "content": str("the complete file content"),
			"remove": map[string]any{"type": "boolean"}, "agent": agentProp}, "id", "path")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			err := t.Svc.WriteFile(ctx, agent, sArg(a, "id"), sArg(a, "path"), rawArg(a, "content"), bArg(a, "remove"))
			return map[string]any{"written": sArg(a, "path")}, "call work_run with check test, then work_commit", err
		}},
	{tool: Tool{"work_run", "Run the project's tests, format check or lint in your worktree, so you can check your work before you submit.",
		obj(map[string]any{"id": str("work item ID"), "check": map[string]any{"type": "string", "enum": []string{"test", "format", "lint"}}, "agent": agentProp}, "id")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			r, err := t.Svc.Run(ctx, agent, sArg(a, "id"), sArg(a, "check"))
			hint := "call work_commit, then work_submit"
			if !r.Passed {
				hint = "fix the failures with work_write and run again"
			}
			return r, hint, err
		}},
	{tool: Tool{"work_commit", "Commit all changes in your worktree with a one-line message.",
		obj(map[string]any{"id": str("work item ID"), "message": str("one line"), "agent": agentProp}, "id", "message")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			c, err := t.Svc.Commit(ctx, agent, sArg(a, "id"), sArg(a, "message"))
			return map[string]any{"commit": c}, "call work_submit", err
		}},
	{tool: Tool{"work_heartbeat", "Renew your claim lease on a work item.", obj(map[string]any{"id": str("work item ID"), "agent": agentProp}, "id")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			return map[string]any{"renewed": true}, "continue, then call work_submit", t.Svc.Heartbeat(ctx, agent, sArg(a, "id"))
		}},
	{tool: Tool{"work_release", "Give up your claim on a work item.", obj(map[string]any{"id": str("work item ID"), "reason": str("why"), "agent": agentProp}, "id", "reason")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			return map[string]any{"released": true}, "call maid_next", t.Svc.Release(ctx, agent, sArg(a, "id"), sArg(a, "reason"))
		}},
	{tool: Tool{"work_submit", "Submit your work. Either commit in your worktree first, or send the changes here as patch (a unified diff) or files (path to complete content) and maid commits them. The merge queue then rebases on main, runs the checks and the tests, and merges the work or returns a rejection report.",
		obj(map[string]any{"id": str("work item ID"), "patch": str("optional unified diff relative to the repository root"),
			"files":   map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": "optional map of path to complete new file content"},
			"message": str("commit message for the patch or files"),
			"model":   str("your model name, for scoring history"), "cost": str("cost you observed, if known"), "agent": agentProp}, "id")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, agent string) (any, string, error) {
			files := map[string]string{}
			if m, ok := a["files"].(map[string]any); ok {
				for k, v := range m {
					if sv, ok := v.(string); ok {
						files[k] = sv
					}
				}
			}
			rep, err := t.Svc.Submit(ctx, work.Submission{Agent: agent, ID: sArg(a, "id"), Model: sArg(a, "model"), Cost: sArg(a, "cost"),
				Patch: rawArg(a, "patch"), Files: files, Message: sArg(a, "message")})
			hint := "accepted; call maid_next for more work"
			if !rep.Accepted {
				hint = "rejected; read the report, then call maid_next"
			}
			return rep, hint, err
		}},
	{tool: Tool{"work_result", "The attempts on a work item and their merge queue reports.", obj(map[string]any{"id": str("work item ID")}, "id")},
		buildOK: true,
		run: func(ctx context.Context, t *Tools, a map[string]any, _ string) (any, string, error) {
			att, err := t.Svc.State.Attempts(sArg(a, "id"))
			return att, "call maid_next", err
		}},
}

// Phase returns the current phase name, for list_changed notifications.
func (t *Tools) Phase(ctx context.Context) string {
	// Purpose: compute the phase without failing the caller.
	n, err := guide.Compute(ctx, t.Svc, guide.RoleDesign, "")
	if err != nil {
		return "unknown"
	}
	return n.Phase
}

// List returns every tool. Clients may not reload the list when the phase
// changes, so phase rules are applied when a tool is called (REQ-PH-004).
func (t *Tools) List(ctx context.Context) []Tool {
	// Purpose: list all tools in every phase.
	out := make([]Tool, 0, len(defs))
	for _, d := range defs {
		out = append(out, d.tool)
	}
	return out
}

// Call runs a tool and returns an MCP tool result.
func (t *Tools) Call(ctx context.Context, name string, raw json.RawMessage, defaultAgent string) map[string]any {
	// Purpose: run one tool and format its result with a next-step hint (REQ-PH-002).
	args := map[string]any{}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &args); err != nil {
			return errorResult("invalid arguments: " + err.Error())
		}
	}
	agent := defaultAgent
	if a := sArg(args, "agent"); a != "" {
		agent = a
	}
	for _, d := range defs {
		if d.tool.Name != name {
			continue
		}
		if d.buildOK {
			if n, err := guide.Compute(ctx, t.Svc, guide.RoleDesign, agent); err == nil && n.Phase < "5" {
				return errorResult(fmt.Sprintf("%s is for the plan and build phases; the project is in phase %s.\nnext step: %s\nnext: call maid_next", name, n.Phase, n.Step))
			}
		}
		v, hint, err := d.run(ctx, t, args, agent)
		if err != nil {
			return errorResult(err.Error() + "\nnext: call maid_next")
		}
		b, _ := json.MarshalIndent(v, "", "  ")
		return map[string]any{"content": []map[string]any{{"type": "text", "text": string(b) + "\nnext: " + hint}}}
	}
	return errorResult("unknown tool " + name)
}

func errorResult(msg string) map[string]any {
	// Purpose: build an MCP tool error result.
	return map[string]any{"isError": true, "content": []map[string]any{{"type": "text", "text": msg}}}
}

func sArg(a map[string]any, k string) string {
	// Purpose: read a string argument.
	s, _ := a[k].(string)
	return strings.TrimSpace(s)
}

func rawArg(a map[string]any, k string) string {
	// Purpose: read a string argument without trimming inner whitespace.
	s, _ := a[k].(string)
	return s
}

func writeHint(r design.WriteResult) string {
	// Purpose: point the agent at check findings after a write.
	if len(r.Findings) > 0 {
		return fmt.Sprintf("fix %d findings in %s, then call maid_next", len(r.Findings), r.Path)
	}
	return "call maid_next"
}

func bArg(a map[string]any, k string) bool {
	// Purpose: read a boolean argument.
	b, _ := a[k].(bool)
	return b
}

func lArg(a map[string]any, k string) []string {
	// Purpose: read a list of strings, or one string, as a list.
	switch v := a[k].(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func remarshal(in any, out any) error {
	// Purpose: convert a decoded JSON value into a typed struct.
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
