---
id: ADR-0004
title: Acceptance criteria from the design agent, tests from a separate test agent
part: decisions
status: review
style: ste
owner: paul
origin: agent
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["test agents often write tests that pass wrong implementations", "the extra work item per code item costs more than it saves"]
summary: >
  The design agent writes short Gherkin acceptance criteria for each work
  item. A separate, cheap test agent turns them into failing tests. A third
  agent implements the work item and cannot change those tests. Small models do
  the volume of test code; independence keeps acceptance honest.
links: [work-items, acceptance, phases]
---

# ADR-0004 Acceptance criteria from the design agent, tests from a separate test agent

## Context and problem statement

Acceptance is automatic, so the tests decide what is accepted. Test code is often as large as the implementation. Small models can write tests well. But if the agent that implements a work item also writes its acceptance tests, it grades its own work. Who writes the acceptance tests?

## Decision drivers

- The implementer must not define its own acceptance.
- Large models are expensive, so they should write as little code as possible.
- Small models can write test code reliably from a precise scenario.

## Considered options

- The design agent writes complete acceptance tests.
- The design agent writes Gherkin acceptance criteria. A separate test agent writes the tests. A different build agent implements.
- The build agent writes its own tests.

## Decision outcome

Chosen option: "The design agent writes Gherkin acceptance criteria. A separate test agent writes the tests. A different build agent implements."

The large agent writes only what needs judgement: the scenarios, which are short. A cheap agent writes the test code. Independence comes from the separation of agents, not from the size of the model.

Rules:

- Each `code` work item has a `tests` work item, and depends on it (REQ-WI-017, REQ-WI-018).
- The agent that wrote the tests cannot implement the work item (REQ-WI-019).
- A `tests` submission is accepted when the tests compile, each test fails on the main branch, and each scenario is covered by a tagged test (REQ-WI-020).
- A build submission that changes its acceptance tests is rejected (REQ-ACC-010).
- At the rigorous depth level, the design agent or a human reviews the tests (REQ-WI-021).

### Consequences

- Good, because the large agent writes much less code.
- Good, because acceptance does not depend on the implementer.
- Bad, because each `code` work item adds a `tests` work item and one more merge.
- Bad, because a test that passes a wrong implementation is not caught mechanically. Scenario tags and rigorous review reduce this risk.
