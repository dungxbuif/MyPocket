from __future__ import annotations

import asyncio
import uuid

from langgraph.checkpoint.postgres.aio import AsyncPostgresSaver
from langgraph.store.postgres.aio import AsyncPostgresStore

from .agent import AgentRuntime, OpenAICompatibleModel
from .channel_adapters import ChannelAdapterService, PostgresAdapterRepository
from .config import get_settings
from .credentials import PostgresCredentialVault
from .memory import DreamWorker, LangMemDreamer, MemoryService
from .migrate import migrate
from .model_provider import build_chat_model
from .mypocket import LiveFinanceContext, MyPocketClient
from .pairing import PostgresPairingRegistry
from .persistence import PostgresPersistence
from .run_jobs import AssistantRunJob, AssistantRunQueue, PostgresAssistantRunQueue
from .telegram_adapter import TelegramAdapter
from .telegram_jobs import PostgresTelegramJobQueue
from .workspace import WorkspaceProjector


async def process_assistant_run(
    runtime: AgentRuntime,
    queue: AssistantRunQueue,
    job: AssistantRunJob,
    worker_id: str,
    *,
    poll_interval: float = 0.5,
) -> bool:
    task = asyncio.create_task(runtime.run_turn(
        job.owner_id,
        job.request,
        delegated_credential=job.delegated_credential,
        run_id=job.run_id,
    ))
    try:
        while not task.done():
            done, _ = await asyncio.wait({task}, timeout=poll_interval)
            if done:
                break
            if await queue.cancellation_requested(job, worker_id):
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass
                await queue.mark_cancelled(job, worker_id)
                return False
        response = await task
        return await queue.complete(job, worker_id, response)
    except asyncio.CancelledError:
        task.cancel()
        raise
    except Exception:
        await queue.fail(job, worker_id)
        return False


async def run_once() -> int:
    settings = get_settings()
    api_key = settings.model_api_key.get_secret_value()
    if not api_key:
        raise RuntimeError("FINANCE_ASSISTANT_MODEL_API_KEY is required for dreaming")
    persistence = PostgresPersistence(settings.database_url, settings.database_schema)
    await persistence.open()
    worker_id = str(uuid.uuid4())
    chat_model = build_chat_model(settings)
    memory = MemoryService(persistence, WorkspaceProjector(settings.data_root))
    dream_worker = DreamWorker(persistence, memory, LangMemDreamer(chat_model))
    adapter_repository = PostgresAdapterRepository(persistence.pool)
    adapters = ChannelAdapterService(adapter_repository, settings.credential_encryption_key.get_secret_value(), settings.public_base_url)
    credentials = PostgresCredentialVault(persistence.pool, settings.credential_encryption_key.get_secret_value())
    mypocket = MyPocketClient(settings.mypocket_base_url)
    jobs = PostgresTelegramJobQueue(persistence.pool)
    run_jobs = PostgresAssistantRunQueue(persistence.pool, settings.credential_encryption_key.get_secret_value())
    processed = 0
    try:
        async with AsyncPostgresSaver.from_conn_string(settings.database_url) as checkpointer:
            async with AsyncPostgresStore.from_conn_string(settings.database_url) as store:
                await checkpointer.setup()
                await store.setup()
                runtime = AgentRuntime(
                    persistence,
                    memory,
                    OpenAICompatibleModel(chat_model),
                    max_context_messages=settings.max_context_messages,
                    checkpointer=checkpointer,
                    store=store,
                    finance_context=LiveFinanceContext(mypocket, credentials),
                )
                telegram = TelegramAdapter(runtime, PostgresPairingRegistry(persistence.pool))
                leased_runs = await run_jobs.lease(worker_id)
                if leased_runs:
                    outcomes = await asyncio.gather(*(
                        process_assistant_run(runtime, run_jobs, job, worker_id)
                        for job in leased_runs
                    ))
                    processed += sum(outcomes)
                for job in await jobs.lease(worker_id):
                    try:
                        adapter = await adapter_repository.get(job.adapter_id)
                        if adapter is None or adapter.status != "active" or adapter.owner_id != job.owner_id:
                            await jobs.complete(job, worker_id)
                            continue
                        await telegram.handle_update(
                            job.payload,
                            adapter_id=adapter.id,
                            owner_id=adapter.owner_id,
                            bot_token=adapters.credential(adapter),
                        )
                        await jobs.complete(job, worker_id)
                        processed += 1
                    except Exception:
                        await jobs.complete(job, worker_id, retry=True)
                for owner_id in await persistence.lease_dreams(worker_id):
                    try:
                        for channel_id, conversation_id in await persistence.recent_conversations(owner_id):
                            await dream_worker.review_owner(owner_id, channel_id, conversation_id)
                        await persistence.complete_dream(owner_id, worker_id)
                        processed += 1
                    except Exception:
                        await persistence.complete_dream(owner_id, worker_id, retry=True)
    finally:
        await mypocket.close()
        await persistence.close()
    return processed


async def main() -> None:
    await migrate()
    while True:
        await run_once()
        await asyncio.sleep(30)


if __name__ == "__main__":
    asyncio.run(main())
