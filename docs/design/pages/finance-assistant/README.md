---
artifact_type: screen_spec
id: FINANCE_ASSISTANT_SCREEN
status: implemented_local
updated: 2026-09-22
---

# Finance Assistant `/assistant`

The first implementation is a read-only conversation tab. It keeps the conversation in PostgreSQL and stores the conversation ID in the browser only as a pointer for history reload; ledger data remains the backend source of truth.

## Contract

- The composer uses `AssistantComposer mode="advisor"`, hides receipt uploads and keeps the shared `BaseTextArea`/`BaseButton` controls.
- User messages are submitted to `POST /api/v1/ai/advisor/messages` with a client request ID; the backend persists a run and assistant message before returning.
- History loads from `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...`.
- V1 is read-only. No confirmation, transaction mutation, bank connection, payment, or arbitrary model URL is exposed.
- When the AI provider is missing, the screen shows a configuration state instead of fake finance numbers.

## Navigation

The bottom navigation has six measured slots: overview, transactions, Add, budgets, assistant and account. The Add FAB remains owned by `BaseFab`; the assistant slot is a normal `BaseNavigationItem`.
