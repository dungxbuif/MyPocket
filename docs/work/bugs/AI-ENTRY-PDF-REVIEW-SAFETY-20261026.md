# AI entry PDF review safety, 2026-10-26

Status: in_review (local regression verified; production review-only acceptance pending). Scope: backend AI entry extraction and normalization, focused backend tests, and documentation. Parent observation: [October Techcombank reconciliation](../tickets/TECHCOMBANK-OCTOBER-RECONCILIATION-20261026.md). This work does not approve proposals, import statements, change production data, or deploy.

## Incident and root cause

The review-only production PDF response returned 17 proposals in about 14 seconds. Twelve proposals used 2026-10-26; one foreign-currency line was proposed as 222 VND; repeated same-amount VJ rows were missing. No proposal was approved. The private source PDF and credentials stay outside the repository. The response evidence was inspected only for aggregate dates and the presence of foreign-currency cues, not copied here. The 222 row carries an "ngoại tệ" question without an ISO currency code, so the deterministic guard handles that cue too.

`AIEntryService.Send` changes missing or unparseable dates to the account's current instant, yielding an approvable RFC3339 value. `mergeExactAIDrafts` drops all but one draft with matching type, amount, wallet, category, jar, timestamp, note, and report flag, even when separate statement lines happen to match. The model response has no explicit source-currency field; the existing prompt and validator allow a foreign decimal amount to become a positive VND integer. These are independent paths to unsafe or missing review rows.

## Intended behavior and decision

- A missing or invalid source date stays empty in the proposal and gets a specific question. Approval already rejects a blank `occurred_at`; the user must save a real date before approval. A parseable source date remains unchanged (date-only values normalize to account-local midnight with a question).
- The extraction contract records per-row source currency as VND, foreign, or unknown. Foreign/unknown amounts remain zero with a question; no automatic conversion or decimal truncation. Explicit foreign-currency cues in model note/questions also fail closed if the model contradicts itself. Approval already requires a positive amount, so the user must enter a verified VND amount. The public proposal JSON shape remains unchanged.
- Preserve separate drafts, including exact same-field repeats, because backend equality cannot prove they are overlapping copies. The model may merge only clearly repeated source references across overlapping views. Separate PDF/statement lines stay separate even with matching amount, narrative, and date. Any unclear overlap remains a review question.

## Boundaries and verification plan

Touched modules: `backend/internal/usecase/ai_entry.go`, AI extraction prompt/parser/schema and internal draft type, focused tests. Direct dependencies: approval validator and repository persistence were inspected; they already block empty dates and zero amounts. No database schema, public route, authentication, mobile UI, attachment upload, or deployment configuration change is intended. Unknown: model accuracy on the private PDF cannot be established from the saved response alone; this fix will be proven with synthetic, non-sensitive fixtures and no live production call.

Test first: reproduce blank-date default, foreign amount despite currency evidence, and duplicate collapse with failing backend tests. Then verify synthetic dated VND entries remain reviewable, incomplete rows cannot pass approval validation, model schema accepts source-currency evidence, and focused backend suites pass. Reconcile the human/agent-readable AI entry API reference and this note against the code. Production acceptance and deployment remain separate.

## Local verification and docs review

- Red: three usecase regressions failed for invented date, collapsed equal drafts, and a 222 VND foreign-currency proposal. The parser rejected a source-currency field. A further incident-shaped test failed when the only foreign cue was "ngoại tệ" in the review question.
- Green: `go test ./... -count=1` from `backend/` passed, 657 tests across 18 packages. This includes AI entry normalization, provider schema, repository, and HTTP packages. No real PDF, live provider, or production API was used in verification.
- Public/agent-readable reference: [AI entry API policy](../../architecture/API.md) now describes blank dates, zero foreign/uncertain amounts, and preserved separate rows. The existing assistant page already says missing dates and amounts have no default; no screen edit was needed. The business rule already prohibits invented dates/amounts. No public JSON shape, database model, route, authentication, or runtime configuration changed, so API schema, ERD, architecture, and deployment docs do not need edits. This note captures the internal provider contract and implementation decision.
- The older review-first work note and synthetic fixture metadata record the prior fallback/merge behavior as historical evidence; neither is an active public contract. Production model accuracy and row completeness on the private PDF remain unverified until a separate review-only run after deployment. Deployment is required for production behavior to change; none was performed here.
