---
artifact_type: implementation_audit
id: FINANCE-ASSISTANT-COMPLETION-AUDIT-20261111
status: active
owner: shared
---

# Finance Assistant completion audit

## Intended behavior

The external Python service remains the channel-neutral Finance Assistant for
the native app and Telegram. MyPocket stays the authentication and ledger
authority. Every owner has an opaque durable workspace, app and channel
transcripts remain separate, trusted memory is shared, and PostgreSQL remains
authoritative over Markdown projections.

This completion pass adds the missing model-selected read-tool loop. The model
may choose only `home`, `wallets`, `transactions`, `budgets`, and `jars`.
Owner identity and the short-lived `mpa_` credential are injected by the
service and never supplied by model arguments. No write tool or direct ledger
mutation is exposed.

## Reconciliation matrix

| Requirement | Current evidence | Completion action |
|---|---|---|
| Separate Python/LangGraph service | Deployed API/worker and PostgreSQL checkpointer/store | Keep architecture unchanged |
| App plus multiple Telegram adapters | Backend proxy, native adapter UI, owner-scoped adapter rows, production Telegram E2E | Re-run regression suite and retain channel-neutral contracts |
| Per-owner memory, dreaming, Markdown workspace | Owner-scoped memories/dream candidates and atomic mode-0600 projections | Add no new authority; verify regression and production-safe evidence |
| Agent-selected finance reads | Current runtime eagerly fetches `home` and `wallets`; keyword-selects `transactions`; capabilities reports no tools | Implement bounded allow-listed tool loop and publish read-tool catalog |
| Financial writes | No client mutation API; reserved draft schema only | Keep disabled and test that model cannot request writes |
| Durable runs and isolation | PostgreSQL run/Telegram queues, signed assertion, delegated credential vault | Run full unit and PostgreSQL integration suites |
| Backup and restore | Scripts and runbook exist; production restore drill not claimed | Validate scripts statically and run an isolated disposable restore drill before claiming completion |
| Production state | `prod-1.3.19` Telegram/model/ledger path passed | Release only through one new protected `prod-<semver>` tag after all gates; user must rotate disclosed bot token |

## Affected contracts

- `finance-assistant/src/finance_assistant/agent.py`: bounded tool-selection loop.
- `finance-assistant/src/finance_assistant/mypocket.py`: explicit read-tool catalog and execution result conversion.
- `finance-assistant/src/finance_assistant/api.py`: truthful capability catalog.
- Tests: tool selection, parameter validation, maximum rounds, error behavior,
  rich parts, read-only enforcement, owner/delegation isolation.
- Architecture, user guide, operations runbook, validation matrix and
  AI-readable Markdown references: reconcile candidate wording with what is
  actually verified.

## Verification plan

1. Add failing focused tests for a model selecting one or more read tools,
   invalid/unknown/write calls, bounded rounds and typed response parts.
2. Implement the minimal LangGraph-compatible tool loop and pass focused tests.
3. Run the complete Finance Assistant test suite and backend proxy tests.
4. Run PostgreSQL integration and an isolated database-plus-workspace restore
   drill without touching production data.
5. Reconcile docs and release through the frozen homelab tag/receipt contract.
6. Verify native app and Telegram read paths, memory recall, channel separation,
   health, queues and unrelated service health. Never print secrets or financial
   payloads in evidence.

