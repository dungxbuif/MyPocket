from __future__ import annotations

from dataclasses import dataclass
from datetime import timedelta
import httpx
from datetime import UTC, datetime
import uuid
from fastapi import Depends, FastAPI, Header, HTTPException, Request, Response, status

from .agent import AgentRuntime
from .channel_adapters import ChannelAdapterService
from .credentials import PostgresCredentialVault
from .identity import AssistantIdentity, AssertionErrorCode, AssertionVerifier
from .models import ChannelAdapterRevokeResponse, ChannelAdapterView, PairingResponse, TelegramAdapterCreate, TurnRequest, TurnResponse
from .mypocket import FINANCE_READ_TOOLS
from .pairing import PairingRegistry
from .run_jobs import AssistantRunQueue, TERMINAL_RUN_STATUSES
from .telegram_jobs import TelegramJobIntake


@dataclass(slots=True)
class Services:
    verifier: AssertionVerifier
    runtime: AgentRuntime
    pairings: PairingRegistry
    adapters: ChannelAdapterService | None = None
    credentials: PostgresCredentialVault | None = None
    telegram: TelegramJobIntake | None = None
    runs: AssistantRunQueue | None = None


def create_app(services: Services) -> FastAPI:
    app = FastAPI(title="MyPocket Finance Assistant", version="0.1.0")
    app.state.services = services

    async def identity(
        authorization: str | None = Header(default=None),
        x_owner_id: str | None = Header(default=None),
    ) -> AssistantIdentity:
        if x_owner_id is not None:
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="client owner headers are forbidden")
        if not authorization or not authorization.startswith("Bearer "):
            raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="assistant assertion required")
        try:
            value = services.verifier.verify(authorization.removeprefix("Bearer ").strip())
        except AssertionErrorCode as exc:
            raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=exc.code) from exc
        return value

    def require_action(action: str):
        async def required(actor: AssistantIdentity = Depends(identity)) -> AssistantIdentity:
            if action not in actor.actions:
                raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="assistant action not permitted")
            return actor
        return required

    @app.get("/healthz")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/v1/turns", response_model=TurnResponse)
    async def turns(
        payload: TurnRequest,
        actor: AssistantIdentity = Depends(require_action("assistant:chat")),
        x_mypocket_delegation: str | None = Header(default=None),
    ) -> TurnResponse:
        if payload.channel_id != "app":
            raise HTTPException(status_code=400, detail="backend intake only accepts app channel")
        return await services.runtime.run_turn(actor.owner_id, payload, delegated_credential=x_mypocket_delegation)

    @app.get("/v1/conversations/{channel_id}/{conversation_id}/messages")
    async def messages(channel_id: str, conversation_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        rows = await services.runtime.persistence.list_messages(actor.owner_id, channel_id, conversation_id)
        return [{"role": row.role, "text": row.text, "created_at": row.created_at.isoformat()} for row in rows]

    @app.post("/v1/conversations/{channel_id}/{conversation_id}/reset", status_code=204)
    async def reset(channel_id: str, conversation_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))) -> None:
        await services.runtime.persistence.reset_messages(actor.owner_id, channel_id, conversation_id)

    @app.post("/v1/channels/{adapter_id}/pairings", response_model=PairingResponse)
    async def pairing(adapter_id: str, actor: AssistantIdentity = Depends(require_action("assistant:channels"))) -> PairingResponse:
        if services.adapters is None:
            raise HTTPException(status_code=503, detail="channel adapters unavailable")
        adapter = await services.adapters.repository.get(adapter_id)
        if adapter is None or adapter.owner_id != actor.owner_id or adapter.status != "active":
            raise HTTPException(status_code=404, detail="channel adapter not found")
        ttl = timedelta(minutes=10)
        code = await services.pairings.issue(actor.owner_id, adapter_id, ttl=ttl)
        return PairingResponse(code=code, expires_in_seconds=int(ttl.total_seconds()), telegram_deep_link=f"https://t.me/{adapter.external_identity}?start={code}")

    @app.post("/v1/channels/telegram", response_model=ChannelAdapterView, status_code=201)
    async def create_telegram(
        payload: TelegramAdapterCreate,
        response: Response,
        actor: AssistantIdentity = Depends(require_action("assistant:channels")),
        x_mypocket_delegation: str | None = Header(default=None),
        x_mypocket_delegation_id: str | None = Header(default=None),
    ) -> ChannelAdapterView:
        if services.adapters is None:
            raise HTTPException(status_code=503, detail="channel adapters unavailable")
        if services.credentials is None or not x_mypocket_delegation or not x_mypocket_delegation_id:
            raise HTTPException(status_code=503, detail="delegated backend credential unavailable")
        try:
            adapter, _ = await services.adapters.create_telegram(actor.owner_id, payload.label, payload.bot_token)
            try:
                previous_id = await services.credentials.put(actor.owner_id, x_mypocket_delegation_id, x_mypocket_delegation)
            except Exception:
                await services.adapters.revoke_telegram(actor.owner_id, adapter.id)
                raise
        except (ValueError, httpx.HTTPError) as exc:
            raise HTTPException(status_code=400, detail="invalid Telegram bot credential") from exc
        if previous_id:
            response.headers["X-MyPocket-Previous-Delegation-ID"] = previous_id
        return ChannelAdapterView(id=adapter.id, provider="telegram", label=adapter.label, external_identity=adapter.external_identity, status="active", created_at=adapter.created_at.isoformat())

    @app.get("/v1/channels", response_model=list[ChannelAdapterView])
    async def list_channels(actor: AssistantIdentity = Depends(require_action("assistant:channels"))) -> list[ChannelAdapterView]:
        if services.adapters is None:
            return []
        rows = await services.adapters.repository.list_owner(actor.owner_id)
        return [ChannelAdapterView(id=row.id, provider="telegram", label=row.label, external_identity=row.external_identity, status=row.status, created_at=row.created_at.isoformat()) for row in rows]

    @app.delete("/v1/channels/{adapter_id}", response_model=ChannelAdapterRevokeResponse)
    async def revoke_channel(
        adapter_id: str,
        response: Response,
        actor: AssistantIdentity = Depends(require_action("assistant:channels")),
    ) -> ChannelAdapterRevokeResponse:
        if services.adapters is None or services.credentials is None:
            raise HTTPException(status_code=503, detail="channel adapters unavailable")
        try:
            revoked = await services.adapters.revoke_telegram(actor.owner_id, adapter_id)
        except httpx.HTTPError as exc:
            raise HTTPException(status_code=502, detail="Telegram webhook revocation failed") from exc
        if not revoked:
            raise HTTPException(status_code=404, detail="channel adapter not found")
        await services.pairings.revoke_adapter(actor.owner_id, adapter_id)
        if await services.adapters.repository.count_active(actor.owner_id) == 0:
            credential_id = await services.credentials.revoke(actor.owner_id)
            if credential_id:
                response.headers["X-MyPocket-Revoke-Delegation-ID"] = credential_id
        return ChannelAdapterRevokeResponse(revoked=True)

    @app.post("/v1/channels/telegram/{adapter_id}/webhook", status_code=202)
    async def telegram_webhook(
        adapter_id: str,
        request: Request,
        x_telegram_bot_api_secret_token: str | None = Header(default=None),
    ) -> dict[str, bool]:
        if services.adapters is None or not x_telegram_bot_api_secret_token:
            raise HTTPException(status_code=401, detail="invalid webhook secret")
        try:
            adapter = await services.adapters.authenticate_webhook(adapter_id, x_telegram_bot_api_secret_token)
        except PermissionError as exc:
            raise HTTPException(status_code=401, detail="invalid webhook secret") from exc
        if services.telegram is None:
            raise HTTPException(status_code=503, detail="telegram disabled")
        await services.telegram.enqueue(adapter.id, adapter.owner_id, await request.json())
        return {"accepted": True}

    # Compatibility facade consumed only through the authenticated MyPocket
    # backend proxy. It keeps native DTOs stable while the service owns state.
    @app.get("/api/v1/ai/advisor/capabilities")
    async def compat_capabilities(actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        tools = [tool["function"]["name"] for tool in FINANCE_READ_TOOLS]
        return {"data": {"enabled": True, "write": False, "tools": tools, "direct_writes": False, "proposal_actions": False, "confirmation_required": True, "parts_version": 1, "run_timeout_seconds": int(services.runtime.run_timeout_seconds), "grounding_version": 1, "entry_integration": False, "memory": True, "channel_adapters": True}}

    @app.get("/api/v1/ai/advisor/conversation/messages")
    async def compat_messages(conversation_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        rows = await services.runtime.persistence.list_messages(actor.owner_id, "app", conversation_id)
        data = []
        for index, row in enumerate(rows, 1):
            parts = row.parts if row.parts else [{"type": "text", "text": row.text}]
            data.append({"id": f"external-{index}", "conversation_id": conversation_id, "generation": 1, "seq": index, "run_id": "external", "role": row.role, "parts_version": 1, "parts": parts, "created_at": row.created_at.isoformat()})
        return {"data": data, "meta": {"conversation_id": conversation_id, "generation": 1, "server_time": datetime.now(UTC).isoformat()}}

    @app.post("/api/v1/ai/advisor/conversation/reset")
    async def compat_reset(payload: dict | None = None, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        conversation_id = str((payload or {}).get("conversation_id", "default"))
        if payload and payload.get("conversation_id"):
            await services.runtime.persistence.reset_messages(actor.owner_id, "app", conversation_id)
        else:
            await services.runtime.persistence.reset_channel(actor.owner_id, "app")
        return {"data": {"conversation_id": conversation_id, "generation": 1, "server_time": datetime.now(UTC).isoformat(), "replayed": False}}

    def compat_request(payload: dict) -> TurnRequest:
        text = str(payload.get("text", "")).strip()
        if not text:
            raise HTTPException(status_code=400, detail="text is required")
        conversation_id = str(payload.get("conversation_id", "default"))
        request_id = str(payload.get("client_request_id") or uuid.uuid4())
        return TurnRequest(conversation_id=conversation_id, channel_id="app", platform_message_id=request_id, text=text)

    async def compat_submit(payload: dict, actor: AssistantIdentity, delegated_credential: str | None) -> TurnResponse:
        return await services.runtime.run_turn(actor.owner_id, compat_request(payload), delegated_credential=delegated_credential)

    @app.post("/api/v1/ai/advisor/messages")
    async def compat_submit_sync(payload: dict, actor: AssistantIdentity = Depends(require_action("assistant:chat")), x_mypocket_delegation: str | None = Header(default=None)):
        result = await compat_submit(payload, actor, x_mypocket_delegation)
        return {"data": result.model_dump()}

    @app.post("/api/v1/ai/advisor/messages/async", status_code=202)
    async def compat_submit_async(payload: dict, actor: AssistantIdentity = Depends(require_action("assistant:chat")), x_mypocket_delegation: str | None = Header(default=None)):
        if services.runs is None or not x_mypocket_delegation:
            raise HTTPException(status_code=503, detail="durable assistant runs unavailable")
        try:
            result = await services.runs.enqueue(actor.owner_id, compat_request(payload), x_mypocket_delegation)
        except ValueError as exc:
            if "conflict" in str(exc):
                raise HTTPException(status_code=409, detail="client_request_id conflict") from exc
            raise HTTPException(status_code=400, detail="invalid assistant request") from exc
        return {"data": {"run_id": result["run_id"], "conversation_id": result["conversation_id"], "status": result["status"]}}

    async def compat_run_payload(owner_id: str, run_id: str) -> dict | None:
        if services.runs is not None:
            return await services.runs.get(owner_id, run_id)
        return await services.runtime.persistence.get_run(owner_id, run_id)

    @app.get("/api/v1/ai/advisor/runs/{run_id}")
    async def compat_run(run_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        result = await compat_run_payload(actor.owner_id, run_id)
        if result is None:
            raise HTTPException(status_code=404, detail="run not found")
        return {"data": {"id": result["run_id"], "conversation_id": result["conversation_id"], "generation": 1, "status": result["status"], "error_code": result.get("error_code"), "last_event_seq": 1 if result["status"] in TERMINAL_RUN_STATUSES else 0, "created_at": result.get("created_at"), "updated_at": result.get("updated_at")}}

    @app.get("/api/v1/ai/advisor/runs/{run_id}/events")
    async def compat_events(run_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        result = await compat_run_payload(actor.owner_id, run_id)
        if result is None:
            raise HTTPException(status_code=404, detail="run not found")
        events = []
        if result["status"] in TERMINAL_RUN_STATUSES:
            events.append({"seq": 1, "run_id": run_id, "type": result["status"], "data": result, "created_at": result.get("updated_at") or datetime.now(UTC).isoformat()})
        return {"data": events}

    @app.post("/api/v1/ai/advisor/runs/{run_id}/cancel")
    async def compat_cancel(run_id: str, actor: AssistantIdentity = Depends(require_action("assistant:chat"))):
        if services.runs is None:
            raise HTTPException(status_code=503, detail="durable assistant runs unavailable")
        result = await services.runs.cancel(actor.owner_id, run_id)
        if result is None:
            raise HTTPException(status_code=404, detail="run not found")
        return {"data": {"id": run_id, "status": result["status"]}}

    return app
