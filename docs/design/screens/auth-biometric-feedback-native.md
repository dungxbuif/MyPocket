# Auth, biometric unlock and feedback changelog surfaces

Status: reference-density wave (2026-10-16).

Biometric unlock and the feedback changelog now use the grouped reference
surface family. Sign-in, Face ID enrollment, logout, feedback resolution and
changelog API behavior remain unchanged; only spacing/radius composition is
adjusted.

## Verification

Existing `DeviceSignInTests` and `MoreFeedbackReferenceTests` remain the
behavioral coverage. Base-component/project guards and the Release simulator
build are required before this wave is reported complete.
