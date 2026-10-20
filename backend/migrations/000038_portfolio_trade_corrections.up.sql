ALTER TABLE portfolio_assets ADD COLUMN IF NOT EXISTS version bigint NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS portfolio_trade_revisions (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 asset_id text NOT NULL REFERENCES portfolio_assets(id) ON DELETE CASCADE,
 source_trade_id text NOT NULL,
 revision bigint NOT NULL CHECK (revision > 0),
 operation text NOT NULL CHECK (operation IN ('replace','void')),
 side text,
 quantity_scaled bigint,
 unit_price bigint,
 fee bigint NOT NULL DEFAULT 0,
 occurred_at timestamptz,
 note text,
 reason text NOT NULL,
 actor_id text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(owner_id, source_trade_id, revision)
);
CREATE INDEX IF NOT EXISTS portfolio_trade_revisions_asset_idx ON portfolio_trade_revisions(owner_id, asset_id, source_trade_id, revision DESC);
