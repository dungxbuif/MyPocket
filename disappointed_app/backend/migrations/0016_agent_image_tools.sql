CREATE TABLE IF NOT EXISTS agent_tool_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_run_id uuid NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    receipt_object_id uuid REFERENCES receipt_objects(id) ON DELETE SET NULL,
    provider text NOT NULL CHECK (provider = 'ocr'),
    provider_document_id text,
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','submitting','processing','completed','failed','cancelled','expired')),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    result jsonb,
    result_expires_at timestamptz,
    error_code text,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (agent_run_id, receipt_object_id)
);
CREATE INDEX IF NOT EXISTS agent_tool_runs_due_idx ON agent_tool_runs (next_attempt_at, id)
    WHERE status IN ('queued','submitting','processing');
