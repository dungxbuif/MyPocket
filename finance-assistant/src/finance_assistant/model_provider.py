from langchain_openai import ChatOpenAI

from .config import Settings


def build_chat_model(settings: Settings) -> ChatOpenAI:
    """Build the same OpenAI-compatible provider boundary used by MyPocket."""
    api_key = settings.model_api_key.get_secret_value()
    if not api_key:
        raise RuntimeError("FINANCE_ASSISTANT_MODEL_API_KEY is required")
    return ChatOpenAI(
        model=settings.model_name,
        api_key=api_key,
        base_url=settings.model_base_url,
        temperature=0,
        max_tokens=settings.model_max_tokens,
        timeout=settings.model_timeout_seconds,
        max_retries=0,
    )
