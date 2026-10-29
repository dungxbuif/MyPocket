---
artifact_type: detail_design
id: FEEDBACK-20261019-NOTIFICATIONS
status: implementation
updated: 2026-10-19
---

# In-app budget notification producer

The current inbox is owner-scoped and deduplicated, but no backend operation
creates rows. This slice adds a bounded, in-app producer only; APNs/device
registration remains a separate external-signing integration.

Contract for this slice:

- inspect active budgets in the account timezone;
- emit one notification when a budget reaches 80% and one when it reaches
  100% of its effective limit;
- use `(owner_id, budget_id, local period, threshold)` as the stable dedupe
  identity, so worker retries and multiple worker instances are harmless;
- never emit for another owner, future/ended budgets, zero limits, transfers,
  excluded transactions, or credit-payment rows;
- preserve the existing inbox API; no APNs claim is made.

Verification: repository insert-or-ignore tests, worker threshold/timezone and
retry tests against disposable PostgreSQL, owner isolation, and the existing
backend race suite. The threshold wording is intentionally versioned in the
notification kind/body so a later product decision can change it without
reinterpreting old rows.
