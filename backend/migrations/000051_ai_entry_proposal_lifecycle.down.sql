-- Disposable rollback only. Legacy rows must be normalized before reverting.
ALTER TABLE ai_entry_proposals
    DROP CONSTRAINT IF EXISTS ai_entry_proposals_status_check;

ALTER TABLE ai_entry_proposals
    ADD CONSTRAINT ai_entry_proposals_status_check
    CHECK (status IN ('pending', 'approved', 'rejected'));
