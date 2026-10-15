# Native Assistant numeric safety

Status: verified_locally — focused native regression green; signed device/UAT pending.
Parent: [full native goal](NATIVE-FULL-COMPLETION-20261014.md), approved chat/card
correctness. This is not a new finance feature, API or confirmation design.
Previous goal turn is progress: semantic reveal correction, probe/walletCRUD
proof and273PASS/2FAIL business gate. No test/fault proxy currently running.

## Current reconciliation — 2026-10-15

The bounded correction is present in the current source: monetary card values
require finite exact `Int64` conversion within the finance-safe JSON integer
range, progress targets reject unsafe values before renderer conversion, counts
use exact conversion/fallback, and masked fallback cards do not expose raw
structured payloads. The focused signed simulator run executed seven relevant
`DesignSystemTests` cases with 0 failures:
`/tmp/mypocket-numeric-safety-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_20-23-53-+0700.xcresult`.
No API, provider, database, production, device or release state changed. The
historical diagnostic/RED notes below describe the pre-correction failure and
are not the current implementation status.

## Intent and current evidence

Preserve exact financial values and a usable Assistant even when old/malformed
structured card data is received. Keep palette, Manrope, existing shared bases,
privacy, navigation and explicit confirmation. Never invent a zero or silently
truncate a financial amount. Owner approved exact parsing, unavailable invalid
values, nontrapping JSON fallback and existing masking; preserve bases/palette.

Read scope: AssistantViews.swift AssistantCardType/AssistantStructuredCard,
Core/API/Envelope.swift JSONValue, DesignSystemTests, assistant screen spec,
AI_ADVISOR_TOOLS catalog. No production or provider request.

Source facts:
- `money` uses `Int64(number(value) ?? 0)`: missing values become0, fractions
  truncate, out-of-range/nonfinite values may trap.
- `progressTarget` admits `Double(Int64.max)` because its upper bound is rounded
  to the same Double; progress renderer then calls `Int64(target)`.
- `transactionCount` has the analogous Double(Int.max) bound before `Int(count)`.
- JSONValue decoder accepts finite Double JSON values such as1e100; fallback
  `prettyDescription` calls `Int(value)` whenever rounded()==value, including
  unrepresentable integral Doubles. This matters even for unknown card views.
- Overview already uses exact values; do not change its verified behavior or
  equate its correctness with every structured card.

Diagnostic command,2026-10-15 exit0 (not rendered regression or crash proof):
```sh
rtk proxy swift -e 'import Foundation; let value = Double(Int64.max); print("upper_bound_guard_accepts=\(value <= Double(Int64.max))"); print("exact_int64_exists=\(Int64(exactly: value) != nil)"); let parsed = try JSONDecoder().decode([String: Double].self, from: Data("{\"amount\":1e100}".utf8)); print("large_json_number_decodes=\(parsed["amount"]?.isFinite == true)"); print("fraction_to_integer=\(Int64(1.5))")'
```
Output: upper_bound_guard_accepts=true, exact_int64_exists=false,
large_json_number_decodes=true, fraction_to_integer=1.

### Actual-source diagnostics — 2026-10-15

No product/test source edited. A Node diagnostic reads actual Envelope.swift
into Swift interpreter input, appends JSON decode→prettyDescription, and isolates
each case in a child process. Valid12500 exits0/prints12500; valid JSON1e100 exits
by SIGTRAP: `Double value cannot be converted to Int because the result would
be greater than Int.max`. No timeout/provider/DB/app action. This is actual
production-function crash reproduction in macOS Swift CLI, NOT iOS card rendering
E2E verification or a shipped fix.

Second diagnostic extracts exact current transactionCount/progressTarget bodies
(only removes static to invoke standalone) alongside actual Envelope.swift.
Count42 exits0/prints42; Double(Int.max) count SIGTRAPs with same conversion
error. progressTarget(Double(Int64.max)) returns nonnil but Int64(exactly:target)
is nil. Root is guarded conversion, not malformed JSON decoding. Actual native
RED/tests remain required after approval; interpreter proof is not simulator proof.

Source reachability: AssistantStructuredCard fallback calls prettyDescription in
unknown-type/missing-view and malformed-items branches. These branches also ignore
masked (AssistantViews.swift522/525/575/607/633). This is a privacy gap within the
same fallback, not yet rendered exposure proof. Public guide must not claim all
fallback cards are masked. Proposed correction applies existing privacy text
before fallback rendering, preserves valid masked money and proves both masked
malformed/unknown states. No new authorization/auth contract. Design approval
remains pending. No product or test source changed in this diagnostic turn.

## Bounded approach presented in chat

Use shared exact monetary parsing in AssistantCardType within the existing
API-safe integer range ±9007199254740991. Keep negative balance/net/delta values;
progress targets and counts remain nonnegative. Preserve legitimate0. Absent/
wrong-type/fractional/nonfinite/overflow money renders existing “Chưa có dữ liệu”
rather than0 or a truncated value. Masking still returns the common masked value.
Draw progress only with valid spent+target, preserving zero-target behavior and
existing target-key precedence. Count conversion must use exact integer conversion,
not a Double-rounded max guard. Fallback JSON pretty formatting must not trap;
unrepresentable raw numbers can remain textual, never treated as money.

Alternatives rejected: clamp invalid financial values; silently strip fractions;
remove structured cards; treat malformed provider output as a finance write.
No new card/base, hidden field, transport, library, auth/schema or provider limit.
ui-ux-pro-max searches for missing chart values returned unrelated results twice;
no verified recommendation adopted. Fall back to existing unavailable-data spec
and common base/error rules, not a fabricated skill recommendation.

## Verification before any completion claim

Final48656exit0:287/287PASS0skip/no runtime warnings (279unit+8native chatUI),
`/tmp/mypocket-numeric-chat-final-20261015.xcresult`. Restored masked fallback and
goal exact-value guard bothGREEN; valid overview, draft persistence, keyboard,
landscape and largest-type chat regressionsPASS. CapturesCA7766AF/5CFE537F inspected:
normal overview/source+composer and large-text keyboard remain usable; no whole-app
privacy/visual acceptance inferred. Base guard/tests/project/diffPASS.
Release production-API build18960 is live waiting codesign; phone currently
unavailable. No updated install/release/production backend change claimed.

Approved implementation checkpoint: range test35797exit0,1/1GREEN. Expanded
suite93815:277PASS/1literalFAIL; expected money strings corrected to existing
NumberFormatter NBSP, no shared formatter change. Business gate53977 contains
all278unitPASS plus3UI PASS/1categoryFAIL (tracked separately).

Horizontal numeric check found goal-progress finite-but-unsafe amounts could
fabricate0/100percent despite unavailable money. Actual native test42731RED1/1
(`/tmp/mypocket-goal-progress-invalid-red-20261015.xcresult`); goal helper now
uses the same exact finance-safe validation before ratio calculation. Existing
valid signed/current-zero/clamping behavior retained.

Privacy regression sensitivity verified by restoring only old unmasked fallback
behavior temporarily:76862exit65,1FAIL, hosted unknown-card height385.5 instead
of masked66points and other malformed-card assertions fail. Mutation removed
immediately; final fallback again masks before pretty formatting. Evidence
`/tmp/mypocket-privacy-fallback-mutation-red-20261015.xcresult`. This is not a
delivered unmasked build. Final279units+8chatUI gate48656 now running; do not
claim finalGREEN until terminal. No new production or iPhone install this turn.

Native RED now verified: `/tmp/mypocket-native-progress-range-red-20261015.xcresult`,
exit65,1testFAIL0skip/no runtime warnings. Actual progressTarget accepts
9007199254740992 (outside finance safe range); test fails XCTAssertNil as intended,
not compile/environment error. No trapping renderer invoked and no product fix.
This is one failing test, not two distinct tests despite duplicate log lines.
The previously installed Release binary is unchanged; this diagnostic was only
compiled for simulator. Approval remains pending for presentation/privacy changes.

2026-10-15 next diagnostic: add one native regression invoking actual
`AssistantCardType.progressTarget` without rendering/trapping conversion.
Expect nil for2^63 and values outside the finance API safe-integer range, valid
zero/safe maximum retained. This converts CLI-only evidence into native XCTest
proof; no product fix, data deletion, model call or changed approval inferred.
Run only this test first and retain expected RED. Product design remains pending.

After approval, TDD tests must catch unsafe range acceptance, truncation and
fallback formatting using actual helpers/card rendering. Cases: nil/null/string,
fractional/negative/zero, ±safe boundary, outside safe boundary, Double(Int64.max),
1e100/nonfinite, valid signed amounts, malformed count with list-count fallback,
valid and invalid budget/jar progress, privacy masking. Keep all financial/API
assertions. Watch actual RED before product fix, then all native units and normal/
accessibility/landscape/draft/APIoverview chat regressions. Inspect card captures.
No real model/provider, production, app reset or financial write required.

## Reconciliation

Affected internal docs: this ticket, full goal, context/backlog, validation, plan
ledger, Assistant card catalog/screen. After implementation reconcile public
[native Assistant guide](../../guides/finance-assistant-ios.md) and its plain
Markdown agent surface/catalog. API/OpenAPI/ERD/config/auth/ADR unchanged because
this is defensive native presentation, not a wire/ownership change. No release
publication/feedback resolution before owner acceptance. Test/refinement gates
for pinned controls/category and other blocked slices remain independent.

Links: [validation](../VALIDATION_MATRIX.md), [backlog](../BACKLOG.md),
[UI plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[screen](../../design/pages/finance-assistant/README.md),
[catalog](../../architecture/AI_ADVISOR_TOOLS.md).
