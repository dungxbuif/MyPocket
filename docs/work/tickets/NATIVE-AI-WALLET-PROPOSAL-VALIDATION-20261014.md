# Native wallet proposal validation

Status: in_review. Existing approved financial chat/action-card correctness.
Small task exemption: yes — enforce existing proposal contract at renderer.
Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

Source evidence: AssistantStructuredCard.walletInput defaults unknown/missing
type to basic and malformed opening balance to0; goal/credit required amounts
are not checked. Backend propose_create_wallet requires supported type, integer
money and positive goal/credit limit. Thus stale/malformed structured data can
offer confirmation for an intent different from the supplied proposal.
Hypothesis: permissive native conversion causes invalid cards to appear valid.

Before implementation: rendered unit compares malformed proposal with existing
shared invalid-data status, including unsupported type and fractional balance;
watch actual RED. Then enforce type/money/type-required fields through a pure
native decoder used by the card, preserving backend valid defaults and optional
null fields. Reject invalid data; do not coerce it into another financial intent.
Keep base/status/buttons/palette/font and existing explicit confirmation state.
No new prepared-action endpoint, replay guarantee or automatic mutation.

Gates: decoder boundary/valid-payload tests, rendered invalid card, all native
units, normal chat keyboard/draft/APIoverview, guards/project/diff. Update human
and agent assistant guide/catalog/screen, context/backlog/validation/full goal.
No Swagger/ERD because wire/schema unchanged. Durable confirmed actions remain
separate gap; no prod/device/release or owner feedback acceptance.

RED18261exit65: all four malformed rendered cards150.5pt versus expected shared
invalid status86pt; actual confirmations offered for invalid type/fractional
balance/missing goal target/missing credit limit. Bundle
`/tmp/mypocket-wallet-proposal-red-20261014.xcresult`.
Root cause confirmed permissive walletInput defaults and unchecked type-required
amounts. AssistantWalletProposal.decode now supplies the same valid missing
opening0/totaltrue defaults, but validates present values, supported type and
required amounts, uses exact safe integer conversion, valid calendar date and
credit days. Card calls it; removed only obsolete private conversion helpers.
FirstGREEN89518exit0:269/269nativeunitsPASS0skip/no runtime warnings. Additional
valid/null/UTF8/type/money/nonfinite/overflow/date/day cases and eight chat UI
regressions now running, fresh final bundle. Not real action API/replay proof.

Proactive search found other structured-card direct Int64 conversions in money/
progress rendering. Record separate followup rather than silently claim all AI
numeric safety fixed. No speculative patch bundled; those paths need own RED.

Final8709exit0:279/279PASS0skip (271nativeunits+8chatUI), no runtime warnings.
`/tmp/mypocket-wallet-proposal-final-20261014.xcresult`. Includes rendered four
malformed cards, valid basic/goal/credit preservation, null/defaults, type/action/
name/description limits, nonfinite/fractional/overflow/string amounts, calendar
and credit-day rejection. Chat keyboard/AX/landscape/draft/APIoverview regressions
pass. Baseguards/guardtests/project/diff pass. No live create confirmation/replay
or provider E2E claim. The follow-up confirmation-summary slice now renders
goal/credit/total fields locally; see
`NATIVE-AI-WALLET-CONFIRMATION-SUMMARY-20261015.md`. Persisted proposal
`action_id` replay is covered separately; the full prepared-action dispatcher
and live provider proof remain separate.
Docs reconciled human+agent guide/catalog/screen/context/backlog/validation.
No API/schema/auth/runtime change, so Swagger/ERD/deployment docs unchanged.
No release publication/owner acceptance/production/device install.

Links: [goal](NATIVE-FULL-COMPLETION-20261014.md),
[catalog](../../architecture/AI_ADVISOR_TOOLS.md),
[screen](../../design/pages/finance-assistant/README.md),
[validation](../VALIDATION_MATRIX.md).
