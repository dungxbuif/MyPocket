-- Canonical category-icon seed. Every default category carries its icon key
-- from deployment; no frontend-only fallback is required to populate the DB.
WITH icon_seed(system_key, icon_key) AS (
  SELECT system_key, system_key
  FROM categories
  WHERE owner_id IS NULL AND system_key IS NOT NULL
)
UPDATE categories AS category
SET icon_key = icon_seed.icon_key
FROM icon_seed
WHERE category.system_key = icon_seed.system_key;
