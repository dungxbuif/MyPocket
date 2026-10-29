# Native authentication

Entry: unauthenticated AppRoot; email/password submit creates normal session.
Returning enrolled owner may use Face ID even after logout via independent
device credential. Restored active session can enter biometric lock screen.

Bases: MPTextField/MPSecureField, MPButton, MPStatusMessage, shared typography.
Email/password manager/paste remain supported. Pending prevents duplicate
submit; failure retains email, shows non-enumerating error. No Google login UI.
Face ID cancellation/network error does not erase enrollment; invalid/expired
credential asks password sign-in. Enrollment/forget rules: [guide](../../../guides/face-id.md).
Current Face ID missing report requires physical-state reproduction, not
silent security relaxation. [Entries and proof](../../platforms/ios-screen-registry.md).
