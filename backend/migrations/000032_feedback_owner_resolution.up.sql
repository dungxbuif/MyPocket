ALTER TABLE feedback ADD COLUMN resolved boolean NOT NULL DEFAULT false;
ALTER TABLE feedback ADD COLUMN resolved_at timestamptz;
