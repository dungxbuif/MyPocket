CREATE TABLE IF NOT EXISTS budgets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    period_type text NOT NULL CHECK (period_type IN ('weekly', 'monthly', 'quarterly', 'yearly', 'custom')),
    amount_vnd bigint NOT NULL CHECK (amount_vnd > 0),
    all_categories boolean NOT NULL DEFAULT true,
    custom_start date,
    custom_end date,
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (period_type = 'custom' AND custom_start IS NOT NULL AND custom_end IS NOT NULL AND custom_end >= custom_start)
        OR (period_type <> 'custom' AND custom_start IS NULL AND custom_end IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS budgets_user_active_idx
    ON budgets (user_id, archived_at, created_at);

CREATE TABLE IF NOT EXISTS budget_categories (
    budget_id uuid NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (budget_id, category_id)
);

CREATE INDEX IF NOT EXISTS budget_categories_user_idx
    ON budget_categories (user_id, category_id);

CREATE TABLE IF NOT EXISTS budget_alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    budget_id uuid NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    threshold integer NOT NULL CHECK (threshold IN (80, 100)),
    period_start date NOT NULL,
    period_end date NOT NULL,
    triggered_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (budget_id, threshold, period_start)
);

CREATE INDEX IF NOT EXISTS budget_alerts_user_idx
    ON budget_alerts (user_id, triggered_at DESC);
