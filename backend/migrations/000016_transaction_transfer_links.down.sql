DROP INDEX IF EXISTS transactions_transfer_id_idx;
ALTER TABLE transactions DROP COLUMN IF EXISTS transfer_id;
