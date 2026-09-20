---
artifact_type: ticket
id: UI-EMPTY-01
status: in_review
owner: shared
---

# Borderless transaction empty state

Follow-up: owner also requests the same presentation for “Chưa có ví”. Overview and WalletManagementPanel now explicitly use `StatusMessage variant="plain"` for wallet empty states. The initial patch covered only transaction messages; this closes the missed wallet consumers. Verification below is rerun for this follow-up; visual acceptance remains pending.

Owner request 2026-09-20: previously tested bugs are OK; continue work and remove the border around “Chưa có giao dịch”. This statement records the user's reported test outcome, not verification of every unfinished feature.

Small task exemption: yes
Reason: bounded presentation correction explicitly requested by owner; no business behavior changes.
Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no

Root cause: `StatusMessage` always composed `SurfaceCard`, which supplies border/background/shadow. Both empty-state consumers inherit it.
Approved approach: add `plain` presentation to the shared base and opt in only the Overview and Transactions empty messages. Default card/loading/error presentation stays intact.
Reused bases: `app/src/atomic/atoms/StatusMessage.tsx`, `Text.tsx`, `SurfaceCard.tsx`.

Acceptance: empty messages render semantic status text without a card; error feedback retains alert semantics. No new keyboard interaction.

Links: [Backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [screen spec](../../design/screens/transactions/README.md), [base contract](../../design/system/BASE_COMPONENTS.md), [release](../../releases/CHANGELOG.md), [context](../../CONTEXT.md). Phase: none. Separate detail design and ADR: not required for this owner-directed presentation correction. Requirements/API/ERD/architecture: unchanged.

Verification: `rtk proxy npm run check:design`, `rtk proxy npm run test:design`, and `rtk proxy npm run build` pass in `app/`. Shared SSR coverage checks borderless status content and preserved danger alert styling. Visual UAT for the new presentation remains pending owner review in the running app. Docs review: base/screen contracts and trace links updated; no master business/API/schema change or new ADR required.
