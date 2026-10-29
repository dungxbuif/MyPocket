# Organisms

Organisms are stateful, reusable product sections composed from atoms and molecules.
They own fetching, local state, navigation, and composition contracts. Runtime
implementations live in `app/src/atomic/organisms/`; route-specific behavior belongs in
the page contract under `../pages/`.

No standalone organism specification is currently split out here. When an organism has
reusable behavior beyond one route, add its README here before introducing local UI
variants.

[Design gateway](../README.md) · [Base contracts](../system/BASE_COMPONENTS.md) ·
[Page contracts](../pages/README.md)
