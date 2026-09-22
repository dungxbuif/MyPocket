ALTER TABLE "user"
  ADD COLUMN timezone text NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
  ADD COLUMN timezone_confirmed boolean NOT NULL DEFAULT false;

-- Existing accounts retain the legacy app's Vietnam-local calendar interpretation.
UPDATE "user" SET timezone_confirmed = true;

-- Target dates are calendar labels; the legacy endpoint stored them as UTC midnight.
ALTER TABLE wallets
  ALTER COLUMN target_date TYPE date
  USING CASE WHEN target_date IS NULL THEN NULL ELSE (target_date AT TIME ZONE 'UTC')::date END;

DROP INDEX IF EXISTS budgets_owner_period;
ALTER TABLE budgets ADD COLUMN start_date date;
ALTER TABLE budgets ADD COLUMN end_date date;
UPDATE budgets
SET start_date = (start_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,
    end_date = ((end_at AT TIME ZONE 'Asia/Ho_Chi_Minh') - interval '1 microsecond')::date;
ALTER TABLE budgets ALTER COLUMN start_date SET NOT NULL;
ALTER TABLE budgets ALTER COLUMN end_date SET NOT NULL;
ALTER TABLE budgets DROP COLUMN start_at;
ALTER TABLE budgets DROP COLUMN end_at;
ALTER TABLE budgets ADD CONSTRAINT budgets_date_order CHECK (end_date >= start_date);
CREATE INDEX budgets_owner_period ON budgets(owner_id, start_date, end_date);
