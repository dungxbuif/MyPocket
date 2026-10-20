ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_version_check;
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_version_check;
ALTER TABLE transactions DROP COLUMN IF EXISTS version;
ALTER TABLE categories DROP COLUMN IF EXISTS version;
