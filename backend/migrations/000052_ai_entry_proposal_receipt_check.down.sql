ALTER TABLE ai_entry_proposals
    DROP CONSTRAINT IF EXISTS ai_entry_proposals_check;

ALTER TABLE ai_entry_proposals
    ADD CONSTRAINT ai_entry_proposals_check
    CHECK ((status = 'approved') = (transaction_id IS NOT NULL));
