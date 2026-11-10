from __future__ import annotations

from contextlib import asynccontextmanager

from fastapi import FastAPI
from langgraph.checkpoint.postgres.aio import AsyncPostgresSaver
from langgraph.store.postgres.aio import AsyncPostgresStore

from .agent import AgentRuntime, OpenAICompatibleModel, SafeLocalModel
from .api import Services, create_app
from .channel_adapters import ChannelAdapterService, PostgresAdapterRepository
from .config import Settings, get_settings
from .credentials import PostgresCredentialVault
from .identity import AssertionVerifier
from .memory import MemoryService
from .model_provider import build_chat_model
from .mypocket import LiveFinanceContext, MyPocketClient
from .pairing import PairingRegistry, PostgresPairingRegistry
from .persistence import InMemoryPersistence, PostgresPersistence
from .run_jobs import PostgresAssistantRunQueue
from .telegram_jobs import PostgresTelegramJobQueue
from .workspace import WorkspaceProjector


def _model(settings: Settings):
    api_key = settings.model_api_key.get_secret_value()
    if not api_key:
        if settings.environment == "production":
            raise RuntimeError("FINANCE_ASSISTANT_MODEL_API_KEY is required in production")
        return SafeLocalModel()
    return OpenAICompatibleModel(build_chat_model(settings))


def build_app(settings: Settings | None = None) -> FastAPI:
    settings = settings or get_settings()
    placeholder_store = InMemoryPersistence()
    placeholder_memory = MemoryService(placeholder_store, WorkspaceProjector(settings.data_root))
    placeholder_runtime = AgentRuntime(placeholder_store, placeholder_memory, SafeLocalModel())
    pairings = PairingRegistry()
    services = Services(AssertionVerifier(settings.assertion_secret.get_secret_value()), placeholder_runtime, pairings)
    app = create_app(services)

    @asynccontextmanager
    async def lifespan(_: FastAPI):
        persistence = PostgresPersistence(
            settings.database_url, settings.database_schema
        )
        await persistence.open()
        async with AsyncPostgresSaver.from_conn_string(settings.database_url) as checkpointer:
            async with AsyncPostgresStore.from_conn_string(settings.database_url) as store:
                await checkpointer.setup()
                await store.setup()
                memory = MemoryService(persistence, WorkspaceProjector(settings.data_root))
                services.runtime = AgentRuntime(
                    persistence,
                    memory,
                    _model(settings),
                    max_context_messages=settings.max_context_messages,
                    checkpointer=checkpointer,
                    store=store,
                )
                encryption_key = settings.credential_encryption_key.get_secret_value()
                if not encryption_key:
                    raise RuntimeError("FINANCE_ASSISTANT_CREDENTIAL_ENCRYPTION_KEY is required")
                services.adapters = ChannelAdapterService(PostgresAdapterRepository(persistence.pool), encryption_key, settings.public_base_url)
                services.credentials = PostgresCredentialVault(persistence.pool, encryption_key)
                services.pairings = PostgresPairingRegistry(persistence.pool)
                mypocket = MyPocketClient(settings.mypocket_base_url)
                services.runtime.finance_context = LiveFinanceContext(mypocket, services.credentials)
                services.telegram = PostgresTelegramJobQueue(persistence.pool)
                services.runs = PostgresAssistantRunQueue(persistence.pool, encryption_key)
                try:
                    yield
                finally:
                    await mypocket.close()
                    await persistence.close()

    app.router.lifespan_context = lifespan
    return app


app = build_app()
