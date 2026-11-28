---
id: index
title: maid design
part: design
status: accepted
style: ste
owner: paul
summary: Front page of the design and the reading order of its chapters.
links: []
---

# maid design

`design/` is the specification of maid and the source of truth. This page lists the chapters in reading order. Read the summary of a chapter before you read the chapter. Read only what your task needs. `maid_status` shows the status of each chapter. `typst compile design/index.typ` builds the design as one PDF.

## Foundation

1. [Vision](foundation/vision.md)
2. [Glossary](foundation/glossary.md)
3. [Constraints](foundation/constraints.md)

## Method

1. [Phases](method/phases.md)
2. [Work items](method/work-items.md)
3. [Acceptance](method/acceptance.md)
4. [Scoring](method/scoring.md)
5. [Profiles](method/profiles.md)

## Specification

1. [Chapter format](specification/chapter-format.md)
2. [Style registry](specification/style-registry.md)

## Architecture

1. [Architecture overview](architecture/overview.md)
2. [Task store](architecture/task-store.md)

## Interfaces

1. [MCP interface](interfaces/mcp.md)

## Engineering

1. [Conventions](engineering/conventions.md)

## Delivery

1. [Roadmap](delivery/roadmap.md)

## Decisions

1. [ADR-0001 Write maid in Go](decisions/ADR-0001-go.md)
2. [ADR-0002 Pull model](decisions/ADR-0002-pull-model.md)
3. [ADR-0003 Merge queue](decisions/ADR-0003-merge-queue.md)
4. [ADR-0004 Acceptance tests first](decisions/ADR-0004-acceptance-tests-first.md)
5. [ADR-0005 Beads task store](decisions/ADR-0005-beads-task-store.md)
6. [ADR-0006 Chapters and styles](decisions/ADR-0006-chapters-and-styles.md)
7. [ADR-0007 Standard library only](decisions/ADR-0007-stdlib-only.md)
8. [ADR-0008 Everything is a task](decisions/adr-0008-everything-is-a-task.md)

## Styles

1. [Simplified Technical English](styles/ste.md)
2. [EARS](styles/ears.md)
3. [SBVR Structured English](styles/sbvr-se.md)
4. [Gherkin](styles/gherkin.md)
5. [Attempto Controlled English](styles/ace.md)

## Questions

1. [Open questions](questions.md)
