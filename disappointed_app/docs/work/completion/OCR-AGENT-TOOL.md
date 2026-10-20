# Third-Party OCR Agent Tool — Completion Evidence

Date: 2026-09-11

## Boundary

OCR is an external provider used only as an image tool by the MyPocket agent. MyPocket does not expose a generic OCR proxy and does not treat OCR completion as an accounting command. An owned receipt may be attached to an agent run; the OCR result is untrusted context for either analysis or a pending transaction draft.

## Safety properties

- Receipt ownership is checked in the same database transaction that creates the queued tool run.
- Image bytes are fetched only from the private object key recorded for that owned receipt.
- Reads are capped at 15 MiB and the stored SHA-256 checksum must match before provider submission.
- Provider credentials, response bodies, and private image bytes are not included in errors.
- Provider calls occur outside database transactions; jobs use bounded leases and polling.
- `404`, `410`, `429`, timeout, failed, cancelled, expired, oversized, and capability-drift paths are fail-closed or retry-bounded.
- Completed OCR is scoped to its exact user and agent run. Completion alone creates no transaction and changes no wallet balance.

## Verification

```text
env MYPOCKET_TEST_DATABASE_URL=... go test -p 1 ./internal/agent ./internal/worker ./internal/platform/ocr ./internal/platform/objectstore ./internal/platform/httpapi ./internal/platform/config -count=1
PASS

env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL=... MYPOCKET_TEST_REDIS_URL=... MYPOCKET_TEST_S3_*=... go test -race -p 1 ./... -count=1
PASS (all backend packages, PostgreSQL/Redis/S3 integration enabled)

go vet ./...
PASS
```
