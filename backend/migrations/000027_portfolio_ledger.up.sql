CREATE TABLE IF NOT EXISTS portfolio_assets (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  symbol text NOT NULL,
  name text NOT NULL,
  latest_price bigint,
  latest_price_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT portfolio_assets_symbol_check CHECK (char_length(trim(symbol)) > 0),
  CONSTRAINT portfolio_assets_name_check CHECK (char_length(trim(name)) > 0),
  CONSTRAINT portfolio_assets_price_check CHECK (latest_price IS NULL OR latest_price >= 0),
  FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS portfolio_assets_owner_symbol_idx ON portfolio_assets(owner_id, lower(symbol));

CREATE TABLE IF NOT EXISTS portfolio_trades (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  asset_id text NOT NULL,
  side text NOT NULL,
  quantity_scaled bigint NOT NULL,
  unit_price bigint NOT NULL,
  fee bigint NOT NULL DEFAULT 0,
  occurred_at timestamptz NOT NULL,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT portfolio_trades_side_check CHECK (side IN ('buy', 'sell')),
  CONSTRAINT portfolio_trades_quantity_check CHECK (quantity_scaled > 0),
  CONSTRAINT portfolio_trades_price_check CHECK (unit_price >= 0),
  CONSTRAINT portfolio_trades_fee_check CHECK (fee >= 0),
  FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE,
  FOREIGN KEY (asset_id) REFERENCES portfolio_assets(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS portfolio_trades_owner_asset_time_idx ON portfolio_trades(owner_id, asset_id, occurred_at, created_at, id);
