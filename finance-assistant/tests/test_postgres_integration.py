import os
import uuid
from datetime import UTC, datetime, timedelta

import psycopg
import pytest
from cryptography.fernet import Fernet
from langgraph.checkpoint.postgres.aio import AsyncPostgresSaver
from langgraph.store.postgres.aio import AsyncPostgresStore

from finance_assistant.channel_adapters import ChannelAdapter, PostgresAdapterRepository
from finance_assistant.pairing import PostgresPairingRegistry
from finance_assistant.persistence import MemoryRecord, MessageRecord, PostgresPersistence
from finance_assistant.models import TurnRequest
from finance_assistant.run_jobs import PostgresAssistantRunQueue
from finance_assistant.telegram_jobs import PostgresTelegramJobQueue


DATABASE_URL = os.getenv("TEST_FINANCE_ASSISTANT_DATABASE_URL")
pytestmark = pytest.mark.skipif(not DATABASE_URL, reason="TEST_FINANCE_ASSISTANT_DATABASE_URL is not set")


@pytest.mark.asyncio
async def test_postgres_survives_restart_and_isolates_owners() -> None:
    assert DATABASE_URL
    async with await psycopg.AsyncConnection.connect(DATABASE_URL, autocommit=True) as conn:
        await conn.execute("TRUNCATE assistant_run_jobs,telegram_inbound_jobs,telegram_bindings,telegram_pairing_intents,channel_adapters,dream_candidates,memories,agent_runs,conversation_messages,inbound_events CASCADE")

    persistence = PostgresPersistence(DATABASE_URL)
    await persistence.open()
    now = datetime.now(UTC)
    await persistence.append_message(MessageRecord("owner-a", "app", "main", "user", "hello", now))
    await persistence.put_memory(MemoryRecord(str(uuid.uuid4()), "owner-a", "preference", "concise", "owner_message", "e1", 1, None, now))
    run_id = str(uuid.uuid4())
    await persistence.put_run("owner-a", run_id, {"status": "completed"})
    await persistence.put_dream_candidate("owner-a", "candidate", ["app:main:1"])

    adapters = PostgresAdapterRepository(persistence.pool)
    adapter = ChannelAdapter(str(uuid.uuid4()), "owner-a", "telegram", "personal", "owner_a_bot", Fernet.generate_key(), "hash", "active", now)
    await adapters.create(adapter)
    pairings = PostgresPairingRegistry(persistence.pool, now=lambda: now)
    code = await pairings.issue("owner-a", adapter.id, ttl=timedelta(minutes=5))
    binding = await pairings.redeem(code, telegram_user_id=42, telegram_chat_id=99)
    assert binding.owner_id == "owner-a"
    jobs = PostgresTelegramJobQueue(persistence.pool)
    assert await jobs.enqueue(adapter.id, "owner-a", {"update_id": 123, "message": {"text": "hello"}})
    assert not await jobs.enqueue(adapter.id, "owner-a", {"update_id": 123, "message": {"text": "duplicate"}})
    leased = await jobs.lease("worker-a")
    assert [(job.adapter_id, job.owner_id, job.attempts) for job in leased] == [(adapter.id, "owner-a", 1)]
    await jobs.complete(leased[0], "worker-a")
    assert await jobs.lease("worker-b") == []

    run_queue = PostgresAssistantRunQueue(persistence.pool, Fernet.generate_key())
    request = TurnRequest(conversation_id="main", channel_id="app", platform_message_id="request-1", text="balance")
    accepted = await run_queue.enqueue("owner-a", request, "mpa_test.secret")
    replayed = await run_queue.enqueue("owner-a", request, "mpa_test.secret")
    assert replayed["run_id"] == accepted["run_id"]
    run_job = (await run_queue.lease("run-worker"))[0]
    assert run_job.run_id == accepted["run_id"]
    assert run_job.delegated_credential == "mpa_test.secret"
    await run_queue.cancel("owner-a", run_job.run_id)
    assert await run_queue.cancellation_requested(run_job, "run-worker")
    await run_queue.mark_cancelled(run_job, "run-worker")
    assert (await run_queue.get("owner-a", run_job.run_id))["status"] == "cancelled"
    async with persistence.pool.connection() as conn:
        encrypted = await (await conn.execute("SELECT encrypted_delegation FROM assistant_run_jobs WHERE run_id=%s", (run_job.run_id,))).fetchone()
    assert encrypted["encrypted_delegation"] is None
    await persistence.close()

    restarted = PostgresPersistence(DATABASE_URL)
    await restarted.open()
    assert [row.text for row in await restarted.list_messages("owner-a", "app", "main")] == ["hello"]
    assert await restarted.list_messages("owner-b", "app", "main") == []
    assert [row.content for row in await restarted.list_memories("owner-a")] == ["concise"]
    assert await restarted.get_run("owner-a", run_id) == {"status": "completed"}
    assert await restarted.get_run("owner-b", run_id) is None
    assert await restarted.list_dream_candidates("owner-a") == ["candidate"]
    assert (await PostgresPairingRegistry(restarted.pool).resolve(42, adapter.id)).owner_id == "owner-a"
    await restarted.close()

    async with AsyncPostgresSaver.from_conn_string(DATABASE_URL) as checkpointer:
        await checkpointer.setup()
    async with AsyncPostgresStore.from_conn_string(DATABASE_URL) as store:
        await store.setup()
        await store.aput(("memory", "owner-a"), "m1", {"content": "concise"})
        assert (await store.aget(("memory", "owner-a"), "m1")).value["content"] == "concise"
