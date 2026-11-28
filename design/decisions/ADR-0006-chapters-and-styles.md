---
status: accepted
style: ste
owner: paul
origin: user
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["agents regularly need text from many chapters for one task"]
summary: >
  The design is a living specification organised by area: parts are folders,
  chapters are single Markdown files, sections are addressable. Each chapter
  declares one controlled-language style with a style card and, where possible,
  a validator.
links: [chapter-format, style-registry]
---

# ADR-0006 Organise the design as parts and chapters, each with a declared style

## Context and problem statement

Many agents and humans read the design, and each needs only a part of it. Small agents fail on ambiguous text. How should the design be organised and written?

## Decision drivers

- An agent must read only what its task needs (G-006).
- Text must be unambiguous enough for small models.
- Text must be checkable by a program where possible.

## Considered options

- Parts and chapters by area, with a declared style for each chapter.
- One folder for each feature or change, as in Spec Kit and Kiro.
- Free-form documents.

## Decision outcome

Chosen option: "Parts and chapters by area", because the design is a living specification of the whole project, not a record of each change. Sections are addressable, so a context packet can include one section of a large chapter.

Each chapter declares one style. Embedded blocks can use another style. The starting styles are STE, EARS, SBVR Structured English and Gherkin, which have validators. ACE is registered with advice-only enforcement. AGENTS.md is the entry point, because many agents read it already. Change proposals for accepted items follow the idea of OpenSpec changes.

### Consequences

- Good, because context packets can be small.
- Good, because validators can enforce most styles.
- Bad, because writers must learn several styles. Style cards are kept short (REQ-STY-002).
- Bad, because the layout is not compatible with Spec Kit, Kiro or OpenSpec. An importer can be added later.
