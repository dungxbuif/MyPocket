---
artifact_type: deployment_work_note
status: in_progress
approval: owner explicitly approved additive production backend deployment, preserving data
---

# Remembered-device auth rollout

Links: [design](NATIVE-FEEDBACK-CONSISTENCY-AUTH-DETAIL_DESIGN.md),
[auth contract](../../architecture/AUTH_API.md),
[Face ID guide](../../guides/face-id.md), [validation](../VALIDATION_MATRIX.md).

## Scope and safety

Create a reproducible auth-only candidate from deployed source baseline plus
device auth routes, repository, entity, session-family logic and their tests.
Do not include unrelated uncommitted finance/web/UI changes. An isolated Git
index/candidate branch preserves the active checkout, index and branch state.
Regenerate candidate Swagger from that candidate, not the dirty full checkout.
No changelog/release note publication before owner review.

Migration 30 only adds category is_active default true; migration 31 adds
device_credentials. Include 30 to keep the forward migration chain complete;
neither resets or modifies financial history. Before migration, create and
validate a custom-format pg_dump backup, record migration version/dirty flag
and counts of financial tables. Apply forward migrations only, compare counts,
then update only the API to an immutable image digest. Existing worker/web
service state remains unchanged. Never force, drop, truncate or reset prod.

Preflight: API health 200, API/worker 1/1, web 0/0. Current public API is reachable
without the web service; preserve this observed state rather than restarting
web from an older runbook assumption.

## Verification

- Independent auth review found legacy-family and delayed-issuance revocation
  races. Reproduced by deterministic TTL-aware tests; fixed with stable legacy
  migration groups and post-issuance revocation check. Full disposable-DB Go
  race suite passes. First logout of a legacy session can sign out the owner's
  other pre-upgrade sessions once; fresh password/device families are separate.
- Test candidate itself against disposable PostgreSQL before image build.
- Validate pg_dump catalog; retain backup, do not print/store secrets in docs.
- Migration dirty=false, unchanged financial counts, API/worker health, new
  public OpenAPI auth paths and invalid-device rejection after rollout.
- Build signed native app for owner's paired iPhone using current Personal Team
  provisioning, production API URL and no test launch flags. Verify install and
  launch through devicectl. Owner performs physical Face ID acceptance.
- Keep production feedback status unchanged until owner accepts fixes/release.

## Execution record

Signing follow-up after owner unlock: xcodebuild returned exit0, but mandatory
`codesign --verify --deep --strict` failed because CodeResources sealed a stale
`MyPocketIOS.cstemp` from the previously interrupted signing operation. The file
is absent while its hash remains in the manifest. Do not install this bundle.
Re-sign the existing compiled bundle using the same Apple Development identity,
preserving identifier/entitlements/flags, then repeat strict verification before
devicectl installation. No source/profile/Keychain security changes needed.

Follow-up result: same-identity re-sign exited0; strict/deep verification now
reports valid on disk and satisfies Designated Requirement. Bundle Info.plist
confirms `https://money.dungxbuif.com/api/v1`. Installation without uninstall
succeeded on paired iPhone15 Plus `00008120-001C28C02131A01E`; devicectl returned
bundleID `com.mypocket.ios`, installation container
`95BB3CB6-52BB-4310-B58F-CD72BED86661`. Process launch exited0 and reported
application launched. Public API health rechecked200. Physical biometric/visual
owner UAT remains pending; install success is not acceptance of all features.
No release notes or feedback statuses published.

Backend deployed; native app signed, installed and launched; owner UAT pending.

- Auth-only commit `12572c878645aea9d269da0d88de0e754f37f7bf`, pushed to
  `codex/device-auth-rollout-20261014`. Active iOS checkout/index preserved;
  unrelated dirty finance/web changes were not shipped. Candidate Swagger was
  generated from the isolated candidate, which passed `go test -race ./... -count=1`
  against disposable PostgreSQL.
- Validated custom-format backup on vm100:
  `/var/backups/mypocket/device-auth-12572c87/mypocket-pre-auth.dump`
  (109146 bytes, 193 pg_restore catalog entries). Directory is root-only.
- Forward migrations 29 → 31, dirty=false. Before/after: 1 user, 3 wallets,
  22 transactions, 34 feedback. Wallet and ledger checksums unchanged;
  no financial data reset or credential/password change.
- API image `registry.dungxbuif.com/mypocket-api:auth-12572c87`, immutable digest
  `sha256:1e1c686ce90bfd7c333dcc7f9206aff23117ed9f9d95839c8612a1a0fbc26fce`.
  Swarm update completed. Fresh follow-up: public `/api/v1/health` returns 200,
  API 1/1, worker 1/1 on its existing release, web 0/0 unchanged.
- Public Swagger exposes all four device/logout auth routes. Invalid device
  login returns 401 `device_credential_invalid`; no real-user enrollment/login
  was performed by the agent on production.
- Native proof: 169 unit + 4 compact-device UI tests and 8 API-backed UI tests
  pass, including enrollment → logout → protected Keychain device login.
  See [validation](../VALIDATION_MATRIX.md) for artifacts and residual gates.
- Signed Release build targets the paired iPhone 15 Plus with production API,
  current Personal Team and existing provisioning profile. Compilation reached
  codesign initially waited for Mac/Keychain authorization. After owner unlock,
  stale-signature repair and strict verification, installation/launch succeeded
  without uninstall. No password requested in chat; hardware Face ID acceptance
  remains owner-controlled and is not claimed verified.
- Production feedback remains unchanged. Changelog/release notes remain gated
  on owner acceptance; this rollout record is not a release announcement.
