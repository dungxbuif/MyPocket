CREATE TABLE IF NOT EXISTS events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    starts_on date NOT NULL,
    ends_on date NOT NULL,
    note text NOT NULL DEFAULT '',
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_on >= starts_on)
);

CREATE INDEX IF NOT EXISTS events_user_active_idx
    ON events (user_id, archived_at, starts_on DESC);

CREATE TABLE IF NOT EXISTS event_transactions (
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id uuid NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, transaction_id)
);

CREATE INDEX IF NOT EXISTS event_transactions_user_idx
    ON event_transactions (user_id, transaction_id);

CREATE TABLE IF NOT EXISTS obligations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    direction text NOT NULL CHECK (direction IN ('borrowed', 'lent')),
    principal_vnd bigint NOT NULL CHECK (principal_vnd > 0),
    counterparty text NOT NULL CHECK (length(trim(counterparty)) > 0),
    due_on date NOT NULL,
    note text NOT NULL DEFAULT '',
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS obligations_user_active_idx
    ON obligations (user_id, archived_at, due_on);

CREATE TABLE IF NOT EXISTS obligation_repayments (
    obligation_id uuid NOT NULL REFERENCES obligations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id uuid NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (obligation_id, transaction_id)
);

CREATE INDEX IF NOT EXISTS obligation_repayments_user_idx
    ON obligation_repayments (user_id, transaction_id);
