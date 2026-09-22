# AI Entry, Account Navigation, and In-App Feedback Design

**Status:** approved in chat on 2026-09-22

## Goal

Make receipt entry reliable and diagnosable, support up to 20 receipt files with drag-and-drop, move budgets into the account area, align transaction-page styling with the shared design system, and provide a floating in-app feedback chat that can attach a screenshot of the current MyPocket DOM.

## Findings driving the change

- The latest real process created a valid owner-scoped process and three PNG attachments, but all three attachment rows ended in `ocr_status=failed` and no ledger transaction was created.
- The backend currently maps every extractor failure to the same safe `ErrAIProvider` message and does not retain a sanitized provider-stage reason. The UI consequently shows a generic failure and its ambiguous-recovery copy.
- The AI entry handler and OCR validator both hard-code a maximum of three files. The upload control has no drag-and-drop handlers.
- Budgets are currently a primary bottom-navigation destination while Account already owns wallet, group, and feedback management.
- Feedback is currently JSON-only and has no private screenshot attachment lifecycle.

## Scope and non-goals

In scope:

1. Add stage-aware, secret-free OCR diagnostics and a synthetic live smoke test path; never automatically resubmit the user's files.
2. Raise the AI entry limit to 20 files, retain a 5 MiB per-file limit, raise the multipart envelope limit accordingly, and add drag/drop with deterministic validation and accessible status.
3. Make Account the canonical home for budgets while preserving a compatibility route for `/budgets`.
4. Normalize the Transactions screen using existing base components and one shared screen-header composition.
5. Add a floating feedback button and modal composer. Capture only the authenticated MyPocket content root with a DOM-to-image library; browser chrome and other tabs are never captured. The screenshot is optional and removable before submission.
6. Store feedback screenshots in the existing private S3-compatible storage under a feedback-specific key namespace. User and agent APIs expose only owner-scoped or short-lived signed access.
7. Update API/design/operations docs and add unit, integration, and browser tests.

Out of scope:

- Browser-wide or operating-system screen capture, which would require a permission prompt and cannot be silent.
- Automatic retries of a failed OCR/provider request.
- A per-user AI usage cap.
- Any bank, payment, or financial write action from the feedback flow.

## Architecture

The AI entry path keeps the existing idempotent process model. The extractor returns a typed provider-stage error (`storage`, `ocr_submit`, `ocr_poll`, `model`, or `schema`) with a safe public message and a redacted log event containing request/process identifiers and HTTP status when available. The session stores a stable `error_code` alongside the user-safe error. The UI displays the safe stage message and uses the existing read-only request recovery path.

Feedback screenshots reuse the existing S3 client with the reserved `batch=feedback` namespace (`.../owners/<owner>/batches/feedback/<feedback-id>/screen.png`). The feedback service writes the feedback row and screenshot metadata together; if storage fails, the feedback is not created with a misleading attachment. The authenticated owner can read metadata and a short-lived screenshot URL. The agent endpoint can fetch the same owner-agnostic item and request a short-lived signed URL, while audit events contain identifiers only.

Account navigation remains a five-destination shell after removing Budgets from the bottom bar. `/account/budgets` renders `BudgetsPanel`; `/budgets` remains a compatibility route that renders the same screen and canonical link metadata. Shared atoms/molecules provide the header, card, buttons, and status states used by Transactions and Budgets.

## Acceptance criteria

- A failed OCR process records a non-secret `error_code` and logs the provider stage; the user sees a useful Vietnamese message and no transaction is inserted.
- A valid submission with 20 files of 5 MiB or less passes request validation; file 21, a file over 5 MiB, a MIME mismatch, or an oversized multipart body returns a stable 400 without calling OCR.
- The file picker and drag/drop both produce the same validated `File[]`; duplicate files are not silently duplicated and the UI announces the rejection reason.
- A user can reach budgets from Account, and existing `/budgets` links continue to work.
- Transactions and Budgets use the same base screen-header/card/status composition and pass design checks.
- The floating feedback composer submits text plus an optional DOM screenshot, stores it privately, and does not capture browser chrome or other tabs. Removing the screenshot omits it from the request.
- Owner and agent access checks prevent cross-owner screenshot access; signed URLs expire and are never written to Redis audit payloads.
- Backend tests, frontend type/design checks, API integration tests, and browser tests pass.

## Documentation surfaces

- `docs/architecture/API.md`, `docs/architecture/FEEDBACK_API.md`, and `docs/architecture/OCR_API.md`
- `docs/design/screens/budgets/README.md`, `docs/design/screens/transactions/README.md`, and a new feedback floating-composer screen note
- `docs/standards/DEBUGGING.md` and `docs/releases/CHANGELOG.md`
- generated API docs under `backend/docs/` after route/schema changes
