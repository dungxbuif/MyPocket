from datetime import UTC, datetime, timedelta

import pytest

from finance_assistant.pairing import PairingError, PairingRegistry


@pytest.mark.asyncio
async def test_pairing_code_is_single_use_and_owner_bound() -> None:
    now = datetime.now(UTC)
    registry = PairingRegistry(now=lambda: now)
    code = await registry.issue("owner-a", ttl=timedelta(minutes=5))
    binding = await registry.redeem(code, telegram_user_id=42, telegram_chat_id=99)
    assert binding.owner_id == "owner-a"
    with pytest.raises(PairingError, match="invalid_or_used"):
        await registry.redeem(code, telegram_user_id=42, telegram_chat_id=99)


@pytest.mark.asyncio
async def test_expired_pairing_code_is_rejected() -> None:
    now = datetime.now(UTC)
    registry = PairingRegistry(now=lambda: now)
    code = await registry.issue("owner-a", ttl=timedelta(seconds=1))
    registry.now = lambda: now + timedelta(seconds=2)
    with pytest.raises(PairingError, match="expired"):
        await registry.redeem(code, telegram_user_id=42, telegram_chat_id=99)
