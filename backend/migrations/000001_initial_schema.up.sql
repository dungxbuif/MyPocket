DO $$
BEGIN
  IF to_regclass('public.app_users') IS NOT NULL AND to_regclass('public."user"') IS NULL THEN
    ALTER TABLE app_users RENAME TO "user";
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS "user" (
  id text PRIMARY KEY,
  google_subject text UNIQUE NOT NULL DEFAULT '',
  name text NOT NULL DEFAULT '',
  email text UNIQUE NOT NULL,
  email_verified boolean NOT NULL DEFAULT false,
  avatar_url text NOT NULL DEFAULT '',
  password_hash text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallets (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  name text NOT NULL,
  type text NOT NULL,
  currency text NOT NULL DEFAULT 'VND',
  opening_balance bigint NOT NULL DEFAULT 0,
  is_in_total boolean NOT NULL DEFAULT true,
  description text,
  target_amount bigint,
  target_date timestamptz,
  credit_limit bigint,
  last_statement_balance bigint,
  statement_day integer,
  payment_due_day integer,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT wallets_type_check CHECK (type IN ('basic', 'goal', 'credit')),
  CONSTRAINT wallets_currency_check CHECK (currency = 'VND')
);
CREATE INDEX IF NOT EXISTS wallets_owner_id_idx ON wallets(owner_id);

CREATE TABLE IF NOT EXISTS categories (
  id text PRIMARY KEY,
  owner_id text,
  parent_id text,
  kind text NOT NULL,
  name text NOT NULL,
  system_key text UNIQUE,
  is_system boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT categories_kind_check CHECK (kind IN ('expense', 'income', 'debt'))
);
CREATE INDEX IF NOT EXISTS categories_owner_id_idx ON categories(owner_id);
CREATE INDEX IF NOT EXISTS categories_parent_id_idx ON categories(parent_id);

CREATE TABLE IF NOT EXISTS category_wallets (
  category_id text NOT NULL,
  wallet_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (category_id, wallet_id),
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  wallet_id text NOT NULL,
  category_id text,
  type text NOT NULL,
  amount bigint NOT NULL CHECK (amount > 0),
  occurred_at timestamptz NOT NULL,
  note text,
  included_in_reports boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT transactions_type_check CHECK (type IN ('income', 'expense')),
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS transactions_owner_id_idx ON transactions(owner_id);
CREATE INDEX IF NOT EXISTS transactions_wallet_id_idx ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS transactions_category_id_idx ON transactions(category_id);
CREATE INDEX IF NOT EXISTS transactions_occurred_at_idx ON transactions(occurred_at);
