# Business Correctness Evidence — 2026-09-11

Status: local verification complete; production provider smoke remains a release gate.

## Independently Calculated Ledger

`backend/internal/analytics/business_ledger_test.go` creates one user and two zero-balance wallets, then executes income, expense, transfer, balance adjustment, an excluded expense, an edited transaction that moves wallet and category, an archived transaction, an obligation repayment, and a gold trade with a manual price.

The expectations are declared independently from the reporting queries:

| Result | Expected |
| --- | ---: |
| Cash wallet | 590,000 VND |
| Bank wallet | 270,000 VND |
| Reported income | 1,000,000 VND |
| Reported expense | 190,000 VND |
| Reported net income | 810,000 VND |
| Food category | 100,000 VND |
| Shopping category | 90,000 VND |
| Wallet net worth | 860,000 VND |
| Investment market value | 200,000 VND |
| Combined net worth | 1,060,000 VND |

The transfer and adjustment affect balances but not income or expense. The excluded and archived expenses do not affect reports. Editing reverses the original wallet/category effect before applying the replacement. Linking the repayment does not duplicate its transaction effect. Portfolio trades do not mutate wallet balances.

## Six-Period Analytics

The comparison endpoint now returns six contiguous equal-duration periods, ordered oldest to current, while preserving the selected wallet scope. The prior-period object is derived from the period immediately before the current period. `ReportsPanel` renders the server-provided values in the accessible `So sánh 6 kỳ` table; it does not reconstruct authoritative totals from visible transaction rows.

The browser accounting journey verifies the API period count, current-period values, and the mounted six-row comparison on mobile Chromium, desktop Chromium, and mobile WebKit.

## Web Push Delivery

- Delivery uses `github.com/SherClockHolmes/webpush-go` v1.4.0 with VAPID and an injected HTTP client.
- Production configuration requires `WEB_PUSH_ENABLED=true` plus complete public key, private key, and subject values.
- The frontend image receives only `VITE_WEB_PUSH_PUBLIC_KEY`; the private key remains server-side.
- Delivery errors expose only generic transport or HTTP status information and do not include endpoint or subscription key material.
- HTTP 404 and 410 expire the subscription. Transient failures retain the bounded retry schedule.
- A subscription is eligible only when a newer durable inbox notice exists. Successful delivery advances its delivery cursor so an unchanged notice is not sent again.
- The browser journey mounts the real notification settings UI and exercises the real subscription API while mocking only the browser/provider push primitive.

Provider delivery against production credentials is intentionally deferred to the final deployment task so secrets are never copied into local evidence.

## Verification

All commands used a PostgreSQL 16 test database at `127.0.0.1:55433`; Redis used the isolated MyPocket instance at `127.0.0.1:6380`.

| Gate | Result |
| --- | --- |
| Ledger integration fixture | pass |
| Finance, planning, portfolio, analytics, notification, config, and HTTP integration packages | pass |
| Full backend `go test -race -p 1 ./... -count=1` | pass |
| Frontend Vitest | pass: 20 files, 163 tests |
| Focused accounting and notification E2E | pass: 12 tests |
| Full Playwright E2E | pass: 90 tests across mobile Chromium, desktop Chromium, and mobile WebKit |

The E2E suite uses ordinary interactions without forced clicks. It includes wallet and transaction CRUD, exact transfer edit/archive reversal, report neutrality, excluded-report behavior, recurring-draft confirmation boundaries, planning links, offline synchronization, user isolation, and Web Push subscription UI.
