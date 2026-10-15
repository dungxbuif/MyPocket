---
artifact_type: detail_design
id: UI-WALLET-02
status: in_review
owner: shared
approval: owner_requested_reference_reimplementation_and_rules
---

# Add Wallet and wallet type selection

## Current verified implementation — 2026-10-08

The normalized wallet reference now covers both management and creation. Native
management uses `MPWalletSelectionList`: included-only “Tổng cộng”, grouped
included/excluded rows, Sửa/Xong mode, row-to-edit, delete in the editor, footer
actions, and the separately retained Chi tiết extension. Aggregate selection
opens the unscoped transaction ledger; wallet selection opens the scoped
ledger. The API-backed simulator test also caught GORM replacing explicit
`is_in_total=false` with its `default:true`; repository creation now persists
the explicit value transactionally, preserving the API default when omitted.

Verification: backend `go test ./...` against the disposable stage database,
native 130/130, wallet-management UI 1/1 including confirmed API deletion,
Release warnings-as-errors and
project/base guards pass. The rendered phone-width list screenshot was inspected
and its first version's narrow row wrapping was corrected. Owner visual/device
acceptance remains pending. No API, schema or migration change.

Latest owner decision supersedes the picker design below: use BaseSelect while creating, immutable type after save. Canonical [wallet page specification](../../design/pages/wallets/README.md) now covers list/create/edit, base mapping, type-specific behavior and official Money Lover sources. Runtime uses BaseSelect; authenticated owner visual/device UAT remains pending.

Docs normalization: Small task exemption: yes. Reason: reconcile owner decisions in documentation only. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. Docs review completed; docs-only UAT not required. Previous test evidence is not proof of the pending select implementation.

Owner explicitly requests stricter base rules and reimplementation against supplied Add Wallet reference. Scope is existing wallet UI; API, DB, auth and runtime contracts unchanged. Existing uncommitted corrections are preserved. No phase applies.

Source: [Add Wallet](../../design/references/add-wallet/README.md), [wallet selector](../../design/references/wallet-selector/README.md). The second image selects existing wallets, not wallet types; type selection adapts its list-row composition to the three supported types. Existing-wallet selection is implemented in the current verified follow-up above.

Approach: reuse SurfaceCard, Text, Heading, Divider, IconBadge, BaseButton, BaseTextInput and BaseBottomSheet. Extend sheet with base-owned form header (cancel/title/save), canvas tone and tall presentation. Add BaseSwitch and a type selector composed from existing row buttons/icons/text. Add Wallet groups name, VND and opening balance in one card; type-specific required inputs remain. No note on create. Currency is fixed VND; service integration is disabled with explanation. Existing edit form and API conversion remain compatible. The supplied existing-wallet selector is implemented as a management list, not a type picker.

State: single sheet switches between form and type selection; back/Escape from picker preserves draft, full cancel discards draft; save disabled for blank name/invalid money/missing target or limit and during request. Error stays inside sheet. Switch checked means is_in_total=false. No nesting of modal focus traps.

Rejected: copying export HTML creates duplicate colors/controls; native select alone misses the reference row-selection interaction. Shared theme/radius override export colors/radii.

Rules: extend AST protection to SurfaceCard and other shared bases, reject clickable native wrappers, add regression fixtures. Reference PNG/HTML are immutable input evidence beside required Markdown specs, never runtime code. New source bundles must be explicitly declared in their README; ordinary non-Markdown design artifacts remain rejected.

Proof: design guard regressions, SSR form validity/switch/type rows, check:design, test:design, test:transactions and build; API-backed native management and form flows are recorded in the validation matrix. Owner visual acceptance remains separate.

Docs review/reconciliation: [base contracts](../../design/system/BASE_COMPONENTS.md), [page](../../design/pages/wallets/README.md), [ADR](../../decisions/ADR-003-wallet-reference-enforcement.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [release](../../releases/CHANGELOG.md), [parent ticket](TICKET-01-02-quan-ly-vi.md). The 2026-10-08 persistence correction changes no API, schema or ERD.

## Evidence and handoff

`rtk proxy npm run check:design`, `rtk proxy npm run test:design`, `rtk proxy npm run test:transactions`, `rtk proxy npm run build`: pass in app. Seven AST regression tests, shared SSR assertions (validity, switch semantics, type options) and four ledger helper tests. First guard run revealed five pre-existing card typography overrides in group screens; migrated to Text props, second run passed. The original Add Wallet slice did not change backend behavior; the later persistence correction is documented above.

Both newly supplied references now have normalized Markdown region/base/state contracts; original evidence retained. Browser visual and interactive end-to-end UAT not completed; owner review pending, not done. No save-to-API browser claim is made for this new form. Docs review completed: base/screen/ADR/context/backlog/validation/release links reconciled. The native wallet-management and explicit `is_in_total=false` persistence follow-up is recorded above; no public API/schema changes.
