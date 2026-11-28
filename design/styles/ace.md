---
status: stub
style: ste
owner: paul
summary: >
  Attempto Controlled English for statements that need logical precision. The
  validator is the external APE parser. Enforcement is advice until the APE
  adapter exists. No v1 chapter uses this style.
links: [style-registry]
---

# Style card: Attempto Controlled English (ACE)

ACE is a subset of English that the APE parser translates into first-order logic. A sentence that APE parses has one meaning.

## Status

This card is a stub. The APE adapter is planned after v1 (Q-007). Until it exists, findings for this style are advice only.

## Basic rules

| ID | Rule |
|---|---|
| ACE-1 | Write simple declarative sentences: subject, verb, object. |
| ACE-2 | Use "every", "a", "no", "if … then" for quantification and conditions. |
| ACE-3 | Use a noun once with "a" to introduce it, then "the" to refer to it. |
| ACE-4 | Do not use pronouns that can refer to more than one noun. |

## Example

```ace
Every work item that is ready has no open dependency.
If an agent claims a work item then no other agent claims the work item.
```
