# Native bottom-navigation safe-area contract

## Decision

The persistent native shell bottom navigation needs a shared content inset so
the last card or form control remains scrollable and visible on compact phones.
`MPBottomNavigationInset` owns the reserved height (`2 * MPMetrics.hitTarget`)
and is applied with `safeAreaInset(edge: .bottom)` by scroll-based screens.
Screen files must not invent raw `Color.clear.frame(height:)` spacers for this
purpose.

## Scope

Use the base for Overview, More and secondary feature screens that render under
the persistent shell. Editor save bars with their own safe-area treatment keep
their existing editor-specific inset. No API, schema or navigation contract
changes.

## Verification

Verify the base-component guard, native unit/contract tests, compact simulator
UI route sweep and a fixture-backed Overview screenshot. The Overview UI proof
must scroll to `overview-section-recent-transactions` and confirm its frame is
above the shell navigation; this checks reachability rather than only checking
that the view exists in the accessibility tree. The bottom report card must
have clear space above the shell navigation.
