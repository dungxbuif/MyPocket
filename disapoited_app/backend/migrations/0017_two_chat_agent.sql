ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS agent_runs_kind_check;

ALTER TABLE agent_runs
    ADD CONSTRAINT agent_runs_kind_check
    CHECK (kind IN ('transaction_draft', 'analysis', 'intake', 'advisor'));

DROP INDEX IF EXISTS transaction_drafts_agent_run_idx;

CREATE INDEX IF NOT EXISTS transaction_drafts_agent_run_idx
    ON transaction_drafts (agent_run_id)
    WHERE agent_run_id IS NOT NULL;
