ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS adjustment_direction text;

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_type_check;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_type_check
  CHECK (type IN ('income', 'expense', 'adjustment'));

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_adjustment_direction_check;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_adjustment_direction_check
  CHECK (
    (type = 'adjustment' AND adjustment_direction IN ('increase', 'decrease'))
    OR (type <> 'adjustment' AND adjustment_direction IS NULL)
  );

CREATE INDEX IF NOT EXISTS transactions_adjustment_direction_idx
  ON transactions(adjustment_direction)
  WHERE type = 'adjustment';
