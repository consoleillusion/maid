---
id: questions
title: Open questions
part: design
status: draft
style: ste
owner: paul
summary: >
  Information that the design needs and does not have. Each question names the
  items that it blocks. Answer a question here, then update the blocked items
  and set the question to resolved.
links: []
---

# Open questions

Status values: `open`, `answered`, `resolved`. A question is resolved when the blocked items are updated.

| ID | Question | Blocks | Owner | Status |
|---|---|---|---|---|
| Q-001 | What are the targets for SM-001 (first or second attempt acceptance rate) and SM-002 (cost ratio)? How do we measure them in trials? | vision#SM-001, vision#SM-002, constraints#A-002 | paul | open |
| Q-002 | Must v1 support Windows, or only Linux and macOS? | constraints#C-005 | paul | open |
| Q-003 | Should `maid serve` accept agents on other machines? If yes, how are worktrees and tests run for them? | architecture | paul | open |
| Q-004 | Should the review band hold a submission for human approval in the default preset, or only in the rigorous depth level? | acceptance#REQ-ACC-014 | paul | open |
| Q-005 | How are score weights and normalisation functions calibrated? Proposal: a benchmark set of work items with known good and bad solutions. | scoring | paul | open |
| Q-006 | Which built-in presets does v1 ship? Proposal: `go-cli`, `web-local` (Bun, TypeScript, SQLite), `script-python` (uv, pytest, ruff). | presets | paul | open |
| Q-007 | Is Formalized English (Martin 2002) needed? If yes, which validator? When is the APE adapter for ACE built? | style-registry, style-ace | paul | open |
| Q-008 | Should the state store and the task store become one store after v1? | architecture | paul | open |
| Q-009 | Which Go version is the minimum for maid? | conventions | paul | open |
| Q-010 | Should maid enforce a cost budget for each work item? In the pull model maid knows only what agents report. | work-items, scoring | paul | open |
| Q-011 | Which generator produces `docs/` for each language (for example `gomarkdoc` for Go, TypeDoc for TypeScript)? Is it part of the preset? | phases#REQ-PH-009, phases#REQ-PH-017, presets | paul | open |
| Q-012 | Which optional adapters come first after v1: forge pull requests (Forgejo, GitHub), tracker mirror (Windshift, Jira), or neither? | vision#NG-001, vision#NG-002 | paul | open |
| Q-013 | How does maid isolate test suites that need shared resources (fixed ports, one database) when attempts run in parallel? | constraints#A-004, acceptance | paul | open |
