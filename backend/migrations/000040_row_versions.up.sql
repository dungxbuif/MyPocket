ALTER TABLE transactions
  ADD COLUMN version bigint NOT NULL DEFAULT 1;

ALTER TABLE categories
  ADD COLUMN version bigint NOT NULL DEFAULT 1;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_version_check CHECK (version >= 1);

ALTER TABLE categories
  ADD CONSTRAINT categories_version_check CHECK (version >= 1);
