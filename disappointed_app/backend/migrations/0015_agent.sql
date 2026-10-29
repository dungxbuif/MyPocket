CREATE TABLE IF NOT EXISTS agent_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idempotency_key text NOT NULL CHECK (length(trim(idempotency_key)) BETWEEN 1 AND 200),
    kind text NOT NULL CHECK (kind IN ('transaction_draft', 'analysis')),
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'completed', 'failed')),
    request_text text NOT NULL CHECK (length(trim(request_text)) BETWEEN 1 AND 8000),
    response_text text,
    error_code text,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS agent_runs_due_idx ON agent_runs (next_attempt_at, id)
    WHERE status IN ('queued', 'processing');

ALTER TABLE transaction_drafts
    ADD COLUMN IF NOT EXISTS agent_run_id uuid REFERENCES agent_runs(id) ON DELETE SET NULL;
ALTER TABLE transaction_drafts
    ADD COLUMN IF NOT EXISTS provenance jsonb NOT NULL DEFAULT '{}'::jsonb;
CREATE UNIQUE INDEX IF NOT EXISTS transaction_drafts_agent_run_idx
    ON transaction_drafts (agent_run_id) WHERE agent_run_id IS NOT NULL;
