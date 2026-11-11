import uuid

import pytest

from finance_assistant.models import TurnResponse
from finance_assistant.pairing import PairingRegistry
from finance_assistant.telegram_adapter import TelegramAdapter


class RecordingRuntime:
    def __init__(self):
        self.requests = []

    async def run_turn(self, owner_id, request):
        self.requests.append((owner_id, request))
        return TurnResponse(
            run_id=request.platform_message_id,
            conversation_id=request.conversation_id,
            status="completed",
            text="ok",
        )


class RecordingTelegramAdapter(TelegramAdapter):
    def __init__(self, runtime, pairings):
        super().__init__(runtime, pairings)
        self.sent = []

    async def send(self, bot_token, chat_id, text):
        self.sent.append((bot_token, chat_id, text))


def update(update_id: int, user_id: int, chat_id: int, text: str) -> dict:
    return {
        "update_id": update_id,
        "message": {
            "from": {"id": user_id},
            "chat": {"id": chat_id, "type": "private"},
            "text": text,
        },
    }


@pytest.mark.asyncio
async def test_uuid_adapters_produce_valid_distinct_channel_scopes() -> None:
    runtime = RecordingRuntime()
    pairings = PairingRegistry()
    telegram = RecordingTelegramAdapter(runtime, pairings)
    adapters = [
        str(uuid.UUID("001b7fbd-bb19-4820-b410-2ba9fccd870c")),
        str(uuid.UUID("101b7fbd-bb19-4820-b410-2ba9fccd870c")),
    ]

    for index, adapter_id in enumerate(adapters, start=1):
        code = await pairings.issue("owner-a", adapter_id)
        await pairings.redeem(code, telegram_user_id=index, telegram_chat_id=100 + index)
        await telegram.handle_update(
            update(index, index, 100 + index, "Tôi có bao nhiêu ví?"),
            adapter_id=adapter_id,
            owner_id="owner-a",
            bot_token="test-token",
        )

    channels = [request.channel_id for _, request in runtime.requests]
    assert len(channels) == 2
    assert len(set(channels)) == 2
    assert all(channel.startswith("telegram-") and len(channel) <= 40 for channel in channels)
