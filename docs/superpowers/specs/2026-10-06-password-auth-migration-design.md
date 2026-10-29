---
artifact_type: detail_design
id: AUTH-MIGRATION-01
status: implementation_partial
owner: shared
updated: 2026-10-06
human_fields: [identifier_choice, bootstrap_account, rollout_approval]
ai_fields: [context_loaded, risk_assessment, verification_plan]
shared_fields: [status, trace]
trace:
  auth: ../../architecture/AUTH_API.md
  api: ../../architecture/API.md
  native: 2026-10-05-native-ios-swiftui-design.md
---

# Password authentication migration

## Current evidence

The backend already exposes `POST /api/v1/login` with email/password and
rotating refresh tokens. The `user` table already has `password_hash`. Google
login remains the way many existing users are provisioned, and production
currently does not expose the native Google PKCE exchange route.

The native iOS implementation now uses `POST /api/v1/login` with the existing
`email` and `password` fields. “Username” is resolved to the normalized email;
no schema change is required. The password form, request DTO, Keychain session
save and refresh-on-restore path are covered by native XCTest and the
unauthenticated simulator launch check.

## Safe rollout decision

Do not disable Google or mutate production users in the first deployment. The
safe order is:

1. verify a password-capable account on production;
2. deploy native password UI/API smoke support;
3. run simulator tests against production without writes outside the selected
   disposable test account;
4. provide a rollback window and only then disable Google behind a reversible
   feature flag.

Existing Google-created users with an empty `password_hash` must receive a
verified password setup/reset flow before Google is disabled. A blind seed
password would lock users out and cannot be considered a no-data-loss release.

## Remaining blockers before production rollout

- Identify the exact bootstrap account. Never seed an unspecified production
  user.
- Deliver a one-time bootstrap secret through a secure local mechanism; never
  print a production password in chat, logs, CI output, or release notes.
- Restore authenticated access to the production Swarm/DB control host so a
  backup, migration, service rollout, and rollback can be verified.

## Rollback requirements

Before any production migration: create and validate a `pg_dump -Fc`, record the
current API image/digest and migration version, apply additive migrations only,
run read-only row/count/invariant checks, then canary the API. Keep Google
available until password login is proven for the selected account and existing
session refreshes remain valid.
