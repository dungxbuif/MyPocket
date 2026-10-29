CREATE TABLE device_credentials (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    device_id varchar(128) NOT NULL,
    device_name varchar(100) NOT NULL,
    secret_hash char(64) NOT NULL,
    password_fingerprint char(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (owner_id, device_id)
);
CREATE INDEX device_credentials_owner_idx ON device_credentials(owner_id);
