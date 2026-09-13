-- Wallet deletion is intentionally destructive by product decision; keep the
-- forward behavior during migration rollback rather than silently blocking it.
SELECT 1;
