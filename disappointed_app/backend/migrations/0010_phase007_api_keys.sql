CREATE TABLE IF NOT EXISTS api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 80),
    key_prefix text NOT NULL CHECK (length(trim(key_prefix)) >= 8),
    key_hash text NOT NULL UNIQUE CHECK (length(trim(key_hash)) > 0),
    last_used_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS api_keys_user_idx
    ON api_keys (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS api_keys_active_hash_idx
    ON api_keys (key_hash)
    WHERE revoked_at IS NULL;
