---
artifact_type: bug
id: NATIVE-REPORT-OWNER-SUMMARY-FAILSAFE
status: verified
owner: shared
---

# Native Reports must scope report reads and fail closed on summary errors

## Problem

The native Reports screen has an owner-scoped read context, but the report API
helpers do not pass that owner through `APIRequest.expectedOwnerID`. In
addition, `ReportsView` treats `monthSummary` as optional with `try?`, so an
insider response can render a report whose summary/breakdown/activity data is
missing or empty.

## Intended behavior

- `FinanceAPI.monthSummary` and `FinanceAPI.insider` accept an optional owner ID
  and pass it to the existing `expectedOwnerID` request guard.
- Reports passes its owner ID to both report reads when available.
- A month-summary failure invalidates the whole report load. Reports must clear
  all report sections and show the existing retry/error state; it must not render
  insider-only data or fabricated empty breakdown/activity content.
- Existing encrypted snapshot fallback remains exact-month and owner-scoped;
  no backend/API schema, offline/outbox, or cache format change is required.

## Root cause

`monthSummary` is awaited with `try?` after the insider request succeeds, which
converts a required report component failure into `nil`. The report helper
requests also omit the owner expected by the existing client-side auth
contract.

## Verification

- `FinancialSnapshotCacheTests` focused target passes 12/12 on the signed
  iPhone 16 simulator at `/tmp/mypocket-reports-hardening-signed2`.
- `build-for-testing` passes with `CODE_SIGNING_ALLOWED=NO` at
  `/tmp/mypocket-reports-hardening-build`.
- Regression coverage proves both report reads reject a foreign owner before
  transport and the presentation policy requires both server payloads.
- Backend/API schema, offline/outbox, and encrypted cache format are unchanged.

## Scope

In scope: `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`,
`Core/API/FinanceAPI.swift`, `Core/Offline/FinancialSnapshotCache.swift` only
as needed for report policy tests, and native tests/docs. Out of scope:
backend/API schema, offline/outbox files, deployment, and production data.
