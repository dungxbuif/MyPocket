---
artifact_type: operational_runbook
id: FINANCE-ASSISTANT-SERVICE
status: active
owner: shared
---

# Finance Assistant service operations

## Release authority

Production follows the frozen homelab contract. After reviewed changes are on
MyPocket `main`, the owner creates exactly one `prod-<semver>` source tag. The
app pipeline on the Mac runner tests/builds/pushes immutable API and assistant
digests, writes them to `homelab-ops/main`, and creates the internal
`mypocket-prod-<semver>` receipt tag. Only the homelab-ops pipeline on the
protected `homelab-vm100-deploy` runner migrates, deploys, checks health and
rolls back. Do not deploy through Portainer, SSH, an app job or a second tag.

Before tagging verify the Mac app runner assignment, protected source tag rule,
masked/protected `HOMELAB_OPS_PUSH_TOKEN`, protected ops receipt tags and the
app-specific CI rules. Shared ops jobs must not select every app tag.

## Required production secrets

- `finance_assistant_database_url`
- `finance_assistant_assertion_secret` (same value configured for MyPocket)
- `finance_assistant_credential_key` (Fernet key)
- `finance_assistant_model_api_key`

Never place values in Git, logs, Markdown, the iOS app or Telegram messages.
Revoke any Telegram token exposed outside the one-time adapter form before use.

## Deploy and health

The ops receipt deploy first validates both immutable digests and required
networks/secrets. It runs MyPocket migration `000054` from the API digest and
the assistant schema migration from the assistant digest, updates API/worker,
then deploys stack `mypocket-assistant`. Readiness requires the public MyPocket
health endpoint and private assistant `/healthz`. Failure uses Swarm rollback
policy and returns the existing API/worker to their prior specs.

The public route is limited to
`/v1/channels/telegram/{adapter_id}/webhook` on the money host. All other
assistant endpoints stay on private Swarm networks.

The worker owns both durable app runs and Telegram intake. An app submit is
healthy only when it returns `202 queued`, the run advances through
`running` to a terminal state, and completed history contains the same rich
parts. Queue depth, attempts, expired leases and terminal failures must be
inspected in PostgreSQL; never log decrypted delegations or bot tokens.

The advertised 240-second run timeout covers the complete LangGraph turn,
including all model and finance-tool rounds, and remains below the five-minute
queue lease. A deadline returns a safe completed message instead of allowing a
second worker to reclaim still-running work.

## Backup and restore

Back up the assistant PostgreSQL database with the normal database class and
export external volume `mypocket-finance-assistant-data` as a mode-0600 tar in
the configs class. The export helper mounts the volume read-only with networking
disabled. The classes are not an atomic snapshot: record both timestamps and
treat PostgreSQL as authoritative if they differ.

Restore PostgreSQL first, restore the workspace archive to the named volume,
then run migrations/checkpointer setup with the release digest. Reconcile or
regenerate Markdown projections from PostgreSQL when archive timestamps differ.
Verify owner A
cannot read owner B, one explicit memory survives, app/Telegram histories remain
separate and a revoked adapter cannot process a webhook. Markdown-only restore
is unsupported because projections are not authoritative.

## Current gate

Release `prod-1.3.19` is deployed and its Telegram/model/ledger path is verified
in production. The model-selected read-tool completion is locally verified and
awaits its immutable release receipt. The disclosed test bot token remains an
open owner action: revoke it in BotFather and replace the adapter credential
before treating the bot as long-lived.

An isolated PostgreSQL 17 backup/restore drill passed on 2026-11-11. It restored
one trusted memory, separate app and Telegram histories, owner isolation and the
workspace archive; checksum validation passed, snapshot directory mode was
0700 and restored Markdown files retained mode 0600. This proves the scripts
and data contract on disposable data, not a production disaster-recovery drill.

## Live Telegram enrollment verification — 2026-11-11

The first production adapter will be enrolled for the established primary
owner through the same `ChannelAdapterService`, encrypted credential vault and
Telegram `setWebhook` path used by the authenticated native facade. No token,
webhook secret, owner identifier, pairing code or financial payload may enter
the work note or operator logs. The temporary bot credential supplied for this
check must be revoked after verification because it was previously disclosed.

Verification is boundary-based: confirm adapter and delegated credential rows,
the exact public webhook URL reported by Telegram, one private-chat update in
the durable inbound queue, single-use pairing, worker completion, a Telegram
reply and an owner-scoped finance read. Record only counts, states and redacted
timestamps. A healthy service or successful webhook registration alone is not
an end-to-end pass.

The live check exposed two PostgreSQL type/boundary regressions before any
financial response was returned. Psycopg yields the adapter UUID as a UUID
object even though `ChannelAdapter.id` is a string contract, so a successfully
redeemed pairing was persisted and then incorrectly rejected by the worker's
adapter comparison. The subsequent channel scope used
`telegram-<adapter UUID>`, which is 45 characters and therefore fails the
`TurnRequest.channel_id` maximum of 40 before model execution.

The corrective release must normalize repository adapter IDs to strings and
derive a deterministic, adapter-specific Telegram channel ID that satisfies the
existing validation limit. Regression tests must prove pairing succeeds with a
UUID-backed repository row, two adapters retain distinct channel scopes, and a
real UUID scope validates. Production E2E is complete only after a fresh
single-use pairing returns the success message and a read-only wallet question
returns through Telegram. Existing MyPocket health and all unrelated services
must remain healthy through the normal immutable `prod-<semver>` release path.

## Production E2E record — 2026-11-11

- App pipeline `#877` passed both test jobs, immutable build and ops promotion
  for `prod-1.3.19`; ops pipeline `#879` passed validation, protected VM100
  deployment and production verification.
- The focused Pi5 enrollment added only `finance-assistant-telegram`. The live
  dummy webhook returns `401`, existing MyPocket health remains `200`, and the
  dynamic file retained inode `2475801`; Traefik was not restarted.
- Telegram reports the exact registered webhook with zero pending updates. Its
  retained last-error metadata refers to the pre-route `404`; subsequent durable
  jobs and replies prove successful delivery after enrollment.
- A fresh single-use pairing returned `Đã kết nối Telegram với MyPocket.` A new
  read-only wallet question completed in one worker attempt and received a
  finance-backed Telegram answer. One earlier job remains failed after ten
  pre-release attempts as historical evidence of the rejected 45-character
  channel scope; it is not pending.
- MyPocket API/worker and Finance Assistant API/worker all remained `1/1`.
  Final database evidence is one active binding and the expected user/assistant
  message pair; no token, owner ID, pairing code or financial payload is stored
  in this record.
