---
artifact_type: release_note
id: REL-1.1.0-20261016-ADVISOR-OCR
status: verified
release: 1.1.0 patch
date: 2026-10-16
---

# 1.1.0 patch — async advisor and AI-entry OCR context

## Shipped

- API and worker image:
  `registry.dungxbuif.com/mypocket-api:release-20261016-7846945`
- Image digest:
  `sha256:71fdd4b4155972f3e6380ae02de846264e503347ab63b9693f65d2f0c713db43`
- After the initial image rollout, migrations 033 and 034 were applied from
  the same image after a verified `pg_dump -Fc` backup. Migration version is
  `34` and clean; no financial data or Redis data was reset or deleted.
- `POST /api/v1/ai/advisor/messages/async` is available behind normal auth.
- AI-entry OCR preserves submitted image order and sends explicit document
  boundaries to the text-only model.

## Verification

- `go test ./...` passed.
- `go test -race ./...` passed.
- `go vet ./...` passed.
- Deterministic local AI/OCR fixture contract covers overlapping duplicate
  rows, fixed-clock missing-date review, owner-wallet inference, owner-safe
  advisor arguments and pending proposal/no-auto-approval behavior. The fixture
  test uses synthetic data only and does not exercise a live provider.
- Production `GET /api/v1/health` returned `200`.
- Production API and worker converged at `1/1`.
- Production Swagger returned `78` paths and includes the async route and
  authenticated sync capabilities route.
- The production worker no longer reports the missing
  `last_user_message_at` column after the migration catch-up.
- Local gap-closure verification: backend tests, race tests and vet pass;
  signed iOS Simulator XCTest passes 307/307. Reports owner scoping,
  offline transfer receipt replay assertions and deterministic AI/OCR fixtures
  are local follow-up commits and were not deployed to production.

## Remaining acceptance gates

- Owner-authenticated provider/OCR run and AI-entry proposal review.
- Authenticated `/sync/capabilities` response and native outbox enablement.
- Physical iPhone Face ID, visual and release acceptance.
- Web remains intentionally `0/0` for the native-only rollout.
