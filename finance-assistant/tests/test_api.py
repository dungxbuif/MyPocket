import asyncio
from types import SimpleNamespace
from datetime import UTC, datetime, timedelta

import jwt
from fastapi.testclient import TestClient
import pytest

from finance_assistant.agent import AgentRuntime, SafeLocalModel
from finance_assistant.api import Services, create_app
from finance_assistant.identity import AssertionVerifier
from finance_assistant.memory import MemoryService
from finance_assistant.models import TurnRequest
from finance_assistant.mypocket import FinanceContextResult
from finance_assistant.pairing import PairingRegistry
from finance_assistant.persistence import InMemoryPersistence
from finance_assistant.workspace import WorkspaceProjector


SECRET = "test-finance-assistant-shared-secret-at-least-32-bytes"


def bearer(owner: str = "owner-a", actions: list[str] | None = None) -> str:
    now = datetime.now(UTC)
    token = jwt.encode(
        {"iss": "mypocket", "aud": "finance-assistant", "sub": owner, "iat": now, "exp": now + timedelta(minutes=2), "jti": "req-1", "actions": actions or ["assistant:chat"]},
        SECRET,
        algorithm="HS256",
    )
    return f"Bearer {token}"


def client(tmp_path) -> TestClient:
    persistence = InMemoryPersistence()
    memory = MemoryService(persistence, WorkspaceProjector(tmp_path))
    runtime = AgentRuntime(persistence, memory, SafeLocalModel())
    return TestClient(create_app(Services(AssertionVerifier(SECRET), runtime, PairingRegistry())))


def test_turn_rejects_spoofed_owner_header_before_workspace_access(tmp_path) -> None:
    response = client(tmp_path).post(
        "/v1/turns",
        headers={"Authorization": bearer(), "X-Owner-ID": "owner-b"},
        json={"conversation_id": "main", "channel_id": "app", "platform_message_id": "m1", "text": "hello"},
    )
    assert response.status_code == 400
    assert not (tmp_path / "workspaces").exists()


def test_app_turn_is_idempotent_and_memory_is_owner_scoped(tmp_path) -> None:
    api = client(tmp_path)
    payload = {"conversation_id": "main", "channel_id": "app", "platform_message_id": "m1", "text": "Nhớ rằng tôi thích báo cáo ngắn"}
    first = api.post("/v1/turns", headers={"Authorization": bearer("owner-a")}, json=payload)
    repeated = api.post("/v1/turns", headers={"Authorization": bearer("owner-a")}, json=payload)
    other = api.get("/v1/conversations/app/main/messages", headers={"Authorization": bearer("owner-b")})

    assert first.status_code == 200
    assert first.json()["status"] == "completed"
    assert repeated.json()["status"] == "duplicate"
    assert other.json() == []
    workspaces = list((tmp_path / "workspaces").iterdir())
    assert len(workspaces) == 1
    assert "báo cáo ngắn" in (workspaces[0] / "USER.md").read_text()


def test_pairing_requires_chat_permission(tmp_path) -> None:
    response = client(tmp_path).post("/v1/channels/adapter-a/pairings", headers={"Authorization": bearer(actions=["assistant:pair"])})
    assert response.status_code == 403


def test_owner_correction_supersedes_previous_memory(tmp_path) -> None:
    api = client(tmp_path)
    headers = {"Authorization": bearer("owner-a")}
    api.post("/v1/turns", headers=headers, json={"conversation_id": "main", "channel_id": "app", "platform_message_id": "m1", "text": "Nhớ rằng tôi thích báo cáo ngắn"})
    api.post("/v1/turns", headers=headers, json={"conversation_id": "main", "channel_id": "app", "platform_message_id": "m2", "text": "Từ giờ tôi thích báo cáo chi tiết"})
    workspace = next((tmp_path / "workspaces").iterdir())
    content = (workspace / "USER.md").read_text()
    assert "báo cáo chi tiết" in content
    assert "báo cáo ngắn" not in content


def test_native_capabilities_decode_contract_and_empty_reset(tmp_path) -> None:
    api = client(tmp_path)
    headers = {"Authorization": bearer("owner-a")}
    capabilities = api.get("/api/v1/ai/advisor/capabilities", headers=headers)
    assert capabilities.status_code == 200
    assert capabilities.json()["data"]["tools"] == []
    assert capabilities.json()["data"]["entry_integration"] is False

    api.post("/v1/turns", headers=headers, json={"conversation_id": "main", "channel_id": "app", "platform_message_id": "m1", "text": "hello"})
    reset = api.post("/api/v1/ai/advisor/conversation/reset", headers=headers)
    assert reset.status_code == 200
    history = api.get("/v1/conversations/app/main/messages", headers=headers)
    assert history.json() == []


def test_run_survives_runtime_recreation(tmp_path) -> None:
    persistence = InMemoryPersistence()
    memory = MemoryService(persistence, WorkspaceProjector(tmp_path))
    first = AgentRuntime(persistence, memory, SafeLocalModel())
    response = asyncio.run(first.run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="hello"
    )))
    restarted = AgentRuntime(persistence, memory, SafeLocalModel())
    restored = asyncio.run(restarted.get_run("owner-a", response.run_id))
    assert restored == response
    assert asyncio.run(restarted.get_run("owner-b", response.run_id)) is None


def test_app_turn_passes_ephemeral_delegation_to_finance_context(tmp_path) -> None:
    captured = []

    async def finance_context(owner_id: str, channel_id: str, message: str, token: str | None):
        captured.append((owner_id, channel_id, message, token))
        return FinanceContextResult(parts=[{"type": "notice", "data": {"text": "fresh"}}])

    persistence = InMemoryPersistence()
    runtime = AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        SafeLocalModel(),
        finance_context=finance_context,
    )
    request = TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="balance"
    )
    response = asyncio.run(runtime.run_turn("owner-a", request, delegated_credential="mpa_ephemeral"))
    assert captured == [("owner-a", "app", "balance", "mpa_ephemeral")]
    assert response.parts[0].type == "notice"
    rows = asyncio.run(persistence.list_messages("owner-a", "app", "main"))
    assert rows[-1].parts[0]["type"] == "notice"


def test_model_context_is_bounded_to_recent_messages(tmp_path) -> None:
    class RecordingModel:
        def __init__(self):
            self.calls = []

        async def reply(self, messages, memories, ledger_context):
            self.calls.append(list(messages))
            return "ok"

    persistence = InMemoryPersistence()
    model = RecordingModel()
    runtime = AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        model,
        max_context_messages=4,
    )
    for index in range(4):
        asyncio.run(runtime.run_turn("owner-a", TurnRequest(
            conversation_id="main",
            channel_id="app",
            platform_message_id=f"m{index}",
            text=f"question-{index}",
        )))

    assert len(model.calls[-1]) == 4
    assert model.calls[-1][-1].content == "question-3"


def test_telegram_webhook_persists_job_before_acknowledging(tmp_path) -> None:
    class Adapters:
        async def authenticate_webhook(self, adapter_id, secret):
            assert adapter_id == "adapter-a"
            assert secret == "webhook-secret"
            return SimpleNamespace(id=adapter_id, owner_id="owner-a")

        def credential(self, adapter):
            return "123:test-token"

    class Intake:
        def __init__(self):
            self.enqueued = []
            self.handled = 0

        async def enqueue(self, adapter_id, owner_id, payload):
            self.enqueued.append((adapter_id, owner_id, payload["update_id"]))
            return True

        async def handle_update(self, *args, **kwargs):
            self.handled += 1

    persistence = InMemoryPersistence()
    intake = Intake()
    services = Services(
        AssertionVerifier(SECRET),
        AgentRuntime(persistence, MemoryService(persistence, WorkspaceProjector(tmp_path)), SafeLocalModel()),
        PairingRegistry(),
        adapters=Adapters(),
        telegram=intake,
    )
    response = TestClient(create_app(services)).post(
        "/v1/channels/telegram/adapter-a/webhook",
        headers={"X-Telegram-Bot-Api-Secret-Token": "webhook-secret"},
        json={"update_id": 123, "message": {"text": "hello"}},
    )

    assert response.status_code == 202
    assert intake.enqueued == [("adapter-a", "owner-a", 123)]
    assert intake.handled == 0


def test_failed_model_turn_can_resume_same_inbound_event_without_duplicate_user(tmp_path) -> None:
    class FlakyModel:
        def __init__(self):
            self.calls = 0

        async def reply(self, messages, memories, ledger_context):
            self.calls += 1
            if self.calls == 1:
                raise RuntimeError("temporary model failure")
            return "recovered"

    persistence = InMemoryPersistence()
    model = FlakyModel()
    runtime = AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        model,
    )
    request = TurnRequest(
        conversation_id="main",
        channel_id="telegram-adapter-a",
        platform_message_id="update-1",
        text="hello",
    )
    with pytest.raises(RuntimeError, match="temporary model failure"):
        asyncio.run(runtime.run_turn("owner-a", request))

    response = asyncio.run(runtime.run_turn("owner-a", request))
    rows = asyncio.run(persistence.list_messages("owner-a", "telegram-adapter-a", "main"))
    assert response.status == "completed"
    assert [row.role for row in rows] == ["user", "assistant"]
    assert model.calls == 2
