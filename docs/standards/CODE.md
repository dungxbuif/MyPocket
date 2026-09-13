# Code Standards

Project-specific code standards should be added by humans as the technology stack is selected.

## Default Rules

- Prefer existing local patterns over new abstractions.
- Keep changes scoped to the active ticket or bug.
- Avoid unrelated refactors.
- Add abstractions only when they remove real complexity or match an established pattern.
- Do not add dependencies without documenting the reason and impact.

## Dependency Rules

Adding or replacing a major dependency requires:

- Detail design approval
- ADR
- Update to relevant setup or deployment docs
- Test evidence

## Naming And Structure

- Every API response MUST use shared response helpers; handlers MUST NOT construct ad-hoc JSON envelopes.
- Success responses use `{data, meta}`. Errors use RFC 9457-style Problem Details with stable code constants and `request_id`.
- Repeated UI patterns MUST be configurable atoms/molecules before screen composition.
- Frontend work MUST read `docs/design/README.md` and the matching component contract before JSX. A missing base must be specified and implemented before screen composition. Native controls belong to atoms; screens use shared text/card/badge/progress/feedback bases.
- Base visual changes use named props/variants, never consumer `className` color/font/radius/shadow overrides. Primitive values live only in `app/src/ui/theme.css`; component variants live in `app/src/ui/variants.ts`.
- `npm run check:design` and `npm run test:design` in `app/` are required for UI changes; production build runs the guard. Do not add per-screen bypasses. Review remains required for semantic misuse the AST cannot prove.
- User-visible strings, route fragments, API paths, error codes, and domain kinds MUST be named constants. No magic strings or numbers in handlers/components.
