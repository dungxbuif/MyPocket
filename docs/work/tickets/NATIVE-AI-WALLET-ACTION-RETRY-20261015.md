# Assistant wallet action retry key

Status: local implementation complete. The mounted card now keeps a stable
key, and new persisted proposals carry a server-generated `action_id`; the
full prepared-action dispatcher remains open. A legacy proposal without the
field still uses a mounted-card fallback.

## Decision

Give each mounted wallet confirmation card a stable request key and pass it
through the existing wallet create API. New proposal payloads add a server-
generated `action_id`, which is reused after history reload; legacy cards keep
the mounted fallback. No endpoint, permission or backend migration changes are
required. This closes wallet proposal identity/replay for this action while a
general prepared-action dispatcher remains separate.

## Verification

- API contract test proves a caller-supplied key is sent unchanged.
- Backend and native tests prove the proposal `action_id` is emitted and
  extracted for replay.
- Native assistant build and existing card tests remain green.
- No environment, device, production or database action is part of this fix.

Evidence: `/tmp/mypocket-wallet-action-key-green-20261015.xcresult` passes the
caller-supplied-key API contract. The preceding RED compile is retained at
`/tmp/mypocket-wallet-action-key-red-20261015.xcresult`.
