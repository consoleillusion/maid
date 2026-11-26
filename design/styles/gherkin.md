---
id: style-gherkin
title: "Style card: Gherkin"
part: styles
status: draft
style: ste
owner: paul
summary: >
  How to write acceptance criteria and examples in Gherkin. Use in fenced
  gherkin blocks inside chapters, and in feature files that the design agent
  writes as acceptance tests.
links: [style-registry, acceptance]
---

# Style card: Gherkin

Write Gherkin in a fenced block with the tag `gherkin`. maid parses the block with a Gherkin parser.

## Form

```gherkin
Feature: <capability>

  Scenario: <one behaviour>
    Given <state before>
    And <more state>
    When <one action>
    Then <observable result>
    And <more results>
```

## Rules

| ID | Rule |
|---|---|
| GH-1 | Write one behaviour in each scenario. |
| GH-2 | Write exactly one `When` step in each scenario. |
| GH-3 | Write results that a test can observe. Do not describe internal state. |
| GH-4 | Use glossary terms and concrete example values in quotes. |
| GH-5 | Link the scenario to a requirement with a tag, for example `@REQ-ACC-007`. |
| GH-6 | Use `Scenario Outline` with `Examples` for the same behaviour with different values. |

## Example

```gherkin
@REQ-WI-004
Scenario: Two agents claim the same work item
  Given work item "WI-a1" is ready
  When agent "A" and agent "B" claim "WI-a1" at the same time
  Then exactly one agent holds the claim on "WI-a1"
  And the other agent receives a conflict
```
