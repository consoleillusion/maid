---
id: style-registry
title: Style registry
part: specification
status: accepted
style: ears
owner: paul
summary: >
  The writing standards that chapters can declare, the style card and validator
  for each, and how maid enforces them. Read the style card of a chapter before
  you write in it.
links: [chapter-format, ADR-0006]
---

# Style registry

## Registered styles

| Style ID | Standard | Use for | Card | Validator | v1 enforcement |
|---|---|---|---|---|---|
| `ste` | ASD-STE100 Simplified Technical English | descriptive prose, procedures | `styles/ste.md` | built-in: sentence length, voice, one instruction per sentence, glossary terms | error |
| `ears` | EARS (Easy Approach to Requirements Syntax) | requirements | `styles/ears.md` | built-in: pattern match on each requirement | error |
| `sbvr-se` | SBVR Structured English | vocabulary, business rules | `styles/sbvr-se.md` | built-in: keyword and term check | error |
| `gherkin` | Gherkin | acceptance criteria, examples | `styles/gherkin.md` | Gherkin parser | error |
| `ace` | Attempto Controlled English | statements that need logical precision | `styles/ace.md` | external: APE parser | advice until the adapter exists |

Formalized English (Martin 2002) is not registered, because it has no validator that maid can run (Q-007).

## Requirements

### REQ-STY-001 Registered styles only
maid shall reject a chapter whose `style` value is not a registered style ID.

### REQ-STY-002 Style card
Each registered style shall have a style card that states the rules, gives correct and incorrect examples, and fits in 1500 tokens.

### REQ-STY-003 Validator contract
Each validator shall take a Markdown file or a code block as input and shall return a list of findings, where each finding has a location, a rule ID, a severity and a message.

### REQ-STY-004 Enforcement levels
Each registered style shall have an enforcement level of `error` or `advice`, and only findings from styles at level `error` shall fail a gate or a docs work item.

### REQ-STY-005 Card in context packet
When a context packet includes a section, maid shall include the style card of the chapter style and of each embedded style in that section.

### REQ-STY-006 Scope of a validator
A validator shall check only the content that its style governs: the chapter body for the chapter style, and fenced blocks for embedded styles. Frontmatter, tables and code blocks of other languages are excluded.

### REQ-STY-007 New styles
When a user adds a style, maid shall require a style card and an enforcement level, and shall set the enforcement level to `advice` until a validator is configured.

## Validator rules for EARS requirements

An EARS requirement is the first paragraph after a heading that starts with `REQ-`. It must match one of these patterns. The `<system>` is a named component or `maid`.

| Pattern | Template |
|---|---|
| Ubiquitous | The `<system>` shall `<response>`. |
| Event-driven | When `<trigger>`, the `<system>` shall `<response>`. |
| State-driven | While `<state>`, the `<system>` shall `<response>`. |
| Unwanted behaviour | If `<condition>`, then the `<system>` shall `<response>`. |
| Optional feature | Where `<feature>`, the `<system>` shall `<response>`. |
| Complex | A combination of the keywords above, in the order Where, While, When or If. |
