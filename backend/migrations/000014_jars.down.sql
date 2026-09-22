DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM transactions WHERE jar_id IS NOT NULL)
     OR EXISTS (SELECT 1 FROM jar_month_configs)
     OR EXISTS (SELECT 1 FROM jars) THEN
    RAISE EXCEPTION 'Refusing to drop jar history while jar identities, configurations, or linked transactions exist';
  END IF;
END $$;

DROP INDEX IF EXISTS transactions_owner_jar_idx;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_owner_jar_fk;
ALTER TABLE transactions DROP COLUMN IF EXISTS jar_id;
DROP TABLE IF EXISTS jar_month_configs;
DROP TABLE IF EXISTS jar_months;
DROP TABLE IF EXISTS jars;
