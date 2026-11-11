from __future__ import annotations

import hashlib
import secrets
import uuid
from dataclasses import dataclass, replace
from datetime import UTC, datetime
from typing import Protocol

import httpx
from cryptography.fernet import Fernet
from psycopg.rows import dict_row
from psycopg_pool import AsyncConnectionPool


@dataclass(frozen=True, slots=True)
class ChannelAdapter:
    id: str
    owner_id: str
    provider: str
    label: str
    external_identity: str
    encrypted_credential: bytes
    webhook_secret_hash: str
    status: str
    created_at: datetime


class AdapterRepository(Protocol):
    async def create(self, adapter: ChannelAdapter) -> None: ...
    async def get(self, adapter_id: str) -> ChannelAdapter | None: ...
    async def list_owner(self, owner_id: str) -> list[ChannelAdapter]: ...
    async def revoke(self, owner_id: str, adapter_id: str) -> bool: ...
    async def count_active(self, owner_id: str) -> int: ...


class InMemoryAdapterRepository:
    def __init__(self) -> None:
        self.rows: dict[str, ChannelAdapter] = {}

    async def create(self, adapter: ChannelAdapter) -> None:
        self.rows[adapter.id] = adapter

    async def get(self, adapter_id: str) -> ChannelAdapter | None:
        return self.rows.get(adapter_id)

    async def list_owner(self, owner_id: str) -> list[ChannelAdapter]:
        return [row for row in self.rows.values() if row.owner_id == owner_id]

    async def revoke(self, owner_id: str, adapter_id: str) -> bool:
        row = self.rows.get(adapter_id)
        if row is None or row.owner_id != owner_id:
            return False
        self.rows[adapter_id] = replace(row, status="revoked")
        return True

    async def count_active(self, owner_id: str) -> int:
        return sum(row.owner_id == owner_id and row.status == "active" for row in self.rows.values())


class PostgresAdapterRepository:
    def __init__(self, pool: AsyncConnectionPool) -> None:
        self.pool = pool

    @staticmethod
    def _adapter(row) -> ChannelAdapter:
        values = dict(row)
        values["id"] = str(values["id"])
        return ChannelAdapter(**values)

    async def create(self, adapter: ChannelAdapter) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO channel_adapters(id,owner_id,provider,label,external_identity,encrypted_credential,webhook_secret_hash,status,created_at,updated_at) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
                (adapter.id, adapter.owner_id, adapter.provider, adapter.label, adapter.external_identity, adapter.encrypted_credential, adapter.webhook_secret_hash, adapter.status, adapter.created_at, adapter.created_at),
            )

    async def get(self, adapter_id: str) -> ChannelAdapter | None:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT id,owner_id,provider,label,external_identity,encrypted_credential,webhook_secret_hash,status,created_at FROM channel_adapters WHERE id=%s",
                (adapter_id,),
            )
            row = await result.fetchone()
        return self._adapter(row) if row else None

    async def list_owner(self, owner_id: str) -> list[ChannelAdapter]:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT id,owner_id,provider,label,external_identity,encrypted_credential,webhook_secret_hash,status,created_at FROM channel_adapters WHERE owner_id=%s ORDER BY created_at,id",
                (owner_id,),
            )
            rows = await result.fetchall()
        return [self._adapter(row) for row in rows]

    async def revoke(self, owner_id: str, adapter_id: str) -> bool:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "UPDATE channel_adapters SET status='revoked',updated_at=now() WHERE id=%s AND owner_id=%s AND status='active'",
                (adapter_id, owner_id),
            )
            return result.rowcount == 1

    async def count_active(self, owner_id: str) -> int:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT count(*) AS total FROM channel_adapters WHERE owner_id=%s AND status='active'",
                (owner_id,),
            )
            row = await result.fetchone()
        return int(row["total"])


class ChannelAdapterService:
    def __init__(self, repository: AdapterRepository, encryption_key: str | bytes, public_base_url: str) -> None:
        self.repository = repository
        self.fernet = Fernet(encryption_key)
        self.public_base_url = public_base_url.rstrip("/")

    async def create_telegram(self, owner_id: str, label: str, bot_token: str) -> tuple[ChannelAdapter, str]:
        if not self.public_base_url.startswith("https://"):
            raise ValueError("public_base_url_must_use_https")
        adapter_id = str(uuid.uuid4())
        webhook_secret = secrets.token_urlsafe(32)
        async with httpx.AsyncClient(timeout=15) as client:
            me = await client.get(f"https://api.telegram.org/bot{bot_token}/getMe")
            me.raise_for_status()
            body = me.json()
            if not body.get("ok") or not body.get("result", {}).get("username"):
                raise ValueError("invalid_telegram_bot_token")
            username = str(body["result"]["username"])
            hook = await client.post(
                f"https://api.telegram.org/bot{bot_token}/setWebhook",
                json={
                    "url": f"{self.public_base_url}/v1/channels/telegram/{adapter_id}/webhook",
                    "secret_token": webhook_secret,
                    "allowed_updates": ["message", "callback_query"],
                    "drop_pending_updates": False,
                },
            )
            hook.raise_for_status()
        adapter = ChannelAdapter(
            id=adapter_id,
            owner_id=owner_id,
            provider="telegram",
            label=label.strip(),
            external_identity=username,
            encrypted_credential=self.fernet.encrypt(bot_token.encode()),
            webhook_secret_hash=self._hash(webhook_secret),
            status="active",
            created_at=datetime.now(UTC),
        )
        await self.repository.create(adapter)
        return adapter, webhook_secret

    async def authenticate_webhook(self, adapter_id: str, secret: str) -> ChannelAdapter:
        adapter = await self.repository.get(adapter_id)
        if adapter is None or adapter.status != "active" or not secrets.compare_digest(adapter.webhook_secret_hash, self._hash(secret)):
            raise PermissionError("invalid_webhook")
        return adapter

    async def revoke_telegram(self, owner_id: str, adapter_id: str) -> bool:
        adapter = await self.repository.get(adapter_id)
        if adapter is None or adapter.owner_id != owner_id or adapter.status != "active":
            return False
        token = self.credential(adapter)
        async with httpx.AsyncClient(timeout=15) as client:
            response = await client.post(
                f"https://api.telegram.org/bot{token}/deleteWebhook",
                json={"drop_pending_updates": False},
            )
            response.raise_for_status()
        return await self.repository.revoke(owner_id, adapter_id)

    def credential(self, adapter: ChannelAdapter) -> str:
        return self.fernet.decrypt(adapter.encrypted_credential).decode()

    @staticmethod
    def _hash(value: str) -> str:
        return hashlib.sha256(value.encode()).hexdigest()
