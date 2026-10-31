DROP INDEX IF EXISTS portfolio_assets_price_mode_idx;
ALTER TABLE portfolio_assets DROP CONSTRAINT IF EXISTS portfolio_assets_price_mode_check;
ALTER TABLE portfolio_assets DROP CONSTRAINT IF EXISTS portfolio_assets_unit_label_check;
ALTER TABLE portfolio_assets DROP COLUMN IF EXISTS price_mode;
ALTER TABLE portfolio_assets DROP COLUMN IF EXISTS unit_confirmed;
ALTER TABLE portfolio_assets DROP COLUMN IF EXISTS unit_label;
