# Deterministic Financial Ingestion Design

Status: approved for implementation by owner conversation on 2026-10-26.

## Intent

Keep `DeepSeek-V4.1-Flash`, but stop asking it to be the source of truth for
financial fields that can be read deterministically. The three-file production
benchmark matched only 55/87 canonical external transactions and spent 106.326
of 110.97 seconds in one model response. The improved path must preserve source
rows, make unsafe fields unapprovable, and reduce the model's output burden.

## Goals

- Accept XLSX statements directly instead of converting them to PDF.
- Preserve date, amount, direction, currency, reference and source-row identity
  from structured statements without allowing the model to rewrite them.
- Use the model for semantics: concise note, category suggestion, counterpart,
  transfer/refund hints and review questions.
- Keep every source row traceable to file, sheet and row while avoiding source
  account numbers in operational logs.
- Bound model batches and retain request-level idempotency and review-only
  approval semantics.
- Re-run the supplied three-source evaluation before any release claim.

## Non-goals

- Automatic approval or direct ledger writes.
- A universal parser for every bank layout in the first slice.
- Currency conversion, balance reconciliation or silent internal-transfer
  deletion.
- Production deployment in the implementation task.

## Decision

Use a hybrid pipeline:

```text
XLSX bytes -> strict workbook parser -> canonical source rows ----+
PDF/image -> private storage -> text/OCR -> source excerpts -------+-> bounded LLM semantics
                                                               |
canonical row fields + model semantics -> deterministic merge -> review proposals
```

For structured rows, `occurred_at`, `amount`, transaction direction,
`source_currency` and source identity belong to the parser. The model response
must include the supplied source ID. Unknown or repeated source IDs, missing
rows and any model/parser disagreement fail closed. The model may suggest a
category and note, but it cannot change parser-owned fields.

The first parser targets statement layouts with recognizable date,
description, reference, debit, credit and balance columns, including the two
supplied Techcombank files. It ignores headers/footers and zero-value rows,
preserves separate identical rows, and rejects a row with both debit and credit
or neither. Unrecognized workbooks remain unsupported rather than falling back
to opaque OCR.

PDF/image behavior remains review-only. Existing fail-closed date/currency
work is preserved. A later slice may add PDF table extraction with the same
canonical-row interface; it is not required to ship XLSX safety.

## Interfaces

- `financialsource.ParseWorkbook(name string, data []byte, timezone string) (Document, error)`
- `Document.Rows []Row`, where `Row` contains opaque `SourceID`, source location,
  occurred time, type, VND amount, description and reference.
- `AIExtractInput.StructuredRows []AIStructuredRow` carries parser-owned rows to
  the provider adapter.
- `AIExtractDraft.SourceID` is required only for structured batches and is not
  persisted in the public ledger transaction.
- The use case deterministically overlays parser-owned fields after schema
  validation and adds a review question when category/note semantics remain
  uncertain.

## Security And Privacy

- Validate XLSX ZIP structure and MIME by bytes, not filename alone.
- Keep the existing per-file and request size limits.
- Do not log cell values, account numbers, descriptions or model prompts.
- Store originals under the existing private attachment contract.
- Formula cells are read as stored values only; never execute macros or external
  links. Macro-enabled workbooks are rejected.
- Parse the small OOXML subset with Go's `archive/zip` and `encoding/xml`, with
  hard limits on archive members, uncompressed bytes, worksheet rows and cells.
  Do not add Excelize in this slice: the latest released version on 2026-10-26
  still has malformed-workbook vulnerabilities whose fixes are documented for
  an unreleased follow-up. The parser needs only workbook relationships,
  shared strings and worksheet cell values, not spreadsheet mutation or formula
  calculation.

## Acceptance

1. Both supplied Techcombank XLSX files parse all transaction rows with exact
   date, debit/credit direction, integer VND amount and stable source identity.
2. Separate equal-looking rows remain separate.
3. A model response cannot change amount, date, direction or currency.
4. Missing, duplicate or foreign source IDs fail the structured extraction.
5. Existing JPEG/PNG/PDF behavior and fail-closed tests remain green.
6. Evaluation reports structured-field accuracy separately from semantic
   category quality and model latency.

## Alternatives

- Prompt-only tuning was rejected because the live run fabricated dates and
  spent 18,664 completion tokens.
- OCR concurrency alone was rejected because OCR used only 3.691 seconds.
- Replacing the model was rejected because the owner asked to improve the
  current model and deterministic parsing addresses the highest-risk fields.

## Documentation Impact

Update API/reference MIME limits, AI Entry user guide, architecture integration
flow, validation matrix, backlog, context and changelog. Public docs must state
that XLSX financial fields are parser-owned while all proposals still require
review. OpenAPI/MCP multipart catalogs must expose the XLSX MIME type when the
endpoint contract changes.
