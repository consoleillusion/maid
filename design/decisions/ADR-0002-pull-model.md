---
id: ADR-0002
title: Agents pull work from maid; maid does not start agents
part: decisions
status: accepted
style: ste
owner: paul
origin: user
profile: null
date: 2026-11-27
decision-makers: [paul]
supersedes: null
superseded-by: null
revisit-when: ["users need maid to run a project without a human starting agents"]
summary: >
  maid is a pull-based MCP server. Agents connect, call maid_next and claim
  work. Parallel work means more agents connect. An orchestrator that starts
  agents is out of scope for v1.
links: [vision, constraints, architecture]
---

# ADR-0002 Agents pull work from maid; maid does not start agents

## Context and problem statement

An MCP server cannot start agents. It only answers calls from clients. How does maid get parallel work done?

## Decision drivers

- maid must work with any MCP client.
- Starting agents needs API keys, cost limits and process control.
- v1 is for one developer on one machine.

## Considered options

- Pull: maid serves claimable work. Something outside maid starts agents.
- Push: maid starts agents through model APIs.
- Both, push later.

## Decision outcome

Chosen option: "Pull", because it works with every client and keeps maid small. A user or a script starts N agents and points each at maid. Push can be added later as a separate component.

### Consequences

- Good, because any MCP client works.
- Good, because maid holds no API keys.
- Bad, because maid cannot measure token cost. Agents can report cost when they submit.
- Bad, because each client may start its own maid process, so all shared state must be safe for many processes. See C-007.
