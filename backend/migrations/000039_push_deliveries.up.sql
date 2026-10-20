CREATE TABLE push_deliveries (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    notification_id text NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    push_token_id uuid NOT NULL REFERENCES push_tokens(id) ON DELETE CASCADE,
    environment varchar(16) NOT NULL,
    collapse_id varchar(128) NOT NULL,
    payload jsonb NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'pending',
    attempts integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL,
    lease_until timestamptz,
    last_error text,
    delivered_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT push_deliveries_environment_check CHECK (environment IN ('sandbox', 'production')),
    CONSTRAINT push_deliveries_status_check CHECK (status IN ('pending', 'retry', 'sent', 'revoked', 'failed')),
    CONSTRAINT push_deliveries_attempts_check CHECK (attempts >= 0)
);
CREATE UNIQUE INDEX push_deliveries_dedupe_idx
  ON push_deliveries(notification_id, push_token_id, environment);
CREATE INDEX push_deliveries_queue_idx
  ON push_deliveries(status, available_at, lease_until)
  WHERE status IN ('pending', 'retry');
CREATE INDEX push_deliveries_owner_idx ON push_deliveries(owner_id, created_at DESC);
