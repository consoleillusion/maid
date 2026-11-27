---
id: vision
title: Vision and scope
part: 01-foundation
status: accepted
style: ste
owner: paul
summary: >
  Why maid exists, what it must achieve, what it must not do, and how we
  measure success. Read this first. Every other chapter must agree with it.
links: [glossary, constraints, phases, ADR-0002, ADR-0003, ADR-0004]
---

# Vision and scope

## Problem

AI agents can write code quickly. But they often:

- Add scope that nobody asked for.
- Add dependencies and code without evidence that the project improved.
- Lose context between sessions, and between agents.
- Guess when information is missing.

Large models are more reliable, but they cost much more. Small models are cheap, but they fail when a task is ambiguous or large.

## Solution

maid is an MCP server and a CLI. You install maid in a software project. Then you point an agent at it.

maid guides the agent through a fixed progression:

1. The agent writes the design with the user. The design is in `design/`.
2. A large agent divides the design into small work items. It writes short acceptance criteria for each work item. A separate small agent turns the criteria into failing tests.
3. Small agents take work items and implement them. They can work in parallel.
4. maid accepts each commit automatically when its acceptance tests pass and no other test breaks.

The design is the source of truth. All agents and all humans read the same design. Each agent reads only the sections that its task needs.

## Goals

### G-001 Guided progression
When an agent connects to maid, maid tells the agent the current phase, the next step and the context to read.

### G-002 Design as the source of truth
The design in `design/` is the single specification of the project. Code and tests trace to it. Generated documentation traces to it where the project has a generator.

### G-003 Any amount of input
The user can give as much or as little information as they want. maid applies defaults for decisions that the user does not make. maid records each default as a decision that the user can change.

### G-004 Cheap parallel implementation
A large agent does the design and the decomposition. Smaller agents do the implementation, in parallel. The total cost is a fraction of the cost of a large agent alone.

### G-005 Mechanical acceptance
maid accepts or rejects work by measurement, not by the opinion of an agent. The primary measure is tests. Secondary measures are lines of code, dependencies and benchmarks.

### G-006 Modular context
The design is divided into parts, chapters and sections. An agent can read one section without reading the rest.

### G-007 Progress that humans can see
A human can see the state of the project at any time: what is specified, accepted, implemented, tested and verified.

### G-008 Any agent, any language
maid works with any agent that can use MCP tools. maid works with projects in any programming language.

### G-009 Reuse existing tools
maid uses existing tools where they do the job: git, the task store, formatters, linters, test runners and metrics tools.

## Non-goals

### NG-001 No human project management
maid does not provide boards, sprints, time tracking, comments or customer portals. An optional adapter can copy maid state to such a tool. No adapter is planned for v1 (Q-012).

### NG-002 No git hosting
maid does not host repositories or provide pull requests. maid uses local git. An optional adapter can open pull requests on a forge. No adapter is planned for v1 (Q-012).

### NG-003 No formatter, linter or test framework of its own
maid runs the tools that the project selects.

### NG-004 No agent launching in v1
maid does not start agents in v1. Agents connect to maid and pull work. See ADR-0002.

### NG-005 No multi-user roles in v1
v1 is for one developer and many agents on one machine. The design must not prevent roles later.

## Success measures

| ID | Measure | Target |
|---|---|---|
| SM-001 | Work items accepted on the first or second attempt by a small agent | Set after the calibration benchmark runs (Q-001, REQ-SC-011) |
| SM-002 | Cost of a project built with maid, compared with a large agent alone | Set after the same benchmark project is built both ways (Q-001) |
| SM-003 | Accepted commits that break a previously passing test | 0 |
| SM-004 | Time for a new agent to start useful work after it connects | One `maid_next` call |
