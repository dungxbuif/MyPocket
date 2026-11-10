from __future__ import annotations

import hashlib
from dataclasses import dataclass
from pathlib import Path


def _stable(*parts: str) -> str:
    normalized = "\x1f".join(part.strip() for part in parts)
    return hashlib.sha256(normalized.encode("utf-8")).hexdigest()


@dataclass(frozen=True, slots=True)
class ConversationScope:
    owner_id: str
    channel_id: str
    conversation_id: str

    def __post_init__(self) -> None:
        if not all(value.strip() for value in (self.owner_id, self.channel_id, self.conversation_id)):
            raise ValueError("conversation scope values must be non-empty")

    @property
    def thread_id(self) -> str:
        return f"thread-{_stable(self.owner_id, self.channel_id, self.conversation_id)}"

    @property
    def memory_namespace(self) -> tuple[str, str, str]:
        return ("finance-assistant", _stable(self.owner_id), "memory")


def owner_workspace_path(data_root: Path, owner_id: str) -> Path:
    if not owner_id.strip():
        raise ValueError("owner_id is required")
    return data_root.resolve() / "workspaces" / _stable(owner_id)
