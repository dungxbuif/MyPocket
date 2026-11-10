from __future__ import annotations

from dataclasses import dataclass

import jwt


class AssertionErrorCode(ValueError):
    def __init__(self, code: str) -> None:
        super().__init__(code)
        self.code = code


@dataclass(frozen=True, slots=True)
class AssistantIdentity:
    owner_id: str
    request_id: str
    actions: frozenset[str]
    credential_id: str | None = None


class AssertionVerifier:
    def __init__(
        self,
        secret: str,
        *,
        issuer: str = "mypocket",
        audience: str = "finance-assistant",
    ) -> None:
        if len(secret) < 32:
            raise ValueError("assertion secret must contain at least 32 characters")
        self.secret = secret
        self.issuer = issuer
        self.audience = audience

    def verify(self, token: str) -> AssistantIdentity:
        try:
            claims = jwt.decode(
                token,
                self.secret,
                algorithms=["HS256"],
                audience=self.audience,
                issuer=self.issuer,
                options={"require": ["iss", "aud", "sub", "exp", "iat", "jti"]},
            )
        except jwt.ExpiredSignatureError as exc:
            raise AssertionErrorCode("expired") from exc
        except jwt.InvalidAudienceError as exc:
            raise AssertionErrorCode("invalid_audience") from exc
        except jwt.InvalidIssuerError as exc:
            raise AssertionErrorCode("invalid_issuer") from exc
        except jwt.PyJWTError as exc:
            raise AssertionErrorCode("invalid_assertion") from exc

        owner_id = str(claims.get("sub", "")).strip()
        request_id = str(claims.get("jti", "")).strip()
        raw_actions = claims.get("actions", [])
        if not owner_id:
            raise AssertionErrorCode("invalid_subject")
        if not request_id or not isinstance(raw_actions, list):
            raise AssertionErrorCode("invalid_assertion")
        return AssistantIdentity(
            owner_id=owner_id,
            request_id=request_id,
            actions=frozenset(str(action) for action in raw_actions),
            credential_id=str(claims["credential_id"]) if claims.get("credential_id") else None,
        )
