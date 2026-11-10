from __future__ import annotations

import hashlib
import json
import uuid
from dataclasses import dataclass, field
from datetime import UTC, datetime, timedelta
from typing import Any, Callable


class DraftError(ValueError):
    pass


@dataclass(slots=True)
class ActionDraft:
    id: str
    owner_id: str
    operation: str
    payload: dict[str, Any]
    revision: int
    expires_at: datetime
    status: str = "pending"
    idempotency_key: str = ""
    receipt: dict[str, Any] | None = None
    payload_hash: str = field(init=False)

    def __post_init__(self) -> None:
        canonical = json.dumps(self.payload, sort_keys=True, separators=(",", ":"))
        self.payload_hash = hashlib.sha256(canonical.encode()).hexdigest()
        if not self.idempotency_key:
            self.idempotency_key = f"assistant-{self.id}-r{self.revision}"


class ActionDrafts:
    def __init__(self, *, now: Callable[[], datetime] | None = None) -> None:
        self.now = now or (lambda: datetime.now(UTC))
        self._drafts: dict[str, ActionDraft] = {}

    def create(
        self,
        owner_id: str,
        operation: str,
        payload: dict[str, Any],
        *,
        ttl: timedelta = timedelta(minutes=15),
    ) -> ActionDraft:
        if not owner_id.strip() or not operation.strip() or not payload or ttl <= timedelta(0):
            raise DraftError("invalid_draft")
        draft = ActionDraft(
            id=str(uuid.uuid4()),
            owner_id=owner_id,
            operation=operation,
            payload=dict(payload),
            revision=1,
            expires_at=self.now() + ttl,
        )
        self._drafts[draft.id] = draft
        return draft

    def revise(self, owner_id: str, draft_id: str, revision: int, payload: dict[str, Any]) -> ActionDraft:
        draft = self._get(owner_id, draft_id)
        if draft.revision != revision or draft.status != "pending":
            raise DraftError("stale_revision")
        draft.payload = dict(payload)
        draft.revision += 1
        draft.__post_init__()
        return draft

    def confirm(self, owner_id: str, draft_id: str, revision: int) -> ActionDraft:
        draft = self._get(owner_id, draft_id)
        if self.now() >= draft.expires_at:
            raise DraftError("expired")
        if draft.revision != revision:
            raise DraftError("stale_revision")
        if draft.status == "confirmed":
            return draft
        if draft.status != "pending":
            raise DraftError("not_pending")
        draft.status = "confirmed"
        return draft

    def record_receipt(self, owner_id: str, draft_id: str, receipt: dict[str, Any]) -> ActionDraft:
        draft = self._get(owner_id, draft_id)
        if draft.status != "confirmed":
            raise DraftError("not_confirmed")
        draft.receipt = dict(receipt)
        draft.status = "applied"
        return draft

    def _get(self, owner_id: str, draft_id: str) -> ActionDraft:
        draft = self._drafts.get(draft_id)
        if draft is None:
            raise DraftError("not_found")
        if draft.owner_id != owner_id:
            raise DraftError("owner_mismatch")
        return draft
