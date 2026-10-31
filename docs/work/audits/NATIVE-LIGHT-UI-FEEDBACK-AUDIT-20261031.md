---
artifact_type: ui_feedback_audit
id: NATIVE-LIGHT-UI-FEEDBACK-AUDIT-20261031
status: in_review
owner: shared
scope: native iOS light mode only
updated: 2026-10-31
---

# Native iOS light-mode and feedback audit

This audit checks the current native source, existing design/feedback records,
and the supplied simulator screenshots one screen at a time. Dark mode is
explicitly out of scope. No production feedback status, API, database or
financial data was changed.

The historical reconciliation snapshot contains 48 open records: 32 covered
locally, 13 requiring UAT/device/provider acceptance, and 3 requiring a product
or accounting contract. These are engineering dispositions, not proof that
the production feedback records are closed.

## Global shell

| Area | Feedback/design | Current finding | Status |
| --- | --- | --- | --- |
| Bottom navigation and hit targets | `5d6557d2`, `31b7bcf7` | Shared shell and account shielding exist; focused native proof passes. | Covered locally; device visual UAT remains. |
| Feedback bubble | `b63c7727`, `5ff74529` | Bubble now uses a stable shell dock and Assistant footer reserves a measured slot. Existing screenshots still show overlap, so a fresh populated-screen pass is required. | Code covered; visual/device UAT open. |
| Card/row density | `f4597cd2`, `f42c103d`, `f82b0108` | Shared card/row tokens exist. The audit must still verify every populated route, not only unit geometry. | Code covered; screen-by-screen UAT open. |
| Add action | Earlier P0 finding | `MainTabView` keeps the global Add enabled on Transactions while `TransactionsView` also renders a local `+`. This is still a duplicate action in source and needs one owner. | Open P0. |
| Icon family | `2eb68f7b` | Semantic catalog exists, but `usesVerifiedAssets` is false and the asset catalog contains only AppIcon. Shell/More/field controls still use raw SF Symbols. | Open P1: bundle verified Phosphor subset. |

## Tổng quan

| Feedback/design | Finding | Status |
| --- | --- | --- |
| `aa6ca3cf` | Duplicate notification destination was removed; Overview bell remains the intended entry. | Covered locally; owner/device acceptance open. |
| `4ca27075` | Month summary and owner-timezone calculation are wired; deployed data/timezone readback remains UAT. | Needs UAT. |
| `e11d18e6`, `196110ec` | Wallet scope/detail routing and owner-scoped selection are implemented in the shared wallet flow. | Covered locally. |
| `5ff74529`, `b63c7727` | Bubble is still visually close to the Money Insider ring in the supplied baseline screenshot; current footer/dock code needs a fresh screenshot after the latest build. | Open visual verification. |
| Existing copy review | `Bản lưu ... formatted(date: .abbreviated)` and `Money Insider` remain English/mixed-language in source. | Open P1 copy pass. |

## Giao dịch

| Feedback/design | Finding | Status |
| --- | --- | --- |
| `1db32af3` | Note editing and report-inclusion persistence are implemented and tested locally. Owner confirmation of note placement and physical-device acceptance remain. | Needs UAT. |
| `8eeb848f` | Shared transaction row uses category + note/date rather than redundant wallet name. | Covered locally; owner acceptance open. |
| `89cca866` | Jar selector is expense-only and sends the backend jar ID/month contract. | Covered locally. |
| `2d0811c0` | Shared amount/feedback field bases exist, but full feedback-sheet gutters and clipping still require runtime/device proof. | Needs UAT. |
| `7495cd2f`, `f42c103d` | Shared calendar/money and row spacing bases exist. | Covered locally; populated-screen check open. |
| Earlier P0 duplicate `+` | The local header `+` is still present beside the shell Add button. | Open P0. |

## Trợ lý / AI

| Feedback/design | Finding | Status |
| --- | --- | --- |
| `08246b4b`, `f7887563` | Composer is pinned below history and adapts for keyboard/content size. | Code covered; fresh visual UAT open. |
| `559242fe`, `c86b669d` | Assistant transaction previews now reuse canonical transaction-row/card language. | Code covered; populated visual UAT open. |
| `62e31193` | Markdown rendering and regressions exist; live provider formatting and populated-device acceptance remain. | Needs UAT. |
| `824dea29` | Jar create/update proposals have typed, owner-scoped confirmation; deployed provider and owner verification remain. | Needs UAT. |
| `afee0f86` | Broader autonomous/write action set is intentionally not enabled. | Remaining contract. |
| `1a59aa1b` | Local versioned help exists; external release-aware RAG/source refresh is not implemented. | Remaining contract. |
| Copy/icon review | `Finance Assistant` and raw `message.fill`/SF symbols still appear in source; Vietnamese copy and the verified icon bundle are not complete. | Open P1. |

## Tài khoản / Mở rộng

| Feedback/design | Finding | Status |
| --- | --- | --- |
| `e5428521`, `31b7bcf7` | Duplicate account destination and tab shielding are handled by the shared More route/shell. | Covered locally. |
| `31ef1877`, `d098ff72` | Login/logout/Face ID controls exist; signed-device biometric acceptance is still external. | Needs device UAT. |
| `5ff74529` | Account root Add and feedback dock have local assertions. | Code covered; physical visual/touch UAT open. |
| More menu copy | `Travel Mode` and `Finance Assistant` remain English; raw filled symbols are used for several rows. | Open P1. |
| Disabled bank link | Intentionally informational/disabled, but its low contrast reads like unavailable content rather than a product explanation. | Open P2 copy/contrast review. |

## Màn con: Ví, Hũ, Ngân sách, Báo cáo, Portfolio, Phản hồi

| Area | Feedback/design | Finding | Status |
| --- | --- | --- | --- |
| Ví / savings / credit | `e11d18e6`, `196110ec`, `26e78c7f`, `7ddf77f3`, `019b2e9c` | Routing, masking, wallet scope, cache/outbox contracts exist; device/offline and deployed capability acceptance remain. | Needs UAT. |
| Hũ / tháng | `89cca866`, `4ca27075` | First-class monthly Hũ flow and backend contract are implemented; create/edit/remove/month-copy/cumulative history still need owner-populated UAT. | Needs UAT. |
| Ngân sách / recurring | `52362c1c`, `7495cd2f` | Shared form/card bases and API-backed flows exist; expired/editor and deployed data acceptance remain in the referenced designs. | Needs UAT. |
| Reports / month detail | `4ca27075`, `62e31193` | Server-backed summaries and report hierarchy exist; persisted recalculation, live narrative and owner visual UAT remain. | Needs UAT. |
| Portfolio | `06fb8f5b`, `fc8cae65` | Asset/price/trade workflow exists locally; populated workflow/device acceptance remains. | Needs UAT. |
| Portfolio history correction | `b017284a` | Trade history remains append-only. Edit/delete/correction requires an accounting contract before implementation. | Remaining contract. |
| Feedback screen/overlay | `2d0811c0`, `ecf02904`, `1e12aaa1`, `b63c7727` | Shared sheet/attachment/bubble bases exist; full-sheet gutters, attachment edge cases and physical overlap remain. | Needs UAT. |

## Icon bundle decision

The current implementation is not yet a real Money Lover/Phosphor bundle:

- `MPIconCatalog.usesVerifiedAssets == false`.
- `Assets.xcassets` contains only the application icon set.
- Semantic mapping is useful and safe, but it still resolves to SF Symbols.
- Some system controls can remain SF Symbols; finance/category/navigation icons
  should move to a pinned, licensed Phosphor regular/duotone subset with a
  `NOTICE` file and a fallback only for missing/unavailable assets.

## Audit conclusion

The current UI redesign is not “all done” yet. The high-value local work is in
place, but the audit finds three implementation-level follow-ups before owner
review: remove the duplicate Transactions Add action, complete the verified icon
bundle, and finish the mixed-language/cached-date copy pass. The remaining
feedback records are not one homogeneous UI batch: some are local code proof,
some require device/provider/UAT, and three require explicit product/accounting
decisions.
