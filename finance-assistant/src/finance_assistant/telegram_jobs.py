from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Protocol

from psycopg.types.json import Jsonb
from psycopg_pool import AsyncConnectionPool


@dataclass(frozen=True, slots=True)
class TelegramJob:
    id: int
    adapter_id: str
    owner_id: str
    payload: dict[str, Any]
    attempts: int


class TelegramJobIntake(Protocol):
    async def enqueue(self, adapter_id: str, owner_id: str, payload: dict[str, Any]) -> bool: ...


class PostgresTelegramJobQueue:
    def __init__(self, pool: AsyncConnectionPool) -> None:
        self.pool = pool

    async def enqueue(self, adapter_id: str, owner_id: str, payload: dict[str, Any]) -> bool:
        update_id = str(payload.get("update_id", "")).strip()
        if not adapter_id or not owner_id or not update_id:
            raise ValueError("invalid Telegram update")
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "INSERT INTO telegram_inbound_jobs(adapter_id,owner_id,update_id,payload) "
                "VALUES (%s,%s,%s,%s) ON CONFLICT(adapter_id,update_id) DO NOTHING RETURNING id",
                (adapter_id, owner_id, update_id, Jsonb(payload)),
            )
            return await result.fetchone() is not None

    async def lease(self, worker_id: str, limit: int = 10) -> list[TelegramJob]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "WITH candidates AS ("
                "SELECT id FROM telegram_inbound_jobs "
                "WHERE completed_at IS NULL AND failed_at IS NULL AND available_at<=now() "
                "AND (lease_until IS NULL OR lease_until<now()) "
                "ORDER BY id FOR UPDATE SKIP LOCKED LIMIT %s"
                ") UPDATE telegram_inbound_jobs j SET lease_owner=%s,"
                "lease_until=now()+interval '5 minutes',attempts=j.attempts+1 "
                "FROM candidates c WHERE j.id=c.id "
                "RETURNING j.id,j.adapter_id,j.owner_id,j.payload,j.attempts",
                (limit, worker_id),
            )
            rows = await result.fetchall()
        return [TelegramJob(
            id=int(row["id"]),
            adapter_id=str(row["adapter_id"]),
            owner_id=str(row["owner_id"]),
            payload=dict(row["payload"]),
            attempts=int(row["attempts"]),
        ) for row in rows]

    async def complete(self, job: TelegramJob, worker_id: str, *, retry: bool = False) -> None:
        async with self.pool.connection() as conn:
            if retry and job.attempts < 10:
                await conn.execute(
                    "UPDATE telegram_inbound_jobs SET available_at=now()+interval '30 seconds',"
                    "lease_owner=NULL,lease_until=NULL WHERE id=%s AND lease_owner=%s",
                    (job.id, worker_id),
                )
            elif retry:
                await conn.execute(
                    "UPDATE telegram_inbound_jobs SET failed_at=now(),lease_owner=NULL,lease_until=NULL "
                    "WHERE id=%s AND lease_owner=%s",
                    (job.id, worker_id),
                )
            else:
                await conn.execute(
                    "UPDATE telegram_inbound_jobs SET completed_at=now(),lease_owner=NULL,lease_until=NULL "
                    "WHERE id=%s AND lease_owner=%s",
                    (job.id, worker_id),
                )
