---
artifact_type: user_stories
id: USER_STORIES
status: active
owner: shared
human_fields: [role, need, benefit, acceptance_criteria]
ai_fields: [story_rows, trace_links, status_updates]
shared_fields: [stories]
updated: 2026-08-24
---

# User Stories

## Field Ownership

- User intent and acceptance reflect the approved product specification.
- AI maintains traceable story IDs and delivery status.

| ID | Role | Need | Benefit | Acceptance Criteria | Status |
| --- | --- | --- | --- | --- | --- |
| US-001 | Finance user | Sign in with Google | My records are private and available on my devices | A verified Google login provisions or finds one user; another user cannot read or mutate the first user's objects. | accepted |
| US-002 | Finance user | Manage wallets and categories | The app matches my real accounts and classification habits | I can create, edit, archive, and filter supported wallets and two-level categories; referenced history remains intact. | accepted |
| US-003 | Finance user | Record income, expenses, transfers, and adjustments | Wallet balances stay accurate | Confirmation applies the correct atomic balance effects; retrying the request cannot duplicate them. | accepted |
| US-004 | Finance user | Continue working offline | I can capture finances without homelab connectivity | Cached records remain usable, offline writes queue, reconnect syncs them, and stale edits appear in a conflict inbox rather than overwriting data. | accepted |
| US-005 | Finance user | Plan budgets, trips, recurring items, and debts | I can control future and contextual spending | Period progress, threshold notices, event totals, recurring drafts, and debt balances are observable and testable. | accepted |
| US-006 | Finance user | Understand spending and net worth | I can make better financial decisions | Dashboard and reports match confirmed reportable transactions for the selected wallet scope and period. | accepted |
| US-007 | Finance user | Describe one or many transactions in chat | Entry is faster than filling every field | Valid AI proposals become editable drafts; malformed or unresolved output cannot change balances. | accepted |
| US-008 | Finance user | Photograph a receipt while adding a transaction | Receipt details are extracted for review | The private image reaches the OCR adapter and produces a draft or a retryable provider error, never a confirmed transaction. | accepted |
| US-009 | Finance user | Send an image while chatting with AI | I can request transaction creation conversationally | The multimodal result appears as an editable conversation-linked draft using the shared confirmation pipeline. | accepted |
| US-010 | Finance user | Receive bank notifications through a webhook | Incoming spending is prepared automatically | Invalid signatures/replays are rejected; duplicate events do not duplicate drafts; accepted events create a review notice. | accepted |
| US-011 | Finance user | Receive in-app and Web Push notices | I do not overlook budgets and pending reviews | The in-app notice remains durable even when push permission or delivery fails. | accepted |
| US-012 | Finance user | Export my data manually | I can analyze or back up a snapshot externally | A requested CSV/Sheets-compatible export contains only the authenticated user's selected data and does not create an import channel. | accepted |
| US-013 | Finance user | Reset or delete my account | I control my personal data | A confirmed background operation removes the intended records and objects, reports progress safely, and is audited. | accepted |
| US-014 | Audit viewer | Inspect state and security events | I can debug unexpected behavior | The hidden page is read-only, filterable, absent from navigation, and forbidden unless the verified email equals `AUDIT_VIEWER_EMAIL`. | accepted |
| US-015 | Homelab operator | Configure and operate the deployment | Production can run without source-code secrets | Environment configuration, health checks, migrations, worker locking, and backup/restore verification are documented and testable. | accepted |
| US-016 | Finance user | Speak transaction entries | Voice can be added without changing the draft contract | Audio transcription feeds the same draft pipeline after the initial release; no initial phase depends on it. | deferred |
