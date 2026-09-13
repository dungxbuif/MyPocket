UPDATE categories AS personal
SET icon_key = template.icon_key
FROM categories AS template
WHERE personal.owner_id IS NOT NULL
  AND personal.icon_key = 'tag'
  AND template.owner_id IS NULL
  AND template.is_system = false
  AND template.name = personal.name
  AND template.kind = personal.kind
  AND template.icon_key <> 'tag';
