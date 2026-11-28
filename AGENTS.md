# Agent instructions

This file is for every agent that works on this repository: design agents, build agents, subagents and review agents.

## What this project is

maid is an MCP server and CLI. You install it in a software project. It guides agents from a goal to accepted, tested code. The design comes first. Large agents write the design. Smaller agents implement it in small work items, in parallel. maid accepts a commit only when its tests pass and no other test breaks.

## Source of truth

- `design/` is the specification. It says what maid must be. It is the source of truth.
- `docs/`, where present, is generated from code. It says what maid is now. Do not edit it by hand.
- The code is the present state. If the code and `design/` disagree, `design/` is correct, or `design/` must change first.

## How to start

1. Read `design/index.md`. It lists every part and chapter in reading order. `maid_status` and `design_index` give the status and summary of each chapter.
2. Read only the chapters that your task needs. Each chapter has a `summary` in its frontmatter. Read the summary before you read the chapter.
3. Before you write in a chapter, read its style section in `design/specification/styles.md`. The chapter frontmatter names the style.

## Rules

1. Do not guess. If information is missing, ask the user and record the answer in the relevant design section. Do not create a separate register of questions.
2. Do not change a chapter with `status: accepted`. Write a proposal instead. A proposal is a section `## Proposal: <title>` at the end of the chapter, with `status: proposed` in its first line.
3. You can change chapters with `status: stub` or `status: draft`.
4. Name chapters and sections after what they are. Refer to work items by their titles; task store IDs are opaque handles for the database only. Never renumber parts.
5. Follow `design/implementation/conventions.md` for code and commits.
6. One logical change per commit.

## When maid is available

When the maid MCP server runs in this repository, call `maid_next` before you do anything else. Follow its instructions. They replace steps 1–3 above.
