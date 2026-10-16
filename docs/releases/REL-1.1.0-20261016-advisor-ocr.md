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
- No migration ran; financial data and Redis data were not reset.
- `POST /api/v1/ai/advisor/messages/async` is available behind normal auth.
- AI-entry OCR preserves submitted image order and sends explicit document
  boundaries to the text-only model.

## Verification

- `go test ./...` passed.
- `go test -race ./...` passed.
- `go vet ./...` passed.
- Production `GET /api/v1/health` returned `200`.
- Production API and worker converged at `1/1`.
- Production Swagger returned `78` paths and includes the async route and
  authenticated sync capabilities route.

## Remaining acceptance gates

- Owner-authenticated provider/OCR run and AI-entry proposal review.
- Authenticated `/sync/capabilities` response and native outbox enablement.
- Physical iPhone Face ID, visual and release acceptance.
- Web remains intentionally `0/0` for the native-only rollout.
