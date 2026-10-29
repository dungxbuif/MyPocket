ALTER TABLE feedback
  ADD COLUMN IF NOT EXISTS screenshot_object_key text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS screenshot_mime_type text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS screenshot_size_bytes bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS screenshot_created_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_feedback_screenshot_object_key
  ON feedback (screenshot_object_key)
  WHERE screenshot_object_key <> '';
