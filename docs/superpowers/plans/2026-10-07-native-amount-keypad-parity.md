# Native amount keypad parity

Date: 2026-10-07  
Scope: iOS SwiftUI amount entry used by transaction, wallet, budget, recurring,
credit, transfer, adjustment and portfolio forms.

## Intended behavior

The Stitch reference and the web `AmountField` expose a shared calculator
entry point instead of relying only on the platform number keyboard. The native
`MPAmountField` will expose the same base affordance and a reusable keypad with
digits, `000`, operators, clear, backspace, evaluate and done states. The
calculator returns only a non-negative integer VND amount, matching the
backend transaction contract; division by zero, fractional results, overflow
and malformed expressions remain validation errors.

## Technical decision

- Keep the existing UIKit number-pad text field for direct input and
  accessibility.
- Add `MPAmountCalculator` as a pure, non-`eval` parser/evaluator in the
  DesignSystem layer.
- Add `MPAmountKeypad` as a DesignSystem base. Feature screens only bind the
  resulting string and do not own keypad colors, grid geometry or arithmetic.
- Reuse `MPButton`, `MPSurfaceCard`, `MPReferenceScreenHeader` and existing
  tokens; no screen-local styling or new API/schema is required.
- Expose an accessible calculator button and stable keypad identifiers so UI
  tests can verify the same flow without production data.

## Verification plan

1. Add pure tests for valid arithmetic, operator precedence, invalid/fractional
   results, divide-by-zero and overflow.
2. Add a native UI contract for opening the keypad, entering `1000000+500000`,
   evaluating and committing the result back to the amount field.
3. Run the full native unit target, Release warnings-as-errors build,
   project/base-component/diff guards and web design/build checks.
4. Reconcile this plan, the base-component inventory, validation matrix and
   changelog with the actual result.

## Scope boundary

The keypad does not invent quick-amount suggestions because the current API
does not provide a history-backed suggestion contract. Decimal input remains
visible as a reference key but cannot be committed when the result is not an
integer VND amount.

## Result

- `MPAmountCalculator` and `MPAmountKeypad` are implemented in the shared
  DesignSystem and mounted by `MPAmountField`.
- Native unit/contract target: 120/120 at
  `/tmp/mypocket-native-keypad-final-unit.xcresult`.
- Release warnings-as-errors build: pass at
  `/tmp/mypocket-native-keypad-final-release`.
- Transaction-editor keypad flow: 1/1 at `/tmp/mypocket-keypad-ui6.xcresult`.
- No API, schema or migration change.
