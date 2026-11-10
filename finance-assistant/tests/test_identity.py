from datetime import UTC, datetime, timedelta

import jwt
import pytest

from finance_assistant.identity import AssertionErrorCode, AssertionVerifier


SECRET = "test-finance-assistant-shared-secret-at-least-32-bytes"


def token(**overrides: object) -> str:
    now = datetime.now(UTC)
    claims = {
        "iss": "mypocket",
        "aud": "finance-assistant",
        "sub": "owner-a",
        "iat": now,
        "exp": now + timedelta(minutes=2),
        "jti": "request-1",
        "actions": ["assistant:chat"],
    }
    claims.update(overrides)
    return jwt.encode(claims, SECRET, algorithm="HS256")


def test_valid_assertion_resolves_owner_without_client_owner_input() -> None:
    identity = AssertionVerifier(SECRET).verify(token())
    assert identity.owner_id == "owner-a"
    assert identity.request_id == "request-1"


@pytest.mark.parametrize(
    ("value", "code"),
    [
        (lambda: token(aud="other"), "invalid_audience"),
        (lambda: token(exp=datetime.now(UTC) - timedelta(seconds=1)), "expired"),
        (lambda: token(sub=""), "invalid_subject"),
    ],
)
def test_invalid_assertions_fail_closed(value, code: str) -> None:
    with pytest.raises(AssertionErrorCode) as caught:
        AssertionVerifier(SECRET).verify(value())
    assert caught.value.code == code
