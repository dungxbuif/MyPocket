# Native wallet editor Stitch parity implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Make wallet create/edit use the Stitch form hierarchy while retaining every existing wallet type and validation contract.

**Architecture:** The existing `WalletEditorView.save()` remains unchanged. Only the form composition moves to shared reference header/card/base fields, with a safe-area save action so keyboard and tab navigation cannot obscure it.

**Tech Stack:** SwiftUI, existing MyPocket iOS design system, XCTest/XCUITest.

**Spec:** Stitch state `image.png_4`, `financial_clarity_system/DESIGN.md`, and the wallet API contract.

## Global Constraints

- Preserve basic/goal/credit fields, validation, owner-scoped create/update payloads and edit immutability rules.
- Use `MPReferenceScreenHeader`, `MPSurfaceCard(.reference)`, `MPTextField`, `MPAmountField`, `MPWalletSelector`, `MPBooleanSelector`, `MPDateField`, and `MPButton`.
- Keep existing UI labels/routes discoverable; add stable editor IDs instead of replacing them.
- Do not add provider, persistence, or migration behavior.

## Review Focus

- Empty wallet names and invalid amounts still fail before a request.
- Goal target/date and credit statement fields remain conditional and validated.
- Edit mode must not allow changing the persisted wallet type/opening balance.
- The safe-area save action remains hittable with the keyboard open.
- Cancel must discard draft changes and return to the wallet list/detail route.

### Task 1: Contract

**Files:** `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

- [x] Add and run a failing wallet-editor section-order contract.

### Task 2: Form composition

**Files:** `mobile-ios/MyPocketIOS/Features/Wallets/WalletsView.swift`

- [x] Use the Stitch reference header/card hierarchy and stable field/save identifiers.
- [x] Move save into a safe-area inset; keep `save()` unchanged.

### Task 3: Verification/docs

**Files:** `docs/design/references/STITCH-MY-POCKET.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/work/PORT-100-AUDIT.md`, `docs/releases/CHANGELOG.md`

- [x] Run targeted/full native tests, build, guards and diff check; record actual counts and preserve external visual UAT gates.
