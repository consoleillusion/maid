---
id: style-sbvr-se
title: "Style card: SBVR Structured English"
part: styles
status: draft
style: ste
owner: paul
summary: >
  How to write vocabulary and business rules in SBVR Structured English, adapted
  to Markdown. Use for the glossary and for chapters that define business rules.
links: [style-registry, glossary]
---

# Style card: SBVR Structured English

SBVR Structured English uses fonts to mark the parts of a statement. maid adapts them to Markdown.

| SBVR element | Markdown | Example |
|---|---|---|
| Term (noun concept) | `**term**` | **work item** |
| Verb (fact type) | `_verb_` | _implements_ |
| Name (individual) | `"Name"` | "web-local" |
| Keyword | plain text | It is obligatory that, each |

## Definitions

```text
**<term>**: <definition that uses only defined terms and ordinary words>.
```

## Rule keywords

| Keyword | Meaning |
|---|---|
| It is obligatory that … | Behavioural rule: must be true. |
| It is prohibited that … | Behavioural rule: must not be true. |
| It is permitted that … | Explicit permission. |
| It is necessary that … | Definitional rule: always true by definition. |
| It is possible that … | Definitional possibility. |
| each, some, at least one, at most one, exactly one, at least N | Quantifiers. |

## Rules

| ID | Rule |
|---|---|
| SBVR-1 | Define each term before a rule uses it. |
| SBVR-2 | Start each rule with a rule keyword. |
| SBVR-3 | Use only defined terms in bold. |
| SBVR-4 | Use a quantifier before each term in a rule. |
| SBVR-5 | Write one rule in each list item. |

## Examples

| Incorrect | Correct |
|---|---|
| Work items need tests. | It is obligatory that each **work item** _has_ at least one **acceptance test**. |
| Agents can't share a claim. | It is prohibited that two **agents** _hold_ a **claim** _on_ the same **work item** at the same time. |
