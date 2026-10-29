# Page contracts

Native iOS screens and opening controls are mapped in the
[screen/presentation registry](../platforms/ios-screen-registry.md), including
auth, More, recurring, credit, notifications and feedback native contracts.
The web locations below are historical, not an instruction to maintain web.

Page contracts are route-level specifications formerly stored under `screens/`.
Each page documents route, composition, states, event → effect, validation, permissions,
copy, API dependencies, proof, and known gaps. Runtime page entry points live in
`app/src/atomic/pages/` and route composition may use templates and organisms.

Use the [page contract template](README.md) when adding a new route. Shared visual
evidence is stored separately under [references](../references/README.md).

Implemented route contracts include [Travel Mode](travel/README.md), which is an
event dimension on ordinary transactions rather than a wallet type.

[Design gateway](../README.md) · [Templates](../templates/README.md) ·
[System contracts](../system/DESIGN.md)
