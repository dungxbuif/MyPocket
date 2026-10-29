ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS credit_kind text,
  ADD COLUMN IF NOT EXISTS credit_payment_id text;

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_credit_kind_check;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_credit_kind_check
  CHECK (
    credit_kind IS NULL
    OR (credit_kind IN ('purchase', 'fee', 'interest') AND type = 'expense')
    OR (credit_kind IN ('refund', 'payment') AND type = 'income')
  );

CREATE INDEX IF NOT EXISTS transactions_credit_kind_idx
  ON transactions(credit_kind)
  WHERE credit_kind IS NOT NULL;
CREATE INDEX IF NOT EXISTS transactions_credit_payment_idx
  ON transactions(credit_payment_id)
  WHERE credit_payment_id IS NOT NULL;
