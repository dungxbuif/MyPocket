DROP INDEX IF EXISTS report_analyses_fingerprint_live_uidx;
DROP INDEX IF EXISTS report_analyses_scope_latest_idx;
ALTER TABLE report_analyses DROP COLUMN IF EXISTS analysis_fingerprint;
