---
id: acceptance
title: Acceptance and the merge queue
part: 02-method
status: accepted
style: ears
owner: paul
summary: >
  How maid decides to accept or reject a submission. Defines the merge queue,
  the checks it runs, the acceptance rule, protection of acceptance tests, and
  the rejection report that the build agent receives.
links: [work-items, scoring, ADR-0003, ADR-0004]
---

# Acceptance and the merge queue

## Rule in one sentence

A submission is accepted when, on top of the latest main branch, its acceptance tests pass, no previously passing test fails, its changes stay in its owned paths, and the configured checks pass.

## Requirements

### REQ-ACC-001 Submit
When a build agent calls `work_submit`, maid shall add the attempt branch to the merge queue.

### REQ-ACC-002 One merge at a time
The merge queue shall process one submission at a time for each main branch.

### REQ-ACC-003 Single queue worker
While more than one maid process runs in the same project, exactly one process shall run the merge queue, and the other processes shall only add submissions to it.

### REQ-ACC-004 Rebase
When the merge queue takes a submission, it shall rebase the attempt branch onto the current main branch in a separate worktree.

### REQ-ACC-005 Conflict
If the rebase has a conflict, then the merge queue shall reject the submission with the conflicting paths.

### REQ-ACC-006 Checks
When the rebase succeeds, the merge queue shall run, in order: the format check, the lint check, the owned-path check, the acceptance-test protection check, the acceptance tests of the work item, and the full test suite.

### REQ-ACC-007 Accept
If all checks pass and no test that passed on the main branch fails, then the merge queue shall fast-forward the main branch to the rebased attempt branch.

### REQ-ACC-008 Reject
If a check fails, then the merge queue shall reject the submission and shall keep the main branch unchanged.

### REQ-ACC-009 Rejection report
When the merge queue rejects a submission, it shall return a rejection report that contains the failed check, the failed tests with their output truncated to a limit from the profile, the regressions, and the score.

### REQ-ACC-010 Protected acceptance tests
If a submission changes, deletes or skips an acceptance test of its own work item, then the merge queue shall reject the submission.

### REQ-ACC-011 Test baseline
The merge queue shall record the result of each test on the main branch after each merge, and shall use that record to detect regressions.

### REQ-ACC-012 Flaky tests
If a test result differs between two runs on the same commit, then maid shall mark the test as flaky and shall exclude it from regression detection.

### REQ-ACC-016 Flaky test question
When maid marks a test as flaky, maid shall add a question for the design agent that names the test.

### REQ-ACC-013 Competing attempts
Where more than one attempt on the same work item passes all checks, the merge queue shall accept the attempt with the highest score and shall reject the others as superseded.

### REQ-ACC-014 Review band
Where the depth level is rigorous, if the score of a passing submission is inside the review band of the profile, then the merge queue shall hold the submission for human approval.

### REQ-ACC-015 Record
When the merge queue accepts or rejects a submission, maid shall record the work item, the attempt, the agent, the checks, the score and the commit IDs in the state store.

### REQ-ACC-017 Attempt environment
When maid runs the tests of an attempt, maid shall set the environment variables `MAID_ATTEMPT_ID`, `MAID_PORT_BASE` and `MAID_TMPDIR` to values that are unique to that attempt.

### REQ-ACC-018 Serial tests
Where the project decisions mark a test suite as serial, maid shall not run that suite for two attempts at the same time.

## Examples

```gherkin
Feature: Merge queue acceptance

  Scenario: Passing submission is merged
    Given work item "WI-a1" has acceptance test "TestClaimIsAtomic"
    And "TestClaimIsAtomic" fails on main
    When the build agent submits an attempt where "TestClaimIsAtomic" passes
    And no other test changes from pass to fail
    Then the merge queue fast-forwards main to the attempt

  Scenario: Regression is rejected
    Given test "TestReadyQueue" passes on main
    When the build agent submits an attempt where "TestReadyQueue" fails
    Then the merge queue rejects the attempt
    And the rejection report lists "TestReadyQueue" as a regression

  Scenario: Changed acceptance test is rejected
    Given work item "WI-a1" has acceptance test "TestClaimIsAtomic"
    When the build agent submits an attempt that changes "TestClaimIsAtomic"
    Then the merge queue rejects the attempt
```
