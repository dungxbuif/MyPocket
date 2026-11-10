from __future__ import annotations

from typing import Any, Literal

from pydantic import BaseModel, Field, field_validator


class AttachmentRef(BaseModel):
    id: str = Field(min_length=1, max_length=128)
    media_type: str = Field(min_length=1, max_length=100)


class TurnRequest(BaseModel):
    conversation_id: str = Field(default="default", min_length=1, max_length=128)
    channel_id: Literal["app", "telegram"] | str = "app"
    platform_message_id: str = Field(min_length=1, max_length=128)
    text: str = Field(min_length=1, max_length=20_000)
    attachments: list[AttachmentRef] = Field(default_factory=list, max_length=10)
    locale: str = Field(default="vi-VN", max_length=32)

    @field_validator("channel_id")
    @classmethod
    def valid_channel(cls, value: str) -> str:
        value = value.strip().lower()
        if not value or len(value) > 40 or not value.replace("-", "").isalnum():
            raise ValueError("invalid channel")
        return value


class ResponsePart(BaseModel):
    type: Literal[
        "metric_group",
        "wallet_balances",
        "wallet_summary",
        "transaction_list",
        "review_question",
        "action_draft",
        "notice",
    ]
    data: dict[str, Any]


class AgentResult(BaseModel):
    text: str
    parts: list[ResponsePart] = Field(default_factory=list)
    action_refs: list[str] = Field(default_factory=list)


class TurnResponse(AgentResult):
    run_id: str
    conversation_id: str
    status: Literal["completed", "duplicate"]


class PairingResponse(BaseModel):
    code: str
    expires_in_seconds: int
    telegram_deep_link: str | None = None


class TelegramAdapterCreate(BaseModel):
    label: str = Field(min_length=1, max_length=80)
    bot_token: str = Field(min_length=20, max_length=256)


class ChannelAdapterView(BaseModel):
    id: str
    provider: Literal["telegram"]
    label: str
    external_identity: str
    status: Literal["active", "revoked"]
    created_at: str


class ChannelAdapterRevokeResponse(BaseModel):
    revoked: bool
