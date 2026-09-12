---
artifact_type: business_rules_spec
id: BR-MYPOCKET-V2-DRAFT
status: draft
owner: human
approval: pending_human_review
updated: 2026-09-12
trace:
  product_spec: MYPOCKET-V2-SPEC-DRAFT.md
  master_spec: SPEC.md
  requirements: REQUIREMENTS.md
  parity_matrix: ../research/moneylover/FEATURE-PARITY-2026-09-08.md
  architecture: ../architecture/ARCHITECTURE.md
  api: ../architecture/API.md
  erd: ../architecture/ERD.md
  validation_matrix: ../work/VALIDATION_MATRIX.md
  decisions:
    - ../decisions/ADR-003-offline-sync-conflict-review.md
    - ../decisions/ADR-004-review-first-ingestion.md
    - ../decisions/ADR-006-separate-asset-portfolio-valuation.md
    - ../decisions/ADR-008-atomic-offline-sync-commit.md
---

# MyPocket V2 Business Rules Draft

## Review Status

This is a business-rule draft for owner review. It defines observable invariants and outcomes, not implementation authorization. Existing approved master docs continue to govern the running product until reconciliation.

## Business Vocabulary

| Term | Meaning |
| --- | --- |
| Wallet | A money account whose confirmed transactions change an authoritative VND balance. |
| Transaction | A confirmed income, expense, transfer, or balance-adjustment accounting event. |
| Draft | A reviewable proposal with no accounting effect until explicit confirmation. |
| Jar | A spending limit and period whose progress comes from explicitly assigned expenses. |
| Category | A classification and suggestion signal; it does not implicitly assign an expense to a jar. |
| Goal | A target amount and optional deadline whose progress comes from explicit linked activity. |
| Obligation | Money owed by or to the user, with linked disbursement and repayment activity. |
| Asset position | A non-cash holding with quantity, trades, prices, valuation, and P&L. |
| API key | A revocable bearer credential with owner-selected scopes for a bot or external application. |
| Intake agent | An AI flow that may propose drafts but cannot confirm accounting. |
| Advisor agent | An AI flow that may read through tools but cannot mutate business data. |

## Rule Precedence

1. Ownership, authorization, idempotency, accounting, and data-integrity rules cannot be overridden by UI, API client, automation, AI output, or provider output.
2. Confirmed PostgreSQL state is authoritative over caches, agent memory, summaries, IndexedDB, provider data, and displayed optimistic state.
3. More restrictive rules win when a user role, API-key scope, shared-wallet role, agent mode, or lifecycle state overlap.
4. Historical records are archived or superseded where possible; they are not physically deleted when doing so would break auditability or accounting reconstruction.

## Identity, Ownership, And Sharing

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| ID-01 | Every private business object has an owner or an explicit shared-wallet authorization path. | A second unrelated user cannot enumerate, read, mutate, sync, export, or attach to the object. |
| ID-02 | Authentication identity is resolved by the server and is never accepted from a request body, tool arguments, or model output. | Supplying another `user_id` cannot change the effective owner. |
| ID-03 | A shared wallet has one owner and zero or more accepted members with explicit roles. | Membership records identify inviter, invitee, role, status, and lifecycle timestamps. |
| ID-04 | A member sees only shared-wallet data authorized by their role; unrelated owner data remains private. | Shared access does not grant access to other wallets, jars, assets, agent chats, exports, or API keys. |
| ID-05 | Revocation or leaving blocks new server reads and writes immediately; connected clients purge revoked shared data, while disconnected clients cannot replay queued shared mutations and purge on the next sync/login boundary. | Server access is denied immediately and a formerly offline client cannot apply or retain revoked shared data after reconnect. |
| ID-06 | Shared-wallet reports attribute mutations to the acting member while preserving the wallet owner. | Report and audit fixtures distinguish owner, actor, and wallet. |

## Wallet And Category Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| WAL-01 | A wallet represents cash, bank-account-like manual money, a savings goal behavior, or a credit behavior; an investment asset is not a wallet. | Asset valuation never appears as a wallet transaction or wallet balance mutation. |
| WAL-02 | Wallet balance is derived only from confirmed accounting events plus an approved opening state. | Drafts, OCR, agent messages, price changes, and report queries do not change balance. |
| WAL-03 | Archived wallets remain readable through historical records and cannot accept new transactions unless restored. | Historical reports reconcile before and after archive. |
| WAL-04 | Inclusion in total controls aggregate wallet/net-worth views consistently across dashboard and reports. | Excluded-wallet fixtures produce matching scope across all totals. |
| CAT-01 | Categories are expense, income, or debt/loan compatible and have at most two hierarchy levels. | Invalid depth, cycles, and incompatible transaction types are rejected. |
| CAT-02 | System categories retain stable identity and cannot be renamed or deleted by users. | Mutation attempts fail without changing existing references. |
| CAT-03 | User categories may be archived, merged, reordered, and activated per wallet without deleting transaction history. | Historical transactions retain a resolvable category after lifecycle changes. |

## Transaction And Accounting Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| TX-01 | VND accounting amounts are integer units and a normal transaction amount is strictly positive. | Decimal/floating VND and non-positive amounts are rejected at the boundary. |
| TX-02 | Income increases the source wallet; expense decreases it; transfer decreases source and increases destination atomically. | Golden ledger values match exact expected deltas. |
| TX-03 | A transfer requires two different authorized wallets and is excluded from income/expense reports. | Both balance legs commit or neither commits; cash-flow income/expense totals remain unchanged. |
| TX-04 | A transfer fee is an explicit expense effect and cannot be hidden inside the transferred amount. | Source decrease equals transfer amount plus fee under the approved fee contract. |
| TX-05 | Balance adjustment records the delta needed to reach a target balance and is excluded from ordinary income/expense analysis. | Resulting balance equals the target exactly and reports do not classify the delta as spending. |
| TX-06 | Editing or archiving a confirmed transaction exactly reverses the old effects before applying the new effects in one database transaction. | Retry and failure fixtures never leave half-reversed accounting. |
| TX-07 | Duplicate creates a new transaction identity and requires a new explicit date/idempotency decision; it is not a replay of the original command. | Duplicate and retry are distinguishable and cannot create accidental extra copies. |
| TX-08 | Bulk actions return a result for every selected record and never hide partial failure. | The user can identify applied, rejected, conflicted, and replayed items. |
| TX-09 | Receipt attachment is private evidence and does not determine accounting truth. | Removing or failing OCR does not silently mutate confirmed values. |

## Spending Jar Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| JAR-01 | A jar defines an amount limit, period, lifecycle state, and optional category hints. | A jar is usable without selecting a category. |
| JAR-02 | Each confirmed expense may belong to zero or one active jar. | The database/API reject more than one assignment and cross-user assignments. |
| JAR-03 | Jar progress sums only active, confirmed, report-included expenses explicitly assigned to that jar and occurring inside the selected period. | Unassigned, archived, excluded, out-of-period, income, transfer, and adjustment rows contribute zero. |
| JAR-04 | Category membership never changes jar progress automatically. | Two expenses in the same category can affect different jars or no jar. |
| JAR-05 | Reassigning, editing, archiving, restoring, or changing the date of an expense recalculates affected jar periods exactly once. | Before/after progress reconciles to the authoritative assigned transaction set. |
| JAR-06 | Drafts and recurring schedules may carry a proposed jar assignment, but progress changes only after confirmation or approved auto-post. | Pending/rejected drafts contribute zero. |
| JAR-07 | Threshold notices are deduplicated by jar, threshold, and period. | Reprocessing does not send duplicate 80%/100% notices for the same crossing. |
| JAR-08 | Finished-period history is immutable as a historical view except when an underlying authorized historical transaction is corrected. | Historical corrections are visible and auditable. |

## Planning, Credit, Goal, And Debt Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| PLAN-01 | Recurring schedules have frequency, next occurrence, optional end, active/paused state, and draft or explicit auto-post mode. | Paused/ended schedules create nothing; retry creates at most one occurrence. |
| PLAN-02 | Draft mode is the default for provider- or AI-derived automation. | New automation cannot silently default to confirmed accounting. |
| PLAN-03 | Events/trips group linked transactions for context and totals but do not duplicate accounting effects. | Linking/unlinking changes event totals only. |
| PLAN-04 | An obligation records direction, principal, counterpart, status, and linked disbursement/repayment events. | Outstanding balance equals approved principal effects minus valid repayments. |
| PLAN-05 | A credit wallet exposes limit, statement date, due date, debt used, available credit, and repayment behavior with one documented sign convention. | Numeric examples and tests reconcile purchases, refunds, statements, and repayments. |
| PLAN-06 | A savings goal has target, optional deadline, explicit progress source, and completed/archived lifecycle. | Goal progress cannot be fabricated from wallet type or unrelated balance changes. |

## Asset Portfolio Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| AST-01 | Supported positions include gold, stock, crypto, foreign currency, and other user-defined assets. | Each type validates its identity and unit contract. |
| AST-02 | Quantity uses bounded fixed-precision decimals; VND prices, fees, costs, and totals use integer units. | API floats and excess precision are rejected. |
| AST-03 | Buy and sell trades replay in deterministic order using moving weighted-average cost. | Replaying the same ordered ledger produces identical quantity, basis, and realized P&L. |
| AST-04 | A sell cannot exceed the available quantity at that point in the replay. | Historical edit/archive that would create a negative later quantity is rejected atomically. |
| AST-05 | Price observations are append-only, source-attributed, and ordered by observation time plus deterministic tie-breakers. | Correcting a price adds a replacement observation rather than rewriting history. |
| AST-06 | Missing price produces unknown market value/P&L, never zero; stale price is visibly stale. | Dashboard and API distinguish `missing`, `stale`, and `current`. |
| AST-07 | Unrealized market movement is not income, expense, transfer, or wallet balance change. | Cash-flow reports remain unchanged when prices change. |
| AST-08 | Dashboard exposes wallet total, investment value, and combined net worth separately. | Users can reconcile each component independently. |
| AST-09 | Automatic pricing failure preserves the last known observation and allows a manual snapshot. | Provider outage cannot erase or falsely refresh valuation. |
| AST-10 | MyPocket does not execute asset trades or provide investment recommendations. | No broker/exchange credential or order endpoint exists. |

## Reporting Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| REP-01 | Reports use confirmed, non-archived, report-included records within explicit wallet and date scopes. | Drilldown rows sum to the displayed total. |
| REP-02 | Reporting day/month boundaries use `Asia/Ho_Chi_Minh`; stored timestamps remain timezone-explicit. | Boundary fixtures around midnight and month/year changes pass. |
| REP-03 | Transfers and balance adjustments are excluded from income/expense totals; transfer fees follow the expense rule. | Cash-flow fixtures remain neutral except for explicit fee expense. |
| REP-04 | Drafts, rejected imports, failed OCR, and agent proposals are absent from reports. | Confirmation is the first point at which eligible values appear. |
| REP-05 | Asset value and P&L appear in portfolio/net-worth reports but not cash-flow spending totals. | Changing only an asset price changes portfolio/net worth and no cash-flow report. |
| REP-06 | A numerical advisor answer cites the tool result/source and the exact scope used. | Unsupported numbers without a source cause validation failure or an insufficient-data response. |

## API-Key And External Application Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| KEY-01 | Only an authenticated owner browser session with CSRF may create, list, revoke, or replace API keys. | Bearer keys cannot manage credentials, including themselves. |
| KEY-02 | The plaintext key is generated server-side, shown once, never stored, and represented later only by name, prefix, scopes, timestamps, and status. | Database, logs, audit, API responses, and UI history contain no recoverable secret. |
| KEY-03 | Every API key has explicit scopes; absence of a required scope denies the operation even if the owner account could perform it interactively. | Positive and negative scope matrices pass for every public domain. |
| KEY-04 | Revocation and expiration fail closed and take effect for cache hits as well as database lookups. | A revoked/expired key cannot succeed during Redis outage or stale-cache conditions. |
| KEY-05 | API-key identity is always combined with current ownership, sharing, object lifecycle, and domain validation. | Model/tool-provided or stale object IDs cannot bypass authorization. |
| KEY-06 | Retryable API mutations require stable idempotency keys; updates use the current optimistic version. | Same key/same payload replays; same key/different payload conflicts. |
| KEY-07 | API usage is rate-limited and correlated per user/key without logging the secret. | Limit and audit tests identify a safe key ID/prefix and correlation ID. |
| KEY-08 | External applications receive stable machine-readable errors and documented pagination rather than provider or SQL details. | Contract tests match OpenAPI and public docs. |

## AI Agent And Tool Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| AI-01 | Intake and advisor are separate session kinds with separate tool allowlists and durable messages. | A run cannot change kind or reuse a session of the other kind. |
| AI-02 | The server constructs tool context containing effective user, actor/key, scopes, session, run, deadline, and correlation ID. | The model cannot supply or override security context. |
| AI-03 | A function tool call is a model request, not executed authority; the dispatcher validates name, strict arguments, mode, scope, ownership, limits, and current state before calling a domain service. | Unknown, malformed, forbidden, stale, or cross-user calls are rejected and audited safely. |
| AI-04 | Tools reuse application/domain services and never expose SQL, unrestricted HTTP, shell access, or the entire REST router. | Tool and REST fixtures produce the same business outcome and validation errors. |
| AI-05 | Intake may read owned references/OCR and propose income or expense drafts; it may not confirm, transfer, edit confirmed transactions, or invoke advisor-only tools. | Prompt injection and malformed-output tests produce no confirmed accounting. |
| AI-06 | Intake returns `needs_input` instead of inventing required amounts, dates, wallets, categories, or jar assignments. | Ambiguous fixtures create zero drafts and one bounded clarification request. |
| AI-07 | Advisor tools are read-only and retrieve fresh reports, transactions, jar progress, obligations, assets, and bounded chat history. | Attempts to call draft or mutation tools fail regardless of API-key write scopes. |
| AI-08 | Every account-specific numerical claim is grounded in one or more current tool result IDs; session summaries are never financial truth. | Stale summary conflicts resolve in favor of fresh tool data. |
| AI-09 | Prompt context is ordered as fixed policy, bounded summary, recent messages, current input, and current tool results, with complete tool-call/result pairs retained. | Long-session tests stay within the configured token budget and preserve current intent. |
| AI-10 | A run has bounded tool-call count, duration, retries, output size, and cancellation behavior. | A looping or unavailable provider terminates predictably without duplicate side effects. |
| AI-11 | OCR is deterministic preprocessing when an owned receipt is attached; OCR output remains untrusted input for the same run only. | Foreign, failed, expired, or mismatched OCR cannot leak or produce a draft. |
| AI-12 | Raw prompts, OCR text, tool payloads, secrets, and private financial answers are excluded from normal logs and audit metadata. | Redaction tests cover success, failure, retry, and cancellation paths. |

## Offline, Sync, And Concurrency Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| SYNC-01 | IndexedDB is a user-scoped cache/outbox; it never becomes authoritative over PostgreSQL. | Server resync replaces unsupported or stale cached truth safely. |
| SYNC-02 | Every queued mutation has a stable mutation ID, payload, base version where applicable, and observable pending/conflict state. | Lost-response replay returns the original outcome without duplication. |
| SYNC-03 | A sync batch commits atomically according to the approved atomic-sync contract; conflicts are explicit and replayable. | Failure leaves no partial authoritative batch. |
| SYNC-04 | Logout removes the current user's private offline stores, pending media, chat data, and cached tool results from that browser profile. | A subsequent user cannot read the previous user's data offline. |
| SYNC-05 | Shared-wallet access is revalidated during sync and replay. | Revoked membership prevents queued shared-wallet writes from applying. |

## Notifications, Export, And Lifecycle Rules

| ID | Business rule | Verification outcome |
| --- | --- | --- |
| LIFE-01 | Important review, threshold, due, import, and conflict events create durable in-app notices; push is best effort. | Push denial/outage does not remove the in-app notice. |
| LIFE-02 | Export contains only the authenticated user's authorized selected data and is a snapshot, not a synchronization channel. | Export fixtures exclude foreign and unselected records. |
| LIFE-03 | Reset/delete requires recent authenticated confirmation and runs as an idempotent, observable job. | Retry does not recreate or partially duplicate deletion work. |
| LIFE-04 | Account deletion includes owned database rows, private objects, API keys, auth cache, and offline invalidation according to retention/audit policy. | Completion evidence accounts for every owned storage class. |
| LIFE-05 | Online banking and voice have no active route, worker, tool, UI claim, or runtime credential in this scope. | Route/config/UI/docs inventory finds no enabled capability. |

## Cross-Domain Invariants

- A draft can become at most one confirmed transaction.
- One logical retry can create at most one business effect.
- One confirmed expense can affect at most one spending jar.
- Market-price movement can never affect cash-flow accounting.
- An advisor run can never create a draft or mutate finance data.
- An intake run can never confirm a draft.
- An API key can never increase its own privileges or mint credentials.
- A shared-wallet role can never grant access to unshared owner data.
- Provider failure can never be represented as confirmed user financial truth.

## Review Questions

1. Approve or reject shared wallets as part of the core baseline.
2. Confirm that one expense may belong to at most one jar.
3. Approve granular API-key scopes and optional expiration as the V2 target.
4. Confirm that savings goals and full credit lifecycle are in scope.
5. Confirm that automatic recurring posting remains opt-in while draft mode remains default.
6. Confirm that portfolio uses moving weighted-average cost rather than FIFO/tax lots.

## Acceptance Of This Business Draft

- [ ] Owner reviews every item under `Review Questions` and records an explicit decision.
- [ ] Approved business rules are reconciled into stable requirement IDs and user stories.
- [ ] Any changed API/schema/auth/architecture boundary receives an ADR before implementation.
- [ ] Each approved rule family is mapped to automated proof and UAT in the validation matrix.
- [ ] Conflicts with the current master SPEC are resolved explicitly rather than overwritten silently.
