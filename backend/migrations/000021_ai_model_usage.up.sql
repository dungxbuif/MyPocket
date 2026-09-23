ALTER TABLE ai_entry_sessions
 ADD COLUMN IF NOT EXISTS model_usage jsonb NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE advisor_runs
 ADD COLUMN IF NOT EXISTS model_usage jsonb NOT NULL DEFAULT '{}'::jsonb;
