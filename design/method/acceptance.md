---
status: accepted
style: ears
owner: paul
summary: >
  How maid decides to accept or reject a submission. Defines the merge queue,
  the checks it runs, the acceptance rule, protection of acceptance tests, and
  the rejection report that the build agent receives.
links: [work-items, scoring, merge-queue, acceptance-tests-first]
---

# Acceptance and the merge queue

## Rule in one sentence

A submission is accepted when, on top of the latest main branch, its acceptance tests pass, no previously passing test fails, its changes stay in its owned paths, and the configured checks pass.

## Requirements

### Submit
When a build agent calls `work_submit`, maid shall add the commits in the attempt worktree to the merge queue.

### One merge at a time
The merge queue shall process one submission at a time for each main branch.

### Single queue worker
While more than one maid process runs in the same project, exactly one process shall run the merge queue, and the other processes shall only add submissions to it.

### Rebase
When the merge queue takes a submission, it shall rebase the attempt branch onto the current main branch in a separate worktree.

### Conflict
If the rebase has a conflict, then the merge queue shall reject the submission with the conflicting paths.

### Checks
When the rebase succeeds, the merge queue shall run, in order: the format check, the lint check, the owned-path check, the acceptance-test protection check, the acceptance tests of the work item, and the full test suite.

### Accept
If all checks pass and no test that passed on the maid branch fails, then the merge queue shall add the attempt to the maid branch as one commit whose message ends with the trailer `Maid-Task: <work item ID>`.

### One branch
maid shall commit every accepted attempt and every design change to one branch, the maid branch, whose name is the `branch` value in `maid.json` and is `maid-auto` by default.

### Maid branch creation
If the maid branch does not exist, then maid shall create it at the current commit of the checkout.

### Design commits
When a design change starts and the checkout is not on the maid branch, maid shall switch the checkout to the maid branch if no tracked file has uncommitted changes, and otherwise shall refuse the change and return the command that switches the branch.

### Old attempt branches
When `maid init` runs, maid shall delete the branches that match `maid/*` and remove the worktrees of finished attempts.

### Reject
If a check fails, then the merge queue shall reject the submission and shall keep the main branch unchanged.

### Rejection report
When the merge queue rejects a submission, it shall return a rejection report that contains the failed check, the failed tests with their output truncated to a limit from the profile, the regressions, and the score.

### Protected acceptance tests
If a submission changes, deletes or skips an acceptance test of its own work item, then the merge queue shall reject the submission.

### Test baseline
The merge queue shall record the result of each test on the main branch after each merge, and shall use that record to detect regressions.

### Flaky tests
If a test result differs between two runs on the same commit, then maid shall mark the test as flaky and shall exclude it from regression detection.

### Flaky test report
When maid marks a test as flaky, maid shall report the test through `maid_next` and shall record it in the state store.

### Competing attempts
Where more than one attempt on the same work item passes all checks, the merge queue shall accept the attempt with the highest score and shall reject the others as superseded.

### Review band
Where the depth level is rigorous, if the score of a passing submission is inside the review band of the profile, then the merge queue shall hold the submission for human approval.

### Record
When the merge queue accepts or rejects a submission, maid shall record the work item, the attempt, the agent, the checks, the score and the commit IDs in the state store.

### Attempt environment
When maid runs the tests of an attempt, maid shall set the environment variables `MAID_ATTEMPT_ID`, `MAID_PORT_BASE` and `MAID_TMPDIR` to values that are unique to that attempt.

### Serial tests
Where the project decisions mark a test suite as serial, maid shall not run that suite for two attempts at the same time.

## Examples

```gherkin
Feature: Merge queue acceptance

  Scenario: Passing submission is merged
    Given work item "atomic claim" has acceptance test "TestClaimIsAtomic"
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
    Given work item "atomic claim" has acceptance test "TestClaimIsAtomic"
    When the build agent submits an attempt that changes "TestClaimIsAtomic"
    Then the merge queue rejects the attempt
```

## Environment errors

### Missing tools
If a check cannot run because its command does not exist or cannot start, then the merge queue shall report an environment error with the command and the system error, and shall not count the attempt.

### Tool check at start
When maid starts, maid shall check that the configured test, format and lint commands exist, and maid_next and maid_status shall report each missing command as a warning.

### Redefined acceptance tests
When the merge queue checks a tests item for regressions, it shall not count the acceptance tests of that item, which fail until the code item is done.
