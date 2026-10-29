CREATE TABLE IF NOT EXISTS sync_cursors (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    next_cursor bigint NOT NULL DEFAULT 1 CHECK (next_cursor > 0)
);

CREATE TABLE IF NOT EXISTS sync_changes (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cursor bigint NOT NULL CHECK (cursor > 0),
    entity_type text NOT NULL CHECK (entity_type IN ('wallet', 'category', 'transaction')),
    entity_id text NOT NULL CHECK (length(trim(entity_id)) > 0),
    operation text NOT NULL CHECK (operation IN ('create', 'update', 'archive', 'set_default_ai', 'set_category_active')),
    version bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    payload_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, cursor)
);

CREATE INDEX IF NOT EXISTS sync_changes_user_entity_idx
    ON sync_changes (user_id, entity_type, entity_id, cursor DESC);

CREATE TABLE IF NOT EXISTS sync_mutations (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mutation_id text NOT NULL CHECK (length(trim(mutation_id)) > 0),
    request_hash text NOT NULL CHECK (length(trim(request_hash)) > 0),
    result_json jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, mutation_id)
);
