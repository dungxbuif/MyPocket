# Physical device UAT

Date: 2026-09-11

## Automated mobile-browser evidence

- Build under test: Git `815faaa` plus the uncommitted Task 9 UI consolidation.
- Chromium mobile emulation: install prompt, dismissal persistence, five-tab navigation, 44px touch targets, floating transaction action, safe viewport, transaction sheet action, PWA shell and receipt controls covered by Playwright.
- WebKit mobile emulation: the same install, navigation, touch-target and sheet contracts pass.
- The expanded install-help screenshot was inspected after the floating action overlap fix. The prompt is readable, blurred/dark, the account content remains reachable, the five-tab bar remains visible, and the hidden floating action no longer overlaps the prompt.
- Unit contract: all mounted core screens use the shared interactive primitives and the monochrome token contract.

These checks are browser emulation, not a substitute for physical-device evidence.

The remaining `legacy` source matches belong only to the tested IndexedDB migration that imports and quarantines the previous localStorage outbox format (`offline/db.ts`, `offline/migrations.ts`, and their tests). They are a data-compatibility path, not legacy UI.

## Required real iPhone/Safari evidence

Status: **PENDING — release blocker**

No real iPhone is connected to this execution environment, so the following results are intentionally not fabricated. Record device model, iOS version, Safari version, tested commit, screenshots and pass/fail for every row before production release.

| Check | Expected result | Status |
| --- | --- | --- |
| Open install help in Safari | Shows Share → Add to Home Screen guidance without covering required actions | Pending |
| Add to Home Screen and launch | Opens as standalone MyPocket with correct icon/name | Pending |
| Navigate every tab in standalone mode | No blank page; back/forward and reload retain a valid route | Pending |
| Safe areas and touch targets | Header, bottom navigation, sheets and primary actions avoid notch/home indicator | Pending |
| Receipt camera/file selection | Camera and photo library both attach a previewable file | Pending |
| Airplane-mode transaction save | Local pending state is explicit and the UI remains usable | Pending |
| Reconnect synchronization | Pending operation uploads once and resolves visibly | Pending |
| Exact receipt recovery | Downloaded bytes/checksum match the selected receipt | Pending |
| Agent image submission | Image is scoped to the run; OCR result yields review-only draft suggestions | Pending |
| Keyboard and focus | Fields remain visible; closing a sheet restores focus to its trigger | Pending |
| Error/offline/empty states | No crash or blank screen; retry/action guidance remains visible | Pending |

## Release decision

Task 9 automated UI work can be committed, but production deployment must remain fail-closed until this document contains direct physical-device PASS evidence for every required row.
