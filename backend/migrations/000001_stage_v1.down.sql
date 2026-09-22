-- Stage v1 schema contains user data and is not safely reversible.
-- Reset a disposable database explicitly instead of destructive rollback.
SELECT 1;
