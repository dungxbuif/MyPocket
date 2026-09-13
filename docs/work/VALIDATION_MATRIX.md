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

## Runtime Verification — 2026-09-13

- CORS preflight: `OPTIONS /api/v1/profile` with `Origin: http://localhost:4173` returned `204` with allow-origin, credentials, methods and headers.
- API health: `GET /api/v1/health` returned `{"status":"ok"}`.
- Google OAuth start: `GET /api/v1/auth/google` returned `302` to Google with redirect URI `http://localhost:8080/api/v1/auth/google/callback`.

Policy lives in `docs/standards/VALIDATION.md`. This matrix is runtime project state and should change as work is planned, implemented, changed, or retired.

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

- Add or update a row when a requirement, ticket, bug, public contract, or accepted behavior is created or changed.
- Mark proof columns `yes`, `no`, or `not_required`.
- Link evidence to `docs/templates/TEST_VERIFICATION.md`, ticket verification sections, UAT, docs review, release notes, or command output summaries.
- Do not set `implemented` until required proof has evidence.
- If a proof type is `not_required`, record the reason in the linked ticket, bug, or verification artifact.
