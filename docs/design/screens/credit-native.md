# Native credit wallet screen

Status: in_progress. The credit route preserves the existing credit-wallet
contract (`statement`, credit entry and payment actions) while matching the
reference populated-state anatomy:

1. back header;
2. credit-wallet selector;
3. one compact summary surface with debt, available credit, limit and amount
   due;
4. compact entry/payment actions;
5. one transaction surface with a count and shared `MPTransactionRow` rows.

The summary uses `MPMetricGrid`; rows use the shared transaction glyph,
typography, amount masking and accessibility behavior. The reference surface
token is used only for the grouped credit surfaces. No API, statement/payment
calculation, owner scope or mutation behavior changes.

When statement payment data is not configured, the screen keeps the summary
values and shows the existing configuration guidance. Transaction order and
server-provided amounts remain unchanged.

Verification: `CreditNotificationsReferenceTests` covers the four-item summary,
masking, singular/plural count labels and the populated notification grouping
contract. Parent integration must run signed simulator tests and a populated
credit screenshot before release.
