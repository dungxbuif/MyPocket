# Full owner-account API keys, no scopes

Status: in_review. Owner decision2026-10-14 explicitly supersedes the pending
transactions:write choice and all old fine-grained key scopes: a user API key
can do what that owner's JWT can do. [Completion](NATIVE-FULL-COMPLETION-20261014.md),
[backlog](../BACKLOG.md), [API](../../architecture/API.md),
[auth](../../architecture/AUTH_API.md), [key UI](../../design/pages/account-settings/README.md).

## Before-code design

Consolidate authenticated routing through RequireAuth accepting Bearer JWT/session
OR mpk_ user key. Key-looking credentials never fall back to JWT. Key validation
still checks digest, active expiry/revocation and DB owner; key must never select
another owner from body/path. JWT still checks Redis session. API keys do not
refresh and cannot become another user's session. Infrastructure unavailable fails
closed. Advisor credential identity/revalidation and audit remain intact.

Remove scope enforcement from key creation/authentication, advisor endpoint gates
and feedback wrappers. Existing persisted scope arrays have no authority after
rollout; do not require users to regenerate keys. Keep existing SQL column/data
for non-destructive compatibility, and legacy JSON scopes:[] only if needed by
installed app decoders; it must not imply limited access. New key UI has name/
expiry/copy/revoke only, clearly warns full read/write access to this account.
Unknown/old submitted scopes are ignored, not validated as permissions.

All existing JWT-protected business routes, including financial writes, AI entry,
profile/API-key management, device enrollment/revocation, become API-key-capable.
Domain/password confirmation/business rules still apply equally. This changes
credential authentication, not model tool whitelist or confirmation policy.
Login/refresh/device-login retain their own credential exchange contracts, not
unnecessary API-key exchanges. Logout with JWT retains family revocation; logout
with API key revokes that credential, not unrelated device sessions. Dedicated
service-token internal/admin routes are NOT JWT routes and remain separate;
no global/all-user feedback exposure or service-token privilege via user key.

## Implementation / proof

Before implementation add REDs: create without scopes, existing read/unknown/no
scope key gains same owner route access, full financial create+replay+atomic
transfer/state readback, foreign wallet/transaction/key/feedback denial, key
management and self revoke/logout, password still required for enrollment,
expired/revoked/wrong secret/unavailable fail closed and no JWT fallback. Keep
actual JWT/Redis expiry/logout tests. Replace stale scope-denial assertions with
new owner/lifecycle assertions, not simply remove negative security coverage.
Add native create-wire test with no scopes and API UI proof of no selector/
full-access warning/copy/revoke. Audit all route registrations and key-specific
handler checks; scope removal must not stop at ordinary transactions.

Full Go/race/vet with dedicated PostgreSQL+Redis, native units/UI, generated
Swagger, base/project/diff gates. Reconcile human+AI API/key/auth/offline/feedback/
advisor guides, account-screen spec, ADR, CONTEXT/BACKLOG/VALIDATION. No data wipe
or unnecessary schema migration. No production activation/install/release yet;
rollout expands rights of existing keys and requires explicit deployment note.

Verification refinement: actual foreign-key revoke was denied but returned500
because apiKeyFailure did not map repository.ErrAPIKeyNotFound. Tighten realHTTP
test to404, then map the existing typed error; preserve identical no-mutation
owner predicate. No new permission or resource existence disclosure. Add actual
route401 assertions for revoked/expired keys beyond service-only negative checks.

## Implementation and evidence

Common RequireAuth now accepts full-owner keys throughout user routes; advisor/
feedback scope gates and normalize/HasScope/principal scope state removed.
Deprecated internal scope input/argument ignored; SQL column unchanged, metadata
returns scopes:[] for installed decoders. Native create sends no scopes; selector
removed; MPText warning/full-access row reuses palette/Manrope/base controls.
API-key logout revokes itself; JWT family logout unchanged. No migration/package.

RuntimeRED52399 four service/auth cases→21202GREEN. ActualPG HTTP65981RED only
selflogout401→49557GREEN. Refined foreign-revoke78012RED500→typed404 mapping.
Final full Go race18813exit0/vet0 after generatedSwagger, dedicated55434/16379:
`go test -race ./... -count=1`, `go vet ./...`. New realHTTP proves old persisted
feedback:read key creates/replays ordinary row+transfer (exact3rows,2excluded),
foreign wallet/receipt/read/revoke denied, key create/list/revoke/selflogout and
other owner unaffected. List contains no secret/lookup/foreign IDs or obsolete
rights; old SQLscope array unchanged. JWT/Redis TTL/logout and actual key expiry/
revocation route401 tests remain. No key regeneration or user data changes.

Initial test nilJWT verifier panic corrected before valid401 RED. Cleanup briefly
removed generic entity.ErrAPIKeyInvalid used by repository; restored the generic
input-validation sentinel without scope helpers. Neither compile issue counts as
behavioral proof. New list/negative tests supplement, not remove owner security.

NativeRED59411 wire scopes/selector/missingwarning→final97143exit0,212/212PASS0skip
(211unit+1UI), `/tmp/mypocket-full-key-native-final-20261014.xcresult` via standard
SE LOCAL_API_UI warnings-as-errors gate. Inspected screenshotC1E878F7 under
`/tmp/mypocket-full-key-visual-20261014`: no scope selector, base fields/buttons/
warning present. Feedback covers helper; known protected-content design remains
blocked, no fourth placement guess or whole-screen acceptance. This UI uses a
fake session, not new local API keycreate/copy/revoke proof; that remains next.

Human+agent API/auth/feedback/advisor/offline/key guides, account spec, ADR021,
Swagger/current docs reconciled. Final base guard/guardtests/project/diff check
pass after reconciliation. No production activation/local API restart/signed install/release
or owner feedback resolution. Full all-gap goal remains in_progress.

Local runtime/E2E continuation before action: owned API10344/go-run10340,
worktree/backend cwd, loopback8081. Build current source before stopping only that
child; keep prior binary for rollback. Restart fixture API explicitly bound
127.0.0.1:8081 with dedicatedPG55434/mypocket_feedback and Redis16379, development/
OAuthfixture mode. Seed existing fixture@example.com so startup creates/rewrites
no user. App loads existing .env.local literally (never shell-source/evaluate);
override only fixture connections/mode. Local fixture sessions may re-login.
Local file has AI settings, OCR key empty/no S3; no liveOCR/provider claim.
Preserve prod/8080/user5432/6379. Baseline2users/1key/0wallet/0tx. Actual
testAPIBackedAccountAPIKeyCreateAndRevoke includes copy confirmation/API metadata
lookup and exact synthetic name cleanup; full units/guards follow.

Runtime gate53912 exit0,212/212unit+actualAPIUI. New listener19887/session46368,
health200. Key fixture is revoked via API, metadata intentionally retained.
Source audit found lifecycle risks: revoke removes local row only, but loadAccount
uses all server metadata (including revoked); createdSecret also has no associated
ID and is not cleared after revoke. Extend API UI test: revoke clears secret/copy,
leave/reopen account then revoked row stays absent. Add createdKeyID only in
transient state; filter revoked metadata for native active-list on load/create.
Raw API metadata/history and DB rows unchanged. Test RED before code, full native
units+API UI and rendered post-revoke capture. Not a bubble placement change.

Result: rebuilt local fixture API19887/session46368 healthy200; baseline users
unchanged, no wallets/transactions. Initial realUI53912pass212/212 then extended
51955 validRED secret/copy remain and row resurrects. Transient createdKeyID and
revoked-metadata filters added. Final70976exit0,212/2120skip units+actualAPIUI:
/tmp/mypocket-key-revoke-reload-final-20261014.xcresult. Real API metadata confirms
synthetic keys revoked,0active; raw history retained, no financial writes.
Post-reload FF8AB1F2 screenshot inspected under
/tmp/mypocket-key-revoke-reload-visual2-20261014: no key/secret/copy, native empty
state/warning present. Bubble covers timezone control; known design block remains.
Current-source219/219 gate58452 after owner-refresh/CAS also includes this UI proof.
Account spec/public key guide/validation/current docs reconciled, guard/project/
diff pass. Production/signed install/release/owner feedback resolution unchanged.
