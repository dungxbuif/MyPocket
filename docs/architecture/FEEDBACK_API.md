---
artifact_type: api_reference
id: FEEDBACK_API
status: implemented_local
owner: shared
updated: 2026-09-22
trace:
  design: ../superpowers/specs/2026-09-22-feedback-fix-changelog-design.md
  plan: ../superpowers/plans/2026-09-22-feedback-fix-changelog.md
---

# Feedback → Fix → Changelog API

This reference describes the implemented local API surface. All endpoints are under `/api/v1`. Responses use the shared envelope `{ "data": ..., "meta": ... }`; errors use the existing problem response with `code`, `status`, `detail`, and `request_id`.

## Authentication boundaries

User endpoints require `Authorization: Bearer <JWT>` from the normal MyPocket session. They are owner-scoped: a user can only read their own feedback.

Agent and internal endpoints require the dedicated `FEEDBACK_AGENT_TOKEN` as `Authorization: Bearer <service-token>`. This token is for the local/dev fix agent only. It is not a user API key and has no access to finance endpoints. The server compares the token in constant time and emits redacted Redis audit metadata.

Public changelog endpoints do not expose raw feedback, user IDs, email addresses, descriptions, or account identifiers.

## User endpoints

### `POST /feedback`

Create an owner-scoped feedback item.

```json
{
  "type": "bug",
  "title": "Số dư chưa cập nhật",
  "description": "Sau khi thêm giao dịch, số dư ví vẫn hiển thị giá trị cũ."
}
```

`type` is `bug`, `feature`, or `improvement`. Title is limited to 200 Unicode characters; description to 10,000. The response is `201` with the created feedback in `data` and status `open`.

### `GET /feedback`

List the authenticated user's feedback, newest first. No other user's rows are returned.

### `GET /feedback/{id}`

Read one feedback item owned by the authenticated user. A foreign or missing ID returns the shared not-found problem without disclosing whether another owner has it.

## Agent/internal endpoints

### `GET /agent/feedback?status=&limit=`

Poll feedback for the local fix agent. `status` may be `open`, `triaged`, `in_progress`, `fixed`, or `rejected`; `limit` is 1–100 and defaults to 50. The result is ordered oldest first for triage. Descriptions are available only on this authenticated agent surface.

### `PATCH /internal/feedback/{id}/status`

Advance a feedback item through the lifecycle:

```text
open → triaged → in_progress → fixed | rejected
```

The status endpoint rejects direct `fixed`; publishing a changelog is the only path that finalizes a fix. Backward transitions and terminal transitions return `409 FEEDBACK_CONFLICT`.

```json
{ "status": "in_progress" }
```

### `POST /internal/changelog`

Publish a version and atomically mark referenced `in_progress` feedback rows as `fixed`.

```json
{
  "feedback_ids": ["feedback-id"],
  "version": "1.4.2",
  "title": "Cập nhật phân loại giao dịch",
  "description": "Sửa lỗi phân loại các giao dịch định kỳ."
}
```

The server sorts and locks IDs in a deterministic order, requires every referenced row to be `in_progress`, creates the unique changelog version, and updates all rows in one PostgreSQL transaction. Missing/terminal/duplicate references roll back the entire operation. A Redis audit event is emitted after commit and excludes title, description, credentials, and financial data.

## Public changelog endpoints

### `GET /changelog?limit=`

List published changelog entries, newest first. Limit is 1–100 and defaults to 50.

### `GET /changelog/{id}`

Read one published changelog entry. This never returns the feedback rows that produced it.

## Status shown in the app

Account → Phản hồi uses the authenticated user endpoints. A fixed row shows `Đã xử lý` and, when linked, `Fixed in <version>`. The browser never receives or stores `FEEDBACK_AGENT_TOKEN`.

The app also renders a floating feedback bubble above the bottom navigation on every authenticated screen. Opening it captures the current DOM view in the background (excluding the feedback overlay itself); submitting the form sends that PNG as the optional `screenshot` multipart field. If browser capture fails, text feedback still submits without an attachment.

## AI-agent quick reference

## Screenshot capture

`POST /feedback` also accepts `multipart/form-data` with `type`, `title`, `description`, and one `screenshot` PNG (maximum 1 MiB). The server stores it privately and exposes only `screenshot_available` in feedback JSON. `GET /feedback/{id}/screenshot` is owner-scoped and returns a short-lived `302` signed URL; `GET /agent/feedback/{id}/screenshot` uses the dedicated feedback-agent token. Storage object keys are never returned.

```text
1. GET /api/v1/agent/feedback?status=open with Authorization: Bearer FEEDBACK_AGENT_TOKEN.
2. PATCH /api/v1/internal/feedback/{id}/status -> triaged.
3. PATCH /api/v1/internal/feedback/{id}/status -> in_progress.
4. Fix and test code outside this API.
5. POST /api/v1/internal/changelog with feedback_ids, version, title, description.
6. GET /api/v1/agent/feedback?status=fixed to verify changelog_id/fixed_at.
```

Do not call finance APIs with the service token. Do not publish a changelog for `open` or `triaged` feedback. Do not retry a committed publish with a different version without first reading the feedback state and changelog list.
