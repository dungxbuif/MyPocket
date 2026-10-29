ALTER TABLE users ADD COLUMN disabled_at timestamptz;

CREATE TABLE data_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('import','export','reset','delete')),
    status text NOT NULL CHECK (status IN ('queued','running','awaiting_confirmation','completed','failed')),
    idempotency_key text NOT NULL,
    request jsonb NOT NULL DEFAULT '{}'::jsonb,
    result jsonb,
    result_object_key text,
    attempts integer NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    error_code text,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, kind, idempotency_key)
);

CREATE INDEX data_jobs_due_idx ON data_jobs (next_attempt_at, created_at)
    WHERE status IN ('queued','running');
