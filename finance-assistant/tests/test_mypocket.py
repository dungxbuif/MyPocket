import pytest
import httpx

import finance_assistant.mypocket as mypocket

LiveFinanceContext = mypocket.LiveFinanceContext
MyPocketClient = mypocket.MyPocketClient
MyPocketError = mypocket.MyPocketError


def test_assistant_client_does_not_expose_financial_mutations() -> None:
    assert not hasattr(MyPocketClient, "mutate")
    assert not hasattr(MyPocketClient, "WRITE_OPERATIONS")


class Vault:
    async def get(self, owner_id: str) -> str:
        assert owner_id == "owner-a"
        return "mpa_saved"


class Client:
    def __init__(self):
        self.calls = []

    async def read(self, token, operation, params=None):
        assert token in {"mpa_saved", "mpa_ephemeral"}
        self.calls.append((token, operation, params))
        if operation == "home":
            return {"data": {"greeting": "hello"}}
        if operation == "wallets":
            return {"data": [{"id": "w1", "name": "Cash", "current_balance": 500}]}
        if operation == "transactions":
            return {"data": {"items": [{"id": "t1", "type": "expense", "amount": 120}]}}
        if operation == "budgets":
            return {"data": [{"id": "b1", "name": "Food"}]}
        if operation == "jars":
            return {"data": [{"id": "j1", "name": "Emergency"}]}
        raise AssertionError(operation)


def test_read_tool_catalog_is_exact_and_does_not_expose_identity_or_writes() -> None:
    assert hasattr(mypocket, "FINANCE_READ_TOOLS")
    FINANCE_READ_TOOLS = mypocket.FINANCE_READ_TOOLS
    assert tuple(tool["function"]["name"] for tool in FINANCE_READ_TOOLS) == (
        "home", "wallets", "transactions", "budgets", "jars",
    )
    properties = {
        key
        for tool in FINANCE_READ_TOOLS
        for key in tool["function"]["parameters"]["properties"]
    }
    assert properties == {"limit"}
    assert not properties.intersection({"owner", "owner_id", "token", "path"})
    names = {tool["function"]["name"] for tool in FINANCE_READ_TOOLS}
    assert not names.intersection({"create", "update", "delete", "create_transaction"})


@pytest.mark.asyncio
async def test_executor_bounds_transaction_limit_and_preserves_part_order() -> None:
    client = Client()
    context = LiveFinanceContext(client, Vault())

    result = await context.execute(
        "owner-a",
        [mypocket.FinanceToolCall("wallets"), mypocket.FinanceToolCall("transactions", {"limit": 500})],
        "mpa_ephemeral",
    )

    assert client.calls == [
        ("mpa_ephemeral", "wallets", None),
        ("mpa_ephemeral", "transactions", {"limit": 100}),
    ]
    assert [part["type"] for part in result.parts] == [
        "wallet_balances", "metric_group", "transaction_list",
    ]


@pytest.mark.asyncio
async def test_executor_rejects_unknown_or_write_tool_before_http_request() -> None:
    client = Client()
    context = LiveFinanceContext(client, Vault())

    with pytest.raises(MyPocketError, match="operation_not_allowed"):
        await context.execute("owner-a", [mypocket.FinanceToolCall("create_transaction")], "mpa_ephemeral")

    assert client.calls == []


@pytest.mark.asyncio
async def test_executor_returns_unavailable_without_call_when_delegation_missing() -> None:
    class MissingVault:
        async def get(self, owner_id: str) -> str:
            raise KeyError(owner_id)

    client = Client()
    result = await LiveFinanceContext(client, MissingVault()).execute(
        "owner-a", [mypocket.FinanceToolCall("wallets")], None,
    )

    assert result.prompt_lines == ["ledger: unavailable (delegated credential missing)"]
    assert result.parts == []
    assert client.calls == []


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("failure", "expected"),
    [("timeout", "transport_unavailable"), ("malformed", "invalid_response")],
)
async def test_real_client_normalizes_transport_and_decode_failures(failure, expected) -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        if failure == "timeout":
            raise httpx.ReadTimeout("private detail", request=request)
        return httpx.Response(200, text="not-json", request=request)

    client = MyPocketClient("https://mypocket.invalid")
    await client.client.aclose()
    client.client = httpx.AsyncClient(
        base_url="https://mypocket.invalid", transport=httpx.MockTransport(handler),
    )
    try:
        result = await LiveFinanceContext(client, Vault()).execute(
            "owner-a", [mypocket.FinanceToolCall("wallets")], "mpa_ephemeral",
        )
    finally:
        await client.close()

    assert result.prompt_lines == [f"wallets: unavailable ({expected})"]
    assert result.parts == []


@pytest.mark.asyncio
async def test_finance_context_returns_channel_neutral_rich_parts() -> None:
    context = LiveFinanceContext(Client(), Vault())
    result = await context("owner-a", "app", "chi tiêu tháng này", "mpa_ephemeral")

    assert [part["type"] for part in result.parts] == ["wallet_balances", "metric_group", "transaction_list"]
    metric = result.parts[1]["data"]["view"]
    assert metric == {"income": 0, "expense": 120, "net": -120, "count": 1}
    assert result.parts[2]["data"]["view"]["items"][0]["id"] == "t1"
