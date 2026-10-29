---
artifact_type: design_research
status: in_review
owner: dedicated_rea_ui_task
updated: 2026-10-26
---

# REA and Money Lover UI research

## Intent and scope before tool installation

The owner requested the official `morluto/rea` tool and an evidence-backed
Money Lover UI interaction survey to inform MyPocket native iOS design. This
note owns research and a bounded proposal only. It does not authorize MyPocket
screen edits, financial writes, deployment, or copying Money Lover assets,
source, binaries, or branding.

The official target is installed on the paired iPhone as
`me.moneylover.ios.Money-Lover`, version 8.90.0 (9258), verified with
`xcrun devicectl device info apps --include-default-apps --search Money` on
2026-10-26. The existing [physical survey](2026-10-16-money-lover-iphone.md)
records direct interactions with 8.89.0 (9250); its observations are historical
until reconfirmed on the newer build. The public [Money Lover site](https://moneylover.me/)
and [Finsify App Store listing](https://apps.apple.com/us/app/money-lover-money-manager/id486312413)
are secondary official evidence. Neither public source proves an unvisited app
interaction.

## Technical decision and verification plan

- Install the official npm package `rea-agents` in an isolated prefix with a
  compatible Node version. Do not run REA agent setup or modify global Codex
  configuration. Record package URL, integrity, installed path, version and CLI
  capabilities. The current default Node 24.0.2 does not meet REA's engine
  constraint (`^22.23.4 || ^24.22.1 || >=26.12.1`).
- Check whether REA supports this iOS target directly. Use only a capability
  the official tool actually provides; distinguish any public website analysis
  from native iPhone observation.
- Inspect read-only reference states where the paired-device helper is available.
  Do not save or mutate reference financial data. Record observed version,
  action path, visible state and evidence limits.
- Compare only relevant current MyPocket SwiftUI screen/base source and existing
  feedback. Classify already implemented behavior separately from a real gap.
  Recommend one first slice with named bases, acceptance checks and approval
  boundary.
- Reconcile this note with observed results. Since this is research without a
  public product, API, schema, configuration or setup contract change, public
  guides and agent-facing public references are not expected to change. The
  existing shared survey, `docs/CONTEXT.md`, backlog and validation matrix are
  read-only in this task because other workers own them.

## Installed REA and its boundary

The [official REA repository](https://github.com/morluto/rea) points to the
`rea-agents` npm package. npm metadata for 6.8.0 reports the same repository,
tarball `https://registry.npmjs.org/rea-agents/-/rea-agents-6.8.0.tgz`, and
integrity `sha512-9QDeRnGqEDzABZyr04eLnBXBv/c6I9/TWu/6M0MD4pEfkySI/BIgwTODIbAReX3C1x6hjw9+Jlx4A2yUSPZ9ww==`.
Official Node 24.22.1 was installed by the existing nvm installation with a
matched SHA-256 download checksum. `npm install --global --prefix
/Users/dungxbuif/.local/share/rea-cli rea-agents@6.8.0` completed, and the
binary at `/Users/dungxbuif/.local/share/rea-cli/bin/rea` reports `6.8.0`
when run with `/Users/dungxbuif/.nvm/versions/node/v24.22.1/bin` on `PATH`.
The package root is `/Users/dungxbuif/.local/share/rea-cli/lib/node_modules/rea-agents`.
The machine's default Node remains 24.0.2; no repo dependency, global npm
prefix, or Codex agent registration was changed. npm skipped the optional
Frida install script; this task does not require it.

`rea --help` and `rea capabilities` show macOS native UI observation, browser
CDP inspection, Apple artifact inventory, and Android device/APK analysis.
There is no direct physical-iOS-app UI provider exposed for the installed
Money Lover app. Apple artifact analysis would need a separate local IPA/app
artifact, and macOS AX does not inspect a phone screen. `rea doctor` reports
the compatible Node and CLI package but an unhealthy **audit-wide** setup
because optional analysis engines and agent registrations are absent. That
result does not invalidate the working CLI. No `rea setup`, Hopper, Ghidra,
JADX, Frida helper, mirror, or IPA download was used. REA cannot substantiate
the native iPhone flow claims below; those rely on the direct prior survey and
official product documentation.

## Target and UI evidence

`devicectl` currently identifies a paired iPhone 15 Plus and the official
Money Lover bundle at 8.90.0 (9258). The existing device helper is installed,
but the phone was displaying MyPocket during this task. This pass did not take
over the shared device or claim that 8.89.0 visuals are unchanged in 8.90.0.
The [physical survey](2026-10-16-money-lover-iphone.md) records 8.89.0 direct
AX and visual observations with action IDs; screenshots were kept outside Git.
The [official transaction guide](https://moneylover.zendesk.com/hc/en-us/articles/35483252404889-Create-edit-and-delete-transactions),
[search guide](https://moneylover.zendesk.com/hc/en-us/articles/40818955866649-Search-for-transactions),
and [transaction-history guide](https://moneylover.zendesk.com/hc/en-us/articles/35278612706329-View-transaction-history)
corroborate broad flows, but may describe an earlier release.

| UI flow or state | Money Lover evidence | MyPocket current source/feedback | Design consequence |
| --- | --- | --- | --- |
| Ledger scan and grouping | Survey ML-51/52: wallet capsule, period controls, income/expense/net, day groups, category grouping and overflow. Official history guide describes time range and group options. | `TransactionsView` has `MPWalletScopeSelector`, week/custom selection, `MPWeekNavigator`, `MPMetricRow`, `MPTransactionRow`, grouping and overflow. The transaction screen spec already maps the observed regions. | Existing implementation; compare rendered states before proposing visual changes. Do not relabel it as missing. |
| Ordinary entry | Survey ML-44-49/64-66: segmented type, amount, category, note, date stepper, more-details branch and pinned Save. Official transaction guide lists the same core fields. | `TransactionEditorView` composes `MPSegmentedControl`, `MPFormRow`, `MPAmountField`/`MPAmountKeypad`, `MPDateStepper`, note sheet and `MPCheckboxRow`; the save footer is pinned. | Core interaction already exists. Preserve its MyPocket category, VND, offline-queue and explicit-save rules. |
| Note and report inclusion | Survey ML-44/45; official guide describes exclusion. | [Feedback 1db32af3](../../work/bugs/FEEDBACK-1db32af3-native-transaction-entry.md) records simulator proof of note tap and persisted exclusion. [Feedback 8eeb848f](../../work/bugs/FEEDBACK-8eeb848f-native-transaction-row.md) proves note-first row subtitle without repeated wallet name. | No repeat fix; physical/deployed owner acceptance remains a separate open gate. |
| Search | Survey ML-58-63: ledger search entry, basic query, separate advanced filters. The official search guide confirms category/note basic search and independent advanced search. Sample-mode broad results did not prove exact query semantics. | Aggregate/basic `TransactionsView` has no search action. Goal-wallet detail in `WalletsView` has a local category/note/type search using `MPTextField`; `FinanceAPI.transactions()` loads `/transactions` and `TransactionsViewModel` filters the returned list by wallet/period. | A basic, clearly scoped search is a genuine candidate. Do not imply server-wide history or copy advanced filter behavior from the reference. |
| Wallet selection | Survey ML-11/12 shows total row, selected wallet rows and management route. | `MPWalletScopeSelector` and `MPWalletSelectionList` already own those native interactions. | Keep the existing aggregate/`is_in_total` business policy; no wallet-selector rebuild. |

Money Lover's sample data and account-specific values are not MyPocket
fixtures. Reference color, icons, copy, financial rules and branded assets are
not implementation inputs. The [native base guideline](../platforms/ios-base-guideline.md)
and [base contracts](../system/BASE_COMPONENTS.md) own MyPocket's palette,
geometry and accessibility; MyPocket's Finance Assistant, jars, travel,
portfolio and offline behavior remain intact.

## Proposed bounded first UI slice (approval required)

**Basic search within the ordinary aggregate/basic ledger's current wallet and
period.** Add a search action to that ledger header, show a shared `MPTextField`
with clear/close controls, and filter its already-loaded *confirmed* rows by
category name or note with case-insensitive matching. The scope label must say
which wallet and period are being searched. The existing income/expense/net
summary continues to describe the whole selected period; the result count and
day/category group totals describe only matches. Empty results use
`MPEmptyView`. Changing scope or period reruns the same query locally; closing
search restores the unfiltered list. Pending queue rows, stale-cache warning,
privacy masking, detail navigation, bulk-selection eligibility and existing
group mode stay correct. No network search claim, API/schema change, or new
financial write is in this slice.

This composes `MPScreenHeader`/`MPIconButton`, `MPTextField`, `MPButton`,
`MPEmptyView`, `MPTransactionRow`, the current grouping presentation and
`TransactionsViewModel.applyFilters`. Any missing search-header variant belongs
in the shared base before screen use. Acceptance requires a design mapping for
the action/open/typing/clear/empty/close states, focused filtering tests for
wallet/period/category/note and stale-cache behavior, a signed simulator
interaction check, and populated small-phone/large-text visual comparison.
Source must be rechecked at implementation time because parallel workers are
editing native UI.

Two approaches were considered: this in-view search reuses the existing loaded
ledger and makes its limited scope explicit; a full-history advanced search
would need an API/query and pagination contract plus broader design approval.
The narrow slice is recommended first because it yields the observed entry
interaction without promising reference-app advanced semantics. A purely
visual period-rail change remains separate because the currently supported
week/custom periods do not match the reference's broader choices.

## Open access, approval and docs review

- **Access:** the official app is present, so no target download is needed.
  Fresh 8.90.0 screen comparison needs an agreed read-only window on the shared
  unlocked phone; no current-version visual claim is made here. If that window
  is unavailable, the 8.89.0 survey and official help remain the bounded
  reference evidence.
- **Approval:** the owner must approve the basic-search scope and its design
  artifact before any MyPocket screen/base code is changed. The broader native
  visual redesign is outside this task, particularly while other workers edit
  native UI.
- **Docs:** this internal human/agent-readable research note is the only repo
  file changed. Existing screen specs, context, backlog and validation are
  unchanged because this is a proposal, not shipped behavior. Public human
  guides and AI-agent-readable public surfaces need no change until a user
  workflow actually changes. No API/config/schema, production finance or
  deployment change occurred. Verification here was source/doc inspection,
  package identity and CLI version/capability checks; product UI tests are for
  the approved implementation slice.
