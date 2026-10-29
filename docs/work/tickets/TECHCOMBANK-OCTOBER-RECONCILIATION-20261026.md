# October 2026 Techcombank reconciliation

Status: data import verified; AI and feedback follow-up in progress. Owner request: add missing October activity to the one
existing Techcombank wallet, verify recurring entries, and evaluate AI entry
against the supplied real statements. This work is an owner-authorized data
operation, not a change to bank integration or automatic bank syncing.

## Sources and boundaries

- Two Techcombank XLSX account statements (118 combined rows, October 1-26)
  and one card-history PDF (October and September rows). Source files remain in
  the owner's Downloads folder; do not commit raw statements or account numbers.
- The physical iPhone's MyPocket app is logged in to the owner's production
  account. Ordinary bank and October credit-card activity targets the one
  existing `Techcombank` wallet. The owner subsequently approved one separate
  savings wallet for the 10,000,000 VND `Tich Luy Nhu Y` deposit on October 20.
- Exclude same-owner Techcombank-to-Techcombank movements, including matched
  transfers between the two statements, credit-card bill payments, automatic
  savings reallocations, and card installment rebookings. Preserve external
  spending/income and distinguish interest/refunds from internal principal.
- Never infer an internal transfer solely from the sender name; require matching
  reference/amount or an unambiguous Techcombank product description. Ambiguous
  rows stay in the review queue and are not written silently.
- Represent that 10,000,000 VND deposit as one historical wallet transfer from
  Techcombank to the new savings wallet, excluded from income/expense reports;
  do not add a second standalone bank expense. Do not assign unrelated CCTG or
  `LocVang` movements to the new wallet without evidence.
- Owner requested the new wallet name `Tiết kiệm lâu dài`. The existing
  `Tiết kiệm dài hạn` wallet is separate and must not be renamed or merged.
  The transfer API enforces paired system transfer categories; the custom
  expense category `Tiết Kiệm` cannot be assigned to that pair without
  changing its transfer semantics. Keep the pair excluded from reports and
  identify savings in the wallet name and transfer note.
- Per owner instruction, ignore the small `Sinh loi tu dong` interest credits;
  do not add them as income or include them in import totals. The related
  principal reallocations remain excluded as internal bank movements.

## Execution and verification

1. Parse each source structurally; prove opening + credits - debits = closing
   for both bank statements. Parse the PDF by visual row order across pages and
   only use October card events with VND amounts. Avoid double-counting a
   purchase and its installment rebooking.
2. Obtain a temporary full-owner API key from the logged-in app through Mobile
   MCP; never print or commit it. Read wallets, categories and all October
   transactions through authenticated production API. Revoke the key once the
   verified work is finished.
3. Build a canonical, source-referenced candidate list. Match existing records
   by date, signed VND amount, source description/reference and wallet, with
   manual review for collisions. Submit only missing, unambiguous rows using
   canonical transaction API and stable idempotency IDs where supported.
4. Re-fetch every written row and the October ledger. Record counts for source,
   internal, existing, inserted and held-for-review; a second reconciliation
   pass must propose zero duplicate inserts. Do not adjust the wallet opening
   balance to force a match across multiple underlying bank products.
5. Check recurring occurrence records and bank evidence separately: an app
   schedule firing is not proof that the bank charged the merchant. Test AI
   entry extraction in review-only mode against selected real samples, measure
   latency and correctness, and reject proposals rather than approving any
   unverified financial mutation.

## AI extraction observation

The review-only production PDF run returned in about 14 seconds with 17
proposals. Most proposals used today's date rather than the statement date,
one foreign-currency amount was treated as 222 VND, and legitimate repeated
ticket rows were collapsed into a single proposal. None is approved for
insertion. Canonical statement parsing and source-row matching remain the
authority for this import.

## Production reconciliation, 2026-10-26

- Both bank exports balanced independently against their stated opening and
  closing balances (118 October rows total). The card PDF supplied 31 October
  rows; only 19 were external purchases/refunds/interest after excluding bill
  payments, installment principal and a same-day charge reversal.
- Created `Tiết kiệm lâu dài` as a new basic wallet. A single 10,000,000 VND
  historical transfer on October 20 links it to `Techcombank`; both legs are
  excluded from income/expense reports. The older `Tiết kiệm dài hạn` goal
  wallet was untouched. The paired transfer must retain the system `Tiền
  chuyển đi`/`Tiền chuyển đến` categories. Its savings purpose is visible in
  the wallet name and note; assigning the custom expense category `Tiết Kiệm`
  would misrepresent the transfer and break pair semantics.
- Imported 58 missing bank entries and 14 missing card entries with stable
  source-row idempotency keys. A second pass found 68/68 candidate bank rows
  and 19/19 candidate card rows represented, with zero further inserts.
  Existing manual entries accounted for 10 bank and 5 card rows. Of bank
  rows, 31 remain held for review (self-named transfers without a confirmed
  other account, `LocVang`, large opaque debits, and a 30,000 VND `Nuoc` row
  potentially corresponding to a 15,000 VND manual entry). Across both bank
  and card sources, another 31 rows were excluded as internal principal
  movement, bill payment, same-day card reversal or small interest. Thus all
  149 source rows partition into 87 external candidates, 31 held bank rows and
  31 excluded rows. No held row was silently written. The owner explicitly
  decided to keep all seven `LocVang` rows (215,000 VND total) and the unclear
  self-named transfer rows outside MyPocket pending manual reconciliation; their purpose
  and destination are not to be inferred or imported automatically.
- `Google One` has an October 21 app occurrence and a linked 250,000 VND
  transaction. The supplied bank files do not establish the underlying bank
  charge: account statements end October 26 without that amount, while card
  history ends October 19. `Apple Icloud` next due is November 27, so no
  October scheduled occurrence was expected. This separates app scheduling
  verification from bank settlement.
- A temporary owner API key was created through the logged-in iPhone. During
  UI inspection a stale coordinate reference briefly revoked the existing
  `feedback` key. The same database key record was restored immediately with
  a guarded update; the API confirmed it and the temporary key active. Recheck
  the feedback integration before closing this ticket. The existing key's
  `revoked_at` was confirmed null after restoration. The temporary import key
  was revoked through the API with HTTP 204 after the second-pass and device
  checks; no further production writes may use it.

## Scope and docs

No public API, schema, security policy or app workflow is intended to change.
The internal note will record aggregate evidence and held exceptions without
raw financial statements or credentials. If a code defect is reproduced, open
or update a separate bug/design before implementation and reconcile public and
agent-readable docs only if behavior/contract actually changes. Related
feedback inventory: `docs/work/audits/FEEDBACK-UNRESOLVED-UAT-20261020.md`.

## Category review follow-up — 2026-10-26

The owner asked to infer categories from transaction codes and notes, and to
review both the recently added entries and the held bank rows. Read-only
production reconciliation confirms all 87 external source candidates still
match app transaction IDs. Exactly 72 of these carry the import source marker
in their live note; the other 15 were pre-existing manual entries and are
outside this category edit. Of the 72 imported entries, 21 expenses and 14
incomes currently use an unclassified category.

Before any category mutation, save a private local snapshot of the 72 live
records and active category catalog. Suggest only existing active categories
with matching income/expense kind. An explicit merchant, service, commodity or
transaction purpose can support a high-confidence reassignment; an opaque
numeric/QR/reference code or person's name alone cannot. Preserve each
transaction's amount, type, date, wallet, note and report inclusion. Use the
canonical version-checked transaction PATCH, then re-fetch and compare those
unchanged fields. If live version or data drifted, skip and review manually.
Do not alter the 15 pre-existing manual entries. Do not import or auto-classify
the 31 held bank rows into the ledger: keep their suggested purposes in a
private review list for owner confirmation, consistent with the earlier hold
decision. The 10,000,000 VND savings transfer stays a paired transfer with
system categories and out of income/expense reports.

Verification plan: count changed versus skipped records; re-fetch all 87
candidate transactions and confirm source match; check all 31 held IDs remain
absent; dry-run a second pass for zero additional safe changes; revoke the
temporary owner key and verify it no longer authenticates. Raw statements,
notes and credentials stay outside git. This changes transaction metadata, not
API/schema/runtime behavior. Public and agent-readable product docs need no
change unless a new categorization workflow is shipped.

## Owner review result — 2026-10-26

The authenticated post-check found 100 app transactions, including all 87
external source candidates. Their amount, type, wallet and report flag still
match the canonical source plan. The 72 statement-imported notes match their
source rows. The other 15 candidate matches are older manual entries; 5 rely
on date/amount and a weak or generic note, so source coverage is not the same
as confirmed identity. The 31 held row markers remain absent.

Four previously unclassified expenses with explicit notes were assigned active
categories: one raincoat to `Quần áo`, one Salonpas to `Thuốc`, and two parking
fees to `Gửi xe`. Version-checked PATCH returned success for each. The
post-check found exactly four category changes, zero amount/type/date/wallet/
note/report-flag changes and zero changes to the 15 older manual matches.
There are still 31 unclassified imported entries (17 expenses, 14 incomes).
Thirteen of the latter say only that money was transferred in; their source
and economic purpose remain unproven, so they were not assigned a misleading
income category. A separate review flagged 23 imported entries for human
category confirmation, including 5 that already have a potentially wrong
specific category.

The reportable `Techcombank` wallet currently totals 24,205,499 VND income
and 30,337,815 VND expense across 96 entries. These are *app totals*, not a
confirmed statement total: 31 bank rows remain held (25,706,720 VND debit;
36,000,000 VND credit), 9 reportable app expenses totaling 3,397,000 VND do
not match any of the three supplied files, and the card file stops October 19.
Some app-only 30,000 VND entries may overlap `LocVang` but evidence is
insufficient. The 6,917,499 VND Shopee card refund is currently counted as
`Thu nhập khác`; the owner must decide whether to keep it as refund income or
offset the original expense. `Google One` 250,000 VND remains app-scheduled,
not bank-verified. Two separate wallet-transfer pairs remain excluded from
reports, including the 10,000,000 VND long-term savings transfer.

The detailed, private owner checklist is
`/tmp/mypocket-category-review-20261026/REVIEW.md` (mode 600); raw names and
financial notes are intentionally not committed. Five weak manual matches
and all 31 held rows are listed there by source ID, date and amount. The
temporary review API key was revoked by exact ID after a fresh live snapshot
comparison; its authentication then returned HTTP 401 and the local secret
file was removed. The pre-existing `feedback` key was untouched. A public
docs change is unnecessary because no product API, schema or user workflow
changed; this was an owner-authorized data review and four metadata edits.

## Finalized `LocVang` coffee decision — 2026-10-26

The owner clarified that every `LocVang` source transfer is coffee and
authorized syncing already-finalized rows. The reconciliation workbook maps
exactly seven distinct rows to this description: `OCT26-003`, `OCT26-009`,
`OCT26-012`, `OCT26-016`, `OCT26-023`, `OCT26-029` and `OCT26-063`. Six are
30,000 VND expenses and one is a 35,000 VND expense, totaling 215,000 VND.
Each row has a distinct statement-row reference and evidence fingerprint; no
other held row is covered by this decision.

The pre-write read-only production snapshot found one existing Techcombank
expense with the exact local date, amount and active `Cà phê` category for
both `OCT26-003` and `OCT26-012`. Each match is unique, reportable and has no
source note. Treat these two as already represented and do not duplicate or
edit them. No same-date/same-amount Techcombank expense exists for the other
five finalized rows.

Planned production action: insert only the five missing rows as reportable
Techcombank expenses in the active `Cà phê` category. Preserve each source
date and amount; use a privacy-safe note containing the stable source row code
and a deterministic `techcombank-october-2026-<source-code>` idempotency key.
Do not update the two existing manual matches and do not import any of the
other 24 held rows. Verification must re-fetch all seven mappings, confirm
five source markers plus the two exact existing matches, prove a second pass
proposes zero inserts, and compare October counts/totals before and after.

The same read-only snapshot found one active monthly `Google One` schedule at
250,000 VND, next due November 21 local time, plus one linked reportable
October 21 Techcombank expense for 250,000 VND. This verifies the recurring
app occurrence, not independent bank settlement. No public or agent-readable
product documentation changes are needed because this operation changes only
owner-authorized production data under existing API behavior.

## Finalized `LocVang` production result — 2026-10-26

The guarded canonical API pass inserted five expenses and updated none:
`OCT26-009`, `OCT26-016`, `OCT26-023`, `OCT26-029` and `OCT26-063`.
`OCT26-003` and `OCT26-012` were skipped as the two unique existing
same-date/same-amount `Cà phê` matches. No request replay occurred. A fresh
second pass skipped all seven represented rows and proposed zero inserts.

Before the pass, the Techcombank October ledger had 98 rows, 96 reportable
rows, 24,205,499 VND reportable income and 30,337,815 VND reportable expense.
Afterward it has 103 rows, 101 reportable rows, unchanged 24,205,499 VND income
and 30,492,815 VND expense. October `Cà phê` changed from 8 rows totaling
275,000 VND to 13 rows totaling 430,000 VND. Independent read-only SQL found
exactly one Techcombank coffee match for each of the seven source rows; the
five inserted rows each have exactly one stable source marker, while the two
pre-existing manual matches remain unedited and unmarked.

The 24 bank rows still held outside the ledger are `OCT26-004`, `OCT26-006`,
`OCT26-007`, `OCT26-008`, `OCT26-010`, `OCT26-011`, `OCT26-013`,
`OCT26-017`, `OCT26-021`, `OCT26-022`, `OCT26-024`, `OCT26-028`,
`OCT26-032`, `OCT26-033`, `OCT26-039`, `OCT26-040`, `OCT26-041`,
`OCT26-043`, `OCT26-044`, `OCT26-050`, `OCT26-053`, `OCT26-054`,
`OCT26-055` and `OCT26-058`. They comprise 23 debits totaling 25,491,720 VND
and one 36,000,000 VND credit. Production has zero source markers for these
rows. The workbook also retains 41 already-present app rows for owner review:
17 unclassified expenses, 14 receipts whose economic nature is not proven, 9
app-only rows without supplied-file evidence and 1 specifically categorized
expense needing confirmation.

`Google One` remains one active monthly 250,000 VND schedule, next due November
21 at 07:00 account-local time. Exactly one reportable 250,000 VND October 21
Techcombank expense remains present. This still proves app recurrence only,
not bank settlement beyond the supplied-file cutoff.

Because the saved password could not be exported and the iPhone mirror required
the owner's Mac unlock, the operator path created one two-hour credential using
the application's exact key format and stored-hash contract. The plaintext was
kept only in a mode-600 temporary file. The exact key was revoked through the
canonical API with HTTP 204; the same secret then returned HTTP 401, and a
read-only database check found one revoked record and zero active reconciliation
keys. The existing `feedback` key was untouched. No application source, schema,
deployment, commit, push or public/user-facing documentation changed.

## Fresh read-only visibility recheck — 2026-10-27

A new production snapshot was taken independently of the earlier result through
the documented production database path, inside explicit read-only
transactions. The source dates and amounts were re-derived from the two original
XLSX statements. Exact Techcombank-wallet, active `Cà phê`, expense and
account-local-date matching returned one reportable row for every finalized
source tuple:

| Source | Account-local date | Amount | Exact rows | Source marker |
| --- | --- | ---: | ---: | ---: |
| `OCT26-003` | 2026-10-01 | 30,000 VND | 1 | 0, retained manual match |
| `OCT26-009` | 2026-10-05 | 30,000 VND | 1 | 1 |
| `OCT26-012` | 2026-10-06 | 30,000 VND | 1 | 0, retained manual match |
| `OCT26-016` | 2026-10-08 | 30,000 VND | 1 | 1 |
| `OCT26-023` | 2026-10-12 | 35,000 VND | 1 | 1 |
| `OCT26-029` | 2026-10-14 | 30,000 VND | 1 | 1 |
| `OCT26-063` | 2026-10-26 | 30,000 VND | 1 | 1 |

Both deterministic reconciliation passes returned 7/7 represented, 215,000 VND
source total and zero planned inserts. The ledger snapshot was unchanged across
the two passes: 103 October Techcombank rows, 101 reportable rows, 24,205,499 VND
reportable income and 30,492,815 VND reportable expense. No API or database
mutation was attempted. No credential or local secret artifact was created.

The rows can appear missing because the native **Giao dịch** screen is weekly,
not monthly. On 2026-10-27 it opens the current account-local week, so only the
2026-10-26 row is initially visible. Select the `Techcombank` wallet (or
`Tổng cộng`) and use **Tùy chọn** with 2026-10-01 through 2026-10-26 to show all
seven. In **Theo tuần**, their exact locations are:

- 2026-09-28 through 2026-10-04: October 1;
- 2026-10-05 through 2026-10-11: October 5, 6 and 8;
- 2026-10-12 through 2026-10-18: October 12 and 14;
- 2026-10-26 through 2026-11-01: October 26.

The wallet selector is remembered per account, so another selected wallet hides
these rows. `Cà phê` is not a separate ledger exclusion filter: **Xem theo nhóm**
groups the current wallet/period by category, and search for `Cà phê` searches
only that already-filtered wallet/period. All seven have
`included_in_reports=true`; the report flag does not hide them from the ledger
and they contribute to October reports, while report screens still require the
October period and matching wallet/category scope.

## Finalized workbook decisions applied — 2026-10-27

The owner-annotated workbook contained 19 non-empty decisions. A fresh
read-only production snapshot was captured before the write window. Only the
unambiguous actions supported by those decisions and the owner's recorded
conversation confirmations were applied: eight ordinary transactions were
updated, one owner-confirmed refund row was deleted, and two confirmed missing
rows were created. Updates used the canonical API with `expected_version`; the
two creates used stable idempotency keys and disabled travel auto-linking.

Every update was re-fetched immediately. Its version advanced by one, its
approved category/note matched, and wallet, amount, type, occurrence time,
report inclusion, jar/travel association and transfer/credit state stayed
unchanged. The deletion target was revalidated immediately before deletion and
was absent afterward. Each created source appeared exactly once and was not an
idempotency replay.

The independent second dry pass proposed zero updates, zero deletes and zero
creates. It found no unexpected changed or added rows, no marker for any of the
21 held source rows, and exactly one representation for each of the seven
previously finalized coffee rows. The three ambiguous groups remain inside that
held boundary and were not modified.

The Techcombank October ledger changed from 103 to 104 rows and from 101 to 102
reportable rows. Reportable income changed from 24,205,499 VND to 53,288,000
VND; reportable expense changed from 30,492,815 VND to 30,560,815 VND. These
deltas are fully explained by the one deletion and two creates; category/note
updates did not change financial fields.

One two-hour operator key was created only for this write window. Revocation
returned HTTP 204, the same secret then returned HTTP 401, the database reported
zero active reconciliation keys, and all local plaintext credential artifacts
were deleted. No application source, schema, deployment, commit, push or public
documentation changed. Public and AI-agent documentation do not require an
update because this was an owner-data correction under already documented API
contracts.
