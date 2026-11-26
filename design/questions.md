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

Status values: `open`, `answered`, `resolved`. A question is resolved when the blocked items are updated. New questions get the next free ID.

| ID | Question | Answer | Updated items | Status |
|---|---|---|---|---|
| Q-001 | Targets for SM-001 (small-agent acceptance rate) and SM-002 (cost ratio)? | Measure first. Set both targets after the calibration benchmark runs. | vision#SM-001, vision#SM-002, constraints#A-002 | resolved |
| Q-002 | Must v1 support Windows? | No. Linux and macOS in v1. The design must not prevent Windows later. | constraints#C-005 | resolved |
| Q-003 | Should `maid serve` accept agents on other machines? | Not in v1. Local connections only. | architecture, mcp-interface#REQ-MCP-007 | resolved |
| Q-004 | When does the review band hold a submission for a human? | Only at the rigorous depth level. | acceptance#REQ-ACC-014 | resolved |
| Q-005 | How are score weights calibrated? | With a benchmark set of work items that have known good and bad solutions. | scoring#REQ-SC-011, scoring#REQ-SC-012 | resolved |
| Q-006 | Which built-in profiles does v1 ship? | Several project profiles. maid recommends one when information is missing. The agent can accept, customise or ignore it. | profiles | resolved |
| Q-007 | Formalized English and ACE? | Drop Formalized English. Keep ACE as advice only. Build the APE adapter later. | style-registry, style-ace | resolved |
| Q-008 | Merge the state store and the task store? | No. Keep them separate. | architecture | resolved |
| Q-009 | Minimum Go version? | The two newest Go releases, as Go's release policy supports. | conventions | resolved |
| Q-010 | Enforce a cost budget per work item? | Record only. Limit by attempts. | work-items#REQ-WI-016 | resolved |
| Q-011 | Which generator produces `docs/`? | Generating `docs/` is optional. Where the language has an easy generator, the profile names it. | profiles#REQ-PR-014, phases#REQ-PH-009, phases#REQ-PH-017 | resolved |
| Q-012 | Which optional adapters come first after v1? | Neither yet. | vision#NG-001, vision#NG-002 | resolved |
| Q-013 | How are shared test resources isolated? | Unique environment variables for each attempt, and a serial marker for suites that cannot use them. | acceptance#REQ-ACC-017, acceptance#REQ-ACC-018 | resolved |
