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
revisit-when: ["a project needs human review of every merge", "teams share one main branch on a forge"]
summary: >
  maid accepts a submission when its acceptance tests pass on the latest main
  branch and no previously passing test fails. A local merge queue does this.
  No forge or pull request is needed. The score ranks passing attempts.
links: [acceptance, scoring, vision]
---

# Accept submissions automatically through a local merge queue

## Context and problem statement

Many agents submit commits. Who decides which commits reach the main branch, and how?

## Decision drivers

- Acceptance must be mechanical, not the opinion of an agent.
- No network service is required.
- Parallel submissions must not break each other after merge.

## Considered options

- Local merge queue in maid.
- Pull requests on a forge (Forgejo, GitHub), with CI.
- Agents merge directly after their own test run.

## Decision outcome

Chosen option: "Local merge queue", because the acceptance rule is mechanical, so pull requests add no value in the core. The queue rebases each submission on the latest main branch before it runs the tests. This catches breakage between parallel work items.

The score does not accept. It ranks attempts that pass, and it can hold a submission for human review when the score is in the review band.

### Consequences

- Good, because no forge is needed.
- Good, because a submission that passes alone but breaks after another merge is caught.
- Bad, because merges are serial. Throughput depends on test suite time.
- Neutral, because a forge adapter can still open pull requests where a project requires them .
