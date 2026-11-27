---
id: work-items
title: Work items, claims and context packets
part: 02-method
status: accepted
style: ears
owner: paul
summary: >
  What a work item contains, how agents claim it, how leases and escalation
  work, and what maid puts in the context packet that a build agent receives.
  Read this before you plan or take work.
links: [phases, acceptance, task-store, ADR-0004, ADR-0005]
---

# Work items, claims and context packets

## Work item fields

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Assigned by the task store. |
| `title` | yes | One line. |
| `implements` | yes | Requirement IDs. |
| `acceptance_tests` | yes | Test IDs or test paths. Written before any attempt. |
| `owns` | yes | File path patterns that this work item can change. |
| `reads` | no | File path patterns that the agent can read but not change. |
| `depends_on` | no | Work item IDs. |
| `budget` | yes | Maximum attempts and maximum changed lines. From the profile. |
| `kind` | yes | `code`, `tests` or `docs`. |
| `criteria` | for `code` | Gherkin scenarios written by the design agent. |
| `tests_for` | for `tests` | The `code` work item whose acceptance tests this item writes. |
| `status` | yes | `open`, `claimed`, `submitted`, `accepted`, `rejected`, `escalated`, `closed`. |

## Requirements

### REQ-WI-001 Size
The design agent shall create work items that a build agent can complete in one session, with a changed-line budget from the profile.

### REQ-WI-002 Acceptance tests first
The task store shall not offer a work item as ready until its acceptance tests exist on the main branch and fail.

### REQ-WI-003 Ready work
When an agent asks for work, maid shall offer only work items that have no open dependency and no owned path in common with a claimed work item.

### REQ-WI-004 Atomic claim
When two agents claim the same work item at the same time, maid shall give the claim to exactly one agent and shall return a conflict to the other agent.

### REQ-WI-005 Lease
While an agent holds a claim, the agent shall renew the lease before it expires.

### REQ-WI-006 Expired lease
If a lease expires, then maid shall return the work item to ready and shall record the attempt as abandoned.

### REQ-WI-007 Isolated workspace
When an agent claims a work item, maid shall create a git worktree on a new branch from the current main branch for that attempt.

### REQ-WI-008 Owned paths
If a submission changes a file outside the `owns` patterns of the work item, then the merge queue shall reject the submission.

### REQ-WI-009 Escalation
If a work item has the number of rejected attempts that its budget allows, then maid shall set the work item to `escalated` and shall offer it to a design agent.

### REQ-WI-010 Escalation result
When a design agent takes an escalated work item, the design agent shall split the work item, clarify the linked items, or correct the acceptance tests through a proposal.

### REQ-WI-011 Parallel attempts
Where the profile enables parallel attempts, maid shall allow up to N attempts on the same work item at the same time, and the merge queue shall accept only the attempt with the highest score.

### REQ-WI-012 Docs work items
Where a work item has kind `docs`, its acceptance tests shall be the style validator, the required sections and the link check of the target chapter.

### REQ-WI-016 Cost record
When an agent submits an attempt, maid shall record the attempt count, the elapsed time and any cost that the agent reports, and maid shall limit work items by attempts only.

### REQ-WI-017 Test work item
When the design agent creates a work item of kind `code`, maid shall create a work item of kind `tests` that writes its acceptance tests from its acceptance criteria.

### REQ-WI-018 Order
The task store shall make each `code` work item depend on its `tests` work item.

### REQ-WI-019 Independence
If an agent held the claim on a `tests` work item, then maid shall not give that agent a claim on the `code` work item that it tests.

### REQ-WI-020 Tests acceptance
Where a work item has kind `tests`, its acceptance checks shall be: the tests compile, each test fails on the main branch, and each acceptance criterion is covered by at least one test that carries the criterion tag.

### REQ-WI-021 Review of tests
Where the depth level is rigorous, the merge queue shall hold each accepted `tests` submission for review by the design agent or a human.

## Context packet

### REQ-WI-013 Packet contents
When an agent claims a work item, maid shall return a context packet that contains:

- the work item fields,
- the text of each linked requirement and its section,
- the contracts that the owned code implements or calls,
- the acceptance test files,
- the list of owned and readable paths,
- the style cards and conventions that apply,
- the commands to format, lint and test.

### REQ-WI-014 Packet size
The context packet shall not include whole chapters when a section is enough, and maid shall report the size of the packet in tokens.

### REQ-WI-015 No browsing needed
The context packet shall contain enough information for a build agent to complete the work item without reading other parts of the design.

## Lifecycle

```mermaid
stateDiagram-v2
  [*] --> open
  open --> claimed: claim
  claimed --> open: lease expired
  claimed --> submitted: submit
  submitted --> accepted: merge queue accepts
  submitted --> rejected: merge queue rejects
  rejected --> open: attempts remain
  rejected --> escalated: budget used
  escalated --> closed: split or replaced
  accepted --> closed
```

## Test stubs

In compiled languages, a test that calls code that does not exist does not compile. A stub makes the tests compile and fail. A stub cannot implement the behaviour, because REQ-WI-020 rejects tests that pass on the main branch.

### REQ-WI-022 Stubs
Where the code that the acceptance tests call does not exist on the main branch, the test agent shall add stubs with the names and signatures from the criteria, with bodies that return zero values.

### REQ-WI-023 Tests item ownership
When maid creates a tests work item, maid shall give it ownership of its test files and of the owned files of its code item.

## Work through maid only

An agent can complete a work item with maid's tools alone. It does not need file or shell access on the machine that runs maid.

### REQ-WI-024 Read worktree files
When an agent that holds the claim on a work item asks for a file in its worktree, maid shall return the file content.

### REQ-WI-025 Write owned files
When an agent that holds the claim on a work item writes a file, maid shall write the file in the agent's worktree only if the work item owns the path.

### REQ-WI-026 Run checks
When an agent that holds the claim on a work item asks to run the tests, the format check or the lint, maid shall run the configured command in the agent's worktree and return the result.

### REQ-WI-027 Commit
When an agent that holds the claim on a work item asks to commit, maid shall commit every change in the agent's worktree on its attempt branch.

### REQ-WI-028 Submit a patch
When an agent submits a work item with a patch or a set of files, maid shall apply them in the agent's worktree, commit them and run the merge queue, and shall refuse a patch that changes a path that the work item does not own.
