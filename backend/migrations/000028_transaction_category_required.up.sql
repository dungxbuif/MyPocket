-- Ordinary income/expense rows must always have a visible category. Balance
-- adjustments remain intentionally uncategorized; transfer and credit rows
-- already carry their system/category entries.
UPDATE transactions AS t
SET category_id = c.id
FROM categories AS c
WHERE t.category_id IS NULL
  AND t.type = 'expense'
  AND c.system_key = 'expense_general';

UPDATE transactions AS t
SET category_id = c.id
FROM categories AS c
WHERE t.category_id IS NULL
  AND t.type = 'income'
  AND c.system_key = 'income_other';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM transactions
    WHERE type IN ('income', 'expense')
      AND category_id IS NULL
  ) THEN
    RAISE EXCEPTION 'cannot require transaction categories: unresolved ordinary rows remain';
  END IF;
END $$;

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_category_required_check;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_category_required_check
  CHECK (
    (type = 'adjustment' AND category_id IS NULL)
    OR (type IN ('income', 'expense') AND category_id IS NOT NULL)
  );
