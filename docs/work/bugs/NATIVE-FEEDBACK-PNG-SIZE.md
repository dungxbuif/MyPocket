---
artifact_type: bug
id: NATIVE-FEEDBACK-PNG-SIZE
status: regression_verified
owner: shared
---

# Native feedback screenshot size bound

## Reproduction

The feedback API accepts PNG screenshots only up to 1 MiB. A full-resolution
iPhone window can produce a PNG larger than that limit, even after correcting
the MIME type, causing otherwise valid feedback with an automatic screenshot to
be rejected.

## Impact

The global feedback bubble must be reliable on complex screens. A size
overflow silently turns the required screenshot-assisted feedback flow into a
failed submission.

## Intended behavior

Encode a PNG that is never larger than the API's 1 MiB limit. Downscale the
captured native window as needed; if a platform renderer still cannot produce a
bounded image, return `nil` so the existing text-only fallback can submit.

## Verification result — 2026-10-06

The shared snapshot base now progressively downsizes complex captures and
returns `nil` when a bounded PNG cannot be produced. A noisy 960×1600 image
regression is green under the 1 MiB limit. The final simulator gate is green:
65/65 native unit tests, 16/16 signed `LOCAL_API_UI` tests, Release with
warnings-as-errors, base/project guards, no invalid-frame log fault, and zero
leftover `Fixture*` wallets after the local API probe.
