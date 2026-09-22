CREATE TABLE advisor_conversations (
    id text PRIMARY KEY,
    owner_id text NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,
    generation bigint NOT NULL DEFAULT 1 CHECK (generation > 0),
    next_message_seq bigint NOT NULL DEFAULT 1 CHECK (next_message_seq > 0),
    summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    summary_through_seq bigint NOT NULL DEFAULT 0 CHECK (summary_through_seq >= 0),
    summary_version bigint NOT NULL DEFAULT 0 CHECK (summary_version >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, owner_id)
);

CREATE TABLE advisor_runs (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    conversation_id text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    client_request_id text NOT NULL,
    payload_hash text NOT NULL,
    credential_kind text NOT NULL CHECK (credential_kind IN ('session', 'user_api_key')),
    credential_id text NOT NULL,
    credential_expires_at timestamptz,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled', 'interrupted', 'purged')),
    lease_token text,
    lease_until timestamptz,
    last_event_seq bigint NOT NULL DEFAULT 0 CHECK (last_event_seq >= 0),
    error_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    FOREIGN KEY (conversation_id, owner_id) REFERENCES advisor_conversations(id, owner_id) ON DELETE CASCADE,
    UNIQUE (owner_id, client_request_id)
);

CREATE UNIQUE INDEX advisor_one_active_run ON advisor_runs(conversation_id)
    WHERE status IN ('queued', 'running');
CREATE INDEX advisor_run_lease ON advisor_runs(lease_until)
    WHERE status = 'running';

CREATE TABLE advisor_messages (
    id text PRIMARY KEY,
    conversation_id text NOT NULL REFERENCES advisor_conversations(id) ON DELETE CASCADE,
    generation bigint NOT NULL,
    seq bigint NOT NULL CHECK (seq > 0),
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('user', 'assistant')),
    parts_version integer NOT NULL DEFAULT 1 CHECK (parts_version = 1),
    parts jsonb NOT NULL CHECK (jsonb_typeof(parts) = 'array'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (conversation_id, seq),
    UNIQUE (run_id, role)
);
CREATE INDEX advisor_messages_history ON advisor_messages(conversation_id, seq DESC);

CREATE TABLE advisor_events (
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    seq bigint NOT NULL CHECK (seq > 0),
    type text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, seq)
);

CREATE TABLE advisor_fact_bundles (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    scope jsonb NOT NULL,
    as_of timestamptz NOT NULL,
    facts jsonb NOT NULL CHECK (jsonb_typeof(facts) = 'array'),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX advisor_fact_bundles_run ON advisor_fact_bundles(run_id, created_at DESC);
