CREATE TABLE IF NOT EXISTS recurring_schedules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    frequency text NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly')),
    timezone text NOT NULL CHECK (length(trim(timezone)) > 0),
    starts_at timestamptz NOT NULL,
    next_occurs_at timestamptz NOT NULL,
    transaction_type text NOT NULL CHECK (transaction_type IN ('income', 'expense', 'transfer')),
    source_wallet_id uuid NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    destination_wallet_id uuid REFERENCES wallets(id) ON DELETE RESTRICT,
    category_id uuid REFERENCES categories(id) ON DELETE RESTRICT,
    amount_vnd bigint NOT NULL CHECK (amount_vnd > 0),
    note text NOT NULL DEFAULT '',
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (destination_wallet_id IS NULL OR destination_wallet_id <> source_wallet_id)
);

CREATE INDEX IF NOT EXISTS recurring_schedules_due_idx
    ON recurring_schedules (archived_at, next_occurs_at);

CREATE TABLE IF NOT EXISTS recurring_occurrences (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id uuid NOT NULL REFERENCES recurring_schedules(id) ON DELETE CASCADE,
    occurrence_key text NOT NULL CHECK (length(trim(occurrence_key)) > 0),
    occurs_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (schedule_id, occurrence_key)
);

CREATE TABLE IF NOT EXISTS transaction_drafts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id uuid REFERENCES recurring_schedules(id) ON DELETE SET NULL,
    occurrence_key text NOT NULL CHECK (length(trim(occurrence_key)) > 0),
    transaction_type text NOT NULL CHECK (transaction_type IN ('income', 'expense', 'transfer')),
    source_wallet_id uuid NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    destination_wallet_id uuid REFERENCES wallets(id) ON DELETE RESTRICT,
    category_id uuid REFERENCES categories(id) ON DELETE RESTRICT,
    amount_vnd bigint NOT NULL CHECK (amount_vnd > 0),
    occurred_at timestamptz NOT NULL,
    note text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'rejected')),
    confirmed_transaction_id uuid REFERENCES transactions(id) ON DELETE SET NULL,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, occurrence_key)
);

CREATE INDEX IF NOT EXISTS transaction_drafts_user_status_idx
    ON transaction_drafts (user_id, status, occurred_at DESC);

CREATE TABLE IF NOT EXISTS worker_leases (
    lease_key text PRIMARY KEY CHECK (length(trim(lease_key)) > 0),
    owner text NOT NULL CHECK (length(trim(owner)) > 0),
    expires_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
