from __future__ import annotations

import hashlib
from typing import Any

import httpx

from .agent import AgentRuntime
from .models import TurnRequest, TurnResponse
from .pairing import PairingError, PairingRegistry


class TelegramAdapter:
    def __init__(self, runtime: AgentRuntime, pairings: PairingRegistry) -> None:
        self.runtime = runtime
        self.pairings = pairings

    @staticmethod
    def channel_id(adapter_id: str) -> str:
        digest = hashlib.sha256(adapter_id.encode()).hexdigest()[:24]
        return f"telegram-{digest}"

    async def handle_update(self, payload: dict[str, Any], *, adapter_id: str, owner_id: str, bot_token: str) -> TurnResponse | None:
        update_id = str(payload.get("update_id", ""))
        message = payload.get("message") or payload.get("edited_message")
        if not update_id or not isinstance(message, dict):
            return None
        sender = message.get("from", {})
        chat = message.get("chat", {})
        if chat.get("type") != "private" or not isinstance(sender.get("id"), int) or not isinstance(chat.get("id"), int):
            return None
        text = str(message.get("text", "")).strip()
        if text.startswith("/start "):
            code = text.split(maxsplit=1)[1].strip()
            try:
                binding = await self.pairings.redeem(code, telegram_user_id=sender["id"], telegram_chat_id=chat["id"])
            except PairingError:
                await self.send(bot_token, chat["id"], "Liên kết đã hết hạn hoặc đã được sử dụng. Hãy tạo liên kết mới trong MyPocket.")
                return None
            if binding.adapter_id != adapter_id or binding.owner_id != owner_id:
                await self.send(bot_token, chat["id"], "Liên kết không thuộc bot này. Hãy tạo liên kết mới trong MyPocket.")
                return None
            await self.send(bot_token, chat["id"], "Đã kết nối Telegram với MyPocket.")
            return None
        binding = await self.pairings.resolve(sender["id"], adapter_id)
        if binding is None or binding.owner_id != owner_id or not text:
            await self.send(bot_token, chat["id"], "Hãy kết nối Telegram từ ứng dụng MyPocket trước.")
            return None
        response = await self.runtime.run_turn(
            binding.owner_id,
            TurnRequest(channel_id=self.channel_id(adapter_id), conversation_id=str(chat["id"]), platform_message_id=update_id, text=text),
        )
        await self.send(bot_token, chat["id"], self.render(response))
        return response

    async def send(self, bot_token: str, chat_id: int, text: str) -> None:
        async with httpx.AsyncClient(timeout=15) as client:
            response = await client.post(f"https://api.telegram.org/bot{bot_token}/sendMessage", json={"chat_id": chat_id, "text": text})
            response.raise_for_status()

    @staticmethod
    def render(response: TurnResponse) -> str:
        lines = [response.text]
        for part in response.parts:
            if part.type == "action_draft":
                lines.append("\nXác nhận trong ứng dụng hoặc nút xác nhận an toàn của bot.")
            elif part.type == "notice" and part.data.get("text"):
                lines.append(str(part.data["text"]))
        return "\n".join(lines)[:4096]
