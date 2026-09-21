---
artifact_type: validation_matrix
id: VALIDATION_MATRIX
status: active
owner: shared
human_fields:
  - proof_override
  - acceptance_signoff
ai_fields:
  - proof_recommendation
  - evidence_links
  - status_updates
shared_fields:
  - matrix_rows
  - validation_status
updated: 2026-09-13
---

# Validation Matrix

## Field Ownership

- Human owns proof overrides and acceptance sign-off.
- AI recommends proof types, links evidence, and updates status from verification.

This file maps accepted behavior and work items to proof.

## AI plan review — 2026-09-21

### AI-ENTRY-01 implementation follow-up

- [Approved slice and execution proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [UI contract/proof](../design/screens/assistant/README.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md). Entry scope is implemented for review; this does not complete all REQ-17, financial Q&A, transfer matching or retained receipts.
- Backend red→green tests: missing validator/repository/service/config functionality; actual regressions caught hidden category-template acceptance, too-large history blocking follow-up, and cross-session owner quota bypass. PostgreSQL fixtures prove zero writes before approval, eight concurrent approvals producing one row, stale version conflict, cross-owner rejection, terminal reject, deleted-ledger replay safety and explicit report=false preservation. HTTP fixture test exercises real provider adapter → service → DB → edit/approve/reject and reload.
- Independent review remediation: shared category wallet scopes no longer disclose/overwrite other owners' assignments; OCR source survives model failure; post-claim history includes the previous completion. All three regressions observed red then green; new fixture category/wallets removed after verification.
- Final post-review run: full `go test -race ./... -count=1` with local `TEST_DATABASE_URL` passed; frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` passed. Local documentation scan 184 links/22 Markdown files: pass. Git-visible files contain none of the supplied credentials; ignored local env mode is 0600. Runtime restarted with final source.
- `rtk go run ./cmd/migrate up` pass, schema 000010; `rtk go generate ./cmd/api` pass. Go suite/race with `TEST_DATABASE_URL` runs real DB tests; fixture records are cleaned, owner data untouched.
- Frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` pass; Chrome test fixtures prove hold/cancel/release suppression and review flow. Live Chrome through localhost:4173 verifies normal Add/manual form, AI entry alternative, configured OCR/unconfigured AI display. No artificial proposals were inserted into the owner's account.
- Live proxy auth guard returns 401, Swagger 200 with eight entry paths. OCR key is stored but public capability/auth-format probes do not prove actual OCR extraction. Live AI/OCR quality, physical touch and owner acceptance are pending provider config/sample UAT. S3 env is private/ignored but bucket/region and upload implementation remain pending.

- [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md): docs-only feasibility/design/plan delivered for review. Existing two AI tickets remain draft; product approval, provider quality, ledger integration and UAT are pending.
- Read-only evidence: transaction entity/handler/repository confirm current income/expense-only ledger and missing AI confirmation path. OCR public OpenAPI and capabilities returned HTTP 200; current contract does not list scan endpoints. Authenticated OCR/model calls were not run.
- Docs validation: `rtk proxy node --input-type=module -e '…'` checked 117 local Markdown links across the 10 touched documents: pass, zero missing targets. `rtk git diff --check`: pass. Placeholder scan in the new plan found no TODO/TBD. App/Go tests skipped because no runtime code changed. Planned unit, PostgreSQL concurrency, provider contract, model eval and browser UAT gates are in T0–T7; none are claimed passed by this planning work.

## Runtime Verification — 2026-09-13

### Follow-up 2026-09-20

- Budget review follow-up: real-Postgres disappearing-row update regression red (nil error/resurrection) → green (not found/no insert fallback); metadata-date preservation regression passes with browser timezone UTC. `go test ./...` with TEST_DATABASE_URL executes optional integration test. Production API restarted and full `api-roundtrip.mjs` rerun passes after fixes. `npm run build` passes; `git diff --check` clean. Independent review evidence in [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- API-SCREENS-01: `rtk proxy go run ./cmd/migrate up` pass (000009); `rtk proxy node scripts/api-roundtrip.mjs` in app pass through localhost:4173 to PostgreSQL: goal date/deposit/withdrawal/interest/edit/delete/current_balance, invalid goal catalog rejection; budget CRUD/exact overlap409/child category/report exclusion/recalculation and unauthenticated401. Only test-created IDs cleaned, no owner data removed. `go test ./...` pass; design checks/SSR/build pass. New tests fail before implementation for goal filter, category history, budget derivation/validation, masked amounts and problem+json parsing. Browser Google re-login via FE proxy, budget empty/editor data and grouped transaction form inspected; drafts canceled. Owner acceptance and full browser persistence UAT pending. [Evidence and limitations](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- Savings slice: target-date red/green regression, update omit/clear checks, `go test ./...`, Swagger generation pass. Frontend design, shared SSR (goal progress/selection/form), ledger tests and build pass. Browser confirmed grouped selector, goal form and Save/cancel states without saving test data. Full goal create/reload/transaction UAT pending. Transfer/report/notifications are not verified features. [Work evidence](tickets/TICKET-06-01-DETAIL_DESIGN.md).

- Wallet screen spec normalized with source-backed Money Lover references and owner select/immutable-type decisions. Docs review complete; UAT not required for this docs-only update. Runtime select change and per-type detail UI are pending; earlier tests do not cover them.

- UI-WALLET-02: check:design, test:design (7 guard regressions plus SSR form/switch/type contracts), test:transactions (4), build all pass. Add Wallet uses shared sheet/card/switch/type picker; normalized both supplied references. Browser visual and new-form API UAT pending. [Evidence](tickets/UI-WALLET-02-DETAIL_DESIGN.md).

- Add Wallet note removal: `rtk proxy npm run check:design`, `rtk proxy npm run test:design`, `rtk proxy npm run build` passed in `app/`. SSR confirms creation omits “Ghi chú” and editing retains it. Owner visual UAT pending. Docs review: wallet design, context, backlog and changelog reconciled; API/schema unchanged, no ADR required.

- Wallet empty-state follow-up: Overview and WalletManagementPanel now use the tested `StatusMessage` plain variant as requested; design checks/tests/build rerun, visual UAT pending.

- Owner reports previously tested bugs are OK; this is user-reported acceptance of tested fixes, not exhaustive product verification.
- [UI-EMPTY-01](tickets/UI-EMPTY-01.md): `npm run check:design`, `npm run test:design`, and `npm run build` pass in `app/`. SSR regression checks cover plain status without card decoration and preserved danger feedback. New empty-state visual UAT pending owner review.
- Local Google OAuth start endpoint verified HTTP 302 with `redirect_uri=http://localhost:4173/api/v1/auth/google/callback` via FE proxy; session-only launch configuration, no tracked credentials.

### UI-BASE-01 normalization

- Status: implemented / verified for shared UI foundation and standards. [Exact evidence and docs review](tickets/UI-BASE-01-VERIFICATION.md).
- `rtk proxy npm run test:design` in `app/`: pass (6 guardrail regression fixtures and shared SSR contracts).
- `rtk proxy npm run build` in `app/`: pass (source guard, docs guard, typecheck, Vite). Docs guard: 17 Markdown specs, 57 valid links, no PNG/HTML exports.
- Chrome fixture: theme primary color/min touch height, selected keyboard tab, disabled/loading CTA, dialog open/focus trap/Escape/return-focus and visual card/tree/gauge checks passed.
- Root cause reproduced before migration: unimported theme, duplicate literal token values, native controls outside atoms. Owner authorized specs + guardrail + refactor; this supersedes prior UI-BASE-01 removal proposal.
- Product-level CRUD/OAuth/ledger acceptance is unchanged; fixture is shared UI proof, not a new business UAT sign-off.

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Base component styling | TypeScript + production build and authenticated browser inspection | pass | `npm run typecheck && npm run build` in `app/`; canonical `BaseCategoryTree` now powers both Group management and the Reports adapter. Browser proof at `/account/groups`: `nested` source geometry with global colorful icon catalog, 40px parent / 32px child icons, `pl-10` children, a 2px trunk starting at the first child row and curved branches. Regression proof: `BaseButton` now owns inline flex alignment; Quay lại and Nhóm mới both render icon plus label in one centered row. Base card/form/row radii were reduced one token step; semantic circles and pills remain intact. Browser check at `/`: the center create button is vertically contained in the bottom-navigation frame. |
| Database migrations | Dev PostgreSQL migration CLI | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version` in `backend/`; version `8`, `dirty=false`; SQL migrations own schema and the system icon catalog. |
| Default category catalog | Dev PostgreSQL migration, Go test and browser list | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version && go test ./...` in `backend/`; version `8`, `dirty=false`; migration `000008` assigns catalog icon keys for deployment as well as development. |
| API routing and Swagger generation | Go test + code generation | pass | `go generate ./cmd/api && go test ./...` in `backend/`; generated Swagger includes the implemented applicable-wallet endpoint. |
| Group management CRUD | Use-case tests, FE build and owner closure | pass | `go test ./...`, `go generate ./cmd/api`, `npm run typecheck && npm run build` pass. Use case proves owner scope, level/cycle validation and re-parenting children after a personal-parent deletion. System metadata `PATCH` is refused while the dedicated applicable-wallet endpoint remains permitted. Owner requested ticket closure after UI review. |

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Frontend routing | TypeScript compile check | pass | `npm run typecheck` in `app/` |
| Schema and system catalog | Backend unit test | pass | `go test ./...` in `backend/` |
| Categories API auth guard | HTTP manual check | pass | `GET /api/v1/categories` without bearer returned `401` |
| Unified response contract | Go unit tests | pass | `TestOKUsesDataEnvelope`, `TestFailUsesProblemDetails`; `go test ./...` |
| Group management UI | Frontend production build | pass | `npm run typecheck && npm run build` in `app/`; runtime requires authenticated browser session |

- CORS preflight: `OPTIONS /api/v1/profile` with `Origin: http://localhost:4173` returned `204` with allow-origin, credentials, methods and headers.
- API health: `GET /api/v1/health` returned `{"status":"ok"}`.
- Google OAuth start: `GET /api/v1/auth/google` returned `302` to Google with redirect URI `http://localhost:8080/api/v1/auth/google/callback`.
- Swagger UI: `GET /api/v1/docs/index.html` and `GET /api/v1/docs/doc.json` returned `200`; generated paths cover only the implemented auth/profile/home/health endpoints.

Policy lives in `docs/standards/VALIDATION.md`. This matrix is runtime project state and should change as work is planned, implemented, changed, or retired.

### Wallet and basic ledger runtime verification

- Status: core wallet management and the approved basic income/expense ledger are implemented and verified for review. [Wallet evidence](tickets/TICKET-01-02-VERIFICATION.md) · [Ledger evidence](tickets/TICKET-02-01-VERIFICATION.md).
- `rtk go test ./...` in `backend/`: pass, including owner-scoped handler rules, wallet type validation, category-wallet applicability, credit-ledger exclusion and derived balances.
- `rtk go run ./cmd/migrate up` and real HTTP UAT against dev PostgreSQL: pass. Restricted category returned `400`; create/update/delete balance sequence was 115,000 → 113,000 → 93,000; deleting the wallet cascaded its final transaction.
- `rtk go generate ./cmd/api`: pass; Swagger includes wallet `current_balance` and current transaction endpoints.
- `npm run test:transactions`, `npm run test:design`, and `npm run build` in `app/`: pass.
- Chrome UAT through the Vite proxy: create wallet, create expense, list, edit, synchronized Header/Wallet balance, and cleanup all passed. Two stale cross-panel refresh gaps were found during UAT, fixed, and rechecked.
- UAT boundary: receipt/OCR, jar assignment, transfer/adjustment and credit ledger are not required for the approved basic slice and remain open BA/follow-up scope. Human acceptance sign-off remains human-owned.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted as intended behavior, not implemented |
| in_progress | Actively being built or verified |
| implemented | Implemented and evidence exists |
| changed | Contract or expected proof changed after earlier implementation |
| retired | No longer part of the accepted project contract |

## Matrix

| Requirement | Phase | Ticket/Bug | Contract/Behavior | Unit | Integration | E2E | UAT | Platform/Manual | Docs Review | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| REQ-02 | not_applicable | [TICKET-01-02](tickets/TICKET-01-02-quan-ly-vi.md) | Core wallet CRUD, type fields, permanent-delete impact, derived/current total balance | yes | yes | yes | yes | not_required | yes | implemented | [Wallet verification](tickets/TICKET-01-02-VERIFICATION.md); human review remains open for later statement/adjustment scope |
| REQ-04 | not_applicable | [TICKET-02-01](tickets/TICKET-02-01-ghi-thu-chi.md) | Basic income/expense CRUD with wallet/category compatibility and synchronized balances | yes | yes | yes | yes | not_required | yes | implemented | [Ledger verification](tickets/TICKET-02-01-VERIFICATION.md); receipt/jar acceptance remains outside verified slice |
| REQ-01, NFR-04 | not_applicable | [BA tickets](tickets/README.md) | UTC/date-only/account timezone and fixed VND | yes | yes | yes | yes | not_required | yes | planned | Design review only; TIME rules and US-01/05 timezone case; runtime proof pending |
| REQ-02–05, REQ-09–11 | not_applicable | [BA tickets](tickets/README.md) | Wallet, transaction, budget, savings/credit/debt accounting | yes | yes | yes | yes | not_required | yes | planned | Business rules documented; open linked-deletion/credit scenarios remain |
| REQ-06 | not_applicable | [BA tickets](tickets/README.md) | Optional jar grouping, soft warnings, monthly configuration and cumulative view | yes | yes | yes | yes | not_required | yes | planned | JAR rules, numeric example, REPORTS and US-05/06; runtime proof pending |
| REQ-07–08 | not_applicable | [BA tickets](tickets/README.md) | Ordinary recurring and automatic Travel Mode excluding recurring | yes | yes | yes | yes | not_required | yes | planned | REC/TRV rules; backdating/catch-up decisions still open |
| REQ-12–13 | not_applicable | [BA tickets](tickets/README.md) | Recalculable monthly reports, independent notes, context/AI and Insider | yes | yes | yes | yes | not_required | yes | planned | REP/MONTH rules and REPORTS; runtime proof pending |
| REQ-14 | not_applicable | [BA tickets](tickets/README.md) | Moving weighted-average portfolio | yes | yes | yes | yes | not_required | yes | planned | AST formulas documented; funding/historical deletion open |
| REQ-15–18, NFR-01–03, NFR-05–07 | not_applicable | [BA tickets](tickets/README.md) | Keys/API/AI/offline/data lifecycle | yes | yes | yes | yes | yes | yes | planned | Product contract consolidated; no runtime implementation evidence added |

Proof columns above describe required product validation, not test execution. The current [BA ticket index](tickets/README.md) maps these rules to draft parent/child tickets. Phases remain not applicable; ticket creation does not count as implementation proof. Owner document review is pending; no product UAT has been marked passed.

## Documentation Review — 2026-09-13

- Scope: canonical product/business/report docs, requirements/stories, docs guide, context/queue, roadmap cleanup and standards relocation.
- Docs review: passed for captured decisions, trace targets, obsolete product wording and unchanged relocated standards. Open business questions are explicitly retained in [BUSINESS_RULES.md](../requirements/BUSINESS_RULES.md).
- `rtk git diff --check`: pass, exit 0.
- Read-only Node check: 17 docs inspected, 49 Markdown/frontmatter file targets checked, zero missing targets; six product docs have expected review metadata (README excepted).
- Read-only comparison with Git HEAD: all nine relocated standards are byte-identical to their former tracked files.
- Runtime tests/UAT: not run; this update changes documentation only. Product validation remains planned and owner review remains pending.
- API/ERD/runtime docs: no runtime or schema implementation changed; existing unrelated architecture edits were preserved. Technical ADRs remain a prerequisite where future implementation introduces durable architecture/schema decisions, not evidence supplied by this review.
- Reconciliation: requirements, report contract, stories, context, queue note, roadmap and changelog updated. No implementation ticket/phase created.
- Removed material: two superseded product drafts and one inherited Harness CLI phase example. Canonical product content is in `docs/requirements/`; tracked originals can be recovered from Git history. Standards were moved, not discarded.

### Reproduce Link And Trace Check

```sh
rtk proxy node -e 'const fs=require("fs"),path=require("path");let files=["docs/README.md","docs/work/ROADMAP.md",...fs.readdirSync("docs/standards").map(n=>"docs/standards/"+n),...fs.readdirSync("docs/requirements").map(n=>"docs/requirements/"+n)];let failures=[],count=0;for(const file of files){let s=fs.readFileSync(file,"utf8");for(const m of s.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)){const ref=m[1].split("#")[0];if(ref&&!/^[a-z]+:/i.test(ref)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),ref)))failures.push(file+": "+ref);}}let fm=s.match(/^---\n([\s\S]*?)\n---/);if(fm)for(const m of fm[1].matchAll(/^\s+[a-z_]+:\s+(\S+\.md)\s*$/gm)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),m[1])))failures.push(file+": trace "+m[1]);}}console.log(JSON.stringify({files:files.length,linksAndTraces:count,failures},null,2));if(failures.length)process.exit(1);'
```

## BA ticket review

- Owner request: parent tickets with smaller children, concise BA scope and acceptance criteria.
- Result: 11 parents and 31 children, all `draft`; REQ-01 through REQ-18 covered. Common quality requirements are referenced by the shared ticket guide.
- Check: read-only ticket inventory verified unique IDs, draft status, reciprocal parent/child links, at least three acceptance criteria per child, no placeholders/trailing whitespace, and 106 valid Markdown file links. Pass, zero failures.
- `rtk git diff --check`: pass, exit 0.
- Docs review: scope and acceptance use business language; open questions remain in affected tickets; Reports/AI responsibilities avoid separate conflicting formulas.
- UAT: pending future implementation and owner acceptance. Runtime tests not required for this documentation-only breakdown.
- Reconciliation: ticket index, backlog, context, product docs entry point, roadmap note and changelog updated. No runtime, API/schema or architecture change; no technical ADR needed for the breakdown.

## Rules

### Category UI and ASCII ERD docs review

- Naming review: owner chọn `user`; sơ đồ và FK dùng tên đích, hiện trạng mapping cũ chỉ ghi trong mục migration. Kiểm tra `rtk git diff --check`; chưa chạy migration hoặc runtime tests cho thay đổi tên trong docs.

- Phạm vi: Tài khoản → Quản lý nhóm → chọn nhóm → Sửa; nhóm cha, category và ví áp dụng. [UI design](../design/system/DESIGN.md#account--quản-lý-nhóm) liên kết [ASCII ERD](../architecture/ERD.md#ascii--mô-hình-ví-và-nhóm-để-review) và [ticket](tickets/TICKET-01-03-tong-vi-danh-muc.md).
- Docs review: đã phân biệt user model hiện có và quan hệ nhóm/ví đề xuất; giữ catalog cũ; không tự chốt ý nghĩa category, scope rỗng hoặc chiến lược seed. Base component reuse bắt buộc.
- Kiểm tra: `rtk git diff --check` và kiểm tra file đích của Markdown links bằng Node. Runtime/UAT không chạy vì thay đổi chỉ ở docs; UI/API/migration nhóm chưa implement. Review mô hình dữ liệu còn pending; chưa có quyết định kiến trúc được áp dụng cần ADR implementation.

### Wallet research and docs review — 2026-09-13

- Evidence: official Money Lover support for balance adjustment, wallet management, goal/credit wallets; exact URLs and limits recorded in [DESIGN-01-02](tickets/TICKET-01-02-DETAIL_DESIGN.md).
- Reviewed old category catalog `0011_phase002_category_catalog.sql` and research sections 17–18; pinned reuse in wallet/category tickets.
- Corrected deletion/duplicate-name decisions across design, backlog, ERD, architecture and context. Research is not app UAT; no wallet API/schema/seed executed. Full technical design still needs review.
- Documentation-only verification: diff whitespace and local Markdown link checks; product tests/UAT not run for this research update.

- Add or update a row when a requirement, ticket, bug, public contract, or accepted behavior is created or changed.
- Mark proof columns `yes`, `no`, or `not_required`.
- Link evidence to `docs/templates/TEST_VERIFICATION.md`, ticket verification sections, UAT, docs review, release notes, or command output summaries.
- Do not set `implemented` until required proof has evidence.
- If a proof type is `not_required`, record the reason in the linked ticket, bug, or verification artifact.
