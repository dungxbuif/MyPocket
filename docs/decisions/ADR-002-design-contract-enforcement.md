---
artifact_type: adr
id: ADR-002
status: accepted
owner: shared
---

# Design contracts and base ownership

Owner authorized normalization, guardrails and refactor on 2026-09-13. Trigger: [FB-002](../work/FEEDBACK_LOG.md), [UI-BASE-01](../work/tickets/UI-BASE-01-DETAIL_DESIGN.md). No product phase applies.

## Context

The theme was not imported; atoms duplicated literal colors. Screens and molecules recreated controls and overrode base visuals. Old docs contradicted current radius, tree behavior and inventory status.

## Decision

Keep Markdown contracts in [docs/design](../design/README.md), with a source-extraction register. Remove the 7 PNG and 8 HTML exports after extraction; originals remain in Git at 1bc013d. Screen specifications are written as those screens are implemented; synchronized image/HTML copies are no longer mandatory.

One primitive theme, one component variant registry, semantic props and base-owned controls. Organisms compose reusable atoms/molecules. Missing capability is defined and implemented at base first. Build runs an AST guard against literal colors, native controls outside atoms, local card surfaces, raw headings/paragraphs and visual overrides of bases.

Keep the previously reduced 12px card/control corners; use explicit pill/circle exceptions. Assign brand #4caf50, action #059669 and accent #10b981 distinct roles. This replaces contradictory export palettes, not domain calculations.

## Alternatives

- Docs-only rules: previously present and bypassed; rejected as insufficient.
- Rewrite/delete every working screen: superseded by the latest explicit refactor request; preserve route/API behavior.
- Full component library dependency: unnecessary for these scoped ownership fixes; no dependency added.

## Consequences and limits

Build blocks common source violations. Static analysis is not a visual-equivalence proof: composed class aliases, third-party components and semantic misuse still need review. No per-screen allowlist should be added to bypass checks. New base behavior requires regression proof and consumer review.
Keypad, chart interactions, time-based budget warning and tree collapse remain explicitly partial/planned. Source demo values and labels do not establish product semantics.

Links: [Architecture](../architecture/ARCHITECTURE.md), [Validation](../work/VALIDATION_MATRIX.md), [Release](../releases/CHANGELOG.md), [Backlog](../work/BACKLOG.md).
