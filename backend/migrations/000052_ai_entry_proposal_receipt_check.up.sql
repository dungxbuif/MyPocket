-- Both legacy approved and the canonical applied status carry a ledger
-- receipt. Pending/application/rejection states must not claim one early.
ALTER TABLE ai_entry_proposals
    DROP CONSTRAINT IF EXISTS ai_entry_proposals_check;

ALTER TABLE ai_entry_proposals
    ADD CONSTRAINT ai_entry_proposals_check
    CHECK ((status IN ('approved', 'applied')) = (transaction_id IS NOT NULL));
