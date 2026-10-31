CREATE TABLE IF NOT EXISTS market_products (
  id text PRIMARY KEY,
  provider_key text NOT NULL,
  external_product_id text NOT NULL,
  name text NOT NULL,
  kind text NOT NULL,
  unit text NOT NULL,
  currency text NOT NULL DEFAULT 'VND',
  valuation_side text NOT NULL DEFAULT 'buyback',
  enabled boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT market_products_provider_check CHECK (char_length(trim(provider_key)) > 0),
  CONSTRAINT market_products_external_id_check CHECK (char_length(trim(external_product_id)) > 0),
  CONSTRAINT market_products_name_check CHECK (char_length(trim(name)) > 0),
  CONSTRAINT market_products_unit_check CHECK (char_length(trim(unit)) > 0),
  CONSTRAINT market_products_currency_check CHECK (char_length(trim(currency)) > 0),
  CONSTRAINT market_products_valuation_side_check CHECK (valuation_side = 'buyback'),
  UNIQUE (provider_key, external_product_id)
);
CREATE INDEX IF NOT EXISTS market_products_enabled_provider_idx ON market_products(enabled, provider_key);

CREATE TABLE IF NOT EXISTS market_quotes (
  product_id text PRIMARY KEY,
  price bigint NOT NULL,
  observed_at timestamptz NOT NULL,
  fetched_at timestamptz NOT NULL,
  CONSTRAINT market_quotes_price_check CHECK (price > 0),
  FOREIGN KEY (product_id) REFERENCES market_products(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS market_quotes_observed_at_idx ON market_quotes(observed_at DESC);

ALTER TABLE portfolio_assets
  ADD COLUMN IF NOT EXISTS market_product_id text,
  ADD COLUMN IF NOT EXISTS latest_price_source text;

DO $$
BEGIN
  ALTER TABLE portfolio_assets
    ADD CONSTRAINT portfolio_assets_market_product_fk
    FOREIGN KEY (market_product_id) REFERENCES market_products(id) ON DELETE RESTRICT;
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS portfolio_assets_market_product_idx ON portfolio_assets(market_product_id);
