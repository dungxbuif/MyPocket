# More / Account landing — iOS

Entry: Account tab. MPProfileCard shows current owner name/email and opens
AccountView. MPMenuRow groups wallet, planning, transaction and helper routes;
NavigationStack path owns back action and resets route scroll identity.
MPBottomNavigation remains shell-owned. Account root keeps center Add; it opens
the ordinary transaction editor and returns to Account when saved or canceled.
Account child destinations and Assistant hide global Add. The narrow
selected-tab hit blocker remains intact.

Secondary destinations restore native leading-edge interactive Back through
`MPInteractiveBack`, scoped to the active route. Its original UIKit gesture
delegate is restored on disappearance. Root, transitions and active modals cannot
pop. Shared editors, category/wallet editors, dirty month notes and unsaved
Account fields block the gesture; explicit Cancel/Back controls remain available.

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
