# UI component inventory

Date: 2026-09-11

MyPocket uses a black/white-first product theme. Red is reserved for destructive or error states. Mounted product screens must use the shared controls below instead of raw HTML interactive elements.

| Interaction | Shared component | Notes |
| --- | --- | --- |
| Primary/secondary/destructive button | `ActionButton` / `Button` | One semantic button primitive with variants and focus treatment |
| Text, number, date input | `InputControl` / `Input` | Labels remain visible through their owning form row |
| Select | `Select` | Shared height, focus ring, disabled state |
| Multiline text | `TextAreaControl` | Shared form primitive; provider output is never rendered as HTML |
| File or camera receipt input | `FilePickerInput` | Accessible file input used by receipt and agent image workflows |
| Search | `SearchInputField` | Search-specific label and clear behavior |
| Card/group | `Card` / `GroupedCard` | Neutral borders, no screen-local chromatic surfaces |
| Sheet/dialog | `Sheet` / `SheetFrame` | Focus, safe-area and responsive containment |
| List row | `WalletRow`, `TransactionRow`, `CategoryTreeRow`, form row components | Domain-focused rows built on shared surface tokens |
| Status | `Badge`, `AgentRunStatus` | Neutral by default; state color only where meaning requires it |
| Error/notice | `OperationError`, `NoticeBanner` | Error text is persistent and actionable |
| Empty/unavailable | `EmptyStateView`, `UnavailableAction` | Consistent explanation and next action |
| Install prompt | `PWAInstallPrompt` | User-clicked native prompt or platform-specific fallback instructions |
| Charts/progress | chart components in `components/charts` | Monochrome series and accessible surrounding labels |

## Enforced contract

`frontend/src/test/uiContract.test.ts` scans the mounted application surfaces and fails if raw `button`, `input`, `select`, or `textarea` controls return, or if deprecated chromatic product accents are reintroduced. `frontend/src/test/baseComponents.test.tsx` verifies the behavior of the shared primitives.

The responsive Playwright contract verifies at least 44px primary touch targets, viewport-safe navigation, and a visible transaction-save action on the mobile sheet.

## Removed legacy duplicates

The following unmounted duplicate implementations were proven to have no imports and removed:

- `components/inputs/PhotoAttachmentPicker.tsx`
- `components/navigation/BottomTabBar.tsx`
- `screens/sheets/AddTransactionSheet.tsx`
- `screens/subpages/AccountSecurityScreen.tsx`
- `screens/subpages/CategoryManagementScreen.tsx`
- `screens/subpages/RecurringTransactionsScreen.tsx`
- `screens/subpages/SettingsScreen.tsx`

The mounted transaction sheet remains the implementation in `app/App.tsx`. Deleted files remain recoverable through Git history.
