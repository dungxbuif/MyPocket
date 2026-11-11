from cryptography.fernet import Fernet
from datetime import UTC, datetime
import pytest
import respx
from httpx import Response
import uuid

from finance_assistant.channel_adapters import (
    ChannelAdapterService,
    InMemoryAdapterRepository,
    PostgresAdapterRepository,
)


class _Result:
    def __init__(self, row):
        self.row = row

    async def fetchone(self):
        return self.row


class _Connection:
    def __init__(self, row):
        self.row = row

    async def execute(self, *_args, **_kwargs):
        return _Result(self.row)


class _ConnectionContext:
    def __init__(self, row):
        self.row = row

    async def __aenter__(self):
        return _Connection(self.row)

    async def __aexit__(self, *_args):
        return None


class _Pool:
    def __init__(self, row):
        self.row = row

    def connection(self):
        return _ConnectionContext(self.row)


@pytest.mark.asyncio
@respx.mock
async def test_owner_can_have_multiple_encrypted_telegram_adapters() -> None:
    repository = InMemoryAdapterRepository()
    service = ChannelAdapterService(repository, Fernet.generate_key(), "https://assistant.example.com")
    for token, username, label in (("123:token-one", "first_bot", "Cá nhân"), ("456:token-two", "second_bot", "Gia đình")):
        respx.get(f"https://api.telegram.org/bot{token}/getMe").mock(return_value=Response(200, json={"ok": True, "result": {"username": username}}))
        respx.post(f"https://api.telegram.org/bot{token}/setWebhook").mock(return_value=Response(200, json={"ok": True}))
        adapter, secret = await service.create_telegram("owner-a", label, token)
        assert token.encode() not in adapter.encrypted_credential
        assert service.credential(adapter) == token
        assert secret not in adapter.webhook_secret_hash

    assert len(await repository.list_owner("owner-a")) == 2
    assert await repository.list_owner("owner-b") == []


@pytest.mark.asyncio
@respx.mock
async def test_webhook_secret_and_adapter_owner_are_isolated() -> None:
    repository = InMemoryAdapterRepository()
    service = ChannelAdapterService(repository, Fernet.generate_key(), "https://assistant.example.com")
    token = "123:token-one"
    respx.get(f"https://api.telegram.org/bot{token}/getMe").mock(return_value=Response(200, json={"ok": True, "result": {"username": "first_bot"}}))
    respx.post(f"https://api.telegram.org/bot{token}/setWebhook").mock(return_value=Response(200, json={"ok": True}))
    adapter, secret = await service.create_telegram("owner-a", "Cá nhân", token)

    assert (await service.authenticate_webhook(adapter.id, secret)).owner_id == "owner-a"
    with pytest.raises(PermissionError):
        await service.authenticate_webhook(adapter.id, "wrong")


@pytest.mark.asyncio
@respx.mock
async def test_revoke_removes_telegram_webhook_and_disables_adapter() -> None:
    repository = InMemoryAdapterRepository()
    service = ChannelAdapterService(repository, Fernet.generate_key(), "https://assistant.example.com")
    token = "123:token-one"
    respx.get(f"https://api.telegram.org/bot{token}/getMe").mock(return_value=Response(200, json={"ok": True, "result": {"username": "first_bot"}}))
    respx.post(f"https://api.telegram.org/bot{token}/setWebhook").mock(return_value=Response(200, json={"ok": True}))
    delete = respx.post(f"https://api.telegram.org/bot{token}/deleteWebhook").mock(return_value=Response(200, json={"ok": True}))
    adapter, _ = await service.create_telegram("owner-a", "Cá nhân", token)

    assert await service.revoke_telegram("owner-a", adapter.id) is True
    assert delete.called
    assert (await repository.get(adapter.id)).status == "revoked"
    assert await service.revoke_telegram("owner-b", adapter.id) is False


@pytest.mark.asyncio
async def test_postgres_repository_normalizes_uuid_adapter_id_to_string_contract() -> None:
    adapter_id = uuid.UUID("001b7fbd-bb19-4820-b410-2ba9fccd870c")
    repository = PostgresAdapterRepository(_Pool({
        "id": adapter_id,
        "owner_id": "owner-a",
        "provider": "telegram",
        "label": "Cá nhân",
        "external_identity": "finance_bot",
        "encrypted_credential": b"ciphertext",
        "webhook_secret_hash": "digest",
        "status": "active",
        "created_at": datetime.now(UTC),
    }))

    adapter = await repository.get(str(adapter_id))

    assert adapter is not None
    assert adapter.id == str(adapter_id)
    assert isinstance(adapter.id, str)
