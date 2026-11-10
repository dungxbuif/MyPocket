from functools import lru_cache
from pathlib import Path
import re
from urllib.parse import parse_qsl, urlencode, urlparse, urlsplit, urlunsplit

from pydantic import Field, SecretStr, model_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_prefix="FINANCE_ASSISTANT_", env_file=".env", extra="ignore")

    environment: str = "development"
    database_url: str = "postgresql://finance_assistant:finance_assistant@postgres:5432/finance_assistant"
    database_schema: str = "public"
    data_root: Path = Path("/data")
    assertion_secret: SecretStr = SecretStr("")
    assertion_secret_file: Path | None = None
    mypocket_base_url: str = "http://mypocket-api:8080"
    model_base_url: str = "http://127.0.0.1:1234/v1"
    model_api_key: SecretStr = SecretStr("")
    model_api_key_file: Path | None = None
    model_name: str = "Qwen3.6-35B-A3B-MLX-4bit"
    model_max_tokens: int = Field(default=4096, ge=256, le=32768)
    model_timeout_seconds: float = Field(default=240, ge=5, le=600)
    credential_encryption_key: SecretStr = SecretStr("")
    credential_encryption_key_file: Path | None = None
    database_url_file: Path | None = None
    public_base_url: str = ""
    max_context_messages: int = Field(default=30, ge=4, le=100)
    memory_recall_limit: int = Field(default=5, ge=1, le=20)

    @model_validator(mode="after")
    def load_secret_files(self):
        for value_name, file_name in (
            ("assertion_secret", "assertion_secret_file"),
            ("model_api_key", "model_api_key_file"),
            ("credential_encryption_key", "credential_encryption_key_file"),
        ):
            path = getattr(self, file_name)
            if path:
                setattr(self, value_name, SecretStr(path.read_text(encoding="utf-8").strip()))
        if self.database_url_file:
            self.database_url = self.database_url_file.read_text(encoding="utf-8").strip()
        if not re.fullmatch(r"[a-z_][a-z0-9_]*", self.database_schema):
            raise ValueError("database schema must be a lowercase PostgreSQL identifier")
        database = urlsplit(self.database_url)
        query = [
            (key, value)
            for key, value in parse_qsl(database.query, keep_blank_values=True)
            if key != "options"
        ]
        query.append(("options", f"-csearch_path={self.database_schema}"))
        self.database_url = urlunsplit(
            (
                database.scheme,
                database.netloc,
                database.path,
                urlencode(query),
                database.fragment,
            )
        )
        if len(self.assertion_secret.get_secret_value()) < 32:
            raise ValueError("assertion secret must contain at least 32 characters")
        parsed = urlparse(self.model_base_url)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc or parsed.username or parsed.password or parsed.query or parsed.fragment:
            raise ValueError("model base URL must be an absolute HTTP(S) URL without credentials, query, or fragment")
        self.model_base_url = self.model_base_url.rstrip("/")
        if not self.model_name.strip():
            raise ValueError("model name is required")
        return self


@lru_cache
def get_settings() -> Settings:
    return Settings()  # type: ignore[call-arg]
