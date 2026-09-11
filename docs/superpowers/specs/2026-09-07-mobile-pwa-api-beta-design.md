# Mobile PWA and API Beta Design

## Goal

Ship a mobile-first beta that is installable as a PWA, remains useful after a
network loss, and exposes every user-owned business API through a bearer API
key.  Money Insider must only render server-derived financial values.

## Decisions

- Keep the existing lightweight SVG chart components.  They match the approved
  visual system and do not need a chart dependency for the fixed mobile charts
  in this beta.
- API keys authenticate as their owner with `Authorization: Bearer mpk_...`.
  They may call every user-owned business route.  OAuth, health, and API-key
  create/revoke remain browser/session or unauthenticated operations.
- The service worker owns only the app shell and safe GET cache fallback;
  IndexedDB remains the authoritative offline mirror/outbox boundary.
- The beta install surface must have a web manifest, maskable icons, a stable
  standalone start URL, and visible offline/reconnect state.
- No report, chart, or insight may use invented finance values. Empty and
  loading states are explicit.

## Scope

1. Complete the install metadata and cache the PWA shell safely.
2. Document and prove bearer API-key use across the existing API middleware.
3. Replace placeholder Money Insider chart values in the new Overview screen
   with values derived from its report payload; show an empty chart state when
   no report series is available.
4. Add focused component/platform tests and reconcile the public API docs.

## Non-goals

- API-key scopes, rate limiting, or third-party OAuth.
- A new charting dependency.
- Changing transaction-accounting semantics or recurring behavior.
- Replacing the established IndexedDB sync protocol.

## Data flow

```text
+----------------+      Bearer mpk_...       +-------------------+
| Installed PWA  | ------------------------> | Go API middleware |
| shell + IDB    | <--- user-scoped JSON ---- | cookie/API key    |
+-------+--------+                            +---------+---------+
        |                                               |
        | GET cache fallback                            v
        v                                      +-------------------+
+----------------+                              | PostgreSQL        |
| Service worker |                              | user-owned data   |
+----------------+                              +-------------------+
```

## Verification

- Unit/component: manifest metadata, report-derived chart data, empty states.
- Browser E2E: service-worker registration, offline reload, mobile viewport.
- Backend: bearer API-key authentication plus user-owned route access.
- Build: `npm test -- --run`, `npm run build`, relevant Go HTTP tests.
