---
id: ter-go3r
status: closed
deps: []
links: []
created: 2026-05-24T01:29:14Z
type: chore
priority: 0
assignee: Andre Silva
tags: [sentinel, manual, gate]
---
# SENTINEL — Human review required (do not close)

## Purpose

Permanent blocker ticket. All **entry gate** tickets depend on this
sentinel, which keeps them in the `blocked` state. This prevents any
work from starting on an epic until a human explicitly approves it.

**DO NOT close this ticket.** It exists solely to block entry gates.

## How to unlock an epic for work

1. Remove the sentinel dep from the entry gate: `tk undep <entry-gate> ter-go3r`
2. Close the entry gate: `tk close <entry-gate>` — tasks become ready
3. Agents work on tasks
4. When all tasks are closed, the exit gate becomes ready
5. Review the work, then close the exit gate: `tk close <exit-gate>`
6. Close the epic: `tk close <epic-id>`

## Entry gate tickets blocked by this sentinel

| Entry Gate | Exit Gate  | Epic                        |
|------------|------------|-----------------------------|
| `ter-6z5g` | `ter-v203` | E2 — Domain model & TBX I/O |
| `ter-bedf` | `ter-19rb` | E3 — terminology validate   |
| `ter-ab56` | `ter-m9l4` | E4 — Read commands          |
| `ter-c4ra` | `ter-cfeu` | E5 — Matcher                |
| `ter-ppn9` | `ter-b3ug` | E6 — scan & check           |
| `ter-st7u` | `ter-lyrx` | E7 — Write commands         |
| `ter-6dsf` | `ter-pbt2` | E8 — terminology apply      |
| `ter-hv9n` | `ter-uj1y` | E9 — Hardening              |
| `ter-39xj` | `ter-de52` | E10 — Release               |

