import pytest

from finance_assistant import worker


@pytest.mark.asyncio
async def test_worker_ensures_schema_before_polling(monkeypatch) -> None:
    events: list[str] = []

    async def ensure_schema() -> None:
        events.append("migrate")

    async def stop_after_first_poll() -> int:
        events.append("poll")
        raise RuntimeError("stop")

    monkeypatch.setattr(worker, "migrate", ensure_schema)
    monkeypatch.setattr(worker, "run_once", stop_after_first_poll)

    with pytest.raises(RuntimeError, match="stop"):
        await worker.main()

    assert events == ["migrate", "poll"]
