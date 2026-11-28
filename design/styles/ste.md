---
status: draft
style: ste
owner: paul
summary: >
  How to write descriptive text and procedures in Simplified Technical English,
  based on ASD-STE100. Use for vision, constraints, architecture and other prose
  chapters.
links: [style-registry]
---

# Style card: Simplified Technical English (STE)

Based on ASD-STE100. maid checks the rules marked "checked". The full STE dictionary is not included. Where the project supplies an approved word list, maid also checks words.

## Rules

| ID | Rule | Checked |
|---|---|---|
| STE-1 | Write procedural sentences with at most 20 words. | yes |
| STE-2 | Write descriptive sentences with at most 25 words. | yes |
| STE-3 | Write one instruction in each sentence. | yes |
| STE-4 | Use the active voice. | yes (heuristic) |
| STE-5 | Use the imperative for instructions. | no |
| STE-6 | Use one word for one meaning. Use the terms in the glossary. | yes, for glossary terms |
| STE-7 | Write at most 6 sentences in a paragraph. Put one topic in each paragraph. | yes |
| STE-8 | Use simple tenses: present, past, future. | no |
| STE-9 | Do not use phrasal verbs when a single verb exists ("set up" → "install"). | no |
| STE-10 | Write notes and warnings as separate sentences, not in brackets. | no |

## Examples

| Incorrect | Correct |
|---|---|
| The lease is renewed by the agent before it expires. | The agent renews the lease before it expires. |
| Set up the worktree and then run the tests, making sure nothing fails. | Create the worktree. Run the tests. Make sure that no test fails. |
| Work items (also called tasks or jobs) are claimed atomically. | One agent at a time can claim a work item. |
