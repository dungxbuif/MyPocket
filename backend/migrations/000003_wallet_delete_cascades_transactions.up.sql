ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_wallet_id_fkey;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_wallet_id_fkey
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE;
