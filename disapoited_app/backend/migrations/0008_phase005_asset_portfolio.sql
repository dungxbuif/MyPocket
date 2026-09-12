CREATE TABLE IF NOT EXISTS asset_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('gold', 'stock', 'crypto', 'foreign_currency', 'other')),
    symbol text NOT NULL DEFAULT '',
    exchange text NOT NULL DEFAULT '',
    name text NOT NULL CHECK (length(trim(name)) > 0),
    unit text NOT NULL CHECK (length(trim(unit)) > 0),
    reporting_currency text NOT NULL DEFAULT 'VND' CHECK (reporting_currency = 'VND'),
    pricing_mode text NOT NULL DEFAULT 'manual' CHECK (pricing_mode IN ('manual', 'automatic')),
    provider_key text,
    provider_symbol text,
    include_in_net_worth boolean NOT NULL DEFAULT true,
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        pricing_mode = 'manual'
        OR (provider_key IS NOT NULL AND length(trim(provider_key)) > 0 AND provider_symbol IS NOT NULL AND length(trim(provider_symbol)) > 0)
    ),
    UNIQUE (id, user_id)
);

CREATE INDEX IF NOT EXISTS asset_positions_user_active_idx
    ON asset_positions (user_id, archived_at, created_at DESC);

CREATE INDEX IF NOT EXISTS asset_positions_auto_refresh_idx
    ON asset_positions (provider_key, provider_symbol)
    WHERE archived_at IS NULL AND pricing_mode = 'automatic';

CREATE UNIQUE INDEX IF NOT EXISTS asset_positions_user_symbol_unique
    ON asset_positions (user_id, type, lower(symbol), lower(exchange), unit)
    WHERE archived_at IS NULL AND length(trim(symbol)) > 0;

CREATE TABLE IF NOT EXISTS asset_trades (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id uuid NOT NULL,
    side text NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity numeric(30,12) NOT NULL CHECK (quantity > 0),
    unit_price_vnd bigint NOT NULL CHECK (unit_price_vnd >= 0),
    fee_vnd bigint NOT NULL DEFAULT 0 CHECK (fee_vnd >= 0),
    occurred_at timestamptz NOT NULL,
    quantity_after numeric(30,12) NOT NULL DEFAULT 0 CHECK (quantity_after >= 0),
    cost_basis_after_vnd bigint NOT NULL DEFAULT 0 CHECK (cost_basis_after_vnd >= 0),
    realized_pnl_vnd bigint NOT NULL DEFAULT 0,
    note text NOT NULL DEFAULT '',
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (asset_id, user_id) REFERENCES asset_positions(id, user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS asset_trades_user_asset_order_idx
    ON asset_trades (user_id, asset_id, occurred_at, created_at, id);

CREATE TABLE IF NOT EXISTS asset_price_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id uuid NOT NULL,
    unit_price_vnd bigint NOT NULL CHECK (unit_price_vnd > 0),
    priced_at timestamptz NOT NULL,
    source text NOT NULL CHECK (length(trim(source)) > 0),
    provider_quote_id text,
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (asset_id, user_id) REFERENCES asset_positions(id, user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS asset_price_history_user_asset_latest_idx
    ON asset_price_history (user_id, asset_id, priced_at DESC, created_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS asset_price_history_provider_quote_unique
    ON asset_price_history (source, provider_quote_id)
    WHERE provider_quote_id IS NOT NULL AND length(trim(provider_quote_id)) > 0;
