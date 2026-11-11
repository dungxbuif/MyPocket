---
artifact_type: api_reference
id: FINANCE-ASSISTANT-API
status: active
owner: shared
---

# Finance Assistant API and channel contract

This is the stable human- and agent-readable reference for the external
Finance Assistant. Deployment evidence and remaining credential-rotation gates
are recorded in the operations runbook.

## Native facade

All native routes use the existing MyPocket bearer/session authentication and
base path `/api/v1/ai/advisor`. Gin derives the owner; clients must not send an
owner ID.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/capabilities` | Read-only capability and parts-version discovery |
| `GET` | `/conversation/messages?conversation_id=…` | App-channel history with typed parts |
| `POST` | `/conversation/reset` | Reset one app conversation, or the app channel when no ID is supplied |
| `POST` | `/messages` | Compatibility submit |
| `POST` | `/messages/async` | Persist a durable run and return `queued` before model work |
| `GET` | `/runs/{id}` | Owner-scoped durable run state |
| `GET` | `/runs/{id}/events` | Bounded completed-event representation; not token streaming |
| `POST` | `/runs/{id}/cancel` | Cancel queued work or cooperatively interrupt a running provider request |
| `GET` | `/channels` | List the owner's channel adapters |
| `POST` | `/channels/telegram` | Validate and store one Telegram bot token under a distinct label |
| `POST` | `/channels/{adapter_id}/pairings` | Issue an adapter-scoped, single-use ten-minute pairing code |
| `DELETE` | `/channels/{adapter_id}` | Remove webhook, bindings and adapter access |

Successful channel responses are wrapped by Gin as `{ "data": ... }` for the
native decoder. The Python service itself returns raw channel resources.

## Telegram webhook

`POST /v1/channels/telegram/{adapter_id}/webhook` is the only public service
route. Traefik publishes exactly that prefix on the money host. Telegram must
send `X-Telegram-Bot-Api-Secret-Token`; the adapter ID alone grants no access.
Duplicate update/message IDs are idempotent. A valid webhook returns `202`.
The response means the update is durably queued, not that model processing or
Telegram delivery has completed. The worker leases queued rows and retries a
failed delivery up to ten attempts before marking it failed for operations.

## Private service authentication

All non-webhook service routes require a short-lived MyPocket assertion with
issuer `mypocket`, audience `finance-assistant`, owner subject and an allow-list
containing `assistant:chat` and/or `assistant:channels`. `X-Owner-ID` is rejected.
Finance reads additionally use an owner-bound `mpa_` delegated credential with
exact allowed operations. The service never stores an owner `mpk_` key.

## Response parts

The current `parts_version` is `1`. Supported rich types are text,
`metric_group`, `wallet_balances` and `transaction_list`. Unknown parts must be
ignored or rendered as safe text. Capability flags are `write=false`,
`direct_writes=false`, `proposal_actions=false` and
`confirmation_required=true`.

`tools` advertises exactly five model-selectable read operations: `home`,
`wallets`, `transactions`, `budgets` and `jars`. A turn may execute at most
three tool rounds. The model supplies only the operation and bounded public
arguments; the service injects owner identity and the `mpa_` delegation. An
unknown or write-like operation is rejected before an HTTP request. No
advertised tool can create, update or delete financial data.

The service's own OpenAPI document is available on its private network at
`/openapi.json`; it is not an alternative public authentication surface.
