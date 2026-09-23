ALTER TABLE ai_entry_sessions
  ADD COLUMN IF NOT EXISTS reply text NOT NULL DEFAULT '';
