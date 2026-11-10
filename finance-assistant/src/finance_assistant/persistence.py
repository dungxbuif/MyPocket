from __future__ import annotations

import asyncio
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import UTC, datetime
import hashlib
import uuid
from typing import Any, Protocol

from psycopg.rows import dict_row
from psycopg.types.json import Jsonb
from psycopg_pool import AsyncConnectionPool


@dataclass(frozen=True, slots=True)
class MessageRecord:
    owner_id: str
    channel_id: str
    conversation_id: str
    role: str
    text: str
    created_at: datetime
    parts: list[dict[str, Any]] = field(default_factory=list)


@dataclass(frozen=True, slots=True)
class MemoryRecord:
    id: str
    owner_id: str
    kind: str
    content: str
    source_type: str
    source_id: str
    confidence: float
    superseded_by: str | None
    created_at: datetime


class Persistence(Protocol):
    async def accept_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> bool: ...
    async def release_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> None: ...
    async def append_message(self, record: MessageRecord) -> None: ...
    async def list_messages(self, owner_id: str, channel_id: str, conversation_id: str, limit: int = 50) -> list[MessageRecord]: ...
    async def reset_messages(self, owner_id: str, channel_id: str, conversation_id: str) -> None: ...
    async def reset_channel(self, owner_id: str, channel_id: str) -> None: ...
    async def put_memory(self, record: MemoryRecord) -> None: ...
    async def list_memories(self, owner_id: str, limit: int = 50) -> list[MemoryRecord]: ...
    async def supersede_memory(self, owner_id: str, memory_id: str, replacement_id: str) -> None: ...
    async def schedule_dream(self, owner_id: str) -> None: ...
    async def put_dream_candidate(self, owner_id: str, content: str, source_refs: list[str]) -> None: ...
    async def list_dream_candidates(self, owner_id: str, limit: int = 100) -> list[str]: ...
    async def put_run(self, owner_id: str, run_id: str, payload: dict[str, Any]) -> None: ...
    async def get_run(self, owner_id: str, run_id: str) -> dict[str, Any] | None: ...


class InMemoryPersistence:
    def __init__(self) -> None:
        self.events: set[tuple[str, str, str]] = set()
        self.messages: list[MessageRecord] = []
        self.memories: list[MemoryRecord] = []
        self.runs: dict[tuple[str, str], dict[str, Any]] = {}
        self.dream_candidates: dict[str, list[str]] = defaultdict(list)
        self._lock = asyncio.Lock()

    async def accept_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> bool:
        key = (owner_id, channel_id, platform_message_id)
        async with self._lock:
            if key in self.events:
                return False
            self.events.add(key)
            return True

    async def release_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> None:
        async with self._lock:
            self.events.discard((owner_id, channel_id, platform_message_id))

    async def append_message(self, record: MessageRecord) -> None:
        async with self._lock:
            self.messages.append(record)

    async def list_messages(self, owner_id: str, channel_id: str, conversation_id: str, limit: int = 50) -> list[MessageRecord]:
        rows = [row for row in self.messages if row.owner_id == owner_id and row.channel_id == channel_id and row.conversation_id == conversation_id]
        return rows[-limit:]

    async def reset_messages(self, owner_id: str, channel_id: str, conversation_id: str) -> None:
        async with self._lock:
            self.messages = [row for row in self.messages if (row.owner_id, row.channel_id, row.conversation_id) != (owner_id, channel_id, conversation_id)]

    async def reset_channel(self, owner_id: str, channel_id: str) -> None:
        async with self._lock:
            self.messages = [row for row in self.messages if (row.owner_id, row.channel_id) != (owner_id, channel_id)]

    async def put_memory(self, record: MemoryRecord) -> None:
        async with self._lock:
            self.memories.append(record)

    async def list_memories(self, owner_id: str, limit: int = 50) -> list[MemoryRecord]:
        return [row for row in self.memories if row.owner_id == owner_id and row.superseded_by is None][-limit:]

    async def supersede_memory(self, owner_id: str, memory_id: str, replacement_id: str) -> None:
        async with self._lock:
            self.memories = [
                MemoryRecord(row.id, row.owner_id, row.kind, row.content, row.source_type, row.source_id, row.confidence, replacement_id, row.created_at)
                if row.owner_id == owner_id and row.id == memory_id else row
                for row in self.memories
            ]

    async def schedule_dream(self, owner_id: str) -> None:
        return None

    async def put_dream_candidate(self, owner_id: str, content: str, source_refs: list[str]) -> None:
        async with self._lock:
            if content not in self.dream_candidates[owner_id]:
                self.dream_candidates[owner_id].append(content)

    async def list_dream_candidates(self, owner_id: str, limit: int = 100) -> list[str]:
        return self.dream_candidates[owner_id][-limit:]

    async def put_run(self, owner_id: str, run_id: str, payload: dict[str, Any]) -> None:
        async with self._lock:
            self.runs[(owner_id, run_id)] = dict(payload)

    async def get_run(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        value = self.runs.get((owner_id, run_id))
        return dict(value) if value is not None else None


class PostgresPersistence:
    def __init__(self, database_url: str) -> None:
        self.pool = AsyncConnectionPool(database_url, open=False, kwargs={"autocommit": True, "row_factory": dict_row})

    async def open(self) -> None:
        await self.pool.open()

    async def close(self) -> None:
        await self.pool.close()

    async def accept_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> bool:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "INSERT INTO inbound_events(owner_id, channel_id, platform_message_id) VALUES (%s,%s,%s) ON CONFLICT DO NOTHING RETURNING id",
                (owner_id, channel_id, platform_message_id),
            )
            return await result.fetchone() is not None

    async def release_event(self, owner_id: str, channel_id: str, platform_message_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "DELETE FROM inbound_events WHERE owner_id=%s AND channel_id=%s AND platform_message_id=%s",
                (owner_id, channel_id, platform_message_id),
            )

    async def append_message(self, record: MessageRecord) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO conversation_messages(owner_id,channel_id,conversation_id,role,text,created_at,parts) VALUES (%s,%s,%s,%s,%s,%s,%s)",
                (record.owner_id, record.channel_id, record.conversation_id, record.role, record.text, record.created_at, Jsonb(record.parts)),
            )

    async def list_messages(self, owner_id: str, channel_id: str, conversation_id: str, limit: int = 50) -> list[MessageRecord]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT owner_id,channel_id,conversation_id,role,text,created_at,parts FROM conversation_messages WHERE owner_id=%s AND channel_id=%s AND conversation_id=%s ORDER BY seq DESC LIMIT %s",
                (owner_id, channel_id, conversation_id, limit),
            )
            rows = list(reversed(await result.fetchall()))
        return [MessageRecord(**row) for row in rows]

    async def reset_messages(self, owner_id: str, channel_id: str, conversation_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute("DELETE FROM conversation_messages WHERE owner_id=%s AND channel_id=%s AND conversation_id=%s", (owner_id, channel_id, conversation_id))

    async def reset_channel(self, owner_id: str, channel_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "DELETE FROM conversation_messages WHERE owner_id=%s AND channel_id=%s",
                (owner_id, channel_id),
            )

    async def put_memory(self, record: MemoryRecord) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO memories(id,owner_id,kind,content,source_type,source_id,confidence,superseded_by,created_at) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s) ON CONFLICT(id) DO UPDATE SET content=excluded.content,confidence=excluded.confidence,superseded_by=excluded.superseded_by",
                (record.id, record.owner_id, record.kind, record.content, record.source_type, record.source_id, record.confidence, record.superseded_by, record.created_at),
            )

    async def list_memories(self, owner_id: str, limit: int = 50) -> list[MemoryRecord]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT id,owner_id,kind,content,source_type,source_id,confidence,superseded_by,created_at FROM memories WHERE owner_id=%s AND superseded_by IS NULL ORDER BY created_at DESC LIMIT %s",
                (owner_id, limit),
            )
            rows = await result.fetchall()
        return [MemoryRecord(**row) for row in reversed(rows)]

    async def supersede_memory(self, owner_id: str, memory_id: str, replacement_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute("UPDATE memories SET superseded_by=%s WHERE owner_id=%s AND id=%s AND superseded_by IS NULL", (replacement_id, owner_id, memory_id))

    async def schedule_dream(self, owner_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO dream_jobs(owner_id,available_at,updated_at) VALUES (%s,now() + interval '2 minutes',now()) ON CONFLICT(owner_id) DO UPDATE SET available_at=excluded.available_at,updated_at=now()",
                (owner_id,),
            )

    async def put_dream_candidate(self, owner_id: str, content: str, source_refs: list[str]) -> None:
        content_hash = hashlib.sha256(content.encode()).hexdigest()
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO dream_candidates(id,owner_id,content,content_hash,source_refs,status) VALUES (%s,%s,%s,%s,%s,'pending') ON CONFLICT(owner_id,content_hash) DO UPDATE SET source_refs=excluded.source_refs,updated_at=now()",
                (str(uuid.uuid4()), owner_id, content, content_hash, Jsonb(source_refs)),
            )

    async def list_dream_candidates(self, owner_id: str, limit: int = 100) -> list[str]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT content FROM dream_candidates WHERE owner_id=%s AND status='pending' ORDER BY updated_at,id LIMIT %s",
                (owner_id, limit),
            )
            rows = await result.fetchall()
        return [str(row["content"]) for row in rows]

    async def put_run(self, owner_id: str, run_id: str, payload: dict[str, Any]) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO agent_runs(id,owner_id,payload,updated_at) VALUES (%s,%s,%s,now()) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload,updated_at=now() WHERE agent_runs.owner_id=excluded.owner_id",
                (run_id, owner_id, Jsonb(payload)),
            )

    async def get_run(self, owner_id: str, run_id: str) -> dict[str, Any] | None:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT payload FROM agent_runs WHERE id=%s AND owner_id=%s",
                (run_id, owner_id),
            )
            row = await result.fetchone()
        return dict(row["payload"]) if row else None

    async def lease_dreams(self, worker_id: str, limit: int = 10) -> list[str]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "WITH candidates AS (SELECT owner_id FROM dream_jobs WHERE available_at<=now() AND (lease_until IS NULL OR lease_until<now()) ORDER BY available_at FOR UPDATE SKIP LOCKED LIMIT %s) UPDATE dream_jobs d SET lease_owner=%s,lease_until=now()+interval '5 minutes',attempts=d.attempts+1,updated_at=now() FROM candidates c WHERE d.owner_id=c.owner_id RETURNING d.owner_id",
                (limit, worker_id),
            )
            rows = await result.fetchall()
        return [str(row["owner_id"]) for row in rows]

    async def recent_conversations(self, owner_id: str, limit: int = 5) -> list[tuple[str, str]]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT channel_id,conversation_id,max(created_at) AS last_at FROM conversation_messages WHERE owner_id=%s GROUP BY channel_id,conversation_id ORDER BY last_at DESC LIMIT %s",
                (owner_id, limit),
            )
            rows = await result.fetchall()
        return [(str(row["channel_id"]), str(row["conversation_id"])) for row in rows]

    async def complete_dream(self, owner_id: str, worker_id: str, *, retry: bool = False) -> None:
        async with self.pool.connection() as conn:
            if retry:
                await conn.execute("UPDATE dream_jobs SET available_at=now()+interval '10 minutes',lease_owner=NULL,lease_until=NULL,updated_at=now() WHERE owner_id=%s AND lease_owner=%s", (owner_id, worker_id))
            else:
                await conn.execute("DELETE FROM dream_jobs WHERE owner_id=%s AND lease_owner=%s", (owner_id, worker_id))
