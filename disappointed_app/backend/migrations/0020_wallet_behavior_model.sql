ALTER TABLE wallets
    ADD COLUMN IF NOT EXISTS goal_target_vnd bigint,
    ADD COLUMN IF NOT EXISTS goal_deadline_on date;

ALTER TABLE wallets
    DROP CONSTRAINT IF EXISTS wallets_goal_target_vnd_check,
    DROP CONSTRAINT IF EXISTS wallets_type_check;

DO $$
DECLARE
    constraint_name text;
BEGIN
    FOR constraint_name IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'wallets'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%type = ''credit''%'
    LOOP
        EXECUTE format('ALTER TABLE wallets DROP CONSTRAINT %I', constraint_name);
    END LOOP;

    FOR constraint_name IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'wallets'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%cash%'
    LOOP
        EXECUTE format('ALTER TABLE wallets DROP CONSTRAINT %I', constraint_name);
    END LOOP;
END $$;

UPDATE wallets
SET type = CASE type
    WHEN 'credit' THEN 'credit'
    WHEN 'savings' THEN 'goal'
    ELSE 'basic'
END
WHERE type IN ('cash', 'bank', 'e_wallet', 'debt', 'savings', 'credit');

ALTER TABLE wallets
    ADD CONSTRAINT wallets_type_behavior_check CHECK (type IN ('basic', 'goal', 'credit')),
    ADD CONSTRAINT wallets_credit_metadata_behavior_check CHECK (
        type = 'credit'
        OR (credit_limit_vnd IS NULL AND statement_day IS NULL AND payment_due_day IS NULL)
    ),
    ADD CONSTRAINT wallets_goal_metadata_behavior_check CHECK (
        type = 'goal'
        OR (goal_target_vnd IS NULL AND goal_deadline_on IS NULL)
    ),
    ADD CONSTRAINT wallets_goal_target_vnd_check CHECK (goal_target_vnd IS NULL OR goal_target_vnd > 0);
