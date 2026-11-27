---
id: ADR-0008
title: Everything is a task, and checks guide instead of block
part: decisions
status: accepted
style: ste
owner: paul
summary: >
  All work in a project is a Beads task, including the design. maid init seeds the first tasks: record the current state, record the target state, write the design, plan the tasks and set up the project. Any agent picks up the tasks that are left. Metrics, formatting, tests, documentation and code graphs guide agents; they do not slow development.
links: [phases, work-items, roadmap, vision]
---

# ADR-0008 Everything is a task, and checks guide instead of block

## Context and problem statement

maid must be easy to use, make development cheaper, and give agents a structured progression. A fixed sequence of phases with gates makes the first session long, and it forces one agent to finish each phase. How does maid give structure without that cost?

## Decision drivers

- The user prompts the agent in plain words. The user does not need special syntax.
- An agent can stop at any point. The next agent continues from where the work stopped.
- The roadmap and the requirements can change during the project.
- Checks show agents what a good commit is. Checks must not slow development.

## Decision outcome

All work is a task in Beads. This includes design work.

1. `maid init` seeds the first tasks: record the current state, record the target state, write the design, plan the tasks, and set up the project.
2. The first agent gets these tasks. A capable agent is recommended for them, but maid does not require one.
3. An agent can stop at any time. It can leave a task half done or add new tasks. The next agent picks up the tasks that are left.
4. `maid_next` lists the tasks that are ready: none, one or several. The server instructions tell agents to call `maid_next` when the user asks them to work on the project.
5. Phases stay as the order of the seeded tasks and their dependencies. They are guidance, not gates.
6. Agents can change the roadmap and the tasks when the target changes.

## Guidance, not gates

Metrics, automatic formatting, tests, documentation and code graphs exist to show agents what a successful commit is. Formatting makes line counts deterministic and harder to game. The merge queue still accepts a commit only when its tests pass, no other test breaks and its score is not below the minimum. maid adds no other step that slows development.

## Languages

maid is language independent. Metrics come from adapters. Where no adapter exists for a language, maid uses language-independent measures: changed lines from git and dependency changes from known manifest files. Adapters for more languages are added one at a time.

## Consequences

- Good, because the user only needs plain prompts.
- Good, because work continues across agents and sessions without loss.
- Bad, because design tasks need a check that shows when they are done. maid uses the design check and the presence of the required chapters.
