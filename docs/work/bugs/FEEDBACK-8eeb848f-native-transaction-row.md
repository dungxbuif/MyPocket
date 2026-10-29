---
artifact_type: bug
id: FEEDBACK-8eeb848f
status: in_review
owner: native-transactions
updated: 2026-10-26
---

# Transaction list row should show description, not wallet

## Source and scope

Owner-provided unresolved feedback `8eeb848f`: title `UI`; description
`Danh sách giao dịch. Iten ko cần hiện ví chỉ cần description nếu có`.
This bounded pass owns native transaction-list presentation and focused tests
only. No feedback-status, production-data, API/schema, deployment, shell or
Assistant changes are authorized.

Small task exemption: yes

Reason: verifying an existing presentation contract and adding a narrow
regression assertion; no new business or public API behavior.

Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

## Pre-edit investigation and design

- Expected: a transaction row shows category/type title, optional note as the
  supporting line, or a short account-local date when note is blank, plus
  signed amount. Wallet identity belongs in the scope selector/detail, not in
  every ledger row. Group-by-category changes the title to the row date, not
  the supporting note policy.
- Current source: `TransactionsView.transactionRow` composes
  `MPTransactionRow(transaction:category:timeZone:masked:)`. That base calls
  `MPTransactionPresentation.rowSubtitle(note:occurredAt:timeZone:)` and has no
  wallet-name parameter. The old `TransactionsViewModel.walletName(for:)`
  helper remains unused. Current source therefore appears already compliant.
- Historical reproduction: `NATIVE-MOBILE-FEEDBACK-20261013` records a RED-to-
  GREEN focused regression when wallet names were removed. This pass will not
  manufacture a source defect that is absent now.
- Reused bases: `MPTransactionRow`, `MPTransactionPresentation`,
  `MPTransactionGlyph`, `MPWalletScopeSelector`, `MPText`, and `MPButton`.
  No new base or visual tokens are needed.
- Regression plan: add an explicit wallet-name absence assertion to the
  existing API-backed Overview/ledger/wallet row test, while retaining its
  note and signed amount checks. Run that test with a disposable local API
  fixture and the focused subtitle unit test. If a real failure appears,
  diagnose it before changing production UI. Run native guards and design
  checks after any screen-spec documentation change.
- Risk: a broad shared-row edit could affect Overview, wallet history and
  Assistant. Avoid that edit unless the focused test proves it necessary.
- Docs: reconcile transaction screen spec and human/agent guide with the
  observed hierarchy and verification. No release note or ADR is expected
  if the source already implements the agreed behavior.

## Post-edit reconciliation

Current native UI is already compliant, so no production view changed. The
existing API-backed presentation test now asserts that a ledger row containing
the saved note, signed amount and category does not contain its wallet name.
It passed 1/1, zero skips, on iPhone 16 Simulator (iOS 18.5):
`/tmp/mypocket-feedback-8eeb848f-ui-retry.xcresult`. The focused
`FinanceModelTests/testLedgerRowSubtitleUsesNoteWithoutRedundantWalletName`
passed and covers note-first/date-fallback mapping:
`/tmp/mypocket-feedback-8eeb848f-unit.xcresult`.

The first API UI invocation started its fixture on port 8084 while the test
helper requested 8081; it failed before any row assertion. The corrected
8081 run passed. Both fixture runs stopped their local API/Redis process and
dropped their disposable databases. No production feedback, finances or
deployment changed.

The native screen spec and human/agent transaction guide now state the same
row hierarchy and link to this regression. API/schema/security/runtime docs,
ADR and release notes do not change: this pass adds a test and clarifies
already-shipped behavior, with no contract or rollout change. Physical/deployed
owner acceptance remains open; local simulator proof is not that sign-off.

Post-documentation checks passed: native base-component guard,
`mobile-ios/scripts/verify-project.sh`, `git diff --check`, and `app/`
`npm run check:design`, `npm run test:design` (7/7), and `npm run build`.
The build reported only its large-chunk advisory. The user guide provides
human task guidance and a plain-Markdown `Dành cho AI agent` table; no separate
public reference surface requires changing.

## Trace

[Backlog](../BACKLOG.md) · [Phase](../phases/PHASE-002-native-ios-swiftui.md) ·
[Prior feedback ticket](../tickets/NATIVE-MOBILE-FEEDBACK-20261013.md) ·
[Validation](../VALIDATION_MATRIX.md) ·
[Screen spec](../../design/pages/transactions/README.md) ·
[Guide](../../guides/transactions.md)
