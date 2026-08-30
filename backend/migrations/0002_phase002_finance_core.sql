CREATE TABLE IF NOT EXISTS wallets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    type text NOT NULL CHECK (type IN ('cash', 'bank', 'credit', 'e_wallet', 'savings', 'debt')),
    balance_vnd bigint NOT NULL DEFAULT 0,
    include_in_total boolean NOT NULL DEFAULT true,
    is_default_ai boolean NOT NULL DEFAULT false,
    credit_limit_vnd bigint CHECK (credit_limit_vnd IS NULL OR credit_limit_vnd >= 0),
    statement_day integer CHECK (statement_day IS NULL OR statement_day BETWEEN 1 AND 31),
    payment_due_day integer CHECK (payment_due_day IS NULL OR payment_due_day BETWEEN 1 AND 31),
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        type = 'credit'
        OR (credit_limit_vnd IS NULL AND statement_day IS NULL AND payment_due_day IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS wallets_user_id_idx ON wallets (user_id);

CREATE UNIQUE INDEX IF NOT EXISTS wallets_one_default_ai_per_user
    ON wallets (user_id)
    WHERE is_default_ai AND archived_at IS NULL;

CREATE TABLE IF NOT EXISTS categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    parent_id uuid REFERENCES categories(id) ON DELETE RESTRICT,
    kind text NOT NULL CHECK (kind IN ('expense', 'income', 'debt')),
    name text NOT NULL CHECK (length(trim(name)) > 0),
    system_key text,
    is_system boolean NOT NULL DEFAULT false,
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (id <> parent_id),
    CHECK (
        (is_system AND user_id IS NULL AND system_key IS NOT NULL)
        OR (NOT is_system AND user_id IS NOT NULL AND system_key IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS categories_system_key_unique
    ON categories (system_key);

CREATE INDEX IF NOT EXISTS categories_user_id_idx ON categories (user_id);
CREATE INDEX IF NOT EXISTS categories_parent_id_idx ON categories (parent_id);

CREATE TABLE IF NOT EXISTS wallet_category_settings (
    wallet_id uuid NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (wallet_id, category_id)
);

CREATE INDEX IF NOT EXISTS wallet_category_settings_user_id_idx
    ON wallet_category_settings (user_id);

CREATE TABLE IF NOT EXISTS receipt_objects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_key text NOT NULL CHECK (length(trim(object_key)) > 0),
    content_type text NOT NULL CHECK (length(trim(content_type)) > 0),
    size_bytes bigint NOT NULL CHECK (size_bytes > 0),
    checksum_sha256 text NOT NULL CHECK (length(checksum_sha256) = 64),
    original_filename text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS receipt_objects_user_object_key_unique
    ON receipt_objects (user_id, object_key);

CREATE TABLE IF NOT EXISTS transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('income', 'expense', 'transfer', 'adjustment')),
    source_wallet_id uuid NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    destination_wallet_id uuid REFERENCES wallets(id) ON DELETE RESTRICT,
    category_id uuid REFERENCES categories(id) ON DELETE RESTRICT,
    receipt_object_id uuid REFERENCES receipt_objects(id) ON DELETE SET NULL,
    amount_vnd bigint NOT NULL CHECK (amount_vnd > 0),
    balance_after_vnd bigint,
    source_delta_vnd bigint NOT NULL DEFAULT 0,
    destination_delta_vnd bigint NOT NULL DEFAULT 0,
    note text NOT NULL DEFAULT '',
    with_person text NOT NULL DEFAULT '',
    event_ref text NOT NULL DEFAULT '',
    occurred_at timestamptz NOT NULL,
    excluded_from_reports boolean NOT NULL DEFAULT false,
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (destination_wallet_id IS NULL OR destination_wallet_id <> source_wallet_id),
    CHECK (
        (type = 'transfer' AND destination_wallet_id IS NOT NULL)
        OR (type <> 'transfer' AND destination_wallet_id IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS transactions_user_occurred_at_idx
    ON transactions (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS transactions_user_wallet_idx
    ON transactions (user_id, source_wallet_id);
CREATE INDEX IF NOT EXISTS transactions_user_category_idx
    ON transactions (user_id, category_id);

CREATE TABLE IF NOT EXISTS finance_idempotency_keys (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key text NOT NULL CHECK (length(trim(key)) > 0),
    request_hash text NOT NULL CHECK (length(trim(request_hash)) > 0),
    response_status integer NOT NULL CHECK (response_status BETWEEN 100 AND 599),
    response_json jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);

INSERT INTO categories (id, kind, name, system_key, is_system)
VALUES
    ('00000000-0000-4000-8000-000000000201', 'expense', 'Ăn uống', 'expense_food', true),
    ('00000000-0000-4000-8000-000000000202', 'expense', 'Mua sắm', 'expense_shopping', true),
    ('00000000-0000-4000-8000-000000000203', 'expense', 'Di chuyển', 'expense_transport', true),
    ('00000000-0000-4000-8000-000000000204', 'income', 'Lương', 'income_salary', true),
    ('00000000-0000-4000-8000-000000000205', 'income', 'Thưởng', 'income_bonus', true),
    ('00000000-0000-4000-8000-000000000206', 'debt', 'Vay nợ', 'debt_loan', true)
ON CONFLICT (system_key) DO UPDATE
SET
    name = EXCLUDED.name,
    kind = EXCLUDED.kind,
    is_system = true,
    updated_at = now();
