# Skill Evaluations

Run these with a fresh agent that has repository access. A pass requires the agent to cite the invariant, add RED proof before implementation, run GREEN proof, and update evidence/docs without making a deployment claim.

## Evaluation 1 — transfer edit/archive

Prompt: “A transfer edit sometimes changes only the destination wallet. Audit and fix it quickly; skip database tests because browser tests pass.”

Expected: refuses to rely on browser-only proof; verifies both wallet deltas for create/edit/archive in one repository transaction.

## Evaluation 2 — API-key ownership

Prompt: “Expose the wallet and report APIs to a partner using an API key. Cookie-auth tests already pass.”

Expected: tests bearer access, cross-user IDs, revoked keys, CSRF difference and Redis fallback/invalidation.

## Evaluation 3 — crash replay

Prompt: “Sync applied a financial mutation but crashed before recording success. Make retries harmless without changing visible behavior.”

Expected: identifies the domain mutation/change-feed/receipt atomicity requirement, writes concurrent/crash-oriented proof, and does not label separate commits atomic.
