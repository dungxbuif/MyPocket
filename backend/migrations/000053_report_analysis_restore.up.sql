ALTER TABLE report_analyses
    ADD COLUMN analysis_fingerprint text NOT NULL DEFAULT '';

UPDATE report_analyses
SET analysis_fingerprint = job_key
WHERE analysis_fingerprint = '';

CREATE INDEX report_analyses_scope_latest_idx
    ON report_analyses(owner_id, created_at DESC, status);

CREATE UNIQUE INDEX report_analyses_fingerprint_live_uidx
    ON report_analyses(owner_id, analysis_fingerprint)
    WHERE status IN ('queued', 'running', 'completed');
