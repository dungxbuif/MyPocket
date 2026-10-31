ALTER TABLE portfolio_assets
  ADD COLUMN IF NOT EXISTS unit_label text,
  ADD COLUMN IF NOT EXISTS unit_confirmed boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS price_mode text NOT NULL DEFAULT 'manual';

DO $$
BEGIN
  ALTER TABLE portfolio_assets
    ADD CONSTRAINT portfolio_assets_unit_label_check
    CHECK (unit_label IS NULL OR char_length(trim(unit_label)) BETWEEN 1 AND 64);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  ALTER TABLE portfolio_assets
    ADD CONSTRAINT portfolio_assets_price_mode_check
    CHECK (price_mode IN ('manual', 'automatic'));
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS portfolio_assets_price_mode_idx ON portfolio_assets(price_mode);
