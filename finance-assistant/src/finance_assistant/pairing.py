from __future__ import annotations

import secrets
import hashlib
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta
from typing import Callable
from psycopg_pool import AsyncConnectionPool


class PairingError(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class TelegramBinding:
    owner_id: str
    adapter_id: str
    telegram_user_id: int
    telegram_chat_id: int


@dataclass(slots=True)
class _Intent:
    owner_id: str
    adapter_id: str
    expires_at: datetime
    used: bool = False


class PairingRegistry:
    """Reference repository used by tests and local development.

    Production wiring uses the same contract with PostgreSQL row locks.
    """

    def __init__(self, *, now: Callable[[], datetime] | None = None) -> None:
        self.now = now or (lambda: datetime.now(UTC))
        self._intents: dict[str, _Intent] = {}
        self._bindings: dict[tuple[str, int], TelegramBinding] = {}

    async def issue(self, owner_id: str, adapter_id: str = "default", *, ttl: timedelta = timedelta(minutes=10)) -> str:
        if not owner_id.strip() or not adapter_id.strip() or ttl <= timedelta(0):
            raise PairingError("invalid_pairing_request")
        code = secrets.token_urlsafe(18)
        self._intents[code] = _Intent(owner_id=owner_id, adapter_id=adapter_id, expires_at=self.now() + ttl)
        return code

    async def redeem(self, code: str, *, telegram_user_id: int, telegram_chat_id: int) -> TelegramBinding:
        intent = self._intents.get(code)
        if intent is None or intent.used:
            raise PairingError("invalid_or_used")
        if self.now() >= intent.expires_at:
            raise PairingError("expired")
        key = (intent.adapter_id, telegram_user_id)
        existing = self._bindings.get(key)
        if existing and existing.owner_id != intent.owner_id:
            raise PairingError("telegram_already_bound")
        intent.used = True
        binding = TelegramBinding(intent.owner_id, intent.adapter_id, telegram_user_id, telegram_chat_id)
        self._bindings[key] = binding
        return binding

    async def resolve(self, telegram_user_id: int, adapter_id: str = "default") -> TelegramBinding | None:
        return self._bindings.get((adapter_id, telegram_user_id))

    async def revoke(self, owner_id: str) -> None:
        self._bindings = {
            key: binding
            for key, binding in self._bindings.items()
            if binding.owner_id != owner_id
        }

    async def revoke_adapter(self, owner_id: str, adapter_id: str) -> None:
        self._bindings = {
            key: binding
            for key, binding in self._bindings.items()
            if not (binding.owner_id == owner_id and binding.adapter_id == adapter_id)
        }


class PostgresPairingRegistry:
    def __init__(self, pool: AsyncConnectionPool, *, now: Callable[[], datetime] | None = None) -> None:
        self.pool = pool
        self.now = now or (lambda: datetime.now(UTC))

    @staticmethod
    def _hash(code: str) -> str:
        return hashlib.sha256(code.encode()).hexdigest()

    async def issue(self, owner_id: str, adapter_id: str = "default", *, ttl: timedelta = timedelta(minutes=10)) -> str:
        if not owner_id.strip() or not adapter_id.strip() or ttl <= timedelta(0):
            raise PairingError("invalid_pairing_request")
        code = secrets.token_urlsafe(18)
        async with self.pool.connection() as conn:
            await conn.execute(
                "INSERT INTO telegram_pairing_intents(code_hash,owner_id,adapter_id,expires_at) VALUES (%s,%s,%s,%s)",
                (self._hash(code), owner_id, adapter_id, self.now() + ttl),
            )
        return code

    async def redeem(self, code: str, *, telegram_user_id: int, telegram_chat_id: int) -> TelegramBinding:
        async with self.pool.connection() as conn:
            async with conn.transaction():
                result = await conn.execute(
                    "SELECT owner_id,adapter_id,expires_at,used_at FROM telegram_pairing_intents WHERE code_hash=%s FOR UPDATE",
                    (self._hash(code),),
                )
                row = await result.fetchone()
                if row is None or row["used_at"] is not None:
                    raise PairingError("invalid_or_used")
                if self.now() >= row["expires_at"]:
                    raise PairingError("expired")
                owner_id, adapter_id = str(row["owner_id"]), str(row["adapter_id"])
                existing = await conn.execute("SELECT owner_id FROM telegram_bindings WHERE adapter_id=%s AND telegram_user_id=%s AND revoked_at IS NULL", (adapter_id, telegram_user_id))
                current = await existing.fetchone()
                if current and str(current["owner_id"]) != owner_id:
                    raise PairingError("telegram_already_bound")
                await conn.execute("UPDATE telegram_pairing_intents SET used_at=%s WHERE code_hash=%s", (self.now(), self._hash(code)))
                await conn.execute(
                    "INSERT INTO telegram_bindings(adapter_id,telegram_user_id,telegram_chat_id,owner_id) VALUES (%s,%s,%s,%s) ON CONFLICT(adapter_id,telegram_user_id) DO UPDATE SET telegram_chat_id=excluded.telegram_chat_id,owner_id=excluded.owner_id,revoked_at=NULL",
                    (adapter_id, telegram_user_id, telegram_chat_id, owner_id),
                )
        return TelegramBinding(owner_id, adapter_id, telegram_user_id, telegram_chat_id)

    async def resolve(self, telegram_user_id: int, adapter_id: str = "default") -> TelegramBinding | None:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "SELECT owner_id,adapter_id,telegram_user_id,telegram_chat_id FROM telegram_bindings WHERE adapter_id=%s AND telegram_user_id=%s AND revoked_at IS NULL",
                (adapter_id, telegram_user_id),
            )
            row = await result.fetchone()
        return TelegramBinding(str(row["owner_id"]), str(row["adapter_id"]), int(row["telegram_user_id"]), int(row["telegram_chat_id"])) if row else None

    async def revoke(self, owner_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute("UPDATE telegram_bindings SET revoked_at=%s WHERE owner_id=%s AND revoked_at IS NULL", (self.now(), owner_id))

    async def revoke_adapter(self, owner_id: str, adapter_id: str) -> None:
        async with self.pool.connection() as conn:
            await conn.execute(
                "UPDATE telegram_bindings SET revoked_at=%s WHERE owner_id=%s AND adapter_id=%s AND revoked_at IS NULL",
                (self.now(), owner_id, adapter_id),
            )
