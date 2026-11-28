---
id: chapter-format
title: Chapter format
part: specification
status: accepted
style: ears
owner: paul
summary: >
  How design/ is organised into parts, chapters and sections. Defines the
  frontmatter, stable IDs, status values and write rules that every chapter
  and every agent must follow.
links: [style-registry, ADR-0006]
---

# Chapter format

## Structure

`design/` contains parts. A part is a folder. A part contains chapters. A chapter is one Markdown file. A chapter contains sections. A section is a level-2 heading and its content.

```text
design/
  index.md                    design title and the files in reading order
  index.typ                   builds one PDF of the design from index.md
  questions.md                open questions
  <part-name>/
    <chapter-name>.md
  styles/<style-id>.md        style cards
  decisions/adr-NNNN-<slug>.md
```

Part and chapter names are lowercase words joined by hyphens. They do not contain position numbers. The order of the parts and chapters is the order in `index.md`. The number in an ADR file name is the ID of the decision, not a position.

`index.md` holds only information that no other file holds: the design title and the reading order. It has no frontmatter. Titles, statuses and summaries stay in the frontmatter of each file. Part names come from the folder names.

Projects that use maid use the same structure.

A chapter covers one area of the project. A chapter can be 1 to 20 pages. An agent normally reads sections, not whole chapters.

## Requirements

### REQ-DOC-001 Chapter location
The design store shall contain each chapter as exactly one Markdown file at `design/<part>/<chapter>.md`, where the part and chapter names are lowercase, hyphenated and do not start with a digit.

### REQ-DOC-002 Frontmatter
Each chapter shall start with YAML frontmatter that contains the fields `id`, `title`, `status`, `style`, `owner`, `summary` and `links`.

### REQ-DOC-003 Summary length
The `summary` field shall contain at most 60 words.

### REQ-DOC-004 Declared style
Each chapter shall declare exactly one style in the `style` field, and the value shall match a style ID in the style registry.

### REQ-DOC-005 Embedded styles
Where a chapter contains a fenced code block with a language tag that is a registered style ID, maid shall validate that block against that style instead of the chapter style.

### REQ-DOC-006 Stable IDs
Each requirement, decision, work item and open question shall have an ID that is unique in the project and that never changes.

### REQ-DOC-007 No ID reuse
If an item is removed, then maid shall keep its ID reserved and shall not assign the ID to another item.

### REQ-DOC-008 Addressable sections
The design store shall make each level-2 and level-3 heading addressable as `<chapter-id>#<anchor>`, where the anchor is the item ID when the heading starts with an ID.

### REQ-DOC-009 Status values
The `status` field of a chapter or an item shall have one of these values: `stub`, `draft`, `review`, `accepted`, `superseded`.

### REQ-DOC-010 Write rule for open items
While a chapter or an item has status `stub` or `draft`, maid shall accept a direct write to it from any agent.

### REQ-DOC-011 Write rule for accepted items
While a chapter or an item has status `accepted`, maid shall reject a direct write to it and shall accept a proposal instead.

### REQ-DOC-012 Version check
When maid receives a write to a section, maid shall reject the write if the section changed after the version that the writer read.

### REQ-DOC-013 Unknown information
If a writer does not have the information for a required section, then the writer shall write `TBD (Q-NNN)` and shall add question `Q-NNN` to `questions.md`.

### REQ-DOC-014 Reading order
The file `design/index.md` shall contain only the design title as a level-1 heading and a bulleted list of file paths, relative to `design/`, in reading order.

### REQ-DOC-015 Links
Each ID in the `links` field shall resolve to an existing chapter, item or decision.

### REQ-DOC-016 Complete index
maid shall report an error for each chapter in a part folder that `design/index.md` does not list, and for each entry in `design/index.md` whose file does not exist.

### REQ-DOC-017 New chapters in the index
When an agent creates a chapter, maid shall add the path of the chapter to `design/index.md` after the last path in the same part folder, or at the end of the list when no path is in that folder.

### REQ-DOC-018 PDF build
The file `design/index.typ` shall build the design as one PDF with Typst: it shall read the title and the reading order from `design/index.md`, start a part at each new part folder, number the parts with roman numerals and nothing else, include each file with the cmarker package, use a sans-serif font and start with a title page that shows the title at the bottom left.

### REQ-DOC-022 Unnumbered headings
The design store shall contain no chapter or section heading that starts with a position number; numbered lists are only for sequences such as steps, tasks and roadmap milestones.

### REQ-DOC-019 Index files
When maid creates a chapter, maid shall create `design/index.md` and `design/index.typ` if they do not exist, and shall commit them with the chapter.

### REQ-DOC-021 Index is not a chapter
The design store shall not load `design/index.md` as a chapter.

### REQ-DOC-020 Part field
If a chapter has a `part` field, then its value shall equal the name of the part folder that contains the chapter.

## Frontmatter reference

| Field | Type | Meaning |
|---|---|---|
| `id` | string | Chapter ID. Lowercase, hyphenated. Unique. |
| `title` | string | Human title. |
| `part` | string | Optional. Equal to the part folder name (REQ-DOC-020). |
| `status` | enum | See REQ-DOC-009. |
| `style` | string | Style ID from the style registry. |
| `owner` | string | The person who accepts changes. |
| `summary` | string | At most 60 words. What the chapter covers and who needs it. |
| `links` | list | IDs of related chapters, items and decisions. |

## ID prefixes

| Prefix | Item |
|---|---|
| `REQ-<AREA>-NNN` | Requirement. AREA is a short code for the chapter. |
| `ADR-NNNN` | Decision record. |
| `WI-...` | Work item. The task store assigns the suffix. |
| `Q-NNN` | Open question. |
| `G-NNN` | Goal. |
| `NG-NNN` | Non-goal. |

## Status lifecycle

```mermaid
stateDiagram-v2
  [*] --> stub
  stub --> draft
  draft --> review
  review --> accepted
  review --> draft
  accepted --> superseded
```

A human or a configured review agent moves an item from `review` to `accepted`.
