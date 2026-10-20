# R0 — current mounted-action inventory

Scope: `App.tsx` plus its mounted Overview, Transactions, Budgets, Account screens and inline sheets; source inspection on 2026-09-09, not a claim that every negative path was browser-tested. [Design](../phases/R0-DETAIL_DESIGN.md), [fresh evidence](R0-FUNCTIONAL-STABILITY-2026-09-09.md), [full parity matrix](../../research/moneylover/FEATURE-PARITY-2026-09-08.md), [backlog](../BACKLOG.md).

Keep the current UI style and use base components for changes. Full R0 is **in_progress**, not complete. Having a handler is not proof of correct persistence, error handling or accounting.

Follow-up evidence: [transaction/API-key error feedback and WebKit diagnosis](R0-ERROR-FEEDBACK-2026-09-09.md): frontend 85 pass, Chrome mobile/desktop 26 pass. The WebKit offline-emulation failure also occurs with a minimal cache-only fixture; real server-down reload passes in that probe. Physical Safari remains unverified.

Latest [receipt readback investigation](R0-RECEIPT-READBACK-2026-09-09.md): 92 frontend pass, Chrome 32 pass, targeted receipt 8 pass/1 retained WebKit emulation failure. Same stored bytes recover without rewriting when emulation is disabled; separate HTTP-outage byte/entity checks pass on all three projects. Physical Safari/PWA/provider acceptance remains blocked. No application/storage changes; full R0 is not accepted.

| Surface/action | Current evidence | Remaining slice |
| --- | --- | --- |
| Login/reload/logout | Live fixture E2E passes; PWA obstruction fixed | Real Google/production UAT, logout server-failure handling review |
| Four tabs / add action | Mounted; add blocked with no wallet | Keyboard/device UAT beyond existing tests |
| PWA install/help/dismiss | Expanded help and dismissal browser tests; standalone/appinstalled unit tests | Physical iOS/Android install; WebKit offline reload remains failing |
| Hide/reveal header balance | State/local preference handler exists | Privacy consistency across report and asset subviews |
| Global search | Shared loading/error/retry/clear states; query retained; stale success/failure and clear/close/offline/logout guards; [100 frontend / 6 targeted browser proof](R0-SEARCH-FEEDBACK-2026-09-09.md) | Cached-result freshness and full result navigation/parity; physical-device UAT |
| Notifications enable/read | API handlers exist; permission status feedback | Mark-read failures swallowed; real push/provider acceptance |
| Conflict resync/keep-server/discard/retry | Current offline conflict E2E passes keep-server | Other resolution/error scenarios still require dedicated browser proof |
| Overview wallets / manage / close detail | Mounted real API; basic finance E2E | Wallet-detail fetch errors swallowed; offline row can look actionable but do nothing |
| Money Insider refresh | Handler calls real API | Failure silently ignored; distinguish unavailable versus no data |
| Overview “Xem báo cáo” | A span without navigation/handler | Wire genuine report journey in reports slice; not complete |
| Transactions search/edit | Note/type local filter; real edit opens | Month control has no handler and does not filter; transfer/adjustment row shows zero; complete in finance slice |
| Budget create/edit/archive | Original named-budget live CRUD now passes; keyboard/offline unit proof | Network errors in sheets still need visible feedback |
| Budget summary | Zero/unspent/exhausted/null data regression proof added | Mixed-period/overlapping categories can make summed budgets misleading; financial/reporting design required |
| Budget “Tất cả các nhóm” chip | No callback | Real scope selector belongs to budgets slice |
| Event/debt create/link | Live planning E2E passes | Negative/retry atomicity; debt record is metadata, not complete cashflow |
| Recurring create/archive | Create has live planning proof; archive handler exists | Existing schedule cannot be fully edited; complete create-to-confirm in planning slice |
| Draft “Xem” | No handler; pending records only displayed | Confirm/reject API and real review journey missing; do not pretend a click confirms funds |
| Add expense/income/debt | Live create/offline replay; shared save-error feedback retains form, explicit retry browser proof | Debt is not cashflow-equivalent; separate obligation negative paths remain |
| Add “Với”, location, event, reminder | Shared native disabled actions explain “Chưa hỗ trợ trong form này”; inert chevrons removed | Actual linkage/location/contact/reminder features remain unimplemented; no domain capabilities silently promoted |
| Add receipt | Both actions open one picker; selection survives collapsed details/cancel; remove/reselect and debt guard tested; shared footer clears toolbar overlap; exact byte/entity checks pass under separate HTTP outage on all three projects | **Release gate:** original WebKit offline-emulation readback still fails, isolated in minimal fixture; physical Safari/PWA and remote provider/attachment replay/viewing still need proof |
| Edit transaction / archive | Live CRUD/conflict proof; rejected writes now retain fields and show safe error/correlation ID | Transfer/adjustment editing contracts and additional conflicts |
| Wallet manager create/rename/hide/default-AI | Actual handlers; manager has error state; create/rename covered by live finance test | Per-action error/version/offline checks, no claiming full wallet parity |
| Category create/rename/hide | Existing manager UI/API paths | Hierarchy/activation/merge remain distinct unimplemented contracts |
| Assets add | Mounted add handler | Buy/sell/price/archive callbacks not used by mounted Account; old AssetRow is unmounted; next portfolio slice |
| API key list/create/revoke/copy | Cookie/bearer isolation; list/read retry/create/revoke/copy feedback; confirmed mutations survive refresh failure; stale list/copy response regression proof | Physical clipboard/device UAT; API completeness is separate from key lifecycle UI |
| Hidden audit refresh | Authorization check and real request | Refresh failure swallowed; production role/UAT remains |
| Docs link | `/docs/` link exists | Full machine-validated OpenAPI and third-party examples remain separate API work |
| Account labels | Premium/iPhone/version hardcoded | Replace with factual app/device/build information, without restyling |
| Settings/security/travel/category prototype screens | Not mounted from current App | Do not wire prototype sample data; implement real flows only within approved scope |

## Next execution order within R0

1. Focused receipt investigation is in review; retain the original failing WebKit emulation test and arrange physical Safari/PWA/provider acceptance separately. No storage migration justified by current evidence; no blanket device-ready claim.
2. Continue visible error handling on notification mark-read, wallet-detail/Insider, planning and audit paths; search, add/edit/archive transactions and Account API keys now have bounded verification. Keep drafts/form values on failure; never report success until confirmed.
3. Receipt footer and unsupported quick-add controls are now explicit/wired, not full receipt acceptance. Continue other existing capability gaps without mounting placeholder screens as functional completion.
4. Then proceed to approved finance and reports slices for real month filters/report drilldown and accounting contracts. Update OpenAPI/examples and test cookie/bearer on each changed API.
