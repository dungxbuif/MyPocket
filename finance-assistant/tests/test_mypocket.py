import pytest

from finance_assistant.mypocket import LiveFinanceContext, MyPocketClient


def test_assistant_client_does_not_expose_financial_mutations() -> None:
    assert not hasattr(MyPocketClient, "mutate")
    assert not hasattr(MyPocketClient, "WRITE_OPERATIONS")


class Vault:
    async def get(self, owner_id: str) -> str:
        assert owner_id == "owner-a"
        return "mpa_saved"


class Client:
    async def read(self, token, operation, params=None):
        assert token in {"mpa_saved", "mpa_ephemeral"}
        if operation == "home":
            return {"data": {"greeting": "hello"}}
        if operation == "wallets":
            return {"data": [{"id": "w1", "name": "Cash", "current_balance": 500}]}
        if operation == "transactions":
            return {"data": {"items": [{"id": "t1", "type": "expense", "amount": 120}]}}
        raise AssertionError(operation)


@pytest.mark.asyncio
async def test_finance_context_returns_channel_neutral_rich_parts() -> None:
    context = LiveFinanceContext(Client(), Vault())
    result = await context("owner-a", "app", "chi tiêu tháng này", "mpa_ephemeral")

    assert [part["type"] for part in result.parts] == ["wallet_balances", "metric_group", "transaction_list"]
    metric = result.parts[1]["data"]["view"]
    assert metric == {"income": 0, "expense": 120, "net": -120, "count": 1}
    assert result.parts[2]["data"]["view"]["items"][0]["id"] == "t1"
