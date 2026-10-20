# Deterministic AI/OCR fixtures

`fixtures.json` is a synthetic, provider-free contract for the AI-entry and
Finance Assistant safety boundaries. It is loaded by the Go test
`TestAIEvalFixtures` in `backend/internal/usecase`.

The fixture clock is fixed. The cases cover overlapping OCR duplicate rows,
missing-date review disclosure, owner-wallet inference, strict owner-safe
advisor arguments, and proposal responses that remain pending until explicit
owner confirmation. No provider URL, API key, bearer token, real account ID,
OCR text, or user financial row belongs in this directory.

Run the deterministic check with:

```sh
cd backend
go test ./internal/usecase -run TestAIEvalFixtures -count=1
```

Live OCR/provider evaluation remains a separate, explicit opt-in path through
`AI_EVAL_LIVE=1`; it must use synthetic input and must never be required for
this test to pass. The live harness also needs deployment-managed AI/OCR/S3
credentials and is not a release prerequisite for local deterministic tests.
Use the [provider UAT checklist](../../docs/operations/PROVIDER_UAT.md) before
running it; never point it at a user's receipt or production ledger.
