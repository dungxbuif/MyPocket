-- On-demand grounded analysis jobs. Facts are reconstructed from an immutable
-- snapshot; this table never stores ledger rows, notes, images, or prompts.
CREATE TABLE report_analyses (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    snapshot_id text NOT NULL REFERENCES report_snapshots(id) ON DELETE CASCADE,
    job_key text NOT NULL,
    client_request_id text NOT NULL,
    payload_hash text NOT NULL CHECK (length(trim(payload_hash)) > 0),
    revision bigint NOT NULL CHECK (revision > 0),
    rule_version text NOT NULL,
    prompt_version text NOT NULL,
    provider_profile_version text NOT NULL,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled')),
    error_code text NOT NULL DEFAULT '',
    provider_called boolean NOT NULL DEFAULT false,
    summary text NOT NULL DEFAULT '',
    observations jsonb NOT NULL DEFAULT '[]'::jsonb,
    suggestions jsonb NOT NULL DEFAULT '[]'::jsonb,
    fact_refs jsonb NOT NULL DEFAULT '[]'::jsonb,
    usage jsonb NOT NULL DEFAULT '{}'::jsonb,
    lease_token text NOT NULL DEFAULT '',
    lease_until timestamptz,
    deadline_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    finished_at timestamptz,
    UNIQUE(owner_id, job_key),
    UNIQUE(owner_id, client_request_id)
);
CREATE INDEX report_analyses_worker_idx ON report_analyses(status, lease_until, deadline_at, created_at);
CREATE INDEX report_analyses_retention_idx ON report_analyses(created_at, status);
