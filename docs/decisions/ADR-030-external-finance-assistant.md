---
artifact_type: adr
id: ADR-030
status: accepted
date: 2026-11-10
---

# ADR-030: External owner-isolated Finance Assistant

## Decision

Run Finance Assistant as a private Python service behind the authenticated Go
backend. Use LangGraph, LangMem and PostgreSQL; keep one opaque persistent
workspace per owner; treat native iOS and Telegram as channel adapters with
separate transcripts and shared trusted memory.

MyPocket remains identity and ledger authority. It sends short-lived signed
assertions and revocable exact-operation delegated credentials. The service is
read-only for financial data in its first release. Existing Go-owned proposal
confirmation remains the only assistant-adjacent financial mutation path.

## Consequences

- Native API compatibility and rich cards remain available without exposing
  provider credentials or adding a second client auth system.
- Owners can add multiple Telegram bots; future adapters can reuse the same
  binding and projection interfaces.
- PostgreSQL is the recovery authority. Markdown is a mode-0600 projection for
  inspection/backup, not an execution database.
- Model processing may include personal financial context and therefore needs
  explicit provider/privacy approval.
- The deployment adds an isolated PostgreSQL schema, worker, named volume,
  webhook route, four Docker secrets and coordinated backup/restore
  obligations. It reuses MyPocket's existing database and text owner key.

## Alternatives

Direct client-to-model/service authentication was rejected because it creates a
second public identity boundary. A single cross-channel transcript was rejected
because transport noise would pollute native history. GoClaw was studied as a
behavioral reference but its non-commercial license prevents source copying;
OpenClaw mechanisms are only candidates where the chosen libraries leave a
measured gap.
