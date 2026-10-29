-- Advisor runtime fencing and completion receipt metadata.
-- This migration is additive. Existing rows are backfilled from their server
-- creation time; no provider is replayed and no financial data is rewritten.

ALTER TABLE advisor_runs
    ADD COLUMN deadline_at timestamptz;

UPDATE advisor_runs
SET deadline_at = created_at + interval '240 seconds'
WHERE deadline_at IS NULL;

ALTER TABLE advisor_runs
    ALTER COLUMN deadline_at SET NOT NULL,
    ALTER COLUMN deadline_at SET DEFAULT (NOW() + interval '240 seconds');

ALTER TABLE advisor_runs
    ADD CONSTRAINT advisor_runs_deadline_check CHECK (deadline_at >= created_at);

CREATE INDEX advisor_run_runtime_reaper_idx
    ON advisor_runs(status, lease_until, deadline_at)
    WHERE status IN ('queued', 'running');

-- The stage baseline already has UNIQUE(run_id, role). Keep a named partial
-- guard for deployments that were created from an older compatible schema,
-- but fail before DDL if duplicate assistant rows exist rather than deleting
-- or choosing a winner automatically.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM advisor_messages
        WHERE role = 'assistant'
        GROUP BY run_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'advisor assistant completion duplicates must be remediated before migration 000041';
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS advisor_assistant_completion_unique
    ON advisor_messages(run_id)
    WHERE role = 'assistant';
