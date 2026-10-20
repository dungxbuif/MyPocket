---
artifact_type: detail_design
id: FEEDBACK-20261019
status: in_review
owner: shared
---

# Feedback and backlog reconciliation, 2026-10-19

## Intent and authorization

The owner requested all feedback and backlog fixes, then explicitly requested
parallel planning and test authoring before parallel implementation and a final
integrated verification pass. This ticket records that execution method. Preserve
the existing native palette, shared components, financial semantics, owner
isolation and explicit confirmation of AI financial actions.

Source: authenticated production `GET /api/v1/agent/feedback?limit=100` returned
48 records (45 bugs, two features, one improvement), all developer status `open`.
The latest creation is 2026-10-15T05:32:03Z. Status is not evidence that a report
still reproduces: several reports predate subsequent implementations. Raw user
content and credentials are not copied into versioned documentation.

## Execution design

Three workers own disjoint areas. The coordinator owns integration and shared
documentation. All use branch `codex/feedback-backlog-20261019` from `f10cff45`.

| Lane | Owned implementation | Plan/test artifact |
| --- | --- | --- |
| Native | `mobile-ios/` screens, shared bases, state and tests | [Native](FEEDBACK-20261019-NATIVE-DETAIL_DESIGN.md) |
| Finance | Non-AI backend domain/repository/HTTP/tests | [Finance](FEEDBACK-20261019-FINANCE-DETAIL_DESIGN.md) |
| AI | Backend advisor/entry/OCR and evaluations | [AI](FEEDBACK-20261019-AI-DETAIL_DESIGN.md) |

1. Compare every report with current code and prior verification. Classify it as
   a reproducible defect, an implemented feature needing verification, a missing
   contract/feature, or a deployment/device/provider gate.
2. Write independent regression expectations and run focused tests against the
   existing implementation. A failure must demonstrate the claimed behavior,
   not merely unavailable tooling or an inaccurate test fixture.
3. Review lane plans and dependencies at one coordinator barrier. Implement
   independent fixes concurrently after the test/contract checkpoint. Shared
   contracts are agreed before consumers change; shared files have one writer.
4. Run focused green checks, then a combined backend race/vet and native
   unit/build/base-guard pass. Run API-backed UI tests where behavior requires
   persistence or navigation proof. Only one worker controls a simulator.
5. Reconcile internal/public Markdown docs and retained backlog items with
   actual proof. Keep production resolution and owner acceptance separate.

This preserves the owner's requested batch workflow while retaining evidence
that new regression tests could detect each bug. A strictly simultaneous run
of dependent API/UI changes or multiple simulator drivers would add races.

## Feedback coverage ownership

| Lane | Feedback ID prefixes |
| --- | --- |
| Native | 62e31193, 726943be, e11d18e6, 89cca866, 7495cd2f, 52362c1c, 08246b4b, 559242fe, 2eb68f7b, 5d6557d2, b63c7727, 31b7bcf7, 26e78c7f, 1b3c25d8, 196110ec, 534ad994, 4b9a11a5, f82b0108, fd10b2de, c9fb1fb7, ecf02904, 1e12aaa1, a738ad4a, 1db32af3, 8eeb848f, 31ef1877, 75dee199, e5428521, d098ff72, aa6ca3cf, 2d0811c0, 26fcad87, f4597cd2, f7887563, 1565a9e5, b209e80f, 5ff74529, f42c103d |
| Finance (native consumer coordinated) | 6b8b32e1, 06fb8f5b, 4ca27075, fc8cae65, b017284a |
| AI (native rendering coordinated) | afee0f86, 1a59aa1b, 824dea29 |
| Coordinator and native offline | 7ddf77f3, 019b2e9c |

## Backlog boundaries

The savings/budget implementation and worker already exist. Historical draft
rows do not justify rebuilding them. Audit their existing contracts/tests and
retain deployment prerequisites explicitly. Provider selection, cross-process
cancellation, broader AI write tools, APNs and mutable investment history are
distinct contracts; identify their actual status before selecting a change.
Archived web maintenance is outside the current native-only product direction.

## Verification and reconciliation

Local repair and review follow-ups are implemented. Integrated backend race
verification passes 497 test/subtest results with real disposable PostgreSQL
and Redis; native passes 359 unit + 6 selected UI tests, Release simulator
warnings-as-errors and base/project guards. Vet passes. Fourteen live external
provider/storage probes remain opt-in and were skipped in this gate. The [validation matrix](../VALIDATION_MATRIX.md) records
artifacts and the [remaining map](FEEDBACK-20261019-REMAINING.md) preserves
unfinished feature contracts and external acceptance. No production feedback
was resolved and no whole-backlog completion is claimed.

- Backend: focused regressions, `go test -race ./...`, `go vet ./...`, with
  disposable PostgreSQL for database-dependent cases; skipped DB tests must
  remain visible in results.
- Native: focused simulator regressions, full `MyPocketIOSTests`, Release
  warnings-as-errors build, shared-base and project guards; API/UI state checks
  where affected. Physical biometric/codesigning checks remain distinct.
- No production data reset, provider charge, credential rotation, remote
  release or feedback-resolution write is part of this local fix batch.
- Internal: this ticket, lane designs, `BACKLOG.md`, `CONTEXT.md`,
  `VALIDATION_MATRIX.md`, relevant architecture and screen/base specs.
- Public: affected `docs/guides/*.md` human/agent-readable guides and generated
  OpenAPI only when the actual wire contract changes. Changelog records verified
  local changes without claiming publication or deployment.

Links: [backlog](../BACKLOG.md), [phase](../phases/PHASE-002-native-ios-swiftui.md),
[earlier feedback audit](NATIVE-FEEDBACK-PROD-AUDIT-20261014.md),
[validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).
