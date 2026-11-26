---
id: style-ears
title: "Style card: EARS requirements"
part: styles
status: draft
style: ste
owner: paul
summary: >
  How to write requirements with the Easy Approach to Requirements Syntax.
  Each requirement is one sentence that matches one of six patterns. Use for
  every REQ item.
links: [style-registry]
---

# Style card: EARS requirements

## Form

Each requirement has a heading and one sentence.

```text
### REQ-<AREA>-NNN <short name>
<one EARS sentence>
```

## Patterns

| Pattern | Keyword | Template |
|---|---|---|
| Ubiquitous | none | The `<system>` shall `<response>`. |
| Event-driven | When | When `<trigger>`, the `<system>` shall `<response>`. |
| State-driven | While | While `<state>`, the `<system>` shall `<response>`. |
| Unwanted behaviour | If … then | If `<condition>`, then the `<system>` shall `<response>`. |
| Optional feature | Where | Where `<feature>`, the `<system>` shall `<response>`. |
| Complex | combination | Where …, while …, when …, the `<system>` shall …. |

## Rules

| ID | Rule |
|---|---|
| EARS-1 | Use "shall" for each response. Do not use "should", "may" or "must". |
| EARS-2 | Name one system: `maid`, or a component from the architecture. |
| EARS-3 | Write at most two responses. Join two responses with "and" only when they must happen together. Otherwise write two requirements. |
| EARS-4 | Make each requirement testable. Name a value or an observable result. |
| EARS-5 | Use glossary terms. |
| EARS-6 | Do not explain the reason in the requirement. Put the reason in the chapter text or in a decision. |

## Examples

| Incorrect | Correct |
|---|---|
| maid should be fast when claiming. | When an agent claims a work item, maid shall return the result in less than 1 second. |
| The merge queue rejects bad commits and notifies the agent. | If a check fails, then the merge queue shall reject the submission. |
| Agents must not edit tests. | If a submission changes an acceptance test of its own work item, then the merge queue shall reject the submission. |
