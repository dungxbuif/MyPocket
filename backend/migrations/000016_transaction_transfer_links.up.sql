ALTER TABLE transactions ADD COLUMN IF NOT EXISTS transfer_id text;
CREATE INDEX IF NOT EXISTS transactions_transfer_id_idx ON transactions(transfer_id);
