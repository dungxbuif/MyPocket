---
artifact_type: detail_design
id: API-KEY-COPY-DETAIL_DESIGN
status: in_review
owner: shared
trace:
  ticket: TICKET-10-01-quan-ly-api-key.md
  screen: ../../design/pages/current-ui.md
  validation: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# API key one-time secret copy — detail design

## Problem and root cause

After creating a scoped API key, AccountView shows the one-time secret as text
inside an `MPAssistantCard`, but has no copy action. This makes it difficult
to transfer the credential to the intended integration. The key already exists
in view state; the missing affordance is entirely in the native presentation.

## Intended behavior

- Keep the one-time secret visible exactly as today until the Account screen is
  left/reloaded; do not add persistence, re-fetch, or a second reveal path.
- Add a clearly labeled “Sao chép API key” action beside/after the secret.
- Copy the exact secret to the system clipboard only after an explicit tap.
- Use a local-only pasteboard item that expires after five minutes; do not
  sync the credential through Universal Clipboard or leave it indefinitely.
- Give brief visible and VoiceOver success feedback (“Đã sao chép API key.”).
- Continue using shared `MPAssistantCard`, `MPText`, and `MPButton`; do not add
  screen-local styling or a new base component.
- No backend/API/schema/permission behavior changes. Clipboard is an explicit
  user action and may be visible to other apps according to iOS pasteboard
  behavior; do not log the secret or include it in analytics/docs/tests output.

## Hypothesis

I think the reported inability to copy occurs because `AccountView` renders the
secret only as `MPText` and exposes no clipboard action. A UI regression test
that taps the proposed action and reads the test device clipboard should fail
before implementation and pass after.

## Scope and impact

- API: no; DB: no; security model: no; runtime/config: no.
- Reused bases: `MPAssistantCard`, `MPText`, `MPButton`.
- Changed implementation: native AccountView only.
- Changed contract/docs: API-key ticket, current UI inventory, validation matrix,
  account settings page spec, and project context. Changelog/release note is
  intentionally deferred until owner approval.
- Small task exemption: yes. Reason: local presentation of an existing
  one-time secret only; no persistence, API, schema, architecture, permission,
  or security-policy change. Impact checked: API=no, DB=no, Security=no,
  Runtime=no, Standards=no.

## Verification plan

1. Unit-test the exact clipboard payload and empty-secret no-op; API-backed UI
   regression verifies the secret is visible, taps the copy action, and sees
   accessible confirmation. Do not read the simulator clipboard from the test
   runner because that crosses the iOS pasteboard privacy boundary.
2. Run the focused iOS UI test against the existing disposable local API
   fixture; do not use production credentials/data.
3. Run native unit/contract tests, Release warnings-as-errors build, project
   guard and diff check.
4. Only after the fix and local verification, attempt a signed device build and
   install if a valid Xcode team/profile exists; otherwise record the exact
   signing blocker without altering the user's signing choice.
5. Reconcile linked ticket, UI inventory, validation matrix and context. Defer
   changelog/release note until owner approval. No public API docs change is
   required because the backend contract is unchanged.

## Evidence

- Initial reproduction: the supplied Account screenshot displays the API key
  secret without a copy control.
- RED: clipboard policy tests first failed to compile because
  `APIKeySecretClipboard` was absent.
- GREEN: 150/150 native unit/contract tests pass, including exact clipboard
  payload and empty-secret no-op. The API-backed UI regression passes 1/1 for
  create → reveal → tap copy → confirmation. UI test intentionally does not
  read UIPasteboard from its separate runner process.
- Project verification, base-component guard/tests, and diff check pass.
- Device install remains blocked: Xcode has a valid local development identity
  but no signed-in account for its Team and no provisioning profile for
  `com.mypocket.ios`; neither app install nor launch was attempted. Owner
  iPhone review and release-note approval remain pending.
