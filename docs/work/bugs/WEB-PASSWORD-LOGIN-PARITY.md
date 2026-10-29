---
artifact_type: work_note
id: WEB-PASSWORD-LOGIN-PARITY-01
status: in_progress
updated: 2026-10-06
---

# Temporary web password login for native layout parity

## Intent

The production web shell is currently Google-only, which prevents a direct
authenticated comparison with the native iOS simulator when the Google flow is
not available in the browser session. Add a reversible email/password login
form to the existing web auth gate so the owner can inspect the same account
data while comparing web and native layouts.

## Contract

- Call the existing `POST /api/v1/login` route with `{email, password}`.
- Reuse the existing `LoginOutput` session persistence and profile bootstrap;
  do not add a second token format or expose credentials after submission.
- Keep Google login available; this is a compatibility/testing path, not a
  Google deprecation or password-policy migration.
- Show a generic failure message and preserve the base auth component styling.
- Do not log the password or deploy it in frontend configuration.

## Verification plan

1. Add a red/green frontend auth service test for the request contract.
2. Run design checks, typecheck and production build.
3. Log in with the owner-provided test account through the production web
   origin only after the build is verified; do not mutate financial data.
4. Recheck the native simulator against the same read-only account state.

## Rollback

Removing the password form and its handler restores the previous Google-only
web gate. The backend `/api/v1/login` contract and Google route are unchanged.
