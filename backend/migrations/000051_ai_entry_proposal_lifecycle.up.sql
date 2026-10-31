-- Extend the legacy one-shot AI entry proposal table with the server-owned
-- lifecycle vocabulary used by the unified Assistant. Existing approved,
-- rejected and pending rows remain valid for rolling upgrades.
ALTER TABLE ai_entry_proposals
    DROP CONSTRAINT IF EXISTS ai_entry_proposals_status_check;

ALTER TABLE ai_entry_proposals
    ADD CONSTRAINT ai_entry_proposals_status_check
    CHECK (status IN ('pending', 'applying', 'applied', 'approved', 'rejected', 'expired', 'conflict', 'superseded'));
