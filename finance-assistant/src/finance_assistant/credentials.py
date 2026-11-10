from __future__ import annotations

from cryptography.fernet import Fernet, InvalidToken
from psycopg_pool import AsyncConnectionPool


class CredentialVault:
    """Process-local cache with encrypted values; production records live in PostgreSQL."""

    def __init__(self, encryption_key: str | bytes) -> None:
        self._fernet = Fernet(encryption_key)
        self._values: dict[str, bytes] = {}

    def put(self, owner_id: str, token: str) -> None:
        if not owner_id.strip() or not token.startswith("mpa_"):
            raise ValueError("invalid delegated credential")
        self._values[owner_id] = self._fernet.encrypt(token.encode())

    def get(self, owner_id: str) -> str:
        encrypted = self._values.get(owner_id)
        if encrypted is None:
            raise KeyError("delegation_missing")
        try:
            return self._fernet.decrypt(encrypted).decode()
        except InvalidToken as exc:
            raise KeyError("delegation_invalid") from exc

    def revoke(self, owner_id: str) -> None:
        self._values.pop(owner_id, None)


class PostgresCredentialVault:
    def __init__(self, pool: AsyncConnectionPool, encryption_key: str | bytes) -> None:
        self.pool = pool
        self._fernet = Fernet(encryption_key)

    async def put(self, owner_id: str, credential_id: str, token: str) -> str | None:
        if not owner_id or not credential_id or not token.startswith("mpa_"):
            raise ValueError("invalid delegated credential")
        encrypted = self._fernet.encrypt(token.encode())
        async with self.pool.connection() as conn:
            existing = await conn.execute(
                "SELECT credential_id FROM delegated_credentials WHERE owner_id=%s AND revoked_at IS NULL",
                (owner_id,),
            )
            previous = await existing.fetchone()
            await conn.execute(
                "INSERT INTO delegated_credentials(owner_id,credential_id,encrypted_secret,updated_at) VALUES (%s,%s,%s,now()) ON CONFLICT(owner_id) DO UPDATE SET credential_id=excluded.credential_id,encrypted_secret=excluded.encrypted_secret,revoked_at=NULL,updated_at=now()",
                (owner_id, credential_id, encrypted),
            )
        return str(previous["credential_id"]) if previous and previous["credential_id"] != credential_id else None

    async def get(self, owner_id: str) -> str:
        async with self.pool.connection() as conn:
            result = await conn.execute("SELECT encrypted_secret FROM delegated_credentials WHERE owner_id=%s AND revoked_at IS NULL", (owner_id,))
            row = await result.fetchone()
        if row is None:
            raise KeyError("delegation_missing")
        return self._fernet.decrypt(bytes(row["encrypted_secret"])).decode()

    async def revoke(self, owner_id: str) -> str | None:
        async with self.pool.connection() as conn:
            result = await conn.execute(
                "UPDATE delegated_credentials SET revoked_at=now(),updated_at=now() WHERE owner_id=%s AND revoked_at IS NULL RETURNING credential_id",
                (owner_id,),
            )
            row = await result.fetchone()
        return str(row["credential_id"]) if row else None
