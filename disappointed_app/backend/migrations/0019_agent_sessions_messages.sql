CREATE TABLE IF NOT EXISTS agent_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('intake', 'advisor')),
    title text NOT NULL DEFAULT '',
    context_summary text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, kind)
);

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS session_id uuid REFERENCES agent_sessions(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS agent_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id uuid NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    run_id uuid REFERENCES agent_runs(id) ON DELETE SET NULL,
    role text NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    text text NOT NULL DEFAULT '',
    action jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS agent_messages_session_created_idx
    ON agent_messages (session_id, created_at, id);

CREATE UNIQUE INDEX IF NOT EXISTS agent_messages_user_run_role_idx
    ON agent_messages (user_id, run_id, role)
    WHERE run_id IS NOT NULL;
