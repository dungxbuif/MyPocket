CREATE TABLE IF NOT EXISTS inbound_events (
    id bigserial PRIMARY KEY,
    owner_id text NOT NULL,
    channel_id text NOT NULL,
    platform_message_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(owner_id, channel_id, platform_message_id)
);

CREATE TABLE IF NOT EXISTS conversation_messages (
    seq bigserial PRIMARY KEY,
    owner_id text NOT NULL,
    channel_id text NOT NULL,
    conversation_id text NOT NULL,
    role text NOT NULL CHECK (role IN ('user','assistant','tool')),
    text text NOT NULL,
    parts jsonb NOT NULL DEFAULT '[]'::jsonb,
    platform_message_id text,
    created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS parts jsonb NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS platform_message_id text;
CREATE INDEX IF NOT EXISTS conversation_messages_scope ON conversation_messages(owner_id, channel_id, conversation_id, seq DESC);
CREATE UNIQUE INDEX IF NOT EXISTS conversation_messages_event_role
    ON conversation_messages(owner_id,channel_id,platform_message_id,role)
    WHERE platform_message_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS agent_runs (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS agent_runs_owner ON agent_runs(owner_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS assistant_run_jobs (
    run_id uuid PRIMARY KEY REFERENCES agent_runs(id) ON DELETE CASCADE,
    owner_id text NOT NULL,
    client_request_id text NOT NULL,
    request_hash text NOT NULL,
    request jsonb NOT NULL,
    encrypted_delegation bytea,
    attempts integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_until timestamptz,
    cancel_requested boolean NOT NULL DEFAULT false,
    completed_at timestamptz,
    failed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(owner_id, client_request_id)
);
CREATE INDEX IF NOT EXISTS assistant_run_jobs_ready ON assistant_run_jobs(available_at,created_at)
    WHERE completed_at IS NULL AND failed_at IS NULL AND cancel_requested=false;

CREATE TABLE IF NOT EXISTS memories (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    kind text NOT NULL,
    content text NOT NULL,
    source_type text NOT NULL,
    source_id text NOT NULL,
    confidence double precision NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    superseded_by uuid REFERENCES memories(id),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS memories_owner_active ON memories(owner_id, created_at DESC) WHERE superseded_by IS NULL;

CREATE TABLE IF NOT EXISTS channel_adapters (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    provider text NOT NULL CHECK (provider IN ('telegram')),
    label text NOT NULL,
    external_identity text NOT NULL,
    encrypted_credential bytea NOT NULL,
    webhook_secret_hash text NOT NULL,
    status text NOT NULL CHECK (status IN ('active','revoked')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(owner_id, provider, label)
);
CREATE INDEX IF NOT EXISTS channel_adapters_owner ON channel_adapters(owner_id, created_at, id);

CREATE TABLE IF NOT EXISTS telegram_pairing_intents (
    code_hash text PRIMARY KEY,
    owner_id text NOT NULL,
    adapter_id uuid NOT NULL REFERENCES channel_adapters(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS telegram_bindings (
    adapter_id uuid NOT NULL REFERENCES channel_adapters(id) ON DELETE CASCADE,
    telegram_user_id bigint NOT NULL,
    telegram_chat_id bigint NOT NULL,
    owner_id text NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY(adapter_id, telegram_user_id),
    UNIQUE(adapter_id, telegram_chat_id)
);

CREATE TABLE IF NOT EXISTS telegram_inbound_jobs (
    id bigserial PRIMARY KEY,
    adapter_id uuid NOT NULL REFERENCES channel_adapters(id) ON DELETE CASCADE,
    owner_id text NOT NULL,
    update_id text NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_until timestamptz,
    completed_at timestamptz,
    failed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(adapter_id, update_id)
);
CREATE INDEX IF NOT EXISTS telegram_inbound_jobs_ready ON telegram_inbound_jobs(available_at,id)
    WHERE completed_at IS NULL AND failed_at IS NULL;

CREATE TABLE IF NOT EXISTS delegated_credentials (
    owner_id text PRIMARY KEY,
    credential_id text NOT NULL,
    encrypted_secret bytea NOT NULL,
    expires_at timestamptz,
    revoked_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS action_drafts (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    operation text NOT NULL,
    payload jsonb NOT NULL,
    payload_hash text NOT NULL,
    revision bigint NOT NULL,
    status text NOT NULL,
    idempotency_key text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    receipt jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS action_drafts_owner ON action_drafts(owner_id, created_at DESC);

CREATE TABLE IF NOT EXISTS delivery_outbox (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    channel_id text NOT NULL,
    destination text NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL DEFAULT now(),
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dream_jobs (
    owner_id text PRIMARY KEY,
    available_at timestamptz NOT NULL,
    lease_owner text,
    lease_until timestamptz,
    attempts integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dream_candidates (
    id uuid PRIMARY KEY,
    owner_id text NOT NULL,
    content text NOT NULL,
    content_hash text NOT NULL,
    source_refs jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','promoted','rejected')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(owner_id, content_hash)
);
CREATE INDEX IF NOT EXISTS dream_candidates_owner ON dream_candidates(owner_id, status, updated_at DESC);
