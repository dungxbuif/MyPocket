from __future__ import annotations

import os
import tempfile
from collections.abc import Iterable
from pathlib import Path

from .persistence import MemoryRecord
from .scope import owner_workspace_path


class WorkspaceProjector:
    def __init__(self, data_root: Path) -> None:
        self.data_root = data_root

    def export(self, owner_id: str, memories: Iterable[MemoryRecord], dreams: Iterable[str] = ()) -> Path:
        root = owner_workspace_path(self.data_root, owner_id)
        episodes = root / "memory"
        episodes.mkdir(parents=True, exist_ok=True, mode=0o700)
        rows = list(memories)
        preferences = [row for row in rows if row.kind == "preference"]
        durable = [row for row in rows if row.kind != "preference"]
        self._atomic(root / "USER.md", self._render("User preferences", preferences))
        self._atomic(root / "MEMORY.md", self._render("Durable memory", durable))
        dream_lines = ["# Reviewed dreams", "", *[f"- {dream}" for dream in dreams]]
        self._atomic(root / "DREAMS.md", "\n".join(dream_lines).rstrip() + "\n")
        return root

    def _render(self, title: str, rows: list[MemoryRecord]) -> str:
        lines = [f"# {title}", "", "Generated projection; PostgreSQL is authoritative.", ""]
        for row in rows:
            lines.append(f"- {row.content} <!-- source:{row.source_type}:{row.source_id} confidence:{row.confidence:.2f} -->")
        return "\n".join(lines).rstrip() + "\n"

    @staticmethod
    def _atomic(path: Path, content: str) -> None:
        descriptor, temp_name = tempfile.mkstemp(prefix=f".{path.name}.", dir=path.parent, text=True)
        try:
            os.fchmod(descriptor, 0o600)
            with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
                handle.write(content)
                handle.flush()
                os.fsync(handle.fileno())
            os.replace(temp_name, path)
        finally:
            if os.path.exists(temp_name):
                os.unlink(temp_name)
