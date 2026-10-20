# Mobile PWA and API Beta Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the MyPocket beta installable and mobile-safe while keeping user-owned APIs accessible by bearer API key and reports truthful.

**Architecture:** Retain the existing service-worker/IndexedDB split and global API-key middleware. Harden browser install metadata and derive overview charts exclusively from API report data.

**Tech Stack:** React 19, TypeScript, Vite PWA shell, native SVG charts, Go HTTP middleware.

**Spec:** `docs/superpowers/specs/2026-09-07-mobile-pwa-api-beta-design.md`

## Global Constraints

- Mobile viewport is the primary layout target.
- API keys have full owner access to business endpoints; key management remains cookie plus CSRF only.
- Financial visualizations never render sample values.
- Keep dependencies unchanged.

---

### Task 1: Harden PWA install metadata

**Files:**
- Modify: `frontend/index.html`, `frontend/public/manifest.webmanifest`, `frontend/public/sw.js`
- Test: `frontend/e2e/pwa-shell.spec.ts`

- [x] Add Apple/mobile metadata, an explicit PWA identity, and a manifest scope.
- [x] Cache the root shell and manifest with navigation fallback without caching API responses.
- [x] Extend E2E assertions for standalone metadata and service-worker registration.
- [x] Run `npm run test:e2e -- pwa-shell.spec.ts` and `npm run build`.

### Task 2: Remove invented dashboard-chart data

**Files:**
- Modify: `frontend/src/screens/OverviewScreen.tsx`
- Test: `frontend/src/test/baseComponents.test.tsx`

- [x] Write a component test that verifies report chart labels/amounts come from report data and no sample values are rendered.
- [x] Convert the overview chart props from the API `daily` report data.
- [x] Show a Vietnamese empty state when report series is unavailable.
- [x] Run `npm test -- --run` and `npm run build`.

### Task 3: Publish bearer API-key beta contract

**Files:**
- Modify: `docs/architecture/API.md`, `frontend/docs/docs/api/authentication.mdx`
- Test: `backend/internal/platform/httpapi/auth_test.go`

- [x] State that bearer API keys authenticate every user-owned business route and describe the only excluded endpoint families.
- [x] Verify positive bearer authentication and bearer denial for API-key management.
- [x] Run the focused Go HTTP test suite.
