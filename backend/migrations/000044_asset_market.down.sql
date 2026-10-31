DROP INDEX IF EXISTS portfolio_assets_market_product_idx;
ALTER TABLE portfolio_assets DROP CONSTRAINT IF EXISTS portfolio_assets_market_product_fk;
ALTER TABLE portfolio_assets DROP COLUMN IF EXISTS latest_price_source;
ALTER TABLE portfolio_assets DROP COLUMN IF EXISTS market_product_id;
DROP TABLE IF EXISTS market_quotes;
DROP INDEX IF EXISTS market_products_enabled_provider_idx;
DROP TABLE IF EXISTS market_products;
