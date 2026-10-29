# Assistant wallet confirmation summary

Status: local implementation complete; device/provider acceptance remains open.
The proposal card previously showed only name, type and opening balance, so a
goal or credit wallet could be confirmed without seeing all of the fields that
would be written.

## Contract

Before confirmation, the card now shows the shared base text rows for:

- opening balance, total-asset inclusion and description when present;
- goal target and target date;
- credit limit, statement balance, statement day and payment due day.

Money uses the existing masking formatter. The summary is presentation-only;
the proposal decoder and existing explicit confirmation boundary remain the
source of truth. No API, schema, permission or backend behavior changed.

## Verification

`/tmp/mypocket-wallet-summary-green-20261015.xcresult` passes the goal/credit
summary contract. Native warnings-as-errors build and base-component/project
guards are rerun after the implementation. Proposal `action_id` replay is now
covered by `/tmp/mypocket-wallet-action-id-green-20261015.xcresult`; real-device/
provider acceptance and the broader prepared-action dispatcher remain separate
gates.
