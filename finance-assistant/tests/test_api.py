import asyncio
import json
from types import SimpleNamespace
from datetime import UTC, datetime, timedelta

import jwt
from fastapi.testclient import TestClient
import pytest

import finance_assistant.agent as agent_module
from finance_assistant.agent import AgentRuntime, OpenAICompatibleModel, SafeLocalModel
from finance_assistant.api import Services, create_app
from finance_assistant.identity import AssertionVerifier
from finance_assistant.memory import MemoryService
from finance_assistant.models import TurnRequest
from finance_assistant.mypocket import FinanceContextResult, FinanceToolCall, MyPocketError
from finance_assistant.pairing import PairingRegistry
from finance_assistant.persistence import InMemoryPersistence
from finance_assistant.run_jobs import InMemoryAssistantRunQueue
from finance_assistant.workspace import WorkspaceProjector
from finance_assistant.worker import process_assistant_run


SECRET = "test-finance-assistant-shared-secret-at-least-32-bytes"


@pytest.mark.asyncio
async def test_model_prompt_truthfully_describes_explicit_memory_persistence() -> None:
    class CapturingModel:
        def bind_tools(self, tools):
            return self

        async def ainvoke(self, messages):
            self.messages = messages
            return SimpleNamespace(content="ok", tool_calls=[])

    provider = CapturingModel()
    await OpenAICompatibleModel(provider).reply([], [], [])

    system = provider.messages[0].content
    assert "Explicit remember requests are persisted after a successful turn" in system
    assert "Do not claim that durable memory is unavailable" in system


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
    return TestClient(create_app(Services(
        AssertionVerifier(SECRET), runtime, PairingRegistry(),
        runs=InMemoryAssistantRunQueue(persistence),
    )))


def test_async_submit_is_durable_idempotent_and_cancellable_before_model_work(tmp_path) -> None:
    api = client(tmp_path)
    headers = {"Authorization": bearer("owner-a"), "X-MyPocket-Delegation": "mpa_test.secret"}
    payload = {"conversation_id": "main", "client_request_id": "request-1", "text": "balance"}

    first = api.post("/api/v1/ai/advisor/messages/async", headers=headers, json=payload)
    replay = api.post("/api/v1/ai/advisor/messages/async", headers=headers, json=payload)

    assert first.status_code == 202
    assert first.json()["data"]["status"] == "queued"
    assert replay.status_code == 202
    assert replay.json()["data"]["run_id"] == first.json()["data"]["run_id"]
    run_id = first.json()["data"]["run_id"]
    assert api.get(f"/api/v1/ai/advisor/runs/{run_id}", headers=headers).json()["data"]["status"] == "queued"

    cancelled = api.post(f"/api/v1/ai/advisor/runs/{run_id}/cancel", headers=headers)
    assert cancelled.status_code == 200
    assert cancelled.json()["data"]["status"] == "cancelled"
    assert api.get(f"/api/v1/ai/advisor/runs/{run_id}", headers=headers).json()["data"]["status"] == "cancelled"


def test_async_submit_requires_worker_delegation(tmp_path) -> None:
    response = client(tmp_path).post(
        "/api/v1/ai/advisor/messages/async",
        headers={"Authorization": bearer("owner-a")},
        json={"conversation_id": "main", "client_request_id": "request-1", "text": "balance"},
    )
    assert response.status_code == 503


@pytest.mark.asyncio
async def test_durable_worker_completes_queued_run_with_stable_run_id(tmp_path) -> None:
    persistence = InMemoryPersistence()
    runtime = AgentRuntime(persistence, MemoryService(persistence, WorkspaceProjector(tmp_path)), SafeLocalModel())
    queue = InMemoryAssistantRunQueue(persistence)
    request = TurnRequest(conversation_id="main", channel_id="app", platform_message_id="request-1", text="balance")
    accepted = await queue.enqueue("owner-a", request, "mpa_test.secret")
    job = (await queue.lease("worker-a"))[0]

    assert await process_assistant_run(runtime, queue, job, "worker-a", poll_interval=0.01)

    completed = await queue.get("owner-a", accepted["run_id"])
    assert completed is not None
    assert completed["run_id"] == accepted["run_id"]
    assert completed["status"] == "completed"


@pytest.mark.asyncio
async def test_durable_worker_cooperatively_cancels_running_model(tmp_path) -> None:
    started = asyncio.Event()

    class SlowModel:
        async def reply(self, messages, memories, ledger_context):
            started.set()
            await asyncio.Event().wait()

    persistence = InMemoryPersistence()
    runtime = AgentRuntime(persistence, MemoryService(persistence, WorkspaceProjector(tmp_path)), SlowModel())
    queue = InMemoryAssistantRunQueue(persistence)
    request = TurnRequest(conversation_id="main", channel_id="app", platform_message_id="request-1", text="balance")
    accepted = await queue.enqueue("owner-a", request, "mpa_test.secret")
    job = (await queue.lease("worker-a"))[0]
    processing = asyncio.create_task(process_assistant_run(runtime, queue, job, "worker-a", poll_interval=0.01))
    await started.wait()

    cancellation = await queue.cancel("owner-a", accepted["run_id"])
    assert cancellation is not None and cancellation["status"] == "running"
    assert not await processing
    cancelled = await queue.get("owner-a", accepted["run_id"])
    assert cancelled is not None and cancelled["status"] == "cancelled"
    assert await persistence.list_messages("owner-a", "app", "main") == []


@pytest.mark.asyncio
async def test_durable_worker_recovers_completion_after_post_model_crash_window(tmp_path) -> None:
    persistence = InMemoryPersistence()
    runtime = AgentRuntime(persistence, MemoryService(persistence, WorkspaceProjector(tmp_path)), SafeLocalModel())
    request = TurnRequest(conversation_id="main", channel_id="app", platform_message_id="request-1", text="balance")
    await runtime.run_turn("owner-a", request)
    queue = InMemoryAssistantRunQueue(persistence)
    accepted = await queue.enqueue("owner-a", request, "mpa_test.secret")
    job = (await queue.lease("worker-a"))[0]

    assert await process_assistant_run(runtime, queue, job, "worker-a", poll_interval=0.01)
    recovered = await queue.get("owner-a", accepted["run_id"])
    assert recovered is not None and recovered["status"] == "completed"


@pytest.mark.asyncio
async def test_durable_worker_reclaims_abandoned_event_claim_without_response(tmp_path) -> None:
    persistence = InMemoryPersistence()
    runtime = AgentRuntime(persistence, MemoryService(persistence, WorkspaceProjector(tmp_path)), SafeLocalModel())
    request = TurnRequest(conversation_id="main", channel_id="app", platform_message_id="request-1", text="balance")
    assert await persistence.accept_event("owner-a", "app", request.platform_message_id)
    queue = InMemoryAssistantRunQueue(persistence)
    accepted = await queue.enqueue("owner-a", request, "mpa_test.secret")
    job = (await queue.lease("worker-a"))[0]

    assert await process_assistant_run(runtime, queue, job, "worker-a", poll_interval=0.01)
    recovered = await queue.get("owner-a", accepted["run_id"])
    assert recovered is not None and recovered["status"] == "completed"
    assert [row.role for row in await persistence.list_messages("owner-a", "app", "main")] == ["user", "assistant"]


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
    contract = capabilities.json()["data"]
    assert contract["tools"] == ["home", "wallets", "transactions", "budgets", "jars"]
    assert contract["write"] is False
    assert contract["direct_writes"] is False
    assert contract["entry_integration"] is False

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
    class ToolModel:
        def __init__(self):
            self.calls = 0

        async def reply(self, messages, memories, ledger_context):
            self.calls += 1
            if self.calls == 1:
                return agent_module.ModelReply(tool_calls=[FinanceToolCall("wallets")])
            assert ledger_context == ["wallets: fresh"]
            return agent_module.ModelReply(text="done")

    class FinanceTools:
        def __init__(self):
            self.calls = []

        async def execute(self, owner_id, calls, delegated_credential):
            self.calls.append((owner_id, calls, delegated_credential))
            return FinanceContextResult(
                ["wallets: fresh"],
                [{"type": "notice", "data": {"text": "fresh"}}],
            )

    persistence = InMemoryPersistence()
    finance_tools = FinanceTools()
    runtime = AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        ToolModel(),
        finance_context=finance_tools,
    )
    request = TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="balance"
    )
    response = asyncio.run(runtime.run_turn("owner-a", request, delegated_credential="mpa_ephemeral"))
    assert finance_tools.calls == [
        ("owner-a", [FinanceToolCall("wallets")], "mpa_ephemeral"),
    ]
    assert response.text == "done"
    assert response.parts[0].type == "notice"
    rows = asyncio.run(persistence.list_messages("owner-a", "app", "main"))
    assert rows[-1].parts[0]["type"] == "notice"


def test_model_can_request_multiple_read_tools_in_one_round(tmp_path) -> None:
    class ToolModel:
        def __init__(self):
            self.calls = 0

        async def reply(self, messages, memories, ledger_context):
            self.calls += 1
            if self.calls == 1:
                return agent_module.ModelReply(tool_calls=[
                    FinanceToolCall("budgets"), FinanceToolCall("jars"),
                ])
            return agent_module.ModelReply(text="grounded")

    class FinanceTools:
        async def execute(self, owner_id, calls, delegated_credential):
            assert owner_id == "owner-a"
            assert [call.operation for call in calls] == ["budgets", "jars"]
            return FinanceContextResult(["budgets: []", "jars: []"])

    persistence = InMemoryPersistence()
    response = asyncio.run(AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        ToolModel(),
        finance_context=FinanceTools(),
    ).run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="plans",
    ), delegated_credential="mpa_ephemeral"))

    assert response.text == "grounded"


def test_model_tool_loop_stops_after_three_execution_rounds(tmp_path) -> None:
    class RepeatingModel:
        async def reply(self, messages, memories, ledger_context):
            return agent_module.ModelReply(tool_calls=[FinanceToolCall("wallets")])

    class FinanceTools:
        def __init__(self):
            self.calls = 0

        async def execute(self, owner_id, calls, delegated_credential):
            self.calls += 1
            return FinanceContextResult([f"wallets round {self.calls}"])

    persistence = InMemoryPersistence()
    tools = FinanceTools()
    response = asyncio.run(AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        RepeatingModel(),
        finance_context=tools,
    ).run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="loop",
    ), delegated_credential="mpa_ephemeral"))

    assert tools.calls == 3
    assert "giới hạn" in response.text.lower()


def test_invalid_model_tool_call_becomes_safe_unavailable_context(tmp_path) -> None:
    class ToolModel:
        def __init__(self):
            self.contexts = []

        async def reply(self, messages, memories, ledger_context):
            self.contexts.append(list(ledger_context))
            if len(self.contexts) == 1:
                return agent_module.ModelReply(tool_calls=[FinanceToolCall("delete_transaction")])
            return agent_module.ModelReply(text="cannot do that")

    class FinanceTools:
        async def execute(self, owner_id, calls, delegated_credential):
            raise MyPocketError("operation_not_allowed")

    persistence = InMemoryPersistence()
    model = ToolModel()
    response = asyncio.run(AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        model,
        finance_context=FinanceTools(),
    ).run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="delete it",
    ), delegated_credential="mpa_ephemeral"))

    assert response.text == "cannot do that"
    assert model.contexts[-1] == ["tool: unavailable (operation_not_allowed)"]


def test_tool_work_is_bounded_per_round_per_turn_and_by_result_size(tmp_path) -> None:
    class RepeatingBatchModel:
        def __init__(self):
            self.calls = 0
            self.final_context = []

        async def reply(self, messages, memories, ledger_context):
            self.calls += 1
            if self.calls <= 2:
                return agent_module.ModelReply(tool_calls=[
                    FinanceToolCall("transactions", {"limit": limit}) for limit in range(1, 101)
                ])
            self.final_context = list(ledger_context)
            return agent_module.ModelReply(text="bounded")

    class FinanceTools:
        def __init__(self):
            self.batch_sizes = []

        async def execute(self, owner_id, calls, delegated_credential):
            self.batch_sizes.append(len(calls))
            return FinanceContextResult(
                ["x" * 20_000 for _ in calls],
                [{"type": "notice", "data": {"text": "x" * 50_000}} for _ in calls],
            )

    persistence = InMemoryPersistence()
    model = RepeatingBatchModel()
    tools = FinanceTools()
    response = asyncio.run(AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        model,
        finance_context=tools,
    ).run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="everything",
    ), delegated_credential="mpa_ephemeral"))

    assert tools.batch_sizes == [5, 3]
    assert sum(len(line) for line in model.final_context) <= 40_000
    assert len(json.dumps([part.model_dump() for part in response.parts])) <= 100_000
    assert any("limit" in line for line in model.final_context)


def test_complete_turn_uses_one_deadline_and_returns_safe_timeout(tmp_path) -> None:
    class SlowModel:
        async def reply(self, messages, memories, ledger_context):
            await asyncio.sleep(0.05)
            return "late"

    persistence = InMemoryPersistence()
    response = asyncio.run(AgentRuntime(
        persistence,
        MemoryService(persistence, WorkspaceProjector(tmp_path)),
        SlowModel(),
        run_timeout_seconds=0.01,
    ).run_turn("owner-a", TurnRequest(
        conversation_id="main", channel_id="app", platform_message_id="m1", text="slow",
    )))

    assert response.status == "completed"
    assert "thời gian" in response.text.lower()
    rows = asyncio.run(persistence.list_messages("owner-a", "app", "main"))
    assert [row.role for row in rows] == ["user", "assistant"]


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
