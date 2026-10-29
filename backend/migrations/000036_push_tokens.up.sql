CREATE TABLE push_tokens (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    device_id varchar(128) NOT NULL,
    device_name varchar(100) NOT NULL,
    platform varchar(16) NOT NULL,
    environment varchar(16) NOT NULL,
    token text NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT push_tokens_platform_check CHECK (platform = 'ios'),
    CONSTRAINT push_tokens_environment_check CHECK (environment IN ('sandbox', 'production'))
);
CREATE UNIQUE INDEX push_tokens_owner_device_idx
  ON push_tokens(owner_id, device_id, platform, environment);
CREATE UNIQUE INDEX push_tokens_provider_address_idx
  ON push_tokens(platform, environment, token);
CREATE INDEX push_tokens_owner_idx ON push_tokens(owner_id, updated_at DESC);
