# MyPocket native redesign from the Money Lover interaction reference

Status: owner-directed implementation. The owner requested a full native UI
rewrite based on the installed Money Lover app after the earlier REA task
stopped at research. This document turns that request into a bounded product
contract before broad source edits.

## Intent

Make MyPocket feel as coherent, compact and predictable as the observed Money
Lover iPhone application across the whole native product. Reproduce its useful
information hierarchy and interaction families, while keeping MyPocket's
identity, financial rules, security boundaries and additional capabilities.

Success means a user moving between Overview, Ledger, Add/Edit, Wallets,
Categories, Budgets, Reports and Account encounters the same visual grammar:
one navigation model, one grouped-row form model, one sheet/editor model, one
wallet selector, one period selector and consistent loading/empty/error states.

## Evidence

- Physical iPhone survey of Money Lover 8.89.0 records 77 visited states and
  their geometry/behavior in
  `docs/design/reference-survey/2026-10-16-money-lover-iphone.md`.
- Money Lover 8.90.0 is installed on the paired phone. REA 6.8.0 cannot
  directly inspect a physical iOS application, so REA is used for inventory
  and artifact research only; MobileNext/device evidence remains the source
  for native interactions.
- Existing MyPocket screen registry and design bases remain the authority for
  product capabilities and native implementation ownership.

## Goals

- Redesign every current native screen, not only the transaction ledger.
- Centralize visual changes in shared SwiftUI bases and tokens before feature
  screens adopt them.
- Match the observed reference's compact light surfaces, grouped rows, stable
  headers, floating capsule navigation, elevated add action, wallet scope,
  period rail, ledger grouping, form sheets and report hierarchy.
- Preserve accessibility, safe areas, keyboard reachability, Dynamic Type,
  privacy masking, dark-mode correctness and minimum 44pt touch targets.
- Preserve all MyPocket functionality absent from Money Lover: Finance
  Assistant, AI document entry, jars, month notes, portfolio, travel,
  feedback, API keys, offline cache/queue and conflict recovery.
- Close current owner feedback when the redesign directly addresses it, but
  keep technical implementation, deployment and owner acceptance as separate
  evidence gates.

## Non-goals and anti-goals

- Do not copy Money Lover trademarks, icons, illustrations, copy, screenshots,
  proprietary assets or sample financial values.
- Do not change API, schema, accounting semantics or report calculations solely
  to imitate a reference screen.
- Do not remove or hide MyPocket-only features because the reference lacks an
  equivalent.
- Do not implement screen-local colors, radii, padding or one-off navigation
  workarounds when the behavior belongs to a shared base.
- Do not mark production feedback resolved from simulator screenshots alone.

## Approaches considered

### Literal visual copy

Closest to the owner's phrase “giống hệt”, but it would copy brand-specific
assets and produce brittle screen-level duplication. Rejected.

### Shared-system adaptation (selected)

Recreate the reference's layout and interaction grammar through MyPocket's
tokens and shared bases, then migrate every screen in controlled waves. This
provides high behavioral fidelity while preserving product identity,
accessibility and maintainability.

### Independent screen rewrites

Would show fast isolated progress but repeat components and create drift across
the large screen registry. Rejected except for throwaway visual probes.

## Product-wide design direction

- Light neutral/lavender canvas, flat white grouped surfaces, subtle inset
  separators and restrained semantic income/expense colors.
- System SF typography. Balance and key financial values receive the strongest
  hierarchy; compact panel headings remain compact.
- Rounded grouping belongs to surfaces and presented sheets. Buttons keep
  role-appropriate geometry; text fields are row values rather than a field of
  nested outlined cards.
- Bottom navigation is a safe-area-aware floating capsule with no more than
  five visible destinations and an elevated central Add action. MyPocket's AI
  entry remains reachable without displacing core navigation.
- Editors use cancel/title/save headers, grouped rows, dedicated subroutes for
  complex fields and a stable bottom action outside scroll content.
- Transaction rows use a category glyph, title/note/date hierarchy and one
  aligned signed amount. Day/category groups own their totals and count.
- Charts and summaries reveal actual MyPocket values and calculations; the
  reference's sample data is never used as fixture truth.
- Motion is subtle and state-oriented. No decorative animation, glass effects
  or one-note dark-blue palette from generic design-system suggestions is
  adopted when it conflicts with observed native evidence.

## Shared foundation

The first implementation wave owns semantic tokens, page canvas, grouped
surface, inset divider, row geometry, screen/header variants, floating bottom
navigation, period rail, wallet scope, editor shell, amount/date/category form
rows, empty/loading/error states and transaction row/group composition.

Feature screens may only consume these bases. A missing variant is added to the
base first with focused tests and a reference mapping.

## Migration waves

1. **Foundation and shell:** tokens, base components, navigation, feedback
   clearance, common screen geometry and reference fixtures.
2. **Daily money loop:** Overview, Ledger/Search, Add/Edit transaction,
   transaction detail and wallet selection.
3. **Money organization:** Wallets, Savings, Categories, Jars, recurring,
   bills, debt and transfers.
4. **Planning and insight:** Budgets, Reports, month detail, credit and charts.
5. **MyPocket extensions:** Assistant, AI entry, Portfolio, Travel,
   notifications, feedback and account/API-key/security screens, expressed
   through the same grammar without losing their capabilities.
6. **Integrated release:** full native suites, screenshots at small/large
   phones and Dynamic Type, signed physical-device install, production API
   smoke and owner UAT. Only accepted feedback records are marked resolved.

The REA-derived basic ledger search already in progress belongs to wave 2 and
must integrate with, not dictate, the foundation.

## Verification

- Token/base unit tests for geometry, semantics, accessibility identifiers and
  state preservation.
- Focused signed simulator flows for every migrated screen, including keyboard,
  empty/loading/error, privacy masking, back navigation and destructive
  confirmation.
- Screenshot review at iPhone SE and a current large iPhone, light/dark mode and
  large text. No overlapping controls or clipped text.
- Existing financial, API, offline, AI and portfolio test suites remain green;
  redesign work does not authorize business-logic expectation changes.
- Physical iPhone acceptance compares the implemented interaction to the
  observed reference and validates MyPocket-specific extensions.

## Documentation

Each wave updates internal token/base/screen contracts before implementation
and reconciles them afterward. Human guides change when navigation or a user
workflow changes. Existing Markdown guides are also the stable agent-readable
surface; public API/OpenAPI docs change only if a contract actually changes.
