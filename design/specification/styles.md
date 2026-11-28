---
status: draft
style: ste
owner: paul
summary: >
  The writing standards that chapters can declare, their rules, examples and
  validators. Read the card for a style before you write in that style.
links: [format]
---

# Writing styles

Each style below is registered in the design format. A chapter declares one style in its frontmatter. A fenced code block can declare another registered style with its language tag.

## Simplified Technical English

Based on ASD-STE100. maid checks the rules marked "checked". The full STE dictionary is not included. Where the project supplies an approved word list, maid also checks words.

### Rules

| Rule | Checked |
|---|---|
| Write procedural sentences with at most 20 words. | yes |
| Write descriptive sentences with at most 25 words. | yes |
| Write one instruction in each sentence. | yes |
| Use the active voice. | yes, heuristic |
| Use the imperative for instructions. | no |
| Use one word for one meaning. Use the terms in the glossary. | yes, for glossary terms |
| Write at most 6 sentences in a paragraph. Put one topic in each paragraph. | yes |
| Use simple tenses: present, past, future. | no |
| Do not use phrasal verbs when a single verb exists. | no |
| Write notes and warnings as separate sentences, not in brackets. | no |

### Examples

| Incorrect | Correct |
|---|---|
| The lease is renewed by the agent before it expires. | The agent renews the lease before it expires. |
| Set up the worktree and then run the tests, making sure nothing fails. | Create the worktree. Run the tests. Make sure that no test fails. |
| Work items (also called tasks or jobs) are claimed atomically. | One agent at a time can claim a work item. |

## EARS requirements

Each requirement has a heading and one sentence. The heading names the behaviour. The sentence uses one of the patterns in the design format.

```text
### <short name>
<one EARS sentence>
```

### Rules

- Use "shall" for each response. Do not use "should", "may" or "must".
- Name one system: `maid`, or a component from the architecture.
- Write at most two responses. Join two responses with "and" only when they must happen together. Otherwise write two requirements.
- Make each requirement testable. Name a value or an observable result.
- Use glossary terms.
- Do not explain the reason in the requirement. Put the reason in the chapter text or in a decision.

### Examples

| Incorrect | Correct |
|---|---|
| maid should be fast when claiming. | When an agent claims a work item, maid shall return the result in less than 1 second. |
| The merge queue rejects bad commits and notifies the agent. | If a check fails, then the merge queue shall reject the submission. |
| Agents must not edit tests. | If a submission changes an acceptance test of its own work item, then the merge queue shall reject the submission. |

## SBVR Structured English

SBVR Structured English uses fonts to mark the parts of a statement. maid adapts them to Markdown.

| SBVR element | Markdown | Example |
|---|---|---|
| Term (noun concept) | `**term**` | **work item** |
| Verb (fact type) | `_verb_` | _implements_ |
| Name (individual) | `"Name"` | `"web-local"` |
| Keyword | plain text | It is obligatory that, each |

### Definitions

```text
**<term>**: <definition that uses only defined terms and ordinary words>.
```

### Rule keywords

| Keyword | Meaning |
|---|---|
| It is obligatory that … | Behavioural rule: must be true. |
| It is prohibited that … | Behavioural rule: must not be true. |
| It is permitted that … | Explicit permission. |
| It is necessary that … | Definitional rule: always true by definition. |
| It is possible that … | Definitional possibility. |
| each, some, at least one, at most one, exactly one | Quantifiers. |

### Rules

- Define each term before a rule uses it.
- Start each rule with a rule keyword.
- Use only defined terms in bold.
- Use a quantifier before each term in a rule.
- Write one rule in each list item.

### Examples

| Incorrect | Correct |
|---|---|
| Work items need tests. | It is obligatory that each **work item** _has_ at least one **acceptance test**. |
| Agents can't share a claim. | It is prohibited that two **agents** _hold_ a **claim** _on_ the same **work item** at the same time. |

## Gherkin

Write Gherkin in a fenced block with the language tag `gherkin`. maid parses the block with a Gherkin parser.

```gherkin
Feature: <capability>

  Scenario: <one behaviour>
    Given <state before>
    And <more state>
    When <one action>
    Then <observable result>
    And <more results>
```

### Rules

- Write one behaviour in each scenario.
- Write exactly one `When` step in each scenario.
- Write results that a test can observe. Do not describe internal state.
- Use glossary terms and concrete example values in quotes.
- Name the scenario after the requirement it covers.
- Use `Scenario Outline` with `Examples` for the same behaviour with different values.

### Example

```gherkin
@atomic-claim
Scenario: Two agents claim the same work item
  Given work item "first" is ready
  When agent "A" and agent "B" claim "first" at the same time
  Then exactly one agent holds the claim on "first"
  And the other agent receives a conflict
```

## Attempto Controlled English

ACE is a subset of English that the APE parser translates into first-order logic. A sentence that APE parses has one meaning.

This card is a stub. The APE adapter is planned after v1. Until it exists, findings for this style are advice only.

### Rules

- Write simple declarative sentences: subject, verb, object.
- Use "every", "a", "no", "if … then" for quantification and conditions.
- Use a noun once with "a" to introduce it, then "the" to refer to it.
- Do not use pronouns that can refer to more than one noun.

### Example

```ace
Every work item that is ready has no open dependency.
If an agent claims a work item then no other agent claims the work item.
```
