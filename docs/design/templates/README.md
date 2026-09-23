# Templates

Templates define page layout shells and responsive composition without owning product
data. They consume organisms and shared base contracts. Runtime implementations live in
`app/src/atomic/templates/`.

The current shell is documented by the page contracts and `screens/current-ui.md` (now
`pages/current-ui.md`); extract a dedicated template contract when a second page reuses
the shell with a different composition.

[Design gateway](../README.md) · [Organisms](../organisms/README.md) ·
[Page contracts](../pages/README.md)
