# Three-file financial AI Entry benchmark — 2026-10-26

Status: in_review. Owner requests an actual-model quality and latency check
using the two supplied October Techcombank XLSX statements and credit-card PDF.
This is a review-only benchmark; no AI proposal may be approved into the ledger.
The independent canonical reconciliation and hold/exclusion decisions are in
[the October ticket](TECHCOMBANK-OCTOBER-RECONCILIATION-20261026.md).

## Input and safety contract

`POST /api/v1/ai/entry/process` accepts up to 20 JPEG, PNG or PDF files, but
does not accept XLSX. Create private, temporary PDF renditions of the two XLSX
files without editing the originals. Verify the 118 bank source rows and 31
October card rows remain readable; if a rendition drops or reorders content,
do not use it to score model quality. Submit three PDFs with a unique request
ID through an owner-authenticated, review-only path. Record aggregate timing,
output counts and error classes, not raw account numbers, credentials or source
documents in git. Revoke any temporary owner key after the run. The request's
stored proposals may be rejected/cleaned after evidence capture; none approved.

## Measurement plan

1. Capture file sizes/pages and local PDF-text checks. Establish the independent
   source-row truth set: external candidates, internal movements, duplicate
   cross-account transfers, held rows and October-only card boundaries.
2. Measure client upload-to-response time with one unique request; never retry
   with a new ID after uncertain acceptance. Read the same request ID for
   recovery. Separate request parsing/storage, OCR per file, model, and result
   persistence using timestamped server logs or existing provider diagnostics;
   mark stages without telemetry as unknown rather than inferring timings.
3. Score extraction on row recall, amount/sign/date/currency, duplicate
   preservation, correct internal-transfer exclusion and review questions.
   Treat unverified rows as unsafe to approve even if a total looks plausible.
4. Reproduce any slow or incorrect stage before changing code. Write a bounded
   hypothesis and regression/performance test; rerun the same safe benchmark
   after a controlled local change. Production image remains unchanged unless
   a separate release is approved.

## Known baseline

A prior review-only run of the credit-card PDF alone returned 17 proposals in
about 14 seconds, with wrong dates, a foreign-currency amount treated as VND,
and identical separate rows collapsed. A local backend safety candidate now
keeps unknown dates/amounts unapprovable and preserves repeated rows, but it
has not been deployed. The three-file run will therefore characterize the
**current production** model/pipeline, not claim that local candidate is live.

No public API/schema change is planned for the benchmark. If XLSX import or
runtime performance changes are later approved, update the human AI Entry guide
and stable agent-readable API reference before release.

## Unchanged-production result

The owner-authenticated review-only run on 2026-10-26 used readable temporary
PDF renditions of the two XLSX files (4 and 3 pages) plus the original card PDF
(7 pages). The bank renditions each kept date, description, debit and credit in
the same visual row. This tests the current PDF ingestion path, **not** native
XLSX ingestion. Production API reported AI/OCR/file capabilities configured
and ran `DeepSeek-V4.1-Flash`. There was one submission and no retry under a
new request ID.

| Measure | Result |
| --- | ---: |
| End-to-end request | 110.97 s, HTTP 200 |
| OCR PDF 1 / 2 / card | 2.044 / 0.647 / 1.000 s (3.691 s serial) |
| Model request | 106.326 s |
| Model tokens | 12,858 prompt; 18,664 completion |
| Proposals | 158; 25 `unknown` type; 60 with questions |
| Exact October external match | 55/87 by local date, direction and VND amount (bank 50/68, card 5/19) |
| External misses | 32/87; 13 have the same amount/direction in a proposal on a different date |
| Held/excluded source candidates | 28 held and 20 excluded rows have a remaining proposal with matching date/direction/amount after exact external matches are consumed; source attribution is not certain from those fields alone |
| Date scope | 23 proposals dated October 26 versus only one source bank row that day; 9 September proposals came from September rows in the supplied card PDF and are outside this October import |

The model call dominates observed latency. Startup, storage and persistence
combined fit in roughly one second, but existing timestamps do not separate
those stages precisely. Serial OCR concurrency would save at most a few seconds
for this input and cannot explain the 106-second model generation. The response
size and 18,664 completion tokens are evidence for a large-output bottleneck;
provider-side queueing versus generation time was not independently measured.

The 55/87 exact-match score is a conservative financial correctness gate, not
a semantic-match recall estimate. It prevents treating similar notes or totals
as proof of correct dates and amounts. Matching a held/excluded row by three
fields is a **potential** unsafe suggestion, not a proven model attribution;
two held keys and one excluded key collide with external keys. Regardless, the
batch is unsafe to approve without per-row review. All 158 benchmark proposals
were rejected, zero were approved, and the temporary owner key was revoked;
the post-revocation API call returned HTTP 401. No ledger mutation came from
the benchmark. Raw files, response and key were never added to git.

## Improvement decision

1. Prioritize correctness: source-row provenance, no fabricated date/amount,
   currency checks, October scope and explicit internal-transfer/held review.
   The local PDF fail-closed candidate is not deployed and must be retested.
2. Detail-design bounded model batches (per document or page) with progress,
   idempotent recovery and cross-file reconciliation. Compare quality and
   wall-time against this baseline before release. Parallel model calls need
   provider-rate/cost limits; simply parallelizing OCR is not the fix.
3. Consider native XLSX ingestion for structured bank statements so exact
   cells survive without PDF conversion. This is a new input contract and
   requires separate API/security design and owner approval.

The deployed attachment-cleanup worker currently lacks private-S3 credential
environment names, so automatic deletion of unlinked originals after the
24-hour window is not verified. All proposals were rejected, but source-object
retention remains an operational release gate in
[attachment feedback](../bugs/FEEDBACK-baca2415-ai-entry-attachments.md).
No runtime configuration or production image was changed during this test.
Human-facing and AI-agent API docs do not change yet because this benchmark
ships no new behavior.

## Local deterministic candidate — 2026-10-27

The local candidate now accepts passive XLSX directly and keeps structured
financial fields outside model control. An opt-in private harness parsed both
supplied workbooks as **118/118 rows** (66 + 52), matched debit/credit counts
and totals, and matched a privacy-safe canonical fingerprint over every row's
date, direction, amount, currency, reference and source ID. No proposal was
approved and the harness does not persist data. Full backend verification is
729 tests; focused race verification is 636 tests across parser, AI, use case
and HTTP packages.

This closes the structured-field regression, not semantic/category quality or
live-provider latency. The next live evaluation must run XLSX semantics and the
card PDF as separate requests because mixed XLSX+OCR input intentionally fails
closed. It must record provider batches/latency and semantic review aggregates
without approving proposals or committing raw responses.

## Configured-provider evaluation contract — 2026-10-27

The remaining Task 4 evaluation is a local, opt-in Go test. It reads the three
private source paths only from explicit environment variables and loads the
existing local AI/OCR provider configuration. It does not open a repository,
database, HTTP API, proposal store or approval path. Provider credentials,
source text, account identifiers and model responses must not be logged.

The two XLSX files run together through deterministic parsing and structured
semantic batches. The PDF runs separately through the existing OCR path. A
mixed structured-plus-OCR request is outside the harness and must continue to
fail closed. The harness enables aggregate provider usage so it can report the
configured model name, source and proposal counts, batch count, prompt and
completion token counts, summed provider latency and end-to-end wall time.

The automated semantic score is intentionally bounded to review safety. A
proposal passes when it has a non-empty bounded note, uses only a supplied
evaluation category whose kind matches the parser-owned transaction type, and
does not contain a blank review question. This score measures whether the
provider returned internally valid, reviewable semantics; it does not claim
that every merchant/category choice matches a human-labelled truth set. XLSX
financial-field accuracy and exact source-ID coverage remain separate, strict
118/118 gates. For the PDF, the source control is 31 October rows and the
proposal count is reported without treating count equality as financial-field
accuracy.

Verification plan:

1. Run the opt-in harness with the two supplied XLSX paths and the supplied PDF
   path, using existing local provider credentials without printing them.
2. Confirm two distinct evaluation summaries, zero persistence/approval code,
   strict XLSX source coverage and privacy-safe aggregate output only.
3. Compare structured semantic and PDF OCR wall/provider latency with the
   historical 55/87 exact-match and 110.97-second three-PDF baseline.
4. Run the focused non-live AI tests and the opt-in private workbook truth-set
   test after the live run. Record exact commands and results in this ticket and
   `docs/work/VALIDATION_MATRIX.md`; public and AI-agent API docs remain
   unchanged because this evaluation ships no contract or runtime behavior.

## Configured-provider result — 2026-10-27

The existing local AI and OCR credentials were present and the opt-in harness
completed against `DeepSeek-V4.1-Flash`. The measured run used separate XLSX
and PDF paths and finished in 60.16 seconds. It opened no database, repository,
HTTP API, proposal store or approval path. No proposal was approved, no ledger
or production state was read or written, and no raw source row, account number,
provider response or credential was written to git.

| Path | Sources | Proposals | Automated semantic review | Provider stages | Tokens |
| --- | ---: | ---: | ---: | ---: | ---: |
| Two XLSX files | 118 | 118 | 118/118 | 5 model batches, 48.304 s provider / 48.313 s wall | 20,051 prompt / 14,130 completion |
| Card PDF, October scope | 31 | 33 | 33/33 | 1.006 s OCR + 10.835 s model, 11.842 s end to end | 3,921 prompt / 3,319 completion |
| Combined separate paths | 149 | 151 | 151/151 | 59.139 s model/provider, 60.155 s measured stage wall | 23,972 prompt / 17,449 completion |

The XLSX path preserved exact source coverage and parser-owned financial fields
at **118/118**. The PDF path over-produced two proposals relative to the 31-row
October source control, so it does not pass source-count parity even though all
33 outputs passed the bounded semantic review-safety rubric. The harness does
not contain a private human-labelled merchant/category truth set; therefore the
151/151 score must not be described as human category accuracy. PDF exact
date/type/amount accuracy also remains unscored in this run.

Compared with the historical three-PDF baseline, deterministic XLSX financial
coverage rises from 55/87 exact external matches to 118/118 structured rows,
but the populations differ and this is not a like-for-like recall percentage.
The separate candidate paths used 60.155 seconds versus 110.97 seconds
(50.815 seconds, 45.8% lower), and model/provider time used 59.139 seconds
versus 106.326 seconds (47.187 seconds, 44.4% lower). The baseline made one
large mixed three-PDF model request; the candidate used five bounded XLSX
batches plus a separate PDF request, so these timings characterize this exact
configured-provider run rather than a general provider SLA.

The first live attempt failed locally before network work because the harness
gave both workbooks the same privacy-safe display name, which correctly
produced duplicate opaque source IDs. Distinct generic document names fixed the
harness without exposing real filenames. A subsequent wrapped pass completed,
and the final unfiltered aggregate run above supplied the recorded evidence.

Verification commands:

- `rtk go test ./internal/infrastructure/ai -run 'TestSemanticReviewScore|TestPrivateTechcombankWorkbookTruthSet|TestPrivateFinancialIngestionConfiguredProviderEval' -count=1` — scorer passes; private tests skip without opt-in.
- `MYPOCKET_EVAL_PROVIDER_LIVE=1 ... rtk proxy go test ./internal/infrastructure/ai -run '^TestPrivateFinancialIngestionConfiguredProviderEval$' -count=1 -v` — PASS, 1/1 in 60.16 s; paths and credentials supplied only through the process environment.
- `MYPOCKET_EVAL_TECHCOMBANK_XLSX_A=... MYPOCKET_EVAL_TECHCOMBANK_XLSX_B=... rtk proxy go test ./internal/infrastructure/ai -run '^TestPrivateTechcombankWorkbookTruthSet$' -count=1 -v` — PASS, 1/1 in 0.01 s with the private 118-row fingerprints.
- `rtk go test ./internal/infrastructure/ai -count=1` — PASS, 117 tests.
- `rtk go test -race ./internal/infrastructure/financialsource ./internal/infrastructure/ai -count=1` — PASS, 160 tests across two packages.
- `rtk git diff --check -- <four owned paths>` — PASS for tracked changes; final source inspection found no credential, raw source-row, account-number or model-response output in the harness or evidence.

Public human and AI-agent API docs do not change: this work adds a private,
opt-in test harness and benchmark evidence only. It does not alter the shipped
API, schema, configuration contract, ingestion behavior or deployment state.

## PDF overrun diagnostic plan — 2026-10-27

Before release, extend the opt-in harness with a privacy-safe PDF proposal
audit. The audit must retain proposal values only in memory and emit aggregate
counts. It will classify missing or invalid dates, dates outside October 2026,
exact repeated financial fingerprints, and unknown or summary-like rows. The
duplicate fingerprint is date + type + amount + an in-memory SHA-256 note
digest; neither notes nor digests may be logged or committed.

Because categories can overlap, the audit will also assign one primary class
per proposal in this order: invalid date, outside October, duplicate,
unknown/summary-like, otherwise unexplained. Unit fixtures must prove each
class and the primary-count partition before another private provider call.
Then run only the PDF OCR/model path with existing local credentials and report
counts, latency and tokens. If the two-proposal delta belongs to a class the
prompt can safely exclude, first add a failing live eval assertion for that
class, then make the smallest prompt-only correction and rerun. Do not change
production extraction code, persistence, approval, deployment or fail-closed
validation.

## PDF overrun diagnostic result — 2026-10-27

The privacy-safe audit reproduced 33 proposals for the 31-row October control.
The initial classified run found no outside-October or exact duplicate
financial fingerprints; the two excess proposals were summary-like rows. A
strict live source-parity assertion was added and failed as intended.

A prompt-only evaluation correction was then tested against the same provider.
It still returned 33 proposals and failed the 31-row parity gate. In that run,
24 proposals lacked a usable date, 8 were summary-like in total and 1 summary
row was the primary classification. This confirms prompt wording alone is not
a reliable PDF row boundary or date-recovery mechanism for this statement.

The PDF path therefore remains review-only and is not a release-quality exact
financial import. Missing dates remain unapprovable, summary/count drift is
visible through the opt-in gate, and no PDF proposal was persisted or approved.
The deterministic XLSX path remains the supported exact route at 118/118. A
future PDF exactness phase needs deterministic row segmentation/source IDs or a
human-labelled OCR truth set; repeated prompt tuning is not accepted as the
fix. Public API behavior did not change.
