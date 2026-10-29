---
artifact_type: adr
id: ADR-003
status: accepted
owner: shared
---

# Reference fidelity is part of base-first implementation

Owner requested tighter rules and wallet reference implementation. [Trigger/design](../work/tickets/UI-WALLET-02-DETAIL_DESIGN.md), [page](../design/pages/wallets/README.md), [prior decision](ADR-002-design-contract-enforcement.md).

Using base controls is necessary but not sufficient: screen composition must map reference regions, actions and states to named bases. New shared behaviors are specified before consumers. Visual overrides on cards and clickable native wrappers are blocked by AST checks, with regression fixtures.

New owner-supplied reference bundles may preserve code.html/screen.png alongside a README explicitly marked artifact_source: visual_reference. They are evidence, never copied runtime code. Other non-spec files still fail. This supersedes the blanket asset prohibition for declared source evidence only; no production-source allowlist is introduced.

Rejected: deleting newly supplied evidence, copying export CSS, or relying solely on prose rules. Consequence: static tests catch ownership violations but browser/reference review remains required to assess layout fidelity. API/data contracts are unchanged.
