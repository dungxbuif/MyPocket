CREATE TABLE assistant_delegations (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lookup_id text NOT NULL UNIQUE,
    secret_hash text NOT NULL,
    allowed_operations jsonb NOT NULL,
    expires_at timestamptz,
    revoked_at timestamptz,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX assistant_delegations_owner_active ON assistant_delegations(owner_id, created_at DESC) WHERE revoked_at IS NULL;
