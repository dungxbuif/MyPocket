# Native category form parity and UI-test stability

## Intent

Keep category management behavior unchanged while making the native screen usable
with a long category tree and preserving the shared base-component contract.

## Decision

- Render the create/edit category form directly below the screen header while it
  is active. The form remains owner-scoped and uses the existing category API,
  validation, wallet selector, parent selector, and shared `MP*` controls.
- Keep the category tree below the form so a user never has to scroll through
  the full system category list to edit a draft. Flatten the tree into stable
  depth-aware rows inside `LazyVStack` so XCTest and VoiceOver do not build the
  complete long tree on every accessibility snapshot.
- Make the XCTest `reveal` helper scroll in either direction and use rendered
  bounds as the stable visibility signal for SwiftUI controls inside a long
  `ScrollView`.

## Verification

- Run native unit/contract tests and the Release warnings-as-errors build.
- Run project/base-component guards and `git diff --check`.
- Run the disposable local API category mutation UI test separately when the
  Xcode runner is responsive. The isolated run must verify create, update and
  delete through the real local API and clean its fixture rows.

## Result

The isolated local API category mutation UI test passed 1/1 on the iPhone 16
Simulator after the form placement, bidirectional reveal helper and lazy tree
changes. Result bundle: `/tmp/mypocket-category-lazy-ui/Logs/Test/Test-MyPocketIOS-2026.10.07_09-42-19-+0700.xcresult`.
