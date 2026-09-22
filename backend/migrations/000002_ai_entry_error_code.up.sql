ALTER TABLE ai_entry_sessions
  ADD COLUMN IF NOT EXISTS error_code text NOT NULL DEFAULT '';
