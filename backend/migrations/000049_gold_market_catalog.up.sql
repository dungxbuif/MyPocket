-- Canonical gold products exposed by the two production-configured sources.
-- Values are buyback prices in VND per chỉ; source identity is the provider key
-- plus the stable external product id, never the display name.
INSERT INTO market_products
    (id, provider_key, external_product_id, name, kind, unit, currency, valuation_side, enabled)
VALUES
    ('gold-btmc-vrtl', 'btmc', 'btmc-vrtl-buy', 'Vàng miếng VRTL Bảo Tín Minh Châu', 'gold', 'chỉ', 'VND', 'buyback', true),
    ('gold-btmc-ring', 'btmc', 'btmc-ring-buy', 'Nhẫn tròn trơn Bảo Tín Minh Châu', 'gold', 'chỉ', 'VND', 'buyback', true),
    ('gold-btmh-ring', 'btmh', 'btmh-jewelry-ring-buy', 'Trang sức - Nhẫn tròn Bảo Tín Mạnh Hải', 'gold', 'chỉ', 'VND', 'buyback', true),
    ('gold-btmh-24k', 'btmh', 'btmh-jewelry-24k-buy', 'Vàng trang sức 24K (999.9) Bảo Tín Mạnh Hải', 'gold', 'chỉ', 'VND', 'buyback', true)
ON CONFLICT (provider_key, external_product_id) DO UPDATE SET
    id = EXCLUDED.id,
    name = EXCLUDED.name,
    kind = EXCLUDED.kind,
    unit = EXCLUDED.unit,
    currency = EXCLUDED.currency,
    valuation_side = EXCLUDED.valuation_side,
    enabled = EXCLUDED.enabled,
    updated_at = now();
