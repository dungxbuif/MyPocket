from datetime import UTC, datetime

import pytest

from finance_assistant.memory import DreamWorker, MemoryService
from finance_assistant.persistence import InMemoryPersistence, MessageRecord
from finance_assistant.workspace import WorkspaceProjector


class CapturingDreamer:
    def __init__(self) -> None:
        self.messages = []

    async def propose(self, messages):
        self.messages = messages
        return ["prefers concise summaries"]


@pytest.mark.asyncio
async def test_dream_candidates_persist_and_never_use_assistant_output_as_source(tmp_path) -> None:
    persistence = InMemoryPersistence()
    now = datetime.now(UTC)
    await persistence.append_message(MessageRecord("owner-a", "app", "main", "user", "I prefer concise summaries", now))
    await persistence.append_message(MessageRecord("owner-a", "app", "main", "assistant", "You have a billion dollars", now))
    memory = MemoryService(persistence, WorkspaceProjector(tmp_path))
    dreamer = CapturingDreamer()
    worker = DreamWorker(persistence, memory, dreamer)

    await worker.review_owner("owner-a", "app", "main")
    await memory.remember_explicit("owner-a", "Remember that reports use VND", "m2")

    assert dreamer.messages == [{"role": "user", "content": "I prefer concise summaries"}]
    assert len(await persistence.list_dream_candidates("owner-a")) == 1
    workspace = next((tmp_path / "workspaces").iterdir())
    assert "prefers concise summaries" in (workspace / "DREAMS.md").read_text()
    assert "billion dollars" not in (workspace / "DREAMS.md").read_text()
