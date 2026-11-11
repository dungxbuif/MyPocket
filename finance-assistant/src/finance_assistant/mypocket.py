from __future__ import annotations

from typing import Any
import json
from dataclasses import dataclass, field

import httpx


class MyPocketError(RuntimeError):
    pass


FINANCE_READ_TOOLS = (
    {
        "type": "function",
        "function": {
            "name": "home",
            "description": "Read the owner's current MyPocket overview.",
            "parameters": {"type": "object", "properties": {}, "additionalProperties": False},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "wallets",
            "description": "List the owner's wallets and current balances.",
            "parameters": {"type": "object", "properties": {}, "additionalProperties": False},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "transactions",
            "description": "Read recent owner transactions for analysis.",
            "parameters": {
                "type": "object",
                "properties": {"limit": {"type": "integer", "minimum": 1, "maximum": 100, "default": 20}},
                "additionalProperties": False,
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "budgets",
            "description": "List the owner's budgets.",
            "parameters": {"type": "object", "properties": {}, "additionalProperties": False},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "jars",
            "description": "List the owner's savings jars.",
            "parameters": {"type": "object", "properties": {}, "additionalProperties": False},
        },
    },
)


@dataclass(frozen=True, slots=True)
class FinanceToolCall:
    operation: str
    arguments: dict[str, Any] = field(default_factory=dict)


@dataclass(slots=True)
class FinanceContextResult:
    prompt_lines: list[str] = field(default_factory=list)
    parts: list[dict[str, Any]] = field(default_factory=list)


class MyPocketClient:
    READ_OPERATIONS = {
        "home": ("GET", "/api/v1/home"),
        "wallets": ("GET", "/api/v1/wallets"),
        "transactions": ("GET", "/api/v1/transactions"),
        "budgets": ("GET", "/api/v1/budgets"),
        "jars": ("GET", "/api/v1/jars"),
    }
    def __init__(self, base_url: str, *, timeout: float = 15) -> None:
        self.client = httpx.AsyncClient(base_url=base_url.rstrip("/"), timeout=timeout)

    async def close(self) -> None:
        await self.client.aclose()

    async def read(self, token: str, operation: str, params: dict[str, Any] | None = None) -> Any:
        if operation not in self.READ_OPERATIONS:
            raise MyPocketError("operation_not_allowed")
        method, path = self.READ_OPERATIONS[operation]
        try:
            response = await self.client.request(method, path, headers=self._headers(token), params=params)
        except httpx.RequestError as exc:
            raise MyPocketError("transport_unavailable") from exc
        return self._decode(response)

    @staticmethod
    def _headers(token: str) -> dict[str, str]:
        if not token.startswith("mpa_"):
            raise MyPocketError("invalid_delegation")
        return {"Authorization": f"Bearer {token}", "Accept": "application/json"}

    @staticmethod
    def _decode(response: httpx.Response) -> Any:
        if response.status_code >= 400:
            raise MyPocketError(f"mypocket_{response.status_code}")
        try:
            return response.json()
        except ValueError as exc:
            raise MyPocketError("invalid_response") from exc


class LiveFinanceContext:
    def __init__(self, client: MyPocketClient, credential_vault) -> None:
        self.client = client
        self.credential_vault = credential_vault

    async def __call__(self, owner_id: str, channel_id: str, message: str, delegated_credential: str | None = None) -> FinanceContextResult:
        operations = ["home", "wallets"]
        normalized = message.lower()
        if any(word in normalized for word in ("giao dịch", "chi tiêu", "transaction", "spending")):
            operations.append("transactions")
        return await self.execute(
            owner_id,
            [FinanceToolCall(operation, {"limit": 20} if operation == "transactions" else {}) for operation in operations],
            delegated_credential,
        )

    async def execute(
        self,
        owner_id: str,
        calls: list[FinanceToolCall],
        delegated_credential: str | None = None,
    ) -> FinanceContextResult:
        allowed = {tool["function"]["name"] for tool in FINANCE_READ_TOOLS}
        if any(call.operation not in allowed for call in calls):
            raise MyPocketError("operation_not_allowed")
        try:
            token = delegated_credential or await self.credential_vault.get(owner_id)
        except KeyError:
            return FinanceContextResult(["ledger: unavailable (delegated credential missing)"])
        result = FinanceContextResult()
        for call in calls:
            operation = call.operation
            params = self._safe_params(call)
            try:
                data = await self.client.read(token, operation, params)
                encoded = json.dumps(data, ensure_ascii=False, separators=(",", ":"))
                result.prompt_lines.append(f"{operation}: {encoded[:20_000]}")
                view = data.get("data", data) if isinstance(data, dict) else data
                items = self._items(view)
                if operation == "wallets":
                    result.parts.append({"type": "wallet_balances", "data": {"view": {"items": items}}})
                elif operation == "transactions":
                    income = sum(int(item.get("amount", 0)) for item in items if item.get("type") == "income")
                    expense = sum(int(item.get("amount", 0)) for item in items if item.get("type") == "expense")
                    result.parts.append({"type": "metric_group", "data": {"view": {"income": income, "expense": expense, "net": income - expense, "count": len(items)}}})
                    transaction_view = view if isinstance(view, dict) and isinstance(view.get("items"), list) else {"items": items}
                    result.parts.append({"type": "transaction_list", "data": {"view": transaction_view}})
            except MyPocketError as exc:
                result.prompt_lines.append(f"{operation}: unavailable ({exc})")
        return result

    @staticmethod
    def _safe_params(call: FinanceToolCall) -> dict[str, Any] | None:
        if call.operation != "transactions":
            return None
        raw_limit = call.arguments.get("limit", 20)
        if isinstance(raw_limit, bool) or not isinstance(raw_limit, int):
            raise MyPocketError("invalid_tool_arguments")
        return {"limit": min(max(raw_limit, 1), 100)}

    @staticmethod
    def _items(value: Any) -> list[dict[str, Any]]:
        if isinstance(value, list):
            return [item for item in value if isinstance(item, dict)]
        if isinstance(value, dict) and isinstance(value.get("items"), list):
            return [item for item in value["items"] if isinstance(item, dict)]
        return []
