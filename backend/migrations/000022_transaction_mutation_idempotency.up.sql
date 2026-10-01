CREATE TABLE IF NOT EXISTS transaction_mutations (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  operation text NOT NULL,
  idempotency_key text NOT NULL,
  request_hash text NOT NULL,
  status text NOT NULL,
  response_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT transaction_mutations_status_check CHECK (status IN ('completed', 'pending')),
  CONSTRAINT transaction_mutations_owner_operation_key_unique UNIQUE (owner_id, operation, idempotency_key)
);
CREATE INDEX IF NOT EXISTS transaction_mutations_created_at_idx ON transaction_mutations(created_at);
