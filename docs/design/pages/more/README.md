# More / Account landing — iOS

Entry: Account tab. MPProfileCard shows current owner name/email and opens
AccountView. MPMenuRow groups wallet, planning, transaction and helper routes;
NavigationStack path owns back action and resets route scroll identity.
MPBottomNavigation remains shell-owned; disabled Account center add is hidden.

Profile is now the sole account entry and Overview bell owns notifications;
duplicate More rows were removed with a failing/passing UI regression. Advisor
is now opened primarily from its shell tab, without a duplicate More row.
The separate AI transaction-entry item remains. Typed advisor destination remains
available for programmatic routing; removal affects the visible menu only.
Navigation/draft gate196/196 passes locally; do not infer release acceptance.
Use one flat MPSurfaceCard per group, shared heading/row inset, no per-row cards.
Menu target44 minimum, dynamic height; scroll final action above navbar.
Logout is async server-confirmed; errors remain visible, no fake logout success.
[Registry](../../platforms/ios-screen-registry.md), [guideline draft](../../platforms/ios-base-guideline.md).
