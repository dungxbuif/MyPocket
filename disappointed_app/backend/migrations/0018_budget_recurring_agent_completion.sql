ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS budget_id uuid REFERENCES budgets(id) ON DELETE SET NULL;

ALTER TABLE transaction_drafts
    ADD COLUMN IF NOT EXISTS budget_id uuid REFERENCES budgets(id) ON DELETE SET NULL;

ALTER TABLE recurring_schedules
    ADD COLUMN IF NOT EXISTS budget_id uuid REFERENCES budgets(id) ON DELETE SET NULL;

ALTER TABLE recurring_schedules
    ADD COLUMN IF NOT EXISTS posting_mode text NOT NULL DEFAULT 'draft';

ALTER TABLE recurring_schedules
    ADD COLUMN IF NOT EXISTS paused_at timestamptz;

ALTER TABLE recurring_schedules
    ADD COLUMN IF NOT EXISTS ends_at timestamptz;

CREATE INDEX IF NOT EXISTS transactions_user_budget_period_idx
    ON transactions (user_id, budget_id, occurred_at)
    WHERE budget_id IS NOT NULL AND archived_at IS NULL;

CREATE INDEX IF NOT EXISTS transaction_drafts_user_budget_idx
    ON transaction_drafts (user_id, budget_id)
    WHERE budget_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS recurring_schedules_user_budget_idx
    ON recurring_schedules (user_id, budget_id)
    WHERE budget_id IS NOT NULL AND archived_at IS NULL;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_budget_expense_only;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_budget_expense_only
    CHECK (budget_id IS NULL OR type = 'expense');

ALTER TABLE transaction_drafts
    DROP CONSTRAINT IF EXISTS transaction_drafts_budget_expense_only;

ALTER TABLE transaction_drafts
    ADD CONSTRAINT transaction_drafts_budget_expense_only
    CHECK (budget_id IS NULL OR transaction_type = 'expense');

ALTER TABLE recurring_schedules
    DROP CONSTRAINT IF EXISTS recurring_schedules_budget_expense_only;

ALTER TABLE recurring_schedules
    ADD CONSTRAINT recurring_schedules_budget_expense_only
    CHECK (budget_id IS NULL OR transaction_type = 'expense');

ALTER TABLE recurring_schedules
    DROP CONSTRAINT IF EXISTS recurring_schedules_posting_mode_check;

ALTER TABLE recurring_schedules
    ADD CONSTRAINT recurring_schedules_posting_mode_check
    CHECK (posting_mode IN ('draft', 'auto_post'));
