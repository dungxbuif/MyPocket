-- Disposable fixtures only. Never apply this down migration to production.
DROP INDEX IF EXISTS advisor_assistant_completion_unique;
DROP INDEX IF EXISTS advisor_run_runtime_reaper_idx;
ALTER TABLE advisor_runs DROP CONSTRAINT IF EXISTS advisor_runs_deadline_check;
ALTER TABLE advisor_runs DROP COLUMN IF EXISTS deadline_at;
