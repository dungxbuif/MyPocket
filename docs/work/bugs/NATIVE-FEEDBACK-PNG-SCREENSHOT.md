---
artifact_type: bug
id: NATIVE-FEEDBACK-PNG-SCREENSHOT
status: regression_verified
owner: shared
---

# Native feedback screenshot MIME mismatch

## Reproduction

Open the floating feedback bubble on iOS, submit a feedback item and let the
screen capture be attached. The native client currently serializes the capture
as `screen.jpg` with `image/jpeg`.

## Impact

The feedback API deliberately accepts only a private PNG screenshot of at most
1 MiB. Text-only feedback succeeds, but any native feedback containing the
automatically captured screen is rejected as an invalid screenshot. This
breaks the required silent screen-context feedback flow.

## Intended behavior

- Capture the current native screen as PNG.
- Send `screen.png` with MIME `image/png` in the `screenshot` multipart part.
- Keep the existing text-only fallback when capture is unavailable.

## Verification result — 2026-10-06

The shared snapshot encoder now returns PNG data and the multipart client sends
`screen.png` with `image/png`. Regression tests assert both the PNG signature
and multipart metadata; full native/signed simulator and backend feedback
contract verification is recorded in the validation matrix.
