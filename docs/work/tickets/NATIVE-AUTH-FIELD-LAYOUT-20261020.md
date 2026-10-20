---
artifact_type: detail_design
id: NATIVE-AUTH-FIELD-LAYOUT-20261020
status: verified_on_physical_device
---

# Native auth field layout guard

## Problem

On the physical iPhone, entering a long UAT email caused the shared email
`MPTextField` UIKit wrapper to retain intrinsic width. Its clear button moved
outside the viewport and the password field disappeared from the accessibility
tree. This blocked password login and was reproduced through Mobile MCP.

## Decision

Constrain the editable field and its form row to the available width in the
shared base component. Text remains horizontally scrollable inside the native
control; the clear action keeps its hit target. `MPSecureField` receives the
same finite-width contract. No auth API, validation, or server behavior changes.

## Verification

Run the native design-system tests, then use Mobile MCP on the physical iPhone:
enter a long email, confirm the password field and login button remain visible
and accessible, enter the password, and verify the production UAT login.

## Implementation result

The shared field rows now constrain editable controls to the available width;
the 98-test `DesignSystemTests` target passed with no failures or runtime
warnings. Release `1.2.0` was installed on the paired physical iPhone after it
was unlocked. Mobile MCP entered the dedicated production UAT account, verified
the long-email/password controls, completed password login, and navigated
Overview, Transactions, Account and the floating feedback entry point.
