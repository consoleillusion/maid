---
id: ADR-0004
title: The design agent writes complete acceptance tests before build work
part: decisions
status: review
style: ste
owner: paul
origin: agent
preset: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["design agents cannot write useful acceptance tests for a class of work items"]
summary: >
  For each work item, the design agent writes complete, failing acceptance
  tests before any build agent can claim it. Build agents cannot change those
  tests. Build agents can add their own tests, which protect against
  regressions but do not count for acceptance.
links: [work-items, acceptance, phases]
---

# ADR-0004 The design agent writes complete acceptance tests before build work

## Context and problem statement

Acceptance is automatic. If the build agent writes the tests that decide acceptance, it grades its own work. Should the design agent write complete tests, or only test stubs?

## Decision drivers

- Small agents are cheap but unreliable. Acceptance must not depend on their judgement.
- Tests are usually cheaper to write than the implementation.
- A human can review tests more quickly than code.

## Considered options

- Design agent writes complete, failing acceptance tests.
- Design agent writes test stubs. The build agent writes the assertions.
- Build agent writes all tests.

## Decision outcome

Chosen option: "Design agent writes complete, failing acceptance tests", because with stubs, the build agent writes the assertions and so decides its own acceptance.

Rules:

- A work item is ready only when its acceptance tests exist on the main branch and fail (REQ-WI-002).
- A submission that changes its own acceptance tests is rejected (REQ-ACC-010).
- Build agents can add tests. After merge, these tests protect against regressions.
- A human reviews the acceptance tests at the plan gate (phase 5).

### Consequences

- Good, because acceptance does not depend on the build agent.
- Good, because the plan gate gives the human one clear review point.
- Bad, because the design agent does more work, at a higher price per token.
- Bad, because a wrong acceptance test blocks the work item. Escalation handles this (REQ-WI-009, REQ-WI-010).
