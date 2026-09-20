CREATE TABLE budgets (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 name text NOT NULL CHECK (length(trim(name)) > 0),
 limit_amount bigint NOT NULL CHECK (limit_amount > 0 AND limit_amount <= 9007199254740991),
 wallet_id text REFERENCES wallets(id) ON DELETE CASCADE,
 category_id text REFERENCES categories(id) ON DELETE CASCADE,
 start_at timestamptz NOT NULL,
 end_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK (end_at > start_at)
);
CREATE INDEX budgets_owner_period ON budgets(owner_id, start_at, end_at);
