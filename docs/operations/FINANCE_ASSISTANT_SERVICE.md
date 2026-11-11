---
artifact_type: operational_runbook
id: FINANCE-ASSISTANT-SERVICE
status: candidate
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

Source/image/route contracts are locally tested. Production is blocked until
the feature is merged, the exposed bot token is rotated, protected release
prerequisites are verified and the owner creates the source tag. No live model,
Telegram webhook, production migration or restore drill is claimed.

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
