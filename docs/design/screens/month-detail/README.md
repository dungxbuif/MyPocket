# Chi tiết tháng

Route `/months/{YYYY-MM}`; implements the data/note portions of [TICKET-07-04](../../../work/tickets/TICKET-07-04-tong-ket-thang-ai.md) and [CORE-03](../../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md). Status: implementation in progress.

## Composition

| Region | Base | Behavior |
| --- | --- | --- |
| Header and month navigation | `PageBackHeader`, `FormField`, `BaseSelect`, `Text` | Selects an account-local `YYYY-MM`; title shows a localized month label. |
| Completion state and totals | `SurfaceCard`, `Text`, `StatusMessage` | Shows server-derived current/completed state, included income, expense and net difference, with timezone/range disclosure. |
| Category and jar context | `SurfaceCard`, `Text`, `Progress`, `BaseButton`, `StatusMessage` | Uses only report-provided aggregates; links to `/jars?month=...`. Empty sections are omitted or described plainly. |
| User note | `SurfaceCard`, `FormField`, `BaseTextArea`, `BaseButton`, `StatusMessage` | Draft is saved/deleted explicitly for this account and month; it is independent from calculated totals and any AI output. |

Only existing bases are used. No local palette, custom card, disabled month-close control, or AI narrative placeholder.

## Events and effects

| Event | Effect |
| --- | --- |
| Open a month | GET `/api/v1/months/{YYYY-MM}`; server resolves timezone, half-open range, current/completed state and aggregates. |
| Select another month | Fetch the selected month; if the note draft is dirty, keep navigation disabled until save or discard. |
| Save note | PUT `/api/v1/months/{YYYY-MM}/note`; on success, update the server-owned note and clear dirty state. |
| Clear saved note | Confirm, DELETE the selected month's note, and refresh the report. |
| Ledger changes | Refresh calculated totals and context without overwriting the saved note. |
| API error | Retain current note draft and report the error with retry. |

## States, validation, and source of truth

- Loading/error states never masquerade as a real empty month.
- A successful month with no ledger data displays true zeros and an empty-context message.
- `is_complete` is derived from the account-local current month and is informational only; completed months remain editable and recalculate.
- Note length/validation follows the API contract; whitespace-only content is deleted/cleared, not stored as a misleading blank note.
- The selected month key is a validated `YYYY-MM`; timestamp boundaries are RFC3339 instants generated server-side.
- Note identity is `(owner, YYYY-MM)` and does not change when timezone changes or a wallet filter is used elsewhere.

## Copy and proof

Use `Tổng kết tháng`, `Đã hoàn tất` / `Đang diễn ra`, `Ghi chú của tôi`, `Lưu ghi chú`, and `Bỏ thay đổi`. Clearly label account timezone and the report range; use shared VND formatting.

Implementation update: `/months/{YYYY-MM}` is connected to the real report/note APIs. It shows account-local period boundaries, live totals, category and jar context, and explicit note save/discard/delete; edits remain available after a month is complete.

Automated proof covers month boundary/completion, PostgreSQL note owner-isolation/clear behavior, date navigation helpers, design guards and production build. Runtime/browser proof for persisted total recalculation and the note editor remains; owner UAT is pending. AI-generated monthly narrative, events/travel context, and freeze/close behavior are not implemented here.
