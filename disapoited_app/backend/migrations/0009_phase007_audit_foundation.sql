CREATE TABLE IF NOT EXISTS audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    correlation_id text NOT NULL CHECK (length(trim(correlation_id)) > 0),
    actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    actor_email_hash text NOT NULL DEFAULT '',
    action text NOT NULL CHECK (length(trim(action)) > 0),
    entity_type text NOT NULL DEFAULT '',
    entity_id text NOT NULL DEFAULT '',
    outcome text NOT NULL CHECK (outcome IN ('success', 'failure', 'denied', 'conflict', 'replayed')),
    severity text NOT NULL CHECK (severity IN ('info', 'warn', 'error', 'security')),
    source text NOT NULL CHECK (source IN ('web', 'pwa', 'api', 'worker', 'sync', 'auth', 'system')),
    error_code text NOT NULL DEFAULT '',
    request_method text NOT NULL DEFAULT '',
    request_path text NOT NULL DEFAULT '',
    ip_hash text NOT NULL DEFAULT '',
    user_agent_hash text NOT NULL DEFAULT '',
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_events_occurred_at_idx
    ON audit_events (occurred_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS audit_events_actor_idx
    ON audit_events (actor_user_id, occurred_at DESC)
    WHERE actor_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS audit_events_correlation_idx
    ON audit_events (correlation_id);

CREATE INDEX IF NOT EXISTS audit_events_action_idx
    ON audit_events (action, occurred_at DESC);

CREATE INDEX IF NOT EXISTS audit_events_severity_outcome_idx
    ON audit_events (severity, outcome, occurred_at DESC);
