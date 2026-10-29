---
artifact_type: docs_audit
id: IOS-DOCS-AUDIT-20261020
status: verified_local
owner: shared
reviewed: 2026-10-20
scope:
  - mobile-ios/project.yml
  - mobile-ios/Config/Debug.xcconfig
  - mobile-ios/Config/Release.xcconfig
  - mobile-ios/MyPocketIOS/Resources/Info.plist
  - mobile-ios/MyPocketIOS/Features
  - backend/internal/controller/http
  - backend/internal/config/config.go
---

# Native iOS documentation audit

## Scope and evidence

This audit compares the implemented native routes/features and backend route
composition with the public Markdown guides. It is a repository audit only: it
does not probe production, certify provider credentials, or claim iPhone UAT.
The source of truth for the route inventory is the Go router and registrations
in `backend/internal/controller/http` plus the native screen composition in
`mobile-ios/MyPocketIOS/Features/More/MoreView.swift`.

## Build/configuration status

| Area | Evidence | Documentation status |
| --- | --- | --- |
| iOS target, iOS 17, bundle `com.mypocket.ios` | `mobile-ios/project.yml` | Covered by [iOS build handoff](../../guides/ios-build-handoff.md) |
| Debug/Release API selection | `Config/Debug.xcconfig`, `Config/Release.xcconfig`, `Info.plist` | Covered by handoff; physical device must not use loopback Debug URL |
| Portrait-only behavior | `Info.plist`, `MyPocketIOSApp.swift` | Covered by handoff and native architecture docs |
| Email/password + refresh + Face ID | `AuthView`, `AuthViewModel`, `DeviceSignInStore`, auth routes | Covered by [Face ID guide](../../guides/face-id.md) and [AUTH_API](../../architecture/AUTH_API.md) |
| Personal Team signing | Xcode project plus empty personal entitlements file | Covered by handoff; older operation notes contain historical Team IDs and should not be copied as personal configuration |
| MCP/mobile-MCP | No implementation evidence in this audit | Pending; intentionally not presented as available |

## Implemented surfaces and current documentation coverage

These surfaces have a human guide and/or an agent-readable API contract. The
status below means “implemented in the checked-out source”; it does not mean
the current production image or a physical device has been reverified.

| Surface | Native/API evidence | Public documentation |
| --- | --- | --- |
| Login, refresh, device credentials, logout | `routes_auth.go`, `auth_device_handler.go` | `architecture/AUTH_API.md`, `guides/face-id.md` |
| Transactions, transfers, adjustments, bulk delete | `routes_transaction.go`, `TransactionHandler` | `architecture/API.md`, `guides/transactions.md`, `guides/offline-sync-api.md` |
| Wallets and account profile/timezone | `routes_wallet.go`, `routes_account.go` | `architecture/API.md`, `architecture/AUTH_API.md`, native architecture guide |
| Categories and category archival | `routes_category.go` | API master and AI tool catalog; no focused user guide (gap below) |
| Budgets, recurring budgets, report/drill-down | `routes_budget.go`, `routes_budget_series.go`, `budget_report_handler.go` | `guides/budget-recurrence-api.md`, `guides/budget-report-api.md` |
| Savings reports | `main.go` report registration and `SavingsReportHandler` | `guides/savings-report-api.md` |
| Advisor chat, structured cards, async runs, reset/cancel | `ai_advisor_handler.go`, native `AssistantViews.swift` | `guides/finance-assistant-ios.md`, `architecture/AI_ADVISOR_TOOLS.md` |
| AI receipt/attachment entry | `ai_entry_handler.go`, native AI Entry view, OCR adapter | API master, `architecture/OCR_API.md`, ADR-005/006; focused human workflow guide is still missing |
| Jars and month notes | `jar_handler.go`, `month_handler.go` | API master and AI catalog; focused user/agent guide is still missing |
| Recurring transactions | `recurring_handler.go` | API master; focused guide is still missing |
| Credit ledger and statements | `credit_handler.go` | API master and native architecture; focused user/agent guide is still missing |
| Travel Mode | `travel_handler.go`, transaction travel endpoint | API master and design page; focused user guide is still missing |
| Portfolio and immutable trade correction | `portfolio_handler.go` | `guides/portfolio-ios.md` |
| In-app notifications and device-token registration | `notification_handler.go` | `guides/push-notifications-api.md`; APNs delivery remains explicitly pending |
| API keys and full-owner API-key auth | `api_key_handler.go`, auth middleware | `guides/api-keys.md`, `architecture/API.md` |
| Feedback/changelog | `routes_feedback.go` | `architecture/FEEDBACK_API.md`, `guides/feedback-ios.md` |

## Documentation gaps to close next

1. **AI Entry user guide** — document the native entry route as a review-first
   flow: multipart `POST /api/v1/ai/entry/process`, up to 20 JPEG/PNG/PDF files
   at 5 MiB each, OCR wait/poll behavior, retained request ID after uncertain
   acceptance, proposal edit/approve/reject, and private attachment download.
   The API contract exists in `architecture/API.md`, but a user-facing guide
   is not discoverable from `docs/README.md`.
2. **Jars and month detail guide** — explain monthly configuration, allocation
   modes, report inclusion, note CRUD, and the fact that jars do not create
   independent wallets. The route contract is currently buried in `API.md`.
3. **Credit wallet guide** — explain purchase/refund/fee/interest entries,
   paired payments, statement filters, and why ordinary transaction/transfer
   routes reject credit rows. The API contract is present but no task-oriented
   guide exists.
4. **Category management guide** — document CRUD, wallet applicability,
   archive/restore, and inactive-category behavior. Native UI exists in
   `CategoriesView`, but only the API master/AI catalog describes it.
5. **Recurring transaction and Travel Mode guides** — separate ordinary
   recurring schedules from recurring budget series, and explain active travel
   event auto-linking/clear behavior. Current material is API/design-only.
6. **Native route registry** — expose the actual More destinations (including
   AI Entry, Feedback, Notifications, Credit, Travel, Portfolio, Jars and
   Months) in one agent-readable screen/route table. Existing architecture
   text describes parts of this list but is distributed across files.

## Stale or potentially misleading documentation

These are not silently “fixed” in this audit because they are shared master or
historical records; they need owner/coordinator reconciliation:

- `docs/architecture/ARCHITECTURE.md:59` says S3 is future-only and unused by
  transient OCR, while `backend/cmd/api/main.go:211-215` requires private S3
  outside development and the AI-entry service is wired with attachment
  storage. Reconcile the sentence with the current private receipt flow.
- `docs/architecture/MOBILE_IOS.md:6-30` is dated 2026-10-08 and contains old
  test counts/open items. Treat its historical counts as non-authoritative;
  use dated validation artifacts for current evidence.
- `docs/operations/IOS-PERSONAL-BUILD.md` and
  `docs/operations/IOS_PERSONAL_TEAM_SIGNING.md` contain historical Team IDs
  and different entitlement paths. A user must use the placeholder-based
  [handoff guide](../../guides/ios-build-handoff.md), not copy those IDs.
- `docs/operations/DEPLOYMENT.md` records a historical production route/image
  state and explicitly notes local portfolio route drift. Do not convert those
  records into a current production claim without a fresh read-only audit.
- `docs/guides/finance-assistant-ios.md` and release notes contain deployment
  language tied to dated rollout records. This audit does not revalidate those
  claims against the current production image.

## Backend environment contract audit

The app has no provider secrets. The Go config reads the following server-side
names from process environment or ignored `backend/.env.local`; production
uses its deployment secret store. Values are intentionally omitted:

```text
DATABASE_URL REDIS_URL JWT_SECRET REFRESH_TOKEN_TTL_SECONDS
AI_BASE_URL AI_API_KEY AI_MODEL AI_STORE_USAGE
OCR_API_URL OCR_API_KEY
S3_ENDPOINT S3_REGION S3_BUCKET S3_PREFIX S3_FORCE_PATH_STYLE
S3_ACCESS_KEY_ID S3_SECRET_ACCESS_KEY
GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL
GOOGLE_IOS_CLIENT_ID GOOGLE_IOS_CLIENT_SECRET GOOGLE_IOS_REDIRECT_URL
FEEDBACK_AGENT_TOKEN APNS_ENABLED APNS_TEAM_ID APNS_KEY_ID APNS_TOPIC
APNS_PRIVATE_KEY APNS_ENVIRONMENT CORS_ALLOWED_ORIGINS
ALLOWED_LOGIN_EMAILS SEED_USER_EMAIL SEED_USER_NAME SEED_USER_PASSWORD
```

`backend/.env.local` is ignored by Git. Never move these values into
`Debug.xcconfig`, `Release.xcconfig`, `Info.plist`, Swift source, or the
Personal Team build. The iOS app receives only the API base URL and user
session/API-key credentials entered through the app.

## Verification performed

- Inspected the native project/config/plist, route registrations, API handlers,
  config loader and existing public guides.
- Verified the native Personal Team build handoff against the actual bundle ID,
  configs, URL scheme, entitlements and release preflight script.
- No production endpoint, secret store, Apple account, device, migration or
  MCP service was contacted or changed by this audit.

