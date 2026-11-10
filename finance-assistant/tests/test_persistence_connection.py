from finance_assistant import persistence


def test_postgres_pool_pins_schema_for_every_connection(monkeypatch) -> None:
    captured: dict[str, object] = {}

    class CapturingPool:
        def __init__(self, connection_info, **kwargs):
            captured["connection_info"] = connection_info
            captured.update(kwargs)

    monkeypatch.setattr(persistence, "AsyncConnectionPool", CapturingPool)

    persistence.PostgresPersistence(
        "postgresql://assistant:secret@postgres/mypocket",
        "finance_assistant",
    )

    assert (
        captured["connection_info"]
        == "postgresql://assistant:secret@postgres/mypocket"
    )
    assert captured["kwargs"]["options"] == "-csearch_path=finance_assistant"
    assert callable(captured["configure"])
