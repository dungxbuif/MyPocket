-- Additive entry task lifecycle. Existing advisor runs, entry sessions,
-- proposals and financial rows are not rewritten.
CREATE TABLE advisor_entry_tasks (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    conversation_id text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    client_request_id text NOT NULL CHECK (length(trim(client_request_id)) > 0),
    payload_hash text NOT NULL CHECK (length(trim(payload_hash)) > 0),
    session_id text NOT NULL,
    status text NOT NULL CHECK (status IN ('accepted', 'processing', 'review_ready', 'failed', 'cancelled')),
    stage text NOT NULL DEFAULT '',
    batch_id text,
    batch_version bigint CHECK (batch_version > 0),
    error_code text,
    lease_token text,
    lease_until timestamptz,
    deadline_at timestamptz NOT NULL,
    last_event_seq bigint NOT NULL DEFAULT 0 CHECK (last_event_seq >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    FOREIGN KEY (conversation_id, owner_id) REFERENCES advisor_conversations(id, owner_id) ON DELETE CASCADE,
    FOREIGN KEY (session_id, owner_id) REFERENCES ai_entry_sessions(id, owner_id) ON DELETE CASCADE,
    UNIQUE (id, owner_id),
    UNIQUE (owner_id, conversation_id, generation, client_request_id),
    CHECK ((batch_id IS NULL AND batch_version IS NULL) OR
           (batch_id IS NOT NULL AND length(trim(batch_id)) > 0 AND batch_version IS NOT NULL AND batch_version > 0)),
    CHECK (deadline_at >= created_at)
);

CREATE INDEX advisor_entry_tasks_worker_idx
    ON advisor_entry_tasks(status, lease_until, deadline_at)
    WHERE status IN ('accepted', 'processing');
CREATE INDEX advisor_entry_tasks_owner_recent_idx
    ON advisor_entry_tasks(owner_id, conversation_id, generation, created_at DESC);

CREATE TABLE advisor_entry_events (
    task_id text NOT NULL,
    owner_id text NOT NULL,
    seq bigint NOT NULL CHECK (seq > 0),
    type text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, seq),
    FOREIGN KEY (task_id, owner_id) REFERENCES advisor_entry_tasks(id, owner_id) ON DELETE CASCADE
);

CREATE TABLE advisor_entry_attachment_bindings (
    task_id text NOT NULL,
    owner_id text NOT NULL,
    attachment_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, attachment_id),
    FOREIGN KEY (task_id, owner_id) REFERENCES advisor_entry_tasks(id, owner_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id, owner_id) REFERENCES transaction_attachments(id, owner_id) ON DELETE CASCADE
);
CREATE INDEX advisor_entry_attachment_bindings_attachment_idx
    ON advisor_entry_attachment_bindings(attachment_id);
