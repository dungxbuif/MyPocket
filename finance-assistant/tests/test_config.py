from cryptography.fernet import Fernet
from httpx import Response
import json
from langchain_core.messages import HumanMessage
import pytest
import respx

from finance_assistant.config import Settings
from finance_assistant.model_provider import build_chat_model


def settings(**overrides) -> Settings:
    values = {
        "assertion_secret": "a" * 32,
        "credential_encryption_key": Fernet.generate_key().decode(),
        "model_api_key": "provider-key",
    }
    values.update(overrides)
    return Settings(**values)


def test_model_provider_matches_existing_openai_compatible_profile() -> None:
    value = settings(model_base_url="https://token-api.fpt.ai/v1/", model_name="DeepSeek-V4.1-Flash")
    model = build_chat_model(value)
    assert value.model_base_url == "https://token-api.fpt.ai/v1"
    assert model.model_name == "DeepSeek-V4.1-Flash"
    assert model.max_tokens == 4096
    assert model.max_retries == 0


@pytest.mark.parametrize("url", ["file:///tmp/model", "https://user:pass@example.com/v1", "https://example.com/v1?q=secret"])
def test_model_provider_rejects_unsafe_base_urls(url: str) -> None:
    with pytest.raises(ValueError, match="model base URL"):
        settings(model_base_url=url)


@pytest.mark.asyncio
@respx.mock
async def test_model_provider_calls_chat_completions_with_selected_model_and_bound() -> None:
    route = respx.post("https://token-api.fpt.ai/v1/chat/completions").mock(
        return_value=Response(
            200,
            json={
                "id": "chat-1",
                "object": "chat.completion",
                "created": 1,
                "model": "DeepSeek-V4.1-Flash",
                "choices": [{"index": 0, "message": {"role": "assistant", "content": "ok"}, "finish_reason": "stop"}],
                "usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
            },
        )
    )
    model = build_chat_model(settings(model_base_url="https://token-api.fpt.ai/v1", model_name="DeepSeek-V4.1-Flash"))
    response = await model.ainvoke([HumanMessage(content="hello")])
    request = route.calls.last.request
    payload = json.loads(request.content)

    assert response.content == "ok"
    assert payload["model"] == "DeepSeek-V4.1-Flash"
    assert payload["max_completion_tokens"] == 4096
    assert request.headers["authorization"] == "Bearer provider-key"
