---
status: accepted
style: sbvr-se
owner: paul
summary: >
  The defined terms of maid and the business rules that connect them. Every
  chapter uses these terms with these meanings. If a term is not here, add it
  before you use it in a requirement.
links: [vision, format]
---

# Glossary and business rules

Notation: **term** is a noun concept. _verb phrase_ is a fact type. Keywords are in plain text. See the style chapter at `specification/styles.md#sbvr-structured-english`.

## Terms

**maid**: the tool that this design specifies. Its name is always written in lowercase.

**project**: a software repository in which maid is installed.

**design**: the set of all **chapters** in the `design/` folder of a **project**.

**part**: a folder in the **design** that groups **chapters**.

**chapter**: one Markdown file in a **part** that covers one area of the **project**.

**section**: a level-2 or level-3 heading in a **chapter**, with its content.

**requirement**: a statement in a **chapter** that specifies a behaviour or a quality that the **project** must have.

**decision**: a **chapter** in `design/decisions/` that records a choice, its context and its consequences.

**default decision**: a **decision** that maid makes from a **profile** because the user did not make the choice.

**style**: a named writing standard, such as Simplified Technical English or EARS.

**style card**: a **section** in `specification/styles.md` that describes a **style** for writers.

**validator**: a program that checks text against a **style**.

**profile**: a named, versioned set of defaults for **decisions**, **styles**, tools and **score** weights.

**phase**: a stage of the progression of a **project**.

**gate**: a set of conditions that maid checks before a **project** leaves a **phase**.

**agent**: an AI agent or a human that calls maid.

**design agent**: an agent that writes the **design**, the **work items** and the **acceptance criteria**.

**test agent**: an agent that writes the **acceptance tests** of a **work item** from its **acceptance criteria**.

**build agent**: an agent that implements a **work item**.

**work item**: a unit of work that is small enough for one **build agent** in one session.

**claim**: the exclusive right of one agent to work on one **work item**.

**lease**: the time limit of a **claim**. An agent renews it with a heartbeat.

**context packet**: the set of **sections**, contracts, files and **style cards** that maid gives to an agent for one **work item**.

**attempt**: one try by one **build agent** to complete one **work item**.

**submission**: the git commit or commits of one **attempt**.

**merge queue**: the maid component that tests each **submission** on the latest main branch and merges or rejects it.

**acceptance criterion**: a Gherkin scenario that a **design agent** writes for a **work item**.

**acceptance test**: a test that a **test agent** writes from an **acceptance criterion** before any **attempt** on the **work item**.

**regression**: a test that passed on the main branch and fails after a **submission**.

**baseline**: the state of the main branch when an **attempt** starts.

**score**: a number from 0 to 100 that maid calculates for an **attempt** from weighted measures.

**adapter**: a component that connects maid to an external tool through a defined interface.

**port**: the interface that an **adapter** implements.

## Fact types

- **chapter** _is in_ **part**
- **chapter** _uses_ **style**
- **work item** _implements_ **requirement**
- **work item** _has_ **acceptance criterion**
- **acceptance test** _implements_ **acceptance criterion**
- **work item** _has_ **acceptance test**
- **work item** _depends on_ **work item**
- **work item** _owns_ file path
- **agent** _holds_ **claim** _on_ **work item**
- **attempt** _is for_ **work item**
- **submission** _is of_ **attempt**
- **default decision** _comes from_ **profile**

## Business rules

- It is obligatory that each **chapter** _uses_ exactly one **style**.
- It is obligatory that each **work item** _implements_ at least one **requirement**.
- It is obligatory that each **work item** _has_ at least one **acceptance test** before a **build agent** _holds_ a **claim** _on_ that **work item**.
- It is prohibited that two **agents** _hold_ a **claim** _on_ the same **work item** at the same time.
- It is prohibited that the **test agent** that writes the **acceptance tests** of a **work item** _holds_ a **claim** _on_ that **work item**.
- It is prohibited that a **build agent** changes an **acceptance test** of the **work item** that the **build agent** implements.
- It is obligatory that maid rejects each **submission** that causes a **regression**.
- It is prohibited that two **work items** that run at the same time _own_ the same file path.
- It is obligatory that each **default decision** states the **profile** that it _comes from_.
- It is obligatory that an **agent** asks the user instead of guessing when information is missing.
