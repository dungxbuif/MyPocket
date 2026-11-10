from datetime import UTC, datetime

from finance_assistant.persistence import MemoryRecord
from finance_assistant.workspace import WorkspaceProjector


def test_workspace_projection_has_memory_and_dream_provenance(tmp_path) -> None:
    record = MemoryRecord("m1", "owner-a", "preference", "Tóm tắt ngắn", "owner_message", "event-1", 1.0, None, datetime.now(UTC))
    root = WorkspaceProjector(tmp_path).export("owner-a", [record], ["Prefer concise monthly review (sources: event-1)"])
    assert "source:owner_message:event-1" in (root / "USER.md").read_text()
    assert "sources: event-1" in (root / "DREAMS.md").read_text()
    assert oct((root / "USER.md").stat().st_mode & 0o777) == "0o600"
