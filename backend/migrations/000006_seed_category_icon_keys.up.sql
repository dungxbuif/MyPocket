UPDATE categories
SET icon_key = system_key
WHERE owner_id IS NULL AND system_key IS NOT NULL AND icon_key = 'tag';
