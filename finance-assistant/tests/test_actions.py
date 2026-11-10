from datetime import UTC, datetime, timedelta

import pytest

from finance_assistant.actions import ActionDrafts, DraftError


def test_confirmation_is_owner_bound_and_idempotent() -> None:
    now = datetime.now(UTC)
    drafts = ActionDrafts(now=lambda: now)
    draft = drafts.create("owner-a", "transactions.create", {"amount": 120000})
    with pytest.raises(DraftError, match="owner_mismatch"):
        drafts.confirm("owner-b", draft.id, draft.revision)

    first = drafts.confirm("owner-a", draft.id, draft.revision)
    second = drafts.confirm("owner-a", draft.id, draft.revision)
    assert first.idempotency_key == second.idempotency_key


def test_revised_and_expired_drafts_reject_stale_confirmation() -> None:
    now = datetime.now(UTC)
    drafts = ActionDrafts(now=lambda: now)
    old = drafts.create("owner-a", "transactions.create", {"amount": 120000}, ttl=timedelta(seconds=1))
    initial_revision = old.revision
    revised = drafts.revise("owner-a", old.id, initial_revision, {"amount": 150000})
    with pytest.raises(DraftError, match="stale_revision"):
        drafts.confirm("owner-a", old.id, initial_revision)
    assert revised.revision == 2

    drafts.now = lambda: now + timedelta(minutes=2)
    with pytest.raises(DraftError, match="expired"):
        drafts.confirm("owner-a", old.id, revised.revision)
