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
