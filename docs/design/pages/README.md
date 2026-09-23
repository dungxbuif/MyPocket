# Page contracts

Page contracts are route-level specifications formerly stored under `screens/`.
Each page documents route, composition, states, event → effect, validation, permissions,
copy, API dependencies, proof, and known gaps. Runtime page entry points live in
`app/src/atomic/pages/` and route composition may use templates and organisms.

Use the [page contract template](README.md) when adding a new route. Shared visual
evidence is stored separately under [references](../references/README.md).

[Design gateway](../README.md) · [Templates](../templates/README.md) ·
[System contracts](../system/DESIGN.md)
