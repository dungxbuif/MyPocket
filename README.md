# Harness

Harness is a markdown-first framework for making AI coding agents follow SDLC thinking when they work on software projects.

The idea comes from the Software Development Life Cycle: before code is written, requirements should be understood; before implementation expands, design and impact should be clear; before work is closed, testing, documentation, release state, and maintenance context should be reconciled.

Harness translates that discipline into repository-local contracts for LLM agents.

```text
Requirements -> Design -> Implementation -> Testing -> Release -> Maintenance
```

becomes:

```text
Hydration -> Detail Design -> Execution & Test -> Reconciliation -> Dehydration
```

The detailed framework documentation lives in [docs/README.md](docs/README.md).

## Reference Ideas

Harness borrows ideas from several AI engineering workflows while staying deliberately portable:

- SDLC: phase gates, role separation, artifacts, testing, release, and maintenance.
- [GStack](https://github.com/garrytan/gstack): root-level agent instructions, specialized review perspectives, QA/review/ship mindset.
- [GSD Core](https://github.com/open-gsd/gsd-core): context files, roadmap/phase planning, verification loops, and context rot mitigation.
- [Superpowers](https://github.com/obra/superpowers): design before execution, systematic debugging, test-first pressure, and reviewer-style gates.
- [Repository Harness](https://github.com/hoangnb24/repository-harness): repository-local agent instructions, feature intake, validation expectations, and durable decisions.

Harness does not copy their runtime model. It extracts the process ideas into a markdown framework that can be read by any coding agent.

## Problem

LLM agents can produce code quickly, but they often fail at the engineering process around the code:

- They start coding before understanding requirements.
- They lose context across long sessions or multi-day work.
- They invent assumptions instead of recording decisions.
- They write tests or documentation as claims, not verified evidence.
- They update code without updating architecture, API, ERD, or product docs.
- They over-scan brownfield projects and burn context on irrelevant files.
- They loop on failing tests until the session runs out of tokens.
- They report work as done even when verification, documentation, or release steps are incomplete.

Human teams reduce these risks through SDLC roles, phase gates, artifacts, QA, release discipline, and maintenance practices. Harness adapts those ideas into a lightweight contract that AI agents can follow inside a repository.

## Philosophy

Harness is built around a few principles:

- A repository should carry enough context for a new human or agent to continue the work.
- Chat history is not durable project memory.
- Agents should not jump from prompt to code.
- Plans, tests, decisions, and docs are part of the work, not afterthoughts.
- Brownfield documentation should be pay-as-you-go, not a full-system rewrite.
- "Done" should mean verified, reconciled, traceable, and resumable.
- Human intent and approval must remain distinguishable from AI-generated analysis.

Harness is intentionally markdown-first. It does not require a runtime orchestrator, custom CLI, or specific AI vendor. Stronger enforcement can be layered later through scripts, hooks, or CI.

## Core Techniques

### SDLC As Agent Gates

Harness does not copy SDLC as heavyweight waterfall. It turns SDLC concerns into agent control points:

- Requirements become hydration and backlog context.
- Design becomes explicit detail design for non-small work.
- Implementation is bounded by ticket, bug, phase, and standards.
- Testing becomes execution-based evidence.
- Release becomes changelog, release notes, and operational checks.
- Maintenance becomes context updates, ADRs, and master-doc reconciliation.

### Context Hydration And Dehydration

Agents hydrate from a small set of project files before work starts, then dehydrate the updated state back into repository docs after work ends.

This reduces context rot because important decisions, current focus, backlog state, and next steps survive beyond one chat session.

### Runtime Backlog Queue

Harness treats backlog as a runtime queue, not just a parking lot for ideas.

Requirements, tickets, bugs, maintenance, and framework work can be reordered based on real priority. Urgent bugs can preempt planned feature tickets. The backlog decides what to consider next; executable work still needs enough context through a ticket, bug, requirement slice, or phase.

### Detail Design Gate

Non-small work must pass through a design artifact before code. The design captures problem, context, approach, alternatives, impacted areas, test plan, validation impact, and reconciliation plan.

This prevents agents from silently making architecture, API, data, security, or runtime decisions while coding.

### Execution-Based Verification

Harness requires evidence from real commands, manual checks, UAT, or explicit skip reasons. Markdown claims are not enough.

Validation policy defines what kind of proof is expected. A runtime validation matrix tracks which accepted behaviors have evidence.

### Root-Cause Debugging And Loop Guard

Bug work must start with reproduction and root cause investigation. Agents are not allowed to patch symptoms or stack guesses.

Harness also defines a loop guard: after repeated fix/test failure, the agent must stop, mark the work blocked, summarize attempts, and ask for human/design review instead of burning the session.

### Reconciliation

After implementation, agents must compare actual changes with master docs and decisions.

If behavior, contracts, data model, architecture, or durable tradeoffs changed, the relevant docs must change too. If docs do not change, the reason should be explicit.

### Parseable Markdown

Harness documents use YAML frontmatter for structured parsing and Markdown bodies for human/agent readability. Templates separate human-owned fields, AI-owned fields, and shared fields so authorship and responsibility stay clear.

## What Harness Is Not

Harness is not:

- A full multi-agent orchestrator
- A task runner
- A CI/CD system
- A replacement for human product judgment
- A guarantee that an agent will behave correctly without tool or human enforcement

It is a repository-local operating contract. Its value comes from making the desired process explicit, inspectable, and reusable.

## Documentation

Use [docs/README.md](docs/README.md) for the actual framework guide:

- Folder responsibilities
- Artifact templates
- YAML frontmatter format
- Human vs AI field ownership
- Lifecycle document usage
- Validation matrix
- Diagram rules
- Completion checklist
