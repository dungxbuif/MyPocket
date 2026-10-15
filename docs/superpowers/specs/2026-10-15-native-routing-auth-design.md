# Native iOS Routing and Authorization Contract

**Status:** approved for implementation by the continuing native-app task  
**Date:** 2026-10-15

## Goal

Every native iOS scene must have a stable route so the app can be opened from a
link, while authentication and owner authorization are enforced before any
protected screen or resource is shown.

## Route contract

The canonical in-app route is an `AppRoute` value. The external representations
are:

- URL scheme: `mypocket://app/<path>`
- Universal-link shape: `https://money.dungxbuif.com/app/<path>`

The first route set is:

| Path | Route | Auth | Optional data |
|---|---|---|---|
| `/overview` | overview | required | — |
| `/transactions` | transactions | required | `wallet_id` query |
| `/transactions/new` | new transaction | required | `wallet_id` query |
| `/wallets` | wallets | required | — |
| `/wallets/<wallet-id>` | wallet detail/ledger | required + owner-scoped | wallet id |
| `/budgets` | budgets | required | — |
| `/categories` | categories | required | — |
| `/jars` | jars | required | — |
| `/months` | months | required | — |
| `/recurring` | recurring transactions | required | — |
| `/travel` | travel mode | required | — |
| `/reports` | reports | required | — |
| `/portfolio` | portfolio | required | — |
| `/credit` | credit ledger | required | — |
| `/notifications` | notifications | required | — |
| `/assistant` | assistant | required | — |
| `/ai-entry` | AI transaction entry | required | — |
| `/account` | account | required | — |
| `/feedback` | feedback | required | — |

Unknown paths are rejected without mutating current navigation. Query values
are treated as navigation hints only; the API fetch still scopes by the current
session owner.

## Authentication and authorization

1. `AppRootView` receives URLs and sends them to a pure `AppRouteParser`.
2. The parsed route is retained as `pendingRoute` until the auth phase is ready.
3. Unauthenticated users are sent to password login; after a successful login,
   the pending route is consumed exactly once.
4. A locked session requires Face ID/passcode first; a refreshed or newly
   authenticated session then consumes the route.
5. An expired/invalid session clears the pending route only after the normal
   login flow has failed; it never displays protected content from a stale URL.
6. Resource IDs in links are never trusted as ownership. Services receive the
   authenticated session owner and must return `accountMismatch`/not-found for
   resources owned by another user.

## Safety and UX invariants

- Opening a link while the app is already running replaces only the intended
  route; it must not reset the selected owner or mask state.
- Cold-start links and links received while locked follow the same gate.
- No token, refresh token, API key, or password is accepted in a URL.
- Routes are additive to the existing `MoreDestination` navigation; feature
  views remain responsible for loading their data.

## Verification

- Unit tests cover canonical parsing, query decoding, unknown-route rejection,
  and scheme/host validation.
- Auth-gate tests cover unauthenticated, locked, authenticated and expired
  session phases plus one-shot pending-route consumption.
- UI smoke tests open overview, transactions, wallet detail and assistant via
  URL in a cold-start and authenticated launch.
- Documentation exposes the route table to humans and agents via this Markdown
  contract and `docs/architecture/MOBILE_IOS.md`.
