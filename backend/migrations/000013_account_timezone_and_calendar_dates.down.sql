DROP INDEX IF EXISTS budgets_owner_period;
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS budgets_date_order;
ALTER TABLE budgets ADD COLUMN start_at timestamptz;
ALTER TABLE budgets ADD COLUMN end_at timestamptz;
UPDATE budgets
SET start_at = start_date::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh',
    end_at = (end_date + 1)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh';
ALTER TABLE budgets ALTER COLUMN start_at SET NOT NULL;
ALTER TABLE budgets ALTER COLUMN end_at SET NOT NULL;
ALTER TABLE budgets DROP COLUMN start_date;
ALTER TABLE budgets DROP COLUMN end_date;
ALTER TABLE budgets ADD CONSTRAINT budgets_time_order CHECK (end_at > start_at);
CREATE INDEX budgets_owner_period ON budgets(owner_id, start_at, end_at);

ALTER TABLE wallets
  ALTER COLUMN target_date TYPE timestamptz
  USING CASE WHEN target_date IS NULL THEN NULL ELSE target_date::timestamp AT TIME ZONE 'UTC' END;

-- Keep account timezone preferences on rollback so user settings are not lost.
