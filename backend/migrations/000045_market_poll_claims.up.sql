CREATE TABLE IF NOT EXISTS market_poll_claims (
  provider_key text NOT NULL,
  slot_at timestamptz NOT NULL,
  lease_token text NOT NULL,
  lease_until timestamptz NOT NULL,
  status text NOT NULL,
  error_message text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (provider_key, slot_at),
  CONSTRAINT market_poll_claims_status_check CHECK (status IN ('running', 'completed', 'failed')),
  CONSTRAINT market_poll_claims_provider_check CHECK (char_length(trim(provider_key)) > 0),
  CONSTRAINT market_poll_claims_token_check CHECK (char_length(trim(lease_token)) > 0)
);
CREATE INDEX IF NOT EXISTS market_poll_claims_lease_idx ON market_poll_claims(status, lease_until);
