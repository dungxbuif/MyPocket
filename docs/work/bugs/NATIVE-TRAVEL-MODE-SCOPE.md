# Native Travel Mode scope parity

## Status

`verified_fixed` — 2026-10-06.

## Contract

The native Travel Mode screen must match the web behavior for the existing API:

- explain that Travel Mode is context, not a wallet, and show the active event status;
- show date scope and context for each event;
- require confirmation before deleting an event because deletion only clears links and must not surprise the user;
- show the editor warning that only newly created ordinary income/expense transactions auto-link to the active event (recurring, transfer, and adjustment transactions do not);
- keep required-name and date-range validation, and use shared native base components.

## Implementation plan

Add the missing summary/status card, delete confirmation, and editor warning to `TravelView`. Do not change API semantics or transaction linking rules.

Implemented in `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`:

- the screen now explains Travel Mode and reports the active event;
- event cards show explicit active state and bounded/unbounded date wording;
- deletion uses the shared confirmation dialog and explains that links are cleared while transactions remain;
- the editor repeats the ordinary-income/expense-only auto-link rule.

## Verification plan

- native unit tests remain green;
- Release build with warnings-as-errors;
- base-component and project checks;
- simulator authenticated More-route smoke test reaches Travel Mode and can open the editor.

Verification: `MyPocketIOSTests` 54/54 passed; the Release warnings-as-errors build, base-component guard, project verification, and `git diff --check` passed; and `MyPocketIOSUITests/LaunchTests/testAuthenticatedMoreMenuOpensEveryDestination` passed 1/1 on the iPhone 16 simulator in 122.031s. No API semantics were changed.
