from __future__ import annotations

from dataclasses import dataclass
from datetime import UTC, datetime
import hashlib
import json
from typing import Any, Protocol
import uuid

from cryptography.fernet import Fernet, InvalidToken
from psycopg.types.json import Jsonb
from psycopg_pool import AsyncConnectionPool

from .models import TurnRequest, TurnResponse
from .persistence import Persistence


TERMINAL_RUN_STATUSES = {"completed", "failed", "cancelled"}


def _now() -> str:
    return datetime.now(UTC).isoformat()


def _queued_payload(run_id: str, conversation_id: str) -> dict[str, Any]:
    now = _now()
    return {
        "run_id": run_id,
        "conversation_id": conversation_id,
        "status": "queued",
        "text": "",
        "parts": [],
        "created_at": now,
        "updated_at": now,
    }


def _request_hash(request: TurnRequest) -> str:
    encoded = json.dumps(request.model_dump(mode="json"), sort_keys=True, separators=(",", ":")).encode()
    return hashlib.sha256(encoded).hexdigest()


@dataclass(frozen=True, slots=True)
class AssistantRunJob:
    run_id: str
    owner_id: str
    request: TurnRequest
    delegated_credential: str
    attempts: int


class AssistantRunQueue(Protocol):
    async def enqueue(self, owner_id: str, request: TurnRequest, delegated_credential: str) -> dict[str, Any]: ...
    async def get(self, owner_id: str, run_id: str) -> dict[str, Any] | None: ...
    async def cancel(self, owner_id: str, run_id: str) -> dict[str, Any] | None: ...
    async def lease(self, worker_id: str, limit: int = 5) -> list[AssistantRunJob]: ...
    async def cancellation_requested(self, job: AssistantRunJob, worker_id: str) -> bool: ...
    async def complete(self, job: AssistantRunJob, worker_id: str, response: TurnResponse) -> bool: ...
    async def fail(self, job: AssistantRunJob, worker_id: str, *, error_code: str = "provider_error") -> None: ...
    async def mark_cancelled(self, job: AssistantRunJob, worker_id: str) -> None: ...


class InMemoryAssistantRunQueue:
    def __init__(self, persistence: Persistence) -> None:
        self.persistence = persistence
        self.jobs: dict[str, dict[str, Any]] = {}
        self.requests: dict[tuple[str, str], str] = {}

    async def enqueue(self, owner_id: str, request: TurnRequest, delegated_credential: str) -> dict[str, Any]:
        if not delegated_credential.startswith("mpa_"):
            raise ValueError("invalid delegated credential")
        key = (owner_id, request.platform_message_id)
        existing_id = self.requests.get(key)
        if existing_id:
            existing = self.jobs[existing_id]
            if existing["request_hash"] != _request_hash(request):
                raise ValueError("client_request_id conflict")
            return dict(await self.get(owner_id, existing_id) or {})
        run_id = str(uuid.uuid4())
        payload = _queued_payload(run_id, request.conversation_id)
        self.requests[key] = run_id
        self.jobs[run_id] = {
            "owner_id": owner_id,
            "request": request,
            "request_hash": _request_hash(request),
            "delegated_credential": delegated_credential,
            "attempts": 0,
            "lease_owner": None,
            "cancel_requested": False,
            "terminal": False,
        }
        await self.persistence.put_run(owner_id, run_id, payload)
        return payload

    async def get(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        return await self.persistence.get_run(owner_id, run_id)

    async def cancel(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        payload = await self.get(owner_id, run_id)
        if payload is None:
            return None
        job = self.jobs.get(run_id)
        if payload.get("status") in TERMINAL_RUN_STATUSES:
            return payload
        if job:
            job["cancel_requested"] = True
        if payload.get("status") == "queued":
            payload.update(status="cancelled", updated_at=_now())
            if job:
                job["terminal"] = True
                job["delegated_credential"] = ""
            await self.persistence.put_run(owner_id, run_id, payload)
        return payload

    async def lease(self, worker_id: str, limit: int = 5) -> list[AssistantRunJob]:
        leased: list[AssistantRunJob] = []
        for run_id, value in self.jobs.items():
            if len(leased) >= limit:
                break
            if value["terminal"] or value["cancel_requested"] or value["lease_owner"] is not None:
                continue
            value["lease_owner"] = worker_id
            value["attempts"] += 1
            payload = await self.get(value["owner_id"], run_id) or {}
            payload.update(status="running", updated_at=_now())
            await self.persistence.put_run(value["owner_id"], run_id, payload)
            leased.append(AssistantRunJob(run_id, value["owner_id"], value["request"], value["delegated_credential"], value["attempts"]))
        return leased

    async def cancellation_requested(self, job: AssistantRunJob, worker_id: str) -> bool:
        value = self.jobs.get(job.run_id)
        return value is None or value["lease_owner"] != worker_id or bool(value["cancel_requested"])

    async def complete(self, job: AssistantRunJob, worker_id: str, response: TurnResponse) -> bool:
        value = self.jobs[job.run_id]
        if value["lease_owner"] == worker_id and not value["cancel_requested"]:
            value.update(terminal=True, lease_owner=None, delegated_credential="")
            payload = response.model_dump()
            payload["status"] = "completed"
            await self.persistence.put_run(job.owner_id, job.run_id, payload)
            return True
        return False

    async def fail(self, job: AssistantRunJob, worker_id: str, *, error_code: str = "provider_error") -> None:
        value = self.jobs[job.run_id]
        if value["lease_owner"] != worker_id:
            return
        payload = await self.get(job.owner_id, job.run_id) or _queued_payload(job.run_id, job.request.conversation_id)
        if job.attempts < 3:
            value["lease_owner"] = None
            payload.update(status="queued", updated_at=_now())
        else:
            value.update(terminal=True, lease_owner=None, delegated_credential="")
            payload.update(status="failed", error_code=error_code, updated_at=_now())
        await self.persistence.put_run(job.owner_id, job.run_id, payload)

    async def mark_cancelled(self, job: AssistantRunJob, worker_id: str) -> None:
        value = self.jobs[job.run_id]
        if value["lease_owner"] == worker_id:
            value.update(terminal=True, lease_owner=None, delegated_credential="", cancel_requested=True)
            payload = await self.get(job.owner_id, job.run_id) or _queued_payload(job.run_id, job.request.conversation_id)
            payload.update(status="cancelled", updated_at=_now())
            await self.persistence.put_run(job.owner_id, job.run_id, payload)


class PostgresAssistantRunQueue:
    def __init__(self, pool: AsyncConnectionPool, encryption_key: str | bytes) -> None:
        self.pool = pool
        self._fernet = Fernet(encryption_key)

    async def enqueue(self, owner_id: str, request: TurnRequest, delegated_credential: str) -> dict[str, Any]:
        if not owner_id or not delegated_credential.startswith("mpa_"):
            raise ValueError("invalid delegated credential")
        request_hash = _request_hash(request)
        run_id = str(uuid.uuid4())
        payload = _queued_payload(run_id, request.conversation_id)
        encrypted = self._fernet.encrypt(delegated_credential.encode())
        async with self.pool.connection() as conn:
            async with conn.transaction():
                await conn.execute(
                    "INSERT INTO agent_runs(id,owner_id,payload) VALUES (%s,%s,%s)",
                    (run_id, owner_id, Jsonb(payload)),
                )
                result = await conn.execute(
                    "INSERT INTO assistant_run_jobs(run_id,owner_id,client_request_id,request_hash,request,encrypted_delegation) "
                    "VALUES (%s,%s,%s,%s,%s,%s) ON CONFLICT(owner_id,client_request_id) DO NOTHING RETURNING run_id",
                    (run_id, owner_id, request.platform_message_id, request_hash, Jsonb(request.model_dump(mode="json")), encrypted),
                )
                inserted = await result.fetchone()
                if inserted:
                    return payload
                await conn.execute("DELETE FROM agent_runs WHERE id=%s AND owner_id=%s", (run_id, owner_id))
                existing_result = await conn.execute(
                    "SELECT j.run_id,j.request_hash,r.payload FROM assistant_run_jobs j JOIN agent_runs r ON r.id=j.run_id "
                    "WHERE j.owner_id=%s AND j.client_request_id=%s",
                    (owner_id, request.platform_message_id),
                )
                existing = await existing_result.fetchone()
                if existing is None or str(existing["request_hash"]) != request_hash:
                    raise ValueError("client_request_id conflict")
                return dict(existing["payload"])

    async def get(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        async with self.pool.connection() as conn:
            result = await conn.execute("SELECT payload FROM agent_runs WHERE id=%s AND owner_id=%s", (run_id, owner_id))
            row = await result.fetchone()
        return dict(row["payload"]) if row else None

    async def cancel(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute("SELECT payload FROM agent_runs WHERE id=%s AND owner_id=%s FOR UPDATE", (run_id, owner_id))
                row = await result.fetchone()
                if row is None:
                    return None
                payload = dict(row["payload"])
                if payload.get("status") in TERMINAL_RUN_STATUSES:
                    return payload
                await conn.execute("UPDATE assistant_run_jobs SET cancel_requested=true WHERE run_id=%s AND owner_id=%s", (run_id, owner_id))
                if payload.get("status") == "queued":
                    payload.update(status="cancelled", updated_at=_now())
                    await conn.execute("UPDATE agent_runs SET payload=%s,updated_at=now() WHERE id=%s AND owner_id=%s", (Jsonb(payload), run_id, owner_id))
                    await conn.execute("UPDATE assistant_run_jobs SET completed_at=now(),encrypted_delegation=NULL WHERE run_id=%s", (run_id,))
                return payload

    async def lease(self, worker_id: str, limit: int = 5) -> list[AssistantRunJob]:
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute(
                    "WITH candidates AS (SELECT run_id FROM assistant_run_jobs WHERE completed_at IS NULL AND failed_at IS NULL "
                    "AND cancel_requested=false AND available_at<=now() AND (lease_until IS NULL OR lease_until<now()) "
                    "ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT %s) "
                    "UPDATE assistant_run_jobs j SET lease_owner=%s,lease_until=now()+interval '5 minutes',attempts=j.attempts+1 "
                    "FROM candidates c WHERE j.run_id=c.run_id "
                    "RETURNING j.run_id,j.owner_id,j.request,j.encrypted_delegation,j.attempts",
                    (limit, worker_id),
                )
                rows = await result.fetchall()
                jobs: list[AssistantRunJob] = []
                for row in rows:
                    try:
                        credential = self._fernet.decrypt(bytes(row["encrypted_delegation"])).decode()
                    except (InvalidToken, TypeError) as exc:
                        raise RuntimeError("invalid encrypted run delegation") from exc
                    request = TurnRequest.model_validate(row["request"])
                    jobs.append(AssistantRunJob(str(row["run_id"]), str(row["owner_id"]), request, credential, int(row["attempts"])))
                    payload_result = await conn.execute("SELECT payload FROM agent_runs WHERE id=%s", (row["run_id"],))
                    payload_row = await payload_result.fetchone()
                    payload = dict(payload_row["payload"])
                    payload.update(status="running", updated_at=_now())
                    await conn.execute("UPDATE agent_runs SET payload=%s,updated_at=now() WHERE id=%s", (Jsonb(payload), row["run_id"]))
        return jobs

    async def cancellation_requested(self, job: AssistantRunJob, worker_id: str) -> bool:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT cancel_requested FROM assistant_run_jobs WHERE run_id=%s AND owner_id=%s AND lease_owner=%s",
                (job.run_id, job.owner_id, worker_id),
            )
            row = await result.fetchone()
        return row is None or bool(row["cancel_requested"])

    async def complete(self, job: AssistantRunJob, worker_id: str, response: TurnResponse) -> bool:
        payload = response.model_dump()
        payload["status"] = "completed"
        payload["updated_at"] = _now()
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute(
                    "UPDATE assistant_run_jobs SET completed_at=now(),lease_owner=NULL,lease_until=NULL,encrypted_delegation=NULL "
                    "WHERE run_id=%s AND owner_id=%s AND lease_owner=%s AND cancel_requested=false RETURNING run_id",
                    (job.run_id, job.owner_id, worker_id),
                )
                if await result.fetchone() is None:
                    return False
                await conn.execute("UPDATE agent_runs SET payload=%s,updated_at=now() WHERE id=%s AND owner_id=%s", (Jsonb(payload), job.run_id, job.owner_id))
                return True

    async def fail(self, job: AssistantRunJob, worker_id: str, *, error_code: str = "provider_error") -> None:
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute(
                    "SELECT r.payload FROM agent_runs r JOIN assistant_run_jobs j ON j.run_id=r.id "
                    "WHERE r.id=%s AND r.owner_id=%s AND j.lease_owner=%s FOR UPDATE OF r,j",
                    (job.run_id, job.owner_id, worker_id),
                )
                row = await result.fetchone()
                if row is None:
                    return
                payload = dict(row["payload"])
                if job.attempts < 3:
                    payload.update(status="queued", updated_at=_now())
                    await conn.execute(
                        "UPDATE assistant_run_jobs SET available_at=now()+interval '30 seconds',lease_owner=NULL,lease_until=NULL "
                        "WHERE run_id=%s AND lease_owner=%s",
                        (job.run_id, worker_id),
                    )
                else:
                    payload.update(status="failed", error_code=error_code, updated_at=_now())
                    await conn.execute(
                        "UPDATE assistant_run_jobs SET failed_at=now(),lease_owner=NULL,lease_until=NULL,encrypted_delegation=NULL "
                        "WHERE run_id=%s AND lease_owner=%s",
                        (job.run_id, worker_id),
                    )
                await conn.execute("UPDATE agent_runs SET payload=%s,updated_at=now() WHERE id=%s", (Jsonb(payload), job.run_id))

    async def mark_cancelled(self, job: AssistantRunJob, worker_id: str) -> None:
        payload = await self.get(job.owner_id, job.run_id) or _queued_payload(job.run_id, job.request.conversation_id)
        payload.update(status="cancelled", updated_at=_now())
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute(
                    "UPDATE assistant_run_jobs SET completed_at=now(),cancel_requested=true,lease_owner=NULL,lease_until=NULL,encrypted_delegation=NULL "
                    "WHERE run_id=%s AND owner_id=%s AND lease_owner=%s RETURNING run_id",
                    (job.run_id, job.owner_id, worker_id),
                )
                if await result.fetchone() is not None:
                    await conn.execute("UPDATE agent_runs SET payload=%s,updated_at=now() WHERE id=%s AND owner_id=%s", (Jsonb(payload), job.run_id, job.owner_id))
